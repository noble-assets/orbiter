// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, NASD Inc. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN "AS IS" BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package components

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
	"orbiter.dev/types/router"
)

type AdapterRouter = interfaces.Router[types.ProtocolID, interfaces.ControllerAdapter]

var _ interfaces.AdapterComponent = &AdapterComponent{}

type AdapterComponent struct {
	logger log.Logger
	// router is an adapter controllers router.
	router     AdapterRouter
	bankKeeper types.BankKeeperAdapter
	dispatcher interfaces.PayloadDispatcher
}

func NewAdapterComponent(
	logger log.Logger,
	bankKeeper types.BankKeeperAdapter,
	dispatcher interfaces.PayloadDispatcher,
) (*AdapterComponent, error) {
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	adaptersKeeper := AdapterComponent{
		logger:     logger.With(types.ComponentPrefix, types.AdaptersComponentName),
		router:     router.New[types.ProtocolID, interfaces.ControllerAdapter](),
		bankKeeper: bankKeeper,
		dispatcher: dispatcher,
	}

	return &adaptersKeeper, adaptersKeeper.Validate()
}

// Validate returns an error if the component instance is not valid.
func (c *AdapterComponent) Validate() error {
	if c.logger == nil {
		return types.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if c.bankKeeper == nil {
		return types.ErrNilPointer.Wrap("bank keeper cannot be nil")
	}
	if c.dispatcher == nil {
		return types.ErrNilPointer.Wrap("dispatcher cannot be nil")
	}
	if c.router == nil {
		return types.ErrNilPointer.Wrap("router cannot be nil")
	}
	return nil
}

func (c *AdapterComponent) Logger() log.Logger {
	return c.logger
}

func (c *AdapterComponent) Router() AdapterRouter {
	return c.router
}

func (c *AdapterComponent) SetRouter(ar AdapterRouter) error {
	if c.router != nil && c.router.Sealed() {
		return errors.New("cannot reset a sealed router")
	}

	c.router = ar
	c.router.Seal()
	return nil
}

// ParsePayload implements types.PayloadAdapter.
func (c *AdapterComponent) ParsePayload(
	id types.ProtocolID,
	payloadBz []byte,
) (bool, *types.Payload, error) {
	adapter, found := c.router.Route(id)
	if !found {
		return false, &types.Payload{}, fmt.Errorf("adapter not found for protocol ID: %s", id)
	}

	return adapter.ParsePayload(payloadBz)
}

// BeforeTransferHook implements types.PayloadAdapter.
func (c *AdapterComponent) BeforeTransferHook(
	ctx context.Context,
	sourceOrbitID types.OrbitID,
	payload *types.Payload,
) error {
	adapter, found := c.router.Route(sourceOrbitID.ProtocolID)
	if !found {
		return fmt.Errorf("adapter not found for protocol ID: %s", sourceOrbitID.ProtocolID)
	}

	if err := adapter.BeforeTransferHook(ctx, payload); err != nil {
		return fmt.Errorf("before transfer hook failed: %w", err)
	}

	return c.clearOrbiterBalances(ctx)
}

// AfterTransferHook implements types.PayloadAdapter.
func (c *AdapterComponent) AfterTransferHook(
	ctx context.Context,
	sourceOrbitID types.OrbitID,
	payload *types.Payload,
) (*types.TransferAttributes, error) {
	adapter, found := c.router.Route(sourceOrbitID.ProtocolID)
	if !found {
		return nil, fmt.Errorf("adapter not found for protocol ID: %s", sourceOrbitID.ProtocolID)
	}

	if err := adapter.AfterTransferHook(ctx, payload); err != nil {
		return nil, fmt.Errorf("after transfer hook failed: %w", err)
	}

	balances := c.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)
	if err := c.validateModuleBalance(balances); err != nil {
		return nil, types.ErrValidation.Wrap(err.Error())
	}

	transferAttr, err := types.NewTransferAttributes(
		sourceOrbitID.ProtocolID,
		sourceOrbitID.CounterpartyID,
		balances[0].GetDenom(),
		balances[0].Amount,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating transfer attributes: %w", err)
	}

	return transferAttr, nil
}

// ProcessPayload implements types.PayloadAdapter.
func (c *AdapterComponent) ProcessPayload(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	payload *types.Payload,
) error {
	return c.dispatcher.DispatchPayload(ctx, transferAttr, payload)
}

// clearOrbiterBalances sends all balances of the orbiter module account to
// a sub-account. This method allows to start a forwarding with the module holding
// only the coins the received transaction is transferring.
func (c *AdapterComponent) clearOrbiterBalances(ctx context.Context) error {
	coins := c.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)
	if coins.IsZero() {
		return nil
	}
	return c.bankKeeper.SendCoins(ctx, types.ModuleAddress, types.DustCollectorAddress, coins)
}

func (c *AdapterComponent) validateModuleBalance(coins sdk.Coins) error {
	if coins.IsZero() {
		return errors.New("expected orbiter module to hold coins after transfer")
	}

	if coins.Len() != 1 {
		return errors.New("expected module to hold only one coin")
	}

	if coins[0].IsZero() {
		return errors.New("received coin has zero amount")
	}

	return nil
}
