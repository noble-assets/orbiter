// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 NASD Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package simapp

import (
	"github.com/circlefin/noble-fiattokenfactory/x/blockibc"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	transferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	soloclient "github.com/cosmos/ibc-go/v8/modules/light-clients/06-solomachine"
	tmclient "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"

	"orbiter.dev/entrypoint"
)

func (app *SimApp) RegisterIBCModules() error {
	if err := app.RegisterStores(
		storetypes.NewKVStoreKey(capabilitytypes.StoreKey),
		storetypes.NewMemoryStoreKey(capabilitytypes.MemStoreKey),
		storetypes.NewKVStoreKey(ibcexported.StoreKey),
		storetypes.NewKVStoreKey(transfertypes.StoreKey),
	); err != nil {
		return err
	}

	app.ParamsKeeper.Subspace(ibcexported.ModuleName).
		WithKeyTable(clienttypes.ParamKeyTable().RegisterParamSet(&connectiontypes.Params{}))
	app.ParamsKeeper.Subspace(transfertypes.ModuleName).WithKeyTable(transfertypes.ParamKeyTable())

	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		app.appCodec,
		app.GetKey(capabilitytypes.StoreKey),
		app.GetMemKey(capabilitytypes.MemStoreKey),
	)

	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.IBCKeeper = ibckeeper.NewKeeper(
		app.appCodec,
		app.GetKey(ibcexported.StoreKey),
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
		// exchange flash debris claw calm shine laundry february cousin glad name miss jar neglect
		// reflect blanket
		// orbit clever educate rent inject lounge pupil plastic
		"noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
	)

	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(transfertypes.ModuleName)
	app.TransferKeeper = transferkeeper.NewKeeper(
		app.appCodec,
		app.GetKey(transfertypes.StoreKey),
		app.GetSubspace(transfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		// exchange flash debris claw calm shine laundry february cousin glad name miss jar neglect
		// reflect blanket
		// orbit clever educate rent inject lounge pupil plastic
		"noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
	)

	var transferStack porttypes.IBCModule
	transferStack = transfer.NewIBCModule(app.TransferKeeper)
	transferStack = entrypoint.NewIBCMiddleware(
		transferStack,
		app.IBCKeeper.ChannelKeeper,
		app.OrbiterKeeper.AdapterComponent(),
	)
	transferStack = blockibc.NewIBCMiddleware(transferStack, app.FTFKeeper)

	ibcRouter := porttypes.NewRouter().AddRoute(transfertypes.ModuleName, transferStack)
	app.IBCKeeper.SetRouter(ibcRouter)

	return app.RegisterModules(
		capability.NewAppModule(app.appCodec, *app.CapabilityKeeper, true),
		ibc.NewAppModule(app.IBCKeeper),
		transfer.NewAppModule(app.TransferKeeper),
		tmclient.NewAppModule(),
		soloclient.NewAppModule(),
	)
}
