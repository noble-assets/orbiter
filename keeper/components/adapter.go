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
func (k *AdapterComponent) Validate() error {
	if k.logger == nil {
		return types.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if k.bankKeeper == nil {
		return types.ErrNilPointer.Wrap("bank keeper cannot be nil")
	}
	if k.dispatcher == nil {
		return types.ErrNilPointer.Wrap("dispatcher cannot be nil")
	}
	if k.router == nil {
		return types.ErrNilPointer.Wrap("router cannot be nil")
	}
	return nil
}

func (k *AdapterComponent) Logger() log.Logger {
	return k.logger
}

func (k *AdapterComponent) Router() AdapterRouter {
	return k.router
}

func (k *AdapterComponent) SetRouter(ar AdapterRouter) error {
	if k.router != nil && k.router.Sealed() {
		return errors.New("cannot reset a sealed router")
	}

	k.router = ar
	k.router.Seal()
	return nil
}

// ParsePayload implements types.PayloadAdapter.
func (k *AdapterComponent) ParsePayload(
	id types.ProtocolID,
	payloadBz []byte,
) (bool, *types.Payload, error) {
	adapter, found := k.router.Route(id)
	if !found {
		return false, &types.Payload{}, fmt.Errorf("adapter not found for protocol ID: %s", id)
	}

	return adapter.ParsePayload(payloadBz)
}

// BeforeTransferHook implements types.PayloadAdapter.
func (k *AdapterComponent) BeforeTransferHook(
	ctx context.Context,
	sourceOrbitID types.OrbitID,
	payload *types.Payload,
) error {
	adapter, found := k.router.Route(sourceOrbitID.ProtocolID)
	if !found {
		return fmt.Errorf("adapter not found for protocol ID: %s", sourceOrbitID.ProtocolID)
	}

	if err := adapter.BeforeTransferHook(ctx, payload); err != nil {
		return fmt.Errorf("before transfer hook failed: %w", err)
	}

	return k.clearOrbiterBalances(ctx)
}

// AfterTransferHook implements types.PayloadAdapter.
func (k *AdapterComponent) AfterTransferHook(
	ctx context.Context,
	sourceOrbitID types.OrbitID,
	payload *types.Payload,
) (*types.TransferAttributes, error) {
	adapter, found := k.router.Route(sourceOrbitID.ProtocolID)
	if !found {
		return nil, fmt.Errorf("adapter not found for protocol ID: %s", sourceOrbitID.ProtocolID)
	}

	if err := adapter.AfterTransferHook(ctx, payload); err != nil {
		return nil, fmt.Errorf("after transfer hook failed: %w", err)
	}

	balances := k.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)
	if err := k.validateOrbiterInitialBalance(balances); err != nil {
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
func (k *AdapterComponent) ProcessPayload(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	payload *types.Payload,
) error {
	return k.dispatcher.DispatchPayload(ctx, transferAttr, payload)
}

func (k *AdapterComponent) clearOrbiterBalances(ctx context.Context) error {
	coins := k.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)
	if coins.IsZero() {
		return nil
	}
	return k.bankKeeper.SendCoins(ctx, types.ModuleAddress, types.DustCollectorAddress, coins)
}

func (k *AdapterComponent) validateOrbiterInitialBalance(coins sdk.Coins) error {
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
