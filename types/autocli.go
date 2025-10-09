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

package types

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

func TxCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "SubmitPayload",
			Use:       "submit [payload]",
			Short:     "Submit a payload to be handled by the orbiter",
			Long: `Depending on the cross-chain technology, ` +
				`it is required to submit the Orbiter payload directly to the Noble chain ` +
				`and then referencing the resulting hash in the bridge transaction. ` +
				`This command submits a given payload to be forwarded.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "payload"},
			},
		},
	}
}

func QueryCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "PendingPayloads",
			Use:       "pending",
			Short:     "Query pending payloads",
		},
	}
}
