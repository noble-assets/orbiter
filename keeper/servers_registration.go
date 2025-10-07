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

package keeper

import (
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/noble-assets/orbiter/keeper/component/adapter"
	"github.com/noble-assets/orbiter/keeper/component/dispatcher"
	"github.com/noble-assets/orbiter/keeper/component/executor"
	"github.com/noble-assets/orbiter/keeper/component/forwarder"
	orbitertypes "github.com/noble-assets/orbiter/types"
	adaptertypes "github.com/noble-assets/orbiter/types/component/adapter"
	dispatchertypes "github.com/noble-assets/orbiter/types/component/dispatcher"
	executortypes "github.com/noble-assets/orbiter/types/component/executor"
	forwardertypes "github.com/noble-assets/orbiter/types/component/forwarder"
)

// RegisterMsgServers registers the gRPC message servers for all Orbiter components
// (Forwarder, Executor, and Adapter) with the module configurator.
func RegisterMsgServers(cfg module.Configurator, k *Keeper) {
	ms := cfg.MsgServer()
	orbitertypes.RegisterMsgServer(ms, k)
	forwardertypes.RegisterMsgServer(ms, forwarder.NewMsgServer(k.forwarder, k))
	executortypes.RegisterMsgServer(ms, executor.NewMsgServer(k.executor, k))
	adaptertypes.RegisterMsgServer(ms, adapter.NewMsgServer(k.adapter, k))
}

// RegisterQueryServers registers the gRPC query servers for all Orbiter components
// with the module configurator.
func RegisterQueryServers(cfg module.Configurator, k *Keeper) {
	qs := cfg.QueryServer()

	orbitertypes.RegisterQueryServer(qs, k)
	forwardertypes.RegisterQueryServer(qs, forwarder.NewQueryServer(k.forwarder))
	executortypes.RegisterQueryServer(qs, executor.NewQueryServer(k.executor))
	adaptertypes.RegisterQueryServer(qs, adapter.NewQueryServer(k.adapter))
	dispatchertypes.RegisterQueryServer(qs, dispatcher.NewQueryServer(k.dispatcher))
}
