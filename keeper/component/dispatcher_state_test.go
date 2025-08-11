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
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"orbiter.dev/testutil/mocks"
	dispatchertypes "orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

const UsdcDenom = "uusdc"

func TestGetDispatchedAmount(t *testing.T) {
	// ARRANGE
	dispatcher, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceInfo := core.OrbitID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: "channel-1",
	}
	destinationInfo := core.OrbitID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// ACT: Test getting non-existent dispatch record
	result := dispatcher.GetDispatchedAmount(ctx, sourceInfo, destinationInfo, UsdcDenom)
	require.Equal(t, dispatchertypes.AmountDispatched{
		Incoming: math.ZeroInt(),
		Outgoing: math.ZeroInt(),
	}, result)

	// ARRANGE: Set a dispatch record
	expectedAmount := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(100),
		Outgoing: math.NewInt(50),
	}

	err := dispatcher.SetDispatchedAmount(
		ctx,
		sourceInfo,
		destinationInfo,
		UsdcDenom,
		expectedAmount,
	)
	require.NoError(t, err)

	// ACT: Test getting existing dispatch record
	result = dispatcher.GetDispatchedAmount(ctx, sourceInfo, destinationInfo, UsdcDenom)

	// ASSERT
	require.Equal(t, expectedAmount, result)
}

func TestHasDispatchedAmount(t *testing.T) {
	// ARRANGE
	dispatcher, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceInfo := core.OrbitID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: "channel-1",
	}
	destinationInfo := core.OrbitID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// ACT
	result := dispatcher.HasDispatchedAmount(ctx, sourceInfo, destinationInfo, UsdcDenom)

	// ASSERT
	require.False(t, result)

	// ARRANGE: Set a dispatch record
	amount := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(100),
		Outgoing: math.NewInt(50),
	}
	err := dispatcher.SetDispatchedAmount(ctx, sourceInfo, destinationInfo, UsdcDenom, amount)
	require.NoError(t, err)

	// ACT: Test existing dispatch record
	result = dispatcher.HasDispatchedAmount(ctx, sourceInfo, destinationInfo, UsdcDenom)

	// ASSERT
	require.True(t, result)
}

func TestSetDispatchedAmount(t *testing.T) {
	// ARRANGE
	dispatcher, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceOrbitID := core.OrbitID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: "channel-1",
	}
	destinationOrbitID := core.OrbitID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	amount := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(200),
		Outgoing: math.NewInt(100),
	}

	// ACT: Test setting dispatch record
	err := dispatcher.SetDispatchedAmount(ctx, sourceOrbitID, destinationOrbitID, UsdcDenom, amount)

	// ASSERT
	require.NoError(t, err)

	result := dispatcher.GetDispatchedAmount(ctx, sourceOrbitID, destinationOrbitID, UsdcDenom)
	require.Equal(t, amount, result)

	// ARRANGE: Test updating existing dispatch record
	updatedAmount := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(300),
		Outgoing: math.NewInt(150),
	}

	// ACT
	err = dispatcher.SetDispatchedAmount(
		ctx,
		sourceOrbitID,
		destinationOrbitID,
		UsdcDenom,
		updatedAmount,
	)

	// ASSERT
	require.NoError(t, err)

	// Verify the record was updated
	result = dispatcher.GetDispatchedAmount(ctx, sourceOrbitID, destinationOrbitID, UsdcDenom)
	require.Equal(t, updatedAmount, result)

	// ARRANGE: set a dispatched amount with an invalid destination protocol ID
	invalidDestOrbitID := core.OrbitID{
		ProtocolId:     0,
		CounterpartyId: "ethereum",
	}

	// ACT
	err = dispatcher.SetDispatchedAmount(ctx, sourceOrbitID, invalidDestOrbitID, UsdcDenom, amount)
	require.ErrorContains(t, err, "error parsing destination orbit ID")
}

func TestGetDispatchedAmountByProtocolID(t *testing.T) {
	// ARRANGE
	dispatcher, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	protocolID := core.PROTOCOL_IBC

	// ACT: Test empty protocol (no dispatch records)
	result := dispatcher.GetDispatchedAmountsByProtocolID(ctx, protocolID)

	// ASSERT
	require.NotNil(t, result.ChainsAmount())
	require.Empty(t, result.ChainsAmount())

	// ARRANGE: Set up test data
	sourceInfo1 := core.OrbitID{
		ProtocolId:     protocolID,
		CounterpartyId: "channel-1",
	}
	sourceInfo2 := core.OrbitID{
		ProtocolId:     protocolID,
		CounterpartyId: "channel-2",
	}
	destinationInfo := core.OrbitID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// Set dispatch records
	amount1 := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(100),
		Outgoing: math.NewInt(50),
	}
	amount2 := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(200),
		Outgoing: math.NewInt(100),
	}

	err := dispatcher.SetDispatchedAmount(ctx, sourceInfo1, destinationInfo, UsdcDenom, amount1)
	require.NoError(t, err)
	err = dispatcher.SetDispatchedAmount(ctx, sourceInfo2, destinationInfo, UsdcDenom, amount2)
	require.NoError(t, err)

	// ACT: Test getting protocol total dispatched
	result = dispatcher.GetDispatchedAmountsByProtocolID(ctx, protocolID)

	// ASSERT
	require.NotNil(t, result.ChainsAmount())
	require.Len(t, result.ChainsAmount(), 2)

	// Verify the results contain the expected data
	require.Contains(t, result.ChainsAmount(), "channel-1")
	require.Contains(t, result.ChainsAmount(), "channel-2")
	chainsAmount := result.ChainsAmount()
	channelOne := chainsAmount["channel-1"]
	channelTwo := chainsAmount["channel-2"]
	require.Equal(t, channelOne.OrbitID(), destinationInfo)
	require.Equal(t, channelTwo.OrbitID(), destinationInfo)
	require.Equal(t, channelOne.AmountDispatched(), amount1)
	require.Equal(t, channelTwo.AmountDispatched(), amount2)
}

func TestDispatchedAmountEmptyStates(t *testing.T) {
	dispatcher, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceInfo := core.OrbitID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: "channel-1",
	}
	destinationInfo := core.OrbitID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// Test all methods with empty state
	result := dispatcher.GetDispatchedAmount(ctx, sourceInfo, destinationInfo, UsdcDenom)
	require.Equal(t, dispatchertypes.AmountDispatched{
		Incoming: math.ZeroInt(),
		Outgoing: math.ZeroInt(),
	}, result)

	hasDispatched := dispatcher.HasDispatchedAmount(ctx, sourceInfo, destinationInfo, UsdcDenom)
	require.False(t, hasDispatched)

	totalDispatched := dispatcher.GetDispatchedAmountsByProtocolID(ctx, core.PROTOCOL_IBC)
	require.NotNil(t, totalDispatched.ChainsAmount())
	require.Empty(t, totalDispatched.ChainsAmount())

	// Test iteration with empty state
	called := false
	dispatcher.IterateDispatchedAmountsByProtocolID(
		ctx,
		core.PROTOCOL_IBC,
		func(sourceCounterpartyId string, dispatchedInfo dispatchertypes.ChainAmountDispatched) bool {
			called = true

			return false
		},
	)
	require.False(t, called)
}

func TestDispatchedAmountMultipleProtocolsAndChains(t *testing.T) {
	dispatcher, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	// Test with multiple protocols and chains
	testCases := []struct {
		sourceInfo      core.OrbitID
		destinationInfo core.OrbitID
		denom           string
		amount          dispatchertypes.AmountDispatched
	}{
		{
			sourceInfo: core.OrbitID{
				ProtocolId:     core.PROTOCOL_IBC,
				CounterpartyId: "channel-1",
			},
			destinationInfo: core.OrbitID{
				ProtocolId:     core.PROTOCOL_CCTP,
				CounterpartyId: "0",
			},
			denom: "uusdc",
			amount: dispatchertypes.AmountDispatched{
				Incoming: math.NewInt(100),
				Outgoing: math.NewInt(50),
			},
		},
		{
			sourceInfo: core.OrbitID{
				ProtocolId:     core.PROTOCOL_IBC,
				CounterpartyId: "channel-2",
			},
			destinationInfo: core.OrbitID{
				ProtocolId:     core.PROTOCOL_CCTP,
				CounterpartyId: "1",
			},
			denom: "uusdc",
			amount: dispatchertypes.AmountDispatched{
				Incoming: math.NewInt(200),
				Outgoing: math.NewInt(100),
			},
		},
		{
			sourceInfo: core.OrbitID{
				ProtocolId:     core.PROTOCOL_CCTP,
				CounterpartyId: "0",
			},
			destinationInfo: core.OrbitID{
				ProtocolId:     core.PROTOCOL_IBC,
				CounterpartyId: "channel-3",
			},
			denom: "uusdc",
			amount: dispatchertypes.AmountDispatched{
				Incoming: math.NewInt(300),
				Outgoing: math.NewInt(150),
			},
		},
	}

	// Set all dispatch records
	for _, tc := range testCases {
		err := dispatcher.SetDispatchedAmount(
			ctx,
			tc.sourceInfo,
			tc.destinationInfo,
			tc.denom,
			tc.amount,
		)
		require.NoError(t, err)
	}

	// Verify all records can be retrieved
	for _, tc := range testCases {
		result := dispatcher.GetDispatchedAmount(ctx, tc.sourceInfo, tc.destinationInfo, tc.denom)
		require.Equal(t, tc.amount, result)

		hasDispatched := dispatcher.HasDispatchedAmount(
			ctx,
			tc.sourceInfo,
			tc.destinationInfo,
			tc.denom,
		)
		require.True(t, hasDispatched)
	}

	// Test protocol-specific queries
	ibcTotal := dispatcher.GetDispatchedAmountsByProtocolID(ctx, core.PROTOCOL_IBC)
	require.Len(t, ibcTotal.ChainsAmount(), 2) // channel-1 and channel-2

	cctpTotal := dispatcher.GetDispatchedAmountsByProtocolID(ctx, core.PROTOCOL_CCTP)
	require.Len(t, cctpTotal.ChainsAmount(), 1) // only counterparty "0"
}

func TestSetAndGetDispatchedCounts(t *testing.T) {
	// ARRANGE
	dispatcher, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceOrbitID := core.OrbitID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: "channel-1",
	}
	destinationOrbitID := core.OrbitID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// ACT: Test setting dispatch record
	err := dispatcher.SetDispatchedCounts(ctx, sourceOrbitID, destinationOrbitID, 1)

	// ASSERT
	require.NoError(t, err)

	result := dispatcher.GetDispatchedCounts(ctx, sourceOrbitID, destinationOrbitID)
	require.Equal(t, uint32(1), result)

	// ARRANGE: Test updating existing dispatched counts record

	// ACT
	err = dispatcher.SetDispatchedCounts(ctx, sourceOrbitID, destinationOrbitID, 10)

	// ASSERT
	require.NoError(t, err)

	// Verify the record was updated
	result = dispatcher.GetDispatchedCounts(ctx, sourceOrbitID, destinationOrbitID)
	require.Equal(t, uint32(10), result)

	// ARRANGE: set a dispatched counts with an invalid destination protocol ID
	invalidDestOrbitID := core.OrbitID{
		ProtocolId:     0,
		CounterpartyId: "ethereum",
	}

	// ACT
	err = dispatcher.SetDispatchedCounts(ctx, sourceOrbitID, invalidDestOrbitID, 1)
	require.ErrorContains(t, err, "error parsing destination orbit ID")
}

// ====================================================================================================
// Indexes
// ====================================================================================================

func TestGetDispatchedAmountByDestinationProtocolID(t *testing.T) {
	// ARRANGE
	dispatcher, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	protocolSource1 := core.PROTOCOL_IBC
	protocolSource2 := core.PROTOCOL_HYPERLANE
	protocolDestination := core.PROTOCOL_CCTP

	// ACT: Test empty protocol (no dispatch records)
	result := dispatcher.GetDispatchedAmountsByProtocolID(ctx, protocolSource1)

	// ASSERT
	require.NotNil(t, result.ChainsAmount())
	require.Empty(t, result.ChainsAmount())

	// ARRANGE: Set up test data
	sourceOrbitID1 := core.OrbitID{
		ProtocolId:     protocolSource1,
		CounterpartyId: "channel-1",
	}
	sourceOrbitID2 := core.OrbitID{
		ProtocolId:     protocolSource1,
		CounterpartyId: "channel-2",
	}
	sourceOrbitID3 := core.OrbitID{
		ProtocolId:     protocolSource2,
		CounterpartyId: "ethereum",
	}
	destOrbitID := core.OrbitID{
		ProtocolId:     protocolDestination,
		CounterpartyId: "0",
	}
	denom := "uusdc"

	// Set dispatch records
	amount1 := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(100),
		Outgoing: math.NewInt(50),
	}
	amount2 := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(200),
		Outgoing: math.NewInt(100),
	}

	err := dispatcher.SetDispatchedAmount(ctx, sourceOrbitID1, destOrbitID, denom, amount1)
	require.NoError(t, err)
	err = dispatcher.SetDispatchedAmount(ctx, sourceOrbitID2, destOrbitID, denom, amount2)
	require.NoError(t, err)
	err = dispatcher.SetDispatchedAmount(ctx, sourceOrbitID3, destOrbitID, denom, amount2)
	require.NoError(t, err)

	// ACT: Test getting protocol total dispatched
	result = dispatcher.GetDispatchedAmountsByDestinationProtocolID(ctx, protocolDestination)

	// ASSERT
	require.NotNil(t, result.ChainAmount)
	require.Len(t, result.ChainsAmount(), 3)

	// Verify the results contain the expected data
	chainsAmount := result.ChainsAmount()
	require.Contains(t, chainsAmount, "channel-1")
	require.Contains(t, chainsAmount, "channel-2")
	require.Contains(t, chainsAmount, "ethereum")

	channelOne := result.ChainAmount("channel-1")
	channelTwo := result.ChainAmount("channel-2")
	ethereum := result.ChainAmount("ethereum")
	require.Equal(t, channelOne.OrbitID(), destOrbitID)
	require.Equal(t, channelTwo.OrbitID(), destOrbitID)
	require.Equal(t, ethereum.OrbitID(), destOrbitID)
	require.Equal(t, channelOne.AmountDispatched(), amount1)
	require.Equal(t, channelTwo.AmountDispatched(), amount2)
	require.Equal(t, ethereum.AmountDispatched(), amount2)
}
