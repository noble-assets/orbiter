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
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	orbiterv1 "orbiter.dev/api/v1"
	"orbiter.dev/keeper/component/adapter"
	"orbiter.dev/keeper/component/dispatcher"
	"orbiter.dev/keeper/component/executor"
	"orbiter.dev/keeper/component/forwarder"
	adaptertypes "orbiter.dev/types/component/adapter"
	dispatchertypes "orbiter.dev/types/component/dispatcher"
	executortypes "orbiter.dev/types/component/executor"
	forwardertypes "orbiter.dev/types/component/forwarder"
)

func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: orbiterv1.Msg_ServiceDesc.ServiceName,
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"executor": {
					Service:              executortypes.Msg_serviceDesc.ServiceName,
					Short:                "Actions executor management commands",
					SubCommands:          map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions:    executor.TxCommandOptions(),
					EnhanceCustomCommand: true,
				},
				"forwarder": {
					Service:              forwardertypes.Msg_serviceDesc.ServiceName,
					Short:                "Cross-chain forwarder management commands",
					SubCommands:          map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions:    forwarder.TxCommandOptions(),
					EnhanceCustomCommand: true,
				},
				"adapter": {
					Service:              adaptertypes.Msg_serviceDesc.ServiceName,
					Short:                "Cross-chain adapter management commands",
					SubCommands:          map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions:    adapter.TxCommandOptions(),
					EnhanceCustomCommand: true,
				},
			},
		},
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: orbiterv1.Query_ServiceDesc.ServiceName,
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"adapter": {
					Service:           adaptertypes.Query_serviceDesc.ServiceName,
					Short:             "Cross-chain adapter query commands",
					SubCommands:       map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions: adapter.QueryCommandOptions(),
				},
				"dispatcher": {
					Service:           dispatchertypes.Query_serviceDesc.ServiceName,
					Short:             "Payload dispatch statistics query commands",
					SubCommands:       map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions: dispatcher.QueryCommandOptions(),
				},
				"executor": {
					Service:           executortypes.Query_serviceDesc.ServiceName,
					Short:             "Actions executor query commands",
					SubCommands:       map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions: dispatcher.QueryCommandOptions(),
				},
				"forwarder": {
					Service:           forwardertypes.Query_serviceDesc.ServiceName,
					Short:             "Cross-chain forwarder query commands",
					SubCommands:       map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions: dispatcher.QueryCommandOptions(),
				},
			},
		},
	}
}
