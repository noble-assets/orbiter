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

package component_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"orbiter.dev/keeper/component"
	"orbiter.dev/testutil/mocks"
	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
	dispatchertypes "orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

func TestUpdateStats(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(context.Context, *component.Dispatcher)
		transferAttr   func() *types.TransferAttributes
		forwarding     func() *core.Forwarding
		expErr         string
		expAmounts     map[string]dispatchertypes.AmountDispatched
		expectedCounts uint32
	}{
		{
			name:         "error - nil transfer attributes",
			transferAttr: func() *types.TransferAttributes { return nil },
			forwarding: func() *core.Forwarding {
				return &core.Forwarding{
					ProtocolId: 2,
					Attributes: nil,
				}
			},
			expErr: "nil transfer attributes",
		},
		{
			name: "error - nil forwarding",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)

				return ta
			},
			forwarding: func() *core.Forwarding { return nil },
			expErr:     "nil forwarding",
		},
		{
			name: "error - destination protocol ID is not supported",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)

				return ta
			},
			forwarding: func() *core.Forwarding {
				attr := &testdata.TestForwardingAttr{
					Planet: "ethereum",
				}
				forwarding := core.Forwarding{
					ProtocolId:         0,
					PassthroughPayload: []byte{},
				}
				err := forwarding.SetAttributes(attr)
				require.NoError(t, err)

				return &forwarding
			},
			expErr: "id is not supported",
		},
		{
			name: "error - invalid forwading attributes",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)

				return ta
			},
			forwarding: func() *core.Forwarding {
				return &core.Forwarding{
					ProtocolId: 2,
					Attributes: nil,
				}
			},
			expErr: "forwarding attributes are not set",
		},
		{
			name: "success - same amount and denom",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)

				return ta
			},
			forwarding: func() *core.Forwarding {
				attr := &testdata.TestForwardingAttr{
					Planet: "ethereum",
				}
				orbit, err := core.NewForwarding(2, attr, []byte{})
				require.NoError(t, err)

				return orbit
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(100),
					Outgoing: math.NewInt(100),
				},
			},
			expectedCounts: 1,
		},
		{
			name: "success - same denom and different amount",
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(2, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)
				ta.SetDestinationAmount(math.NewInt(95))

				return ta
			},
			forwarding: func() *core.Forwarding {
				attr := &testdata.TestForwardingAttr{
					Planet: "ethereum",
				}
				orbit, err := core.NewForwarding(1, attr, []byte{})
				require.NoError(t, err)

				return orbit
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
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
			forwarding: func() *core.Forwarding {
				attr := &testdata.TestForwardingAttr{
					Planet: "ethereum",
				}
				orbit, err := core.NewForwarding(1, attr, []byte{})
				require.NoError(t, err)

				return orbit
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
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
			name: "success - different denom and previous stored stats",
			setup: func(ctx context.Context, d *component.Dispatcher) {
				sourceID := core.CrossChainID{
					ProtocolId:     1,
					CounterpartyId: "hyperliquid",
				}

				destID := core.CrossChainID{
					ProtocolId:     1,
					CounterpartyId: "ethereum",
				}

				err := d.SetDispatchedCounts(ctx, sourceID, destID, 10)
				require.NoError(t, err)

				da := dispatchertypes.AmountDispatched{
					Incoming: math.NewInt(1_000),
					Outgoing: math.NewInt(1_000),
				}
				err = d.SetDispatchedAmount(ctx, sourceID, destID, "uusdc", da)
				require.NoError(t, err)

				err = d.SetDispatchedAmount(ctx, destID, sourceID, "uusdc", da)
				require.NoError(t, err)
			},
			transferAttr: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)

				ta.SetDestinationDenom("gwei")
				ta.SetDestinationAmount(math.NewInt(50))

				return ta
			},
			forwarding: func() *core.Forwarding {
				attr := &testdata.TestForwardingAttr{
					Planet: "ethereum",
				}
				forwarding, err := core.NewForwarding(1, attr, []byte{})
				require.NoError(t, err)

				return forwarding
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(1_100),
					Outgoing: math.NewInt(1_000),
				},
				"gwei": {
					Incoming: math.ZeroInt(),
					Outgoing: math.NewInt(50),
				},
			},
			expectedCounts: 11, // 1 from the test + 10 from the setup
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			dispatcher, deps := mocks.NewDispatcherComponent(t)
			ctx := deps.SdkCtx

			if tC.setup != nil {
				tC.setup(ctx, dispatcher)
			}

			transferAttr := tC.transferAttr()
			forwarding := tC.forwarding()
			err := dispatcher.UpdateStats(ctx, transferAttr, forwarding)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)

				// Create expected source and destination info
				sourceID := core.CrossChainID{
					ProtocolId:     transferAttr.SourceProtocolID(),
					CounterpartyId: transferAttr.SourceCounterpartyID(),
				}
				attr, _ := forwarding.CachedAttributes()
				destID := core.CrossChainID{
					ProtocolId:     forwarding.ProtocolID(),
					CounterpartyId: attr.CounterpartyID(),
				}

				// Verify amount stats
				for denom, expectedAmount := range tC.expAmounts {
					actualAmount := dispatcher.GetDispatchedAmount(ctx, sourceID, destID, denom)

					require.Equal(t, expectedAmount.Incoming, actualAmount.Incoming)
					require.Equal(t, expectedAmount.Outgoing, actualAmount.Outgoing)
				}

				// Verify count stats
				actualCounts := dispatcher.GetDispatchedCounts(ctx, sourceID, destID)

				require.Equal(t, tC.expectedCounts, actualCounts)
			}
		})
	}
}

func TestBuildDenomDispatchedAmounts(t *testing.T) {
	testCases := []struct {
		name               string
		transferAttributes func() *types.TransferAttributes
		expAmounts         map[string]dispatchertypes.AmountDispatched
		expErr             string
	}{
		{
			name:               "error - nil transfer attributes",
			transferAttributes: func() *types.TransferAttributes { return nil },
			expErr:             "nil transfer attributes",
		},
		{
			name: "single entry with same denoms",
			transferAttributes: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)

				return ta
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(100),
					Outgoing: math.NewInt(100),
				},
			},
		},
		{
			name: "single entry with same denoms but different amounts",
			transferAttributes: func() *types.TransferAttributes {
				ta, err := types.NewTransferAttributes(1, "hyperliquid", "uusdc", math.NewInt(100))
				require.NoError(t, err)
				ta.SetDestinationAmount(math.NewInt(50))

				return ta
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: math.NewInt(100),
					Outgoing: math.NewInt(50),
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
			expAmounts: map[string]dispatchertypes.AmountDispatched{
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

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			dispatcher, _ := mocks.NewDispatcherComponent(t)

			ddas, err := dispatcher.BuildDenomDispatchedAmounts(tC.transferAttributes())

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				expectedEntries := len(tC.expAmounts)

				require.Len(t, ddas, expectedEntries)

				// Convert result to map for easier verification
				ddaMap := make(map[string]dispatchertypes.AmountDispatched, len(ddas))
				for _, entry := range ddas {
					ddaMap[entry.Denom] = entry.AmountDispatched
				}

				for denom, expectedAmount := range tC.expAmounts {
					actualAmount, exists := ddaMap[denom]
					require.True(t, exists)
					require.Equal(t, expectedAmount.Incoming, actualAmount.Incoming)
					require.Equal(t, expectedAmount.Outgoing, actualAmount.Outgoing)
				}
			}
		})
	}
}
