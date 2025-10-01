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

package orbiter

import (
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	cctpkeeper "github.com/circlefin/noble-cctp/x/cctp/keeper"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	modulev1 "github.com/noble-assets/orbiter/api/module/v1"
	actionctrl "github.com/noble-assets/orbiter/controller/action"
	adapterctrl "github.com/noble-assets/orbiter/controller/adapter"
	forwardingctrl "github.com/noble-assets/orbiter/controller/forwarding"
	"github.com/noble-assets/orbiter/keeper"
	"github.com/noble-assets/orbiter/types"
	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
)

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	Codec        codec.Codec
	AddressCodec address.Codec
	Logger       log.Logger

	EventService event.Service
	StoreService store.KVStoreService

	BankKeeper types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	if in.Config.GetAuthority() == "" {
		panic("authority for x/orbiter module must be set")
	}

	authority := authtypes.NewModuleAddressOrBech32Address(in.Config.GetAuthority())

	k := keeper.NewKeeper(
		in.Codec,
		in.AddressCodec,
		in.Logger,
		in.EventService,
		in.StoreService,
		authority.String(),
		in.BankKeeper,
	)
	m := NewAppModule(k)

	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}
}

type ComponentsInputs struct {
	Orbiters *keeper.Keeper

	BankKeeper bankkeeper.Keeper
	CCTPKeeper *cctpkeeper.Keeper
	WarpKeeper warpkeeper.Keeper
}

func InjectComponents(in ComponentsInputs) {
	InjectActionControllers(in)
	InjectForwardingControllers(in)
	InjectAdapterControllers(in)
}

func InjectForwardingControllers(in ComponentsInputs) {
	cctp, err := forwardingctrl.NewCCTPController(
		in.Orbiters.Forwarder().Logger(),
		cctpkeeper.NewMsgServerImpl(in.CCTPKeeper),
	)
	if err != nil {
		panic(errorsmod.Wrap(err, "error creating CCTP controller"))
	}

	hyperlane, err := forwardingctrl.NewHyperlaneController(
		in.Orbiters.Forwarder().Logger(),
		forwardingtypes.NewHyperlaneHandler(
			warpkeeper.NewMsgServerImpl(in.WarpKeeper),
			warpkeeper.NewQueryServerImpl(in.WarpKeeper),
		),
	)
	if err != nil {
		panic(errorsmod.Wrap(err, "error creating Hyperlane controller"))
	}

	internal, err := forwardingctrl.NewInternalController(
		in.Orbiters.Forwarder().Logger(),
		bankkeeper.NewMsgServerImpl(in.BankKeeper),
	)
	if err != nil {
		panic(errorsmod.Wrap(err, "error creating internal controller"))
	}

	in.Orbiters.SetForwardingControllers(cctp, hyperlane, internal)
}

func InjectActionControllers(in ComponentsInputs) {
	fee, err := actionctrl.NewFeeController(
		in.Orbiters.Executor().Logger(),
		in.Orbiters.Executor().EventService(),
		in.BankKeeper,
	)
	if err != nil {
		panic(errorsmod.Wrap(err, "error creating fee controller"))
	}

	in.Orbiters.SetActionControllers(fee)
}

func InjectAdapterControllers(in ComponentsInputs) {
	ibc, err := adapterctrl.NewIBCAdapter(
		in.Orbiters.Codec(),
		in.Orbiters.Adapter().Logger(),
	)
	if err != nil {
		panic(errorsmod.Wrap(err, "error creating IBC adapter"))
	}

	in.Orbiters.SetAdapterControllers(ibc)
}
