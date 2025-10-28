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
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/noble-assets/orbiter/keeper/component/dispatcher"
	"github.com/noble-assets/orbiter/testutil/mocks"
	"github.com/noble-assets/orbiter/testutil/testdata"
	dispatchertypes "github.com/noble-assets/orbiter/types/component/dispatcher"
	"github.com/noble-assets/orbiter/types/core"
)

func TestUpdateStats(t *testing.T) {
	defaultAttr := func() *core.TransferAttributes {
		t.Helper()
		ta, err := core.NewTransferAttributes(
			core.PROTOCOL_IBC,
			"channel-1",
			"uusdc",
			sdkmath.NewInt(100),
		)
		require.NoError(t, err)

		return ta
	}

	defaultForwarding := func() *core.Forwarding {
		t.Helper()
		attr := &testdata.TestForwardingAttr{
			Planet: "1",
		}
		f, err := core.NewForwarding(core.PROTOCOL_CCTP, attr, []byte{})
		require.NoError(t, err)

		return f
	}

	testCases := []struct {
		name       string
		setup      func(context.Context, *dispatcher.Dispatcher)
		attr       func() *core.TransferAttributes // used to create source ID
		forwarding func() *core.Forwarding         // used to create destination ID
		expErr     string
		expAmounts map[string]dispatchertypes.AmountDispatched
		expCounts  uint64
	}{
		{
			name:       "error - nil transfer attributes",
			attr:       func() *core.TransferAttributes { return nil },
			forwarding: defaultForwarding,
			expErr:     "nil transfer attributes",
		},
		{
			name:       "error - nil forwarding",
			attr:       defaultAttr,
			forwarding: func() *core.Forwarding { return nil },
			expErr:     "nil forwarding",
		},
		{
			name: "error - destination protocol ID is not supported",
			attr: defaultAttr,
			forwarding: func() *core.Forwarding {
				f := defaultForwarding()
				f.ProtocolId = core.PROTOCOL_UNSUPPORTED

				return f
			},
			expErr: "failed to create destination cross-chain ID",
		},
		{
			name: "error - invalid forwarding attributes",
			attr: defaultAttr,
			forwarding: func() *core.Forwarding {
				return &core.Forwarding{
					ProtocolId: 2,
					Attributes: nil,
				}
			},
			expErr: "forwarding attributes are not set",
		},
		{
			name: "error - dispatched counts overflow",
			setup: func(ctx context.Context, d *dispatcher.Dispatcher) {
				sourceID := core.CrossChainID{
					ProtocolId:     core.PROTOCOL_IBC,
					CounterpartyId: "channel-1",
				}

				destID := core.CrossChainID{
					ProtocolId:     core.PROTOCOL_CCTP,
					CounterpartyId: "1",
				}

				err := d.SetDispatchedCounts(ctx, &sourceID, &destID, math.MaxUint64)
				require.NoError(t, err)
			},
			attr: func() *core.TransferAttributes {
				ta := defaultAttr()
				ta.SetDestinationAmount(sdkmath.NewInt(95))

				return ta
			},
			forwarding: defaultForwarding,
			expErr:     "overflow",
		},
		{
			name:       "success - same amount and denom",
			attr:       defaultAttr,
			forwarding: defaultForwarding,
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: sdkmath.NewInt(100),
					Outgoing: sdkmath.NewInt(100),
				},
			},
			expCounts: 1,
		},
		{
			name: "success - same denom and different amount",
			attr: func() *core.TransferAttributes {
				ta := defaultAttr()
				ta.SetDestinationAmount(sdkmath.NewInt(95))

				return ta
			},
			forwarding: defaultForwarding,
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: sdkmath.NewInt(100),
					Outgoing: sdkmath.NewInt(95),
				},
			},
			expCounts: 1,
		},
		{
			name: "success - different denom",
			attr: func() *core.TransferAttributes {
				ta := defaultAttr()
				ta.SetDestinationDenom("gwei")
				ta.SetDestinationAmount(sdkmath.NewInt(50))

				return ta
			},
			forwarding: defaultForwarding,
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: sdkmath.NewInt(100),
					Outgoing: sdkmath.ZeroInt(),
				},
				"gwei": {
					Incoming: sdkmath.ZeroInt(),
					Outgoing: sdkmath.NewInt(50),
				},
			},
			expCounts: 1,
		},
		{
			name: "success - different denom and previous stored stats",
			setup: func(ctx context.Context, d *dispatcher.Dispatcher) {
				sourceID := core.CrossChainID{
					ProtocolId:     core.PROTOCOL_IBC,
					CounterpartyId: "channel-1",
				}

				destID := core.CrossChainID{
					ProtocolId:     core.PROTOCOL_CCTP,
					CounterpartyId: "1",
				}

				err := d.SetDispatchedCounts(ctx, &sourceID, &destID, 10)
				require.NoError(t, err)

				da := dispatchertypes.AmountDispatched{
					Incoming: sdkmath.NewInt(1_000),
					Outgoing: sdkmath.NewInt(1_000),
				}
				err = d.SetDispatchedAmount(ctx, &sourceID, &destID, "uusdc", da)
				require.NoError(t, err)

				err = d.SetDispatchedAmount(ctx, &destID, &sourceID, "uusdc", da)
				require.NoError(t, err)
			},
			attr: func() *core.TransferAttributes {
				ta := defaultAttr()

				ta.SetDestinationDenom("gwei")
				ta.SetDestinationAmount(sdkmath.NewInt(50))

				return ta
			},
			forwarding: defaultForwarding,
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: sdkmath.NewInt(1_100),
					Outgoing: sdkmath.NewInt(1_000),
				},
				"gwei": {
					Incoming: sdkmath.ZeroInt(),
					Outgoing: sdkmath.NewInt(50),
				},
			},
			expCounts: 11, // 1 from the test + 10 from the setup
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			dispatcher, deps := mocks.NewDispatcherComponent(t)
			ctx := deps.SdkCtx

			if tC.setup != nil {
				tC.setup(ctx, dispatcher)
			}

			attr := tC.attr()
			forwarding := tC.forwarding()
			err := dispatcher.UpdateStats(ctx, attr, forwarding)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)

				// Create expected source and destination info
				sourceID := core.CrossChainID{
					ProtocolId:     attr.SourceProtocolID(),
					CounterpartyId: attr.SourceCounterpartyID(),
				}
				attr, _ := forwarding.CachedAttributes()
				destID := core.CrossChainID{
					ProtocolId:     forwarding.ProtocolID(),
					CounterpartyId: attr.CounterpartyID(),
				}

				// Verify amount stats
				for denom, expAmount := range tC.expAmounts {
					da := dispatcher.GetDispatchedAmount(ctx, &sourceID, &destID, denom)

					require.Equal(t, expAmount.Incoming, da.AmountDispatched.Incoming)
					require.Equal(t, expAmount.Outgoing, da.AmountDispatched.Outgoing)
				}

				// Verify count stats
				actualCounts := dispatcher.GetDispatchedCounts(ctx, &sourceID, &destID)

				require.Equal(t, tC.expCounts, actualCounts.Count)
			}
		})
	}
}

func TestBuildDenomDispatchedAmounts(t *testing.T) {
	testCases := []struct {
		name               string
		transferAttributes func() *core.TransferAttributes
		expAmounts         map[string]dispatchertypes.AmountDispatched
		expErr             string
	}{
		{
			name:               "error - nil transfer attributes",
			transferAttributes: func() *core.TransferAttributes { return nil },
			expErr:             "nil transfer attributes",
		},
		{
			name: "single entry with same denoms",
			transferAttributes: func() *core.TransferAttributes {
				ta, err := core.NewTransferAttributes(1, "channel-1", "uusdc", sdkmath.NewInt(100))
				require.NoError(t, err)

				return ta
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: sdkmath.NewInt(100),
					Outgoing: sdkmath.NewInt(100),
				},
			},
		},
		{
			name: "single entry with same denoms but different amounts",
			transferAttributes: func() *core.TransferAttributes {
				ta, err := core.NewTransferAttributes(1, "channel-1", "uusdc", sdkmath.NewInt(100))
				require.NoError(t, err)
				ta.SetDestinationAmount(sdkmath.NewInt(50))

				return ta
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: sdkmath.NewInt(100),
					Outgoing: sdkmath.NewInt(50),
				},
			},
		},
		{
			name: "two entries with different denoms",
			transferAttributes: func() *core.TransferAttributes {
				ta, err := core.NewTransferAttributes(1, "channel-1", "uusdc", sdkmath.NewInt(100))
				require.NoError(t, err)
				ta.SetDestinationDenom("gwei")
				ta.SetDestinationAmount(sdkmath.NewInt(50))

				return ta
			},
			expAmounts: map[string]dispatchertypes.AmountDispatched{
				"uusdc": {
					Incoming: sdkmath.NewInt(100),
					Outgoing: sdkmath.ZeroInt(),
				},
				"gwei": {
					Incoming: sdkmath.ZeroInt(),
					Outgoing: sdkmath.NewInt(50),
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
