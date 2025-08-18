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

package forwarder

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func TxCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "PauseProtocol",
			Use:       "pause-protocol [protocol_id]",
			Short:     "Pause a specific forwarding protocol by ID",
			Long: `Pause a specific forwarding protocol by its identifier. This will 
disable all forwarding operations for the specified protocol until it is unpaused.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "UnpauseProtocol",
			Use:       "unpause-protocol [protocol_id]",
			Short:     "Resume a paused forwarding protocol by ID",
			Long: `Resume operations for a previously paused forwarding protocol by its 
identifier. This will re-enable all forwarding operations for the specified protocol, that 
are not paused based on the granular protocol+counterparty combination.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "PauseCrossChains",
			Use:       "pause-cross-chains [protocol_id] [counterparty_ids...]",
			Short:     "Pause specific counterparties for a protocol",
			Long: `Pause specific counterparties for a forwarding protocol. This 
allows selective pausing of forwarding operations to specific destinations while 
keeping other counterparties active.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
				{ProtoField: "counterparty_ids", Varargs: true},
			},
		},
		{
			RpcMethod: "UnpauseCrossChains",
			Use:       "unpause-cross-chains [protocol_id] [counterparty_ids...]",
			Short:     "Resume specific counterparties for a protocol",
			Long: `Resume forwarding operations for specific counterparties that 
were previously paused. This allows selective resumption of forwarding to 
specific destinations.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
				{ProtoField: "counterparty_ids", Varargs: true},
			},
		},
		{
			RpcMethod: "ReplaceDepositForBurn",
			Use:       "replace-deposit-for-burn [original_message] [original_attestation] [new_destination_caller] [new_mint_recipient]",
			Short:     "Replace a CCTP deposit for burn message",
			Long: `Replace a sent deposit for burn message in the CCTP protocol with updated 
destination caller and mint recipient.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "original_message"},
				{ProtoField: "original_attestation"},
				{ProtoField: "new_destination_caller"},
				{ProtoField: "new_mint_recipient"},
			},
		},
	}
}

func QueryCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "IsProtocolPaused",
			Use:       "is-protocol-paused [protocol_id]",
			Short:     "Check if a protocol is paused",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "PausedProtocols",
			Use:       "paused-protocols",
			Short:     "Get all paused protocol IDs",
		},
		{
			RpcMethod: "IsCrossChainPaused",
			Use:       "is-cross-chain-paused [protocol_id] [counterparty_id]",
			Short:     "Check if the counterparty is paused for the specific protocol",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
				{ProtoField: "counterparty_id"},
			},
		},
		{
			RpcMethod: "PausedCrossChains",
			Use:       "paused-cross-chains [protocol_id]",
			Short:     "Get all paused counterparties for the specific protocol ID",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
	}
}
