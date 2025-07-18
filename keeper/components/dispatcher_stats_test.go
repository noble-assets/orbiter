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

package components

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"orbiter.dev/testutil/mocks"
	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
)

func TestUpdateDispatchedAmountStats(t *testing.T) {
	denom := "uusdc"
	testCases := []struct {
		name               string
		sourceOrbitID      types.OrbitID
		destinationOrbitID types.OrbitID
		amountDispatched   types.AmountDispatched
		twoUpdates         bool
		expError           string
	}{
		{
			name:               "error - default values (default protocol ID is not valid)",
			sourceOrbitID:      types.OrbitID{},
			destinationOrbitID: types.OrbitID{},
			amountDispatched:   *types.NewAmountDispatched(math.ZeroInt(), math.ZeroInt()),
			expError:           "id is not supported",
		},
		{
			name: "success - non default values and zero incoming dispatched amount",
			sourceOrbitID: types.OrbitID{
				ProtocolID:     1,
				CounterpartyID: "noble",
			},
			destinationOrbitID: types.OrbitID{
				ProtocolID:     2,
				CounterpartyID: "ethereum",
			},
			amountDispatched: *types.NewAmountDispatched(math.ZeroInt(), math.NewInt(1)),
			expError:         "",
		},
		{
			name: "success - non default values and zero outgoing dispatched amount",
			sourceOrbitID: types.OrbitID{
				ProtocolID:     1,
				CounterpartyID: "noble",
			},
			destinationOrbitID: types.OrbitID{
				ProtocolID:     2,
				CounterpartyID: "ethereum",
			},
			amountDispatched: *types.NewAmountDispatched(math.NewInt(1), math.ZeroInt()),
			expError:         "",
		},
		{
			name:               "success - second dispatched amount update",
			sourceOrbitID:      types.OrbitID{ProtocolID: 1, CounterpartyID: "noble"},
			destinationOrbitID: types.OrbitID{ProtocolID: 2, CounterpartyID: "ethereum"},
			amountDispatched:   *types.NewAmountDispatched(math.NewInt(1), math.ZeroInt()),
			twoUpdates:         true,
			expError:           "",
		},
	}

	for _, tc := range testCases {
		deps := mocks.NewDependencies(t)
		ctx := deps.SdkCtx

		sb := collections.NewSchemaBuilder(deps.StoreService)
		dispatcher, err := NewDispatcherComponent(
			deps.EncCfg.Codec,
			sb,
			deps.Logger,
			&mocks.OrbitsHandler{},
			&mocks.ActionsHandler{},
		)
		require.NoError(t, err)
		_, err = sb.Build()
		require.NoError(t, err)

		err = dispatcher.updateDispatchedAmountStats(
			ctx,
			&tc.sourceOrbitID,
			&tc.destinationOrbitID,
			denom,
			tc.amountDispatched,
		)

		t.Run(tc.name, func(t *testing.T) {
			if tc.expError != "" {
				require.ErrorContains(t, err, tc.expError)
				da := dispatcher.GetDispatchedAmount(
					ctx,
					tc.sourceOrbitID,
					tc.destinationOrbitID,
					denom,
				)
				require.Equal(t, math.ZeroInt(), da.Incoming)
				require.Equal(t, math.ZeroInt(), da.Outgoing)
			} else {
				require.NoError(t, err)
				da := dispatcher.GetDispatchedAmount(ctx, tc.sourceOrbitID, tc.destinationOrbitID, denom)
				require.Equal(t, tc.amountDispatched.Incoming, da.Incoming)
				require.Equal(t, tc.amountDispatched.Outgoing, da.Outgoing)
			}
		})

		t.Run("SecondUpdate/"+tc.name, func(t *testing.T) {
			err = dispatcher.updateDispatchedAmountStats(
				ctx,
				&tc.sourceOrbitID,
				&tc.destinationOrbitID,
				denom,
				tc.amountDispatched,
			)

			if tc.expError != "" {
				require.ErrorContains(t, err, tc.expError)
				da := dispatcher.GetDispatchedAmount(
					ctx,
					tc.sourceOrbitID,
					tc.destinationOrbitID,
					denom,
				)
				require.Equal(t, math.ZeroInt(), da.Incoming)
				require.Equal(t, math.ZeroInt(), da.Outgoing)
			} else {
				require.NoError(t, err)
				da := dispatcher.GetDispatchedAmount(ctx, tc.sourceOrbitID, tc.destinationOrbitID, denom)
				require.Equal(t, tc.amountDispatched.Incoming.MulRaw(2), da.Incoming)
				require.Equal(t, tc.amountDispatched.Outgoing.MulRaw(2), da.Outgoing)
			}
		})
	}
}

func TestUpdateStats(t *testing.T) {
	testCases := []struct {
		name            string
		transferAttr    func() *types.TransferAttributes
		orbit           func() *types.Orbit
		expectedError   string
		expectedAmounts map[string]types.AmountDispatched // key: denom
		expectedCounts  uint32
	}{
		{
			name: "success - incoming and outgoing with same amount and denom",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)
				return ta
			},
			orbit: func() *types.Orbit {
				attr := &testdata.TestOrbitAttr{
					Planet: "ethereum",
				}
				orbit, err := types.NewOrbit(2, attr, []byte{})
				require.NoError(t, err)
				return orbit
			},
			expectedAmounts: map[string]types.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(100),
					Outgoing: math.NewInt(100),
				},
			},
			expectedCounts: 1,
		},
		{
			name: "success - incoming and outgoing with same denom",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(2, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)
				ta.SetDestinationAmount(math.NewInt(95))
				return ta
			},
			orbit: func() *types.Orbit {
				attr := &testdata.TestOrbitAttr{
					Planet: "ethereum",
				}
				orbit, err := types.NewOrbit(1, attr, []byte{})
				require.NoError(t, err)
				return orbit
			},
			expectedAmounts: map[string]types.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(100),
					Outgoing: math.NewInt(95),
				},
			},
			expectedCounts: 1,
		},
		{
			name: "success - different denom",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)
				ta.SetDestinationDenom("gwei")
				ta.SetDestinationAmount(math.NewInt(50))
				return ta
			},
			orbit: func() *types.Orbit {
				attr := &testdata.TestOrbitAttr{
					Planet: "ethereum",
				}
				orbit, err := types.NewOrbit(1, attr, []byte{})
				require.NoError(t, err)
				return orbit
			},
			expectedAmounts: map[string]types.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(100),
					Outgoing: math.ZeroInt(),
				},
				"gwei": {
					Incoming: math.ZeroInt(),
					Outgoing: math.NewInt(50),
				},
			},
			expectedCounts: 1,
		},
		{
			name: "error - invalid orbit attributes",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)
				return ta
			},
			orbit: func() *types.Orbit {
				return &types.Orbit{
					ProtocolId: 2,
					Attributes: nil,
				}
			},
			expectedError: "orbit attributes are not set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deps := mocks.NewDependencies(t)
			ctx := deps.SdkCtx

			sb := collections.NewSchemaBuilder(deps.StoreService)
			dispatcher, err := NewDispatcherComponent(
				deps.EncCfg.Codec,
				sb,
				deps.Logger,
				&mocks.OrbitsHandler{},
				&mocks.ActionsHandler{},
			)
			require.NoError(t, err)
			_, err = sb.Build()
			require.NoError(t, err)

			transferAttr := tc.transferAttr()
			orbit := tc.orbit()
			err = dispatcher.updateStats(ctx, transferAttr, orbit)

			if tc.expectedError != "" {
				require.ErrorContains(t, err, tc.expectedError)
			} else {
				require.NoError(t, err)

				// Create expected source and destination info
				sourceInfo := types.OrbitID{
					ProtocolID:     transferAttr.SourceProtocolID(),
					CounterpartyID: transferAttr.SourceCounterpartyID(),
				}
				attr, _ := orbit.CachedAttributes()
				destinationInfo := types.OrbitID{
					ProtocolID:     orbit.ProtocolID(),
					CounterpartyID: attr.CounterpartyID(),
				}

				// Verify amount stats
				for denom, expectedAmount := range tc.expectedAmounts {
					actualAmount := dispatcher.GetDispatchedAmount(ctx, sourceInfo, destinationInfo, denom)
					require.Equal(t, expectedAmount.Incoming, actualAmount.Incoming)
					require.Equal(t, expectedAmount.Outgoing, actualAmount.Outgoing)
				}

				// Verify count stats
				actualCounts := dispatcher.GetDispatchedCounts(ctx, sourceInfo, destinationInfo)
				require.Equal(t, tc.expectedCounts, actualCounts)
			}
		})
	}
}

func TestUpdateDispatchedCountsStats(t *testing.T) {
	testCases := []struct {
		name               string
		sourceOrbitID      types.OrbitID
		destinationOrbitID types.OrbitID
		twoUpdates         bool
		expError           string
	}{
		{
			name:               "error - default values (invalid protocol ID)",
			sourceOrbitID:      types.OrbitID{},
			destinationOrbitID: types.OrbitID{},
			expError:           "id is not supported",
		},
		{
			name: "success - non default values and zero incoming dispatched amount",
			sourceOrbitID: types.OrbitID{
				ProtocolID:     1,
				CounterpartyID: "noble",
			},
			destinationOrbitID: types.OrbitID{
				ProtocolID:     2,
				CounterpartyID: "ethereum",
			},
			expError: "",
		},
		{
			name:               "success - second dispatched amount update",
			sourceOrbitID:      types.OrbitID{ProtocolID: 1, CounterpartyID: "noble"},
			destinationOrbitID: types.OrbitID{ProtocolID: 2, CounterpartyID: "ethereum"},
			twoUpdates:         true,
			expError:           "",
		},
	}

	for _, tc := range testCases {
		deps := mocks.NewDependencies(t)
		ctx := deps.SdkCtx

		sb := collections.NewSchemaBuilder(deps.StoreService)
		dispatcher, err := NewDispatcherComponent(
			deps.EncCfg.Codec,
			sb,
			deps.Logger,
			&mocks.OrbitsHandler{},
			&mocks.ActionsHandler{},
		)
		require.NoError(t, err)
		_, err = sb.Build()
		require.NoError(t, err)

		err = dispatcher.updateDispatchedCountsStats(ctx, &tc.sourceOrbitID, &tc.destinationOrbitID)

		t.Run(tc.name, func(t *testing.T) {
			if tc.expError != "" {
				require.ErrorContains(t, err, tc.expError)
				dc := dispatcher.GetDispatchedCounts(ctx, tc.sourceOrbitID, tc.destinationOrbitID)
				require.Equal(t, uint32(0), dc)
			} else {
				require.NoError(t, err)
				dc := dispatcher.GetDispatchedCounts(ctx, tc.sourceOrbitID, tc.destinationOrbitID)
				require.Equal(t, uint32(1), dc)
			}
		})

		t.Run("SecondUpdate/"+tc.name, func(t *testing.T) {
			err = dispatcher.updateDispatchedCountsStats(
				ctx,
				&tc.sourceOrbitID,
				&tc.destinationOrbitID,
			)

			if tc.expError != "" {
				require.ErrorContains(t, err, tc.expError)
				dc := dispatcher.GetDispatchedCounts(ctx, tc.sourceOrbitID, tc.destinationOrbitID)
				require.Equal(t, uint32(0), dc)
			} else {
				require.NoError(t, err)
				dc := dispatcher.GetDispatchedCounts(ctx, tc.sourceOrbitID, tc.destinationOrbitID)
				require.Equal(t, uint32(2), dc)
			}
		})
	}
}

func TestBuildDenomDispatchedAmounts(t *testing.T) {
	testCases := []struct {
		name               string
		transferAttributes func() *types.TransferAttributes
		expectedEntries    int
		expectedAmounts    map[string]types.AmountDispatched // key: denom
	}{
		{
			name: "single entry with same denoms",
			transferAttributes: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)
				return ta
			},
			expectedEntries: 1,
			expectedAmounts: map[string]types.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(100),
					Outgoing: math.NewInt(100),
				},
			},
		},
		{
			name: "two entries with different denoms",
			transferAttributes: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)
				ta.SetDestinationDenom("gwei")
				ta.SetDestinationAmount(math.NewInt(50))
				return ta
			},
			expectedEntries: 2,
			expectedAmounts: map[string]types.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(100),
					Outgoing: math.ZeroInt(),
				},
				"gwei": {
					Incoming: math.ZeroInt(),
					Outgoing: math.NewInt(50),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deps := mocks.NewDependencies(t)

			sb := collections.NewSchemaBuilder(deps.StoreService)
			dispatcher, err := NewDispatcherComponent(
				deps.EncCfg.Codec,
				sb,
				deps.Logger,
				&mocks.OrbitsHandler{},
				&mocks.ActionsHandler{},
			)
			require.NoError(t, err)
			_, err = sb.Build()
			require.NoError(t, err)

			result := dispatcher.buildDenomDispatchedAmounts(tc.transferAttributes())

			// Verify number of entries
			require.Equal(
				t,
				tc.expectedEntries,
				len(result),
				"Expected %d entries, got %d",
				tc.expectedEntries,
				len(result),
			)

			// Convert result to map for easier verification
			resultMap := make(map[string]types.AmountDispatched, len(result))
			for _, entry := range result {
				resultMap[entry.Denom] = entry.AmountDispatched
			}

			// Verify each expected amount
			for denom, expectedAmount := range tc.expectedAmounts {
				actualAmount, exists := resultMap[denom]
				require.True(t, exists, "Expected denom %s not found in result", denom)
				require.Equal(
					t,
					expectedAmount.Incoming,
					actualAmount.Incoming,
					"Incoming amount mismatch for denom %s",
					denom,
				)
				require.Equal(
					t,
					expectedAmount.Outgoing,
					actualAmount.Outgoing,
					"Outgoing amount mismatch for denom %s",
					denom,
				)
			}

			// Verify no extra entries
			require.Equal(
				t,
				len(tc.expectedAmounts),
				len(resultMap),
				"Number of result entries doesn't match expected",
			)
		})
	}
}
