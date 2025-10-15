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
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"

	"cosmossdk.io/core/appmodule"
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/noble-assets/orbiter/keeper"
	"github.com/noble-assets/orbiter/types"
	adaptertypes "github.com/noble-assets/orbiter/types/component/adapter"
	dispatchertypes "github.com/noble-assets/orbiter/types/component/dispatcher"
	executortypes "github.com/noble-assets/orbiter/types/component/executor"
	forwardertypes "github.com/noble-assets/orbiter/types/component/forwarder"
	"github.com/noble-assets/orbiter/types/core"
)

const ConsensusVersion = 1

var _ module.AppModuleBasic = AppModuleBasic{}

var (
	_ module.AppModuleBasic      = AppModule{}
	_ appmodule.AppModule        = AppModule{}
	_ module.HasConsensusVersion = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}

	_ appmodule.HasBeginBlocker = AppModule{}
)

type AppModuleBasic struct{}

func NewAppModuleBasic() AppModuleBasic {
	return AppModuleBasic{}
}

func (a AppModuleBasic) Name() string {
	return core.ModuleName
}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := adaptertypes.RegisterQueryHandlerClient(context.Background(), mux, adaptertypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}

	if err := dispatchertypes.RegisterQueryHandlerClient(context.Background(), mux, dispatchertypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}

	if err := executortypes.RegisterQueryHandlerClient(context.Background(), mux, executortypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}

	if err := forwardertypes.RegisterQueryHandlerClient(context.Background(), mux, forwardertypes.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

func (a AppModuleBasic) RegisterInterfaces(reg codectypes.InterfaceRegistry) {
	RegisterInterfaces(reg)
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	RegisterLegacyAminoCodec(cdc)
}

type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(),
		keeper:         keeper,
	}
}

func (m AppModule) RegisterServices(cfg module.Configurator) {
	keeper.RegisterMsgServers(cfg, m.keeper)
	keeper.RegisterQueryServers(cfg, m.keeper)
}

func (m AppModule) IsAppModule() {}

func (m AppModule) IsOnePerModuleType() {}

func (m AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(
	cdc codec.JSONCodec,
	_ client.TxEncodingConfig,
	bz json.RawMessage,
) error {
	var genesis types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genesis); err != nil {
		return errorsmod.Wrapf(err, "failed to unmarshal x/%s genesis state", core.ModuleName)
	}

	return genesis.Validate()
}

func (m AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) {
	var genesis types.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesis)

	m.keeper.InitGenesis(ctx, genesis)
}

func (m AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genesis := m.keeper.ExportGenesis(ctx)

	return cdc.MustMarshalJSON(genesis)
}

func (m AppModule) BeginBlock(ctx context.Context) error {
	return m.keeper.BeginBlock(ctx)
}
