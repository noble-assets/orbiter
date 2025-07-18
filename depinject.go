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
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	modulev1 "orbiter.dev/api/module/v1"
	actionsctrl "orbiter.dev/controllers/actions"
	"orbiter.dev/keeper"
	"orbiter.dev/types"
	"orbiter.dev/types/controllers/actions"
	"orbiter.dev/types/interfaces"
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

	StoreService store.KVStoreService

	BankKeeper types.BankKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper *keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	if in.Config.Authority == "" {
		panic("authority for x/orbiter module must be set")
	}

	authority := authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)

	k := keeper.NewKeeper(
		in.Codec,
		in.AddressCodec,
		in.Logger,
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

type ComponentsBankKeeper interface {
	actions.BankKeeper
}

type ComponentsInputs struct {
	Orbiters *keeper.Keeper

	BankKeeper ComponentsBankKeeper
}

func InjectComponents(in ComponentsInputs) {
	InjectActionControllers(in)
	InjectOrbitControllers(in)
	InjectAdapterControllers(in)
}

func InjectOrbitControllers(in ComponentsInputs) {
	var controllers []interfaces.ControllerOrbit

	in.Orbiters.SetOrbitControllers(controllers...)
}

func InjectActionControllers(in ComponentsInputs) {
	feeController, err := actionsctrl.NewFeeController(
		in.Orbiters.ActionComponent().Logger(),
		in.BankKeeper,
	)
	if err != nil {
		panic("error creating fee controller")
	}

	in.Orbiters.SetActionControllers(feeController)
}

func InjectAdapterControllers(in ComponentsInputs) {
	var controllers []interfaces.ControllerAdapter

	in.Orbiters.SetAdapterControllers(controllers...)
}
