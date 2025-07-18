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

package subkeepers

import (
	"context"
	"errors"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"orbiter.dev/controllers"
	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

type AdapterControllersRouter = interfaces.Router[types.ProtocolID, interfaces.AdapterController]

var _ interfaces.AdapterSubkeeper = &AdapterKeeper{}

type AdapterKeeper struct {
	logger log.Logger

	controllersRouter AdapterControllersRouter

	bankKeeper types.BankKeeperAdapters
	dispatcher interfaces.PayloadDispatcher
}

func NewAdapterKeeper(
	logger log.Logger,
	bankKeeper types.BankKeeperAdapters,
	dispatcher interfaces.PayloadDispatcher,
) (*AdapterKeeper, error) {
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	adaptersKeeper := AdapterKeeper{
		logger: logger.With(types.SubKeeperPrefix, types.AdaptersKeeperName),

		controllersRouter: controllers.NewRouter[types.ProtocolID, interfaces.AdapterController](),

		bankKeeper: bankKeeper,
		dispatcher: dispatcher,
	}

	return &adaptersKeeper, adaptersKeeper.Validate()
}

func (k *AdapterKeeper) Validate() error {
	if k.logger == nil {
		return errors.New("logger cannot be nil")
	}
	if k.bankKeeper == nil {
		return errors.New("bank keeper cannot be nil")
	}
	if k.dispatcher == nil {
		return errors.New("dispatcher cannot be nil")
	}
	if k.controllersRouter == nil {
		return errors.New("adapters router cannot be nil")
	}
	return nil
}

func (k *AdapterKeeper) Logger() log.Logger {
	return k.logger
}

func (k *AdapterKeeper) Router() AdapterControllersRouter {
	return k.controllersRouter
}

func (k *AdapterKeeper) SetRouter(ar AdapterControllersRouter) {
	if k.controllersRouter != nil && k.controllersRouter.Sealed() {
		panic(errors.New("cannot reset a sealed controller router"))
	}

	k.controllersRouter = ar
	k.controllersRouter.Seal()
}

// ParsePayload implements types.PayloadAdapter.
func (k *AdapterKeeper) ParsePayload(
	id types.ProtocolID,
	payloadBz []byte,
) (bool, *types.Payload, error) {
	adapter, found := k.controllersRouter.Route(id)
	if !found {
		return false, &types.Payload{}, errors.New("adapter not found")
	}

	return adapter.ParsePayload(payloadBz)
}

// BeforeTransferHook implements types.PayloadAdapter.
func (k *AdapterKeeper) BeforeTransferHook(
	ctx context.Context,
	id types.ProtocolID,
	payload *types.Payload,
) error {
	adapter, found := k.controllersRouter.Route(id)
	if !found {
		return errors.New("adapter not found")
	}

	if err := adapter.BeforeTransferHook(ctx, payload); err != nil {
		return errors.New("adapter error")
	}

	return k.clearOrbiterBalances(ctx)
}

// AfterTransferHook implements types.PayloadAdapter.
func (k *AdapterKeeper) AfterTransferHook(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
	payload *types.Payload,
) error {
	adapter, found := k.controllersRouter.Route(protocolID)
	if !found {
		return errors.New("adapter not found")
	}

	if err := adapter.AfterTransferHook(ctx, payload); err != nil {
		return err
	}

	balances := k.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)
	if err := k.validateOrbiterInitialBalance(balances); err != nil {
		return err
	}

	transferAttr, err := types.NewTransferAttributes(
		protocolID,
		counterpartyID,
		balances[0].GetDenom(),
		balances[0].Amount,
	)
	if err != nil {
		return err
	}

	return k.dispatcher.DispatchPayload(ctx, transferAttr, payload)
}

func (k *AdapterKeeper) clearOrbiterBalances(ctx context.Context) error {
	coins := k.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)
	if coins.IsZero() {
		return nil
	}
	return k.bankKeeper.SendCoins(ctx, types.ModuleAddress, types.DustCollectorAddress, coins)
}

func (k *AdapterKeeper) validateOrbiterInitialBalance(coins sdk.Coins) error {
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
