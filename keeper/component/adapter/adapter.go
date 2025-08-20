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

package adapter

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/types"
	adaptertypes "github.com/noble-assets/orbiter/types/component/adapter"
	"github.com/noble-assets/orbiter/types/core"
	"github.com/noble-assets/orbiter/types/router"
)

type AdapterRouter = *router.Router[core.ProtocolID, types.ControllerAdapter]

var _ types.Adapter = &Adapter{}

type Adapter struct {
	logger log.Logger
	// router is an adapter controllers router.
	router     AdapterRouter
	bankKeeper types.BankKeeperAdapter
	dispatcher types.PayloadDispatcher
	params     collections.Item[adaptertypes.Params]
}

func New(
	cdc codec.BinaryCodec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	bankKeeper types.BankKeeperAdapter,
	dispatcher types.PayloadDispatcher,
) (*Adapter, error) {
	if cdc == nil {
		return nil, core.ErrNilPointer.Wrap("codec cannot be nil")
	}
	if sb == nil {
		return nil, core.ErrNilPointer.Wrap("schema builder cannot be nil")
	}
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	adaptersKeeper := Adapter{
		logger:     logger.With(core.ComponentPrefix, core.AdapterName),
		router:     router.New[core.ProtocolID, types.ControllerAdapter](),
		bankKeeper: bankKeeper,
		dispatcher: dispatcher,
		params: collections.NewItem(
			sb,
			core.AdapterParamsPrefix,
			core.AdapterParamsName,
			codec.CollValue[adaptertypes.Params](cdc),
		),
	}

	return &adaptersKeeper, adaptersKeeper.Validate()
}

// Validate returns an error if the component instance is not valid.
func (a *Adapter) Validate() error {
	if a.logger == nil {
		return core.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if a.bankKeeper == nil {
		return core.ErrNilPointer.Wrap("bank keeper cannot be nil")
	}
	if a.dispatcher == nil {
		return core.ErrNilPointer.Wrap("dispatcher cannot be nil")
	}
	if a.router == nil {
		return core.ErrNilPointer.Wrap("router cannot be nil")
	}

	return nil
}

func (a *Adapter) Logger() log.Logger {
	return a.logger
}

func (a *Adapter) Router() AdapterRouter {
	return a.router
}

func (a *Adapter) SetRouter(r AdapterRouter) error {
	if r == nil {
		return core.ErrNilPointer.Wrap("router cannot be nil")
	}

	if a.router != nil && a.router.Sealed() {
		return errors.New("cannot reset a sealed router")
	}

	a.router = r
	a.router.Seal()

	return nil
}

// ParsePayload implements types.PayloadAdapter.
func (a *Adapter) ParsePayload(
	id core.ProtocolID,
	payloadBz []byte,
) (bool, *core.Payload, error) {
	a.logger.Debug("started payload parsing", "src_protocol", id.String())
	adapter, found := a.router.Route(id)
	if !found {
		a.logger.Error("adapter for protocol not found", "src_protocol", id.String())
		return false, nil, fmt.Errorf("adapter not found for protocol ID: %s", id)
	}

	return adapter.ParsePayload(id, payloadBz)
}

// BeforeTransferHook implements types.PayloadAdapter.
func (a *Adapter) BeforeTransferHook(
	ctx context.Context,
	sourceID core.CrossChainID,
	payload *core.Payload,
) error {
	if err := a.commonBeforeTransferHook(ctx, payload.Forwarding.PassthroughPayload); err != nil {
		return errorsmod.Wrap(err, "generic hook failed")
	}

	return nil
}

// AfterTransferHook implements types.PayloadAdapter.
func (a *Adapter) AfterTransferHook(
	ctx context.Context,
	sourceID core.CrossChainID,
	payload *core.Payload,
) (*types.TransferAttributes, error) {
	balances := a.bankKeeper.GetAllBalances(ctx, core.ModuleAddress)
	if err := a.validateModuleBalance(balances); err != nil {
		return nil, core.ErrValidation.Wrap(err.Error())
	}

	transferAttr, err := types.NewTransferAttributes(
		sourceID.GetProtocolId(),
		sourceID.GetCounterpartyId(),
		balances[0].GetDenom(),
		balances[0].Amount,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error creating transfer attributes")
	}

	return transferAttr, nil
}

// ProcessPayload implements types.PayloadAdapter.
func (a *Adapter) ProcessPayload(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	payload *core.Payload,
) error {
	return a.dispatcher.DispatchPayload(ctx, transferAttr, payload)
}

// CheckPassthroughPayloadSize checks that the passthrough payload
// size is not higher than the maximum allowed.
func (a *Adapter) CheckPassthroughPayloadSize(
	ctx context.Context,
	passthroughPayload []byte,
) error {
	params := a.GetParams(ctx)

	maxSize := params.MaxPassthroughPayloadSize
	if uint64(len(passthroughPayload)) > uint64(maxSize) {
		return core.ErrValidation.Wrapf(
			"passthrough payload size %d > max allowed %d bytes",
			len(passthroughPayload),
			maxSize,
		)
	}

	return nil
}

// commonBeforeTransferHook groups all the logic that must be executed
// before completing the cross-chain transfer, regardless the incoming
// protocol used.
func (a *Adapter) commonBeforeTransferHook(
	ctx context.Context,
	passthroughPayload []byte,
) error {
	if err := a.CheckPassthroughPayloadSize(ctx, passthroughPayload); err != nil {
		return err
	}

	if err := a.clearOrbiterBalances(ctx); err != nil {
		return err
	}

	return nil
}

// clearOrbiterBalances sends all balances of the orbiter module account to
// a sub-account. This method allows to start a forwarding with the module holding
// only the coins the received transaction is transferring.
func (a *Adapter) clearOrbiterBalances(ctx context.Context) error {
	coins := a.bankKeeper.GetAllBalances(ctx, core.ModuleAddress)
	if coins.IsZero() {
		return nil
	}

	return a.bankKeeper.SendCoins(ctx, core.ModuleAddress, core.DustCollectorAddress, coins)
}

func (a *Adapter) validateModuleBalance(coins sdk.Coins) error {
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
