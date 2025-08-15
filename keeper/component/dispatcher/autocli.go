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

package dispatcher

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func TxCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{}
}

func QueryCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "DispatchedCounts",
			Use:       "dispatched-counts [source_protocol_id] [source_counterparty_id] [destination_protocol_id] [destination_counterparty_id]",
			Short:     "Get dispatched counts for a specific route",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "source_protocol_id"},
				{ProtoField: "source_counterparty_id"},
				{ProtoField: "destination_protocol_id"},
				{ProtoField: "destination_counterparty_id"},
			},
		},
		{
			RpcMethod: "DispatchedCountsByDestinationProtocolID",
			Use:       "dispatched-counts-by-destination [protocol_id]",
			Short:     "Get all dispatched counts for a destination protocol",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "DispatchedCountsBySourceProtocolID",
			Use:       "dispatched-counts-by-source [protocol_id]",
			Short:     "Get all dispatched counts for a source protocol",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "DispatchedAmounts",
			Use:       "dispatched-amounts [source_protocol_id] [source_counterparty_id] [destination_protocol_id] [destination_counterparty_id] [denom]",
			Short:     "Get dispatched amounts for a specific route",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "source_protocol_id"},
				{ProtoField: "source_counterparty_id"},
				{ProtoField: "destination_protocol_id"},
				{ProtoField: "destination_counterparty_id"},
				{ProtoField: "denom"},
			},
		},
		{
			RpcMethod: "DispatchedAmountsByDestinationProtocolID",
			Use:       "dispatched-amounts-by-destination [protocol_id]",
			Short:     "Get all dispatched amounts for a destination protocol",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "DispatchedAmountsBySourceProtocolID",
			Use:       "dispatched-amounts-by-source [protocol_id]",
			Short:     "Get all dispatched amounts for a source protocol",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
	}
}
