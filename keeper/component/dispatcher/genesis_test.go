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

package dispatcher_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/noble-assets/orbiter/testutil/mocks"
	dispatchertypes "github.com/noble-assets/orbiter/types/component/dispatcher"
	"github.com/noble-assets/orbiter/types/core"
)

func defaultAmounts(
	sourceProtocolID, destProtocolID core.ProtocolID,
	sourceCounterpartyID, destCounterpartyID string,
) *dispatchertypes.DispatchedAmountEntry {
	return &dispatchertypes.DispatchedAmountEntry{
		SourceId: &core.CrossChainID{
			ProtocolId:     sourceProtocolID,
			CounterpartyId: sourceCounterpartyID,
		},
		DestinationId: &core.CrossChainID{
			ProtocolId:     destProtocolID,
			CounterpartyId: destCounterpartyID,
		},
		Denom: "unoble",
		AmountDispatched: dispatchertypes.AmountDispatched{
			Incoming: math.NewInt(2),
			Outgoing: math.NewInt(1),
		},
	}
}

func defaultCounts(
	sourceProtocolID, destProtocolID core.ProtocolID,
	sourceCounterpartyID, destCounterpartyID string,
) *dispatchertypes.DispatchCountEntry {
	return &dispatchertypes.DispatchCountEntry{
		SourceId: &core.CrossChainID{
			ProtocolId:     sourceProtocolID,
			CounterpartyId: sourceCounterpartyID,
		},
		DestinationId: &core.CrossChainID{
			ProtocolId:     destProtocolID,
			CounterpartyId: destCounterpartyID,
		},
		Count: 1,
	}
}

func TestInitGenesis(t *testing.T) {
	testCases := []struct {
		name    string
		genesis func() *dispatchertypes.GenesisState
		expErr  string
	}{
		{
			name:    "error - nil genesis",
			genesis: func() *dispatchertypes.GenesisState { return nil },
			expErr:  "nil",
		},
		{
			name:    "success - default genesis",
			genesis: dispatchertypes.DefaultGenesisState,
			expErr:  "",
		},
		{
			name: "success - custom genesis",
			genesis: func() *dispatchertypes.GenesisState {
				g := dispatchertypes.DefaultGenesisState()
				g.DispatchedAmounts = append(
					g.DispatchedAmounts,
					*defaultAmounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "channel-1", "2"),
					*defaultAmounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "channel-3", "4"),
					*defaultAmounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "11", "channel-12"),
					*defaultAmounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "13", "channel-14"),
				)
				g.DispatchedCounts = append(
					g.DispatchedCounts,
					*defaultCounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "channel-1", "2"),
					*defaultCounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "channel-3", "4"),
					*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "11", "channel-12"),
					*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "13", "channel-14"),
					*defaultCounts(core.PROTOCOL_HYPERLANE, core.PROTOCOL_CCTP, "1", "2"),
					*defaultCounts(core.PROTOCOL_HYPERLANE, core.PROTOCOL_CCTP, "3", "4"),
					*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_HYPERLANE, "11", "12"),
					*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_HYPERLANE, "13", "14"),
				)

				return g
			},
			expErr: "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			d, deps := mocks.NewDispatcherComponent(t)
			ctx := deps.SdkCtx

			g := tC.genesis()
			err := d.InitGenesis(ctx, g)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, g.DispatchedAmounts, d.GetAllDispatchedAmounts(ctx))
				require.ElementsMatch(t, g.DispatchedCounts, d.GetAllDispatchedCounts(ctx))
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	testCases := []struct {
		name              string
		dispatchedAmounts []dispatchertypes.DispatchedAmountEntry
		dispatchedCounts  []dispatchertypes.DispatchCountEntry
		expEmpty          bool
	}{
		{
			name:     "success - empty state",
			expEmpty: true,
		},
		{
			name: "success - not empty state",
			dispatchedAmounts: []dispatchertypes.DispatchedAmountEntry{
				*defaultAmounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "channel-1", "2"),
				*defaultAmounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "channel-3", "4"),
				*defaultAmounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "11", "channel-12"),
				*defaultAmounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "13", "channel-14"),
			},
			dispatchedCounts: []dispatchertypes.DispatchCountEntry{
				*defaultCounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "channel-1", "2"),
				*defaultCounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "channel-3", "4"),
				*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "11", "channel-12"),
				*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "13", "channel-14"),
				*defaultCounts(core.PROTOCOL_HYPERLANE, core.PROTOCOL_CCTP, "1", "2"),
				*defaultCounts(core.PROTOCOL_HYPERLANE, core.PROTOCOL_CCTP, "3", "4"),
				*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_HYPERLANE, "11", "12"),
				*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_HYPERLANE, "13", "14"),
			},
			expEmpty: false,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			d, deps := mocks.NewDispatcherComponent(t)
			ctx := deps.SdkCtx

			var da []dispatchertypes.DispatchedAmountEntry
			if tC.dispatchedAmounts != nil {
				da = tC.dispatchedAmounts
			}
			for _, amount := range da {
				err := d.SetDispatchedAmount(
					ctx,
					amount.SourceId,
					amount.DestinationId,
					amount.Denom,
					amount.AmountDispatched,
				)
				require.NoError(t, err)
			}

			var dc []dispatchertypes.DispatchCountEntry
			if tC.dispatchedCounts != nil {
				dc = tC.dispatchedCounts
			}
			for _, count := range dc {
				err := d.SetDispatchedCounts(
					ctx,
					count.SourceId,
					count.DestinationId,
					count.Count,
				)
				require.NoError(t, err)
			}

			g := d.ExportGenesis(ctx)

			if tC.expEmpty {
				require.Empty(t, g.DispatchedAmounts)
				require.Empty(t, g.DispatchedCounts)
			} else {
				require.ElementsMatch(t, da, g.DispatchedAmounts)
				require.ElementsMatch(t, dc, g.DispatchedCounts)
			}
		})
	}
}
