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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/testutil/mocks"
	forwardertypes "github.com/noble-assets/orbiter/types/component/forwarder"
	"github.com/noble-assets/orbiter/types/core"
)

func TestInitGenesis(t *testing.T) {
	f, deps := mocks.NewForwarderComponent(t)
	ctx := deps.SdkCtx

	// ACT: fail for invalid gen state
	invalidGenState := &forwardertypes.GenesisState{
		PausedProtocolIds: []core.ProtocolID{core.PROTOCOL_UNSUPPORTED},
	}

	err := f.InitGenesis(ctx, invalidGenState)
	require.ErrorContains(t, err, "invalid paused protocol ID")

	// ACT: update correctly for valid gen state
	defaultPausedActionIDs, err := f.GetPausedProtocols(deps.SdkCtx)
	require.NoError(t, err, "failed to get paused protocol IDs")

	defaultPausedCrossChainIDs, err := f.GetAllPausedCrossChainIDs(deps.SdkCtx)
	require.NoError(t, err, "failed to get paused cross-chain IDs")

	updatedGenState := forwardertypes.GenesisState{
		PausedProtocolIds: []core.ProtocolID{core.PROTOCOL_HYPERLANE},
		PausedCrossChainIds: []*core.CrossChainID{
			{ProtocolId: core.PROTOCOL_IBC, CounterpartyId: "channel-42"},
		},
	}
	require.NotEqual(
		t,
		updatedGenState.PausedProtocolIds,
		defaultPausedActionIDs,
		"updated protocol IDs should be different from current",
	)
	require.NotEqual(
		t,
		updatedGenState.PausedCrossChainIds,
		defaultPausedCrossChainIDs,
		"updated cross-chain IDs should be different from current",
	)

	err = f.InitGenesis(ctx, &updatedGenState)
	require.NoError(t, err, "failed to update genesis state")

	pausedProtocolIDs, err := f.GetPausedProtocols(deps.SdkCtx)
	require.NoError(t, err, "failed to get paused protocol IDs")
	require.Equal(
		t,
		updatedGenState.PausedProtocolIds,
		pausedProtocolIDs,
		"paused protocol IDs do not match",
	)

	pausedCrossChainIDs, err := f.GetAllPausedCrossChainIDs(deps.SdkCtx)
	require.NoError(t, err, "failed to get paused cross chain IDs")
	require.Equal(
		t,
		updatedGenState.PausedCrossChainIds,
		pausedCrossChainIDs,
		"paused cross chain IDs do not match",
	)
}

func TestExportGenesis(t *testing.T) {
	fw, deps := mocks.NewForwarderComponent(t)

	expPausedProtocolIDs := []core.ProtocolID{core.PROTOCOL_HYPERLANE}
	require.NoError(
		t,
		fw.SetPausedProtocol(deps.SdkCtx, expPausedProtocolIDs[0]),
		"failed to set paused protocol",
	)

	expPausedCrossChainIDs := []*core.CrossChainID{
		{ProtocolId: core.PROTOCOL_IBC, CounterpartyId: "channel-1"},
		{ProtocolId: core.PROTOCOL_CCTP, CounterpartyId: "7"},
		{ProtocolId: core.PROTOCOL_HYPERLANE, CounterpartyId: "1"},
	}
	for _, id := range expPausedCrossChainIDs {
		err := fw.SetPausedCrossChain(deps.SdkCtx, *id)
		require.NoError(t, err, "failed to set paused cross chain")
	}

	expGenState := forwardertypes.GenesisState{
		PausedProtocolIds:   expPausedProtocolIDs,
		PausedCrossChainIds: expPausedCrossChainIDs,
	}

	genState := fw.ExportGenesis(deps.SdkCtx)
	require.Equal(t, expGenState.String(), genState.String(), "genesis state does not match")
}
