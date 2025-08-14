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

package forwarder_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/keeper/component/forwarder"
	"orbiter.dev/testutil/mocks"
	forwardertypes "orbiter.dev/types/component/forwarder"
	"orbiter.dev/types/core"
)

func TestInitGenesis(t *testing.T) {
	testcases := []struct {
		name     string
		genState *forwardertypes.GenesisState
		expErr   string
	}{
		{
			name:     "success - default genesis state",
			genState: forwardertypes.DefaultGenesisState(),
			expErr:   "",
		},
		{
			name: "success - genesis state with both paused protocols and cross-chain IDs",
			genState: &forwardertypes.GenesisState{
				PausedProtocolIds: []core.ProtocolID{core.PROTOCOL_HYPERLANE},
				PausedCrossChainIds: []*core.CrossChainID{
					{ProtocolId: core.PROTOCOL_IBC, CounterpartyId: "channel-42"},
				},
			},
			expErr: "",
		},
		{
			name:     "error - nil genesis state",
			genState: nil,
			expErr:   "forwarder genesis state: invalid nil pointer",
		},
		{
			name: "error - invalid protocol ID",
			genState: &forwardertypes.GenesisState{
				PausedProtocolIds: []core.ProtocolID{core.PROTOCOL_UNSUPPORTED},
			},
			expErr: "protocol ID is not supported",
		},
		{
			name: "error - nil cross chain ID",
			genState: &forwardertypes.GenesisState{
				PausedCrossChainIds: []*core.CrossChainID{nil},
			},
			expErr: "invalid nil pointer",
		},
		{
			name: "error - invalid cross chain ID",
			genState: &forwardertypes.GenesisState{
				PausedCrossChainIds: []*core.CrossChainID{
					{ProtocolId: core.PROTOCOL_UNSUPPORTED, CounterpartyId: "channel-1"},
				},
			},
			expErr: "invalid paused cross-chain id",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			f, deps := mocks.NewForwarderComponent(t)
			ctx := deps.SdkCtx

			err := f.InitGenesis(ctx, tc.genState)
			if tc.expErr != "" {
				require.ErrorContains(t, err, tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	testcases := []struct {
		name           string
		setup          func(ctx context.Context, k *forwarder.Forwarder)
		expPausedProts []core.ProtocolID
	}{
		{
			name:           "success - export default genesis state",
			expPausedProts: []core.ProtocolID{},
		},
		{
			name: "success - export genesis state with paused protocols",
			setup: func(ctx context.Context, k *forwarder.Forwarder) {
				require.NoError(t, k.SetPausedProtocol(ctx, core.PROTOCOL_IBC))
				require.NoError(t, k.SetPausedProtocol(ctx, core.PROTOCOL_CCTP))
			},
			expPausedProts: []core.ProtocolID{core.PROTOCOL_IBC, core.PROTOCOL_CCTP},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			fw, deps := mocks.NewForwarderComponent(t)
			ctx := deps.SdkCtx

			if tc.setup != nil {
				tc.setup(ctx, fw)
			}

			genState := fw.ExportGenesis(ctx)

			require.NotNil(t, genState)
			require.ElementsMatch(t, tc.expPausedProts, genState.PausedProtocolIds)
			require.Empty(t, genState.PausedCrossChainIds)
		})
	}
}
