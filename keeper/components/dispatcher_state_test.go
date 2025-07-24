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

package components_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"orbiter.dev/keeper/components"
	"orbiter.dev/testutil/mocks"
	"orbiter.dev/types"
)

func newDispatcherComponent(t testing.TB) (*components.DispatcherComponent, *mocks.Dependencies) {
	deps := mocks.NewDependencies(t)

	sb := collections.NewSchemaBuilder(deps.StoreService)
	dispatcher, err := components.NewDispatcherComponent(
		deps.EncCfg.Codec,
		sb,
		deps.Logger,
		&mocks.OrbitsHandler{},
		&mocks.ActionsHandler{},
	)
	require.NoError(t, err)
	_, err = sb.Build()
	require.NoError(t, err)

	return dispatcher, &deps
}

func TestGetDispatched(t *testing.T) {
	// ARRANGE
	dispatcher, deps := newDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceInfo := types.OrbitID{
		ProtocolID:     types.PROTOCOL_IBC,
		CounterpartyID: "channel-1",
	}
	destinationInfo := types.OrbitID{
		ProtocolID:     types.PROTOCOL_CCTP,
		CounterpartyID: "0",
	}
	denom := "uusdc"

	// ACT: Test getting non-existent dispatch record
	result := dispatcher.GetDispatchedAmount(ctx, sourceInfo, destinationInfo, denom)
	require.Equal(t, types.AmountDispatched{
		Incoming: math.ZeroInt(),
		Outgoing: math.ZeroInt(),
	}, result)

	// ARRANGE: Set a dispatch record
	expectedAmount := types.AmountDispatched{
		Incoming: math.NewInt(100),
		Outgoing: math.NewInt(50),
	}

	err := dispatcher.SetDispatchedAmount(ctx, sourceInfo, destinationInfo, denom, expectedAmount)
	require.NoError(t, err)

	// ACT: Test getting existing dispatch record
	result = dispatcher.GetDispatchedAmount(ctx, sourceInfo, destinationInfo, denom)

	// ASSERT
	require.Equal(t, expectedAmount, result)
}

func TestHasDispatchedAmount(t *testing.T) {
	// ARRANGE
	dispatcher, deps := newDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceInfo := types.OrbitID{
		ProtocolID:     types.PROTOCOL_IBC,
		CounterpartyID: "channel-1",
	}
	destinationInfo := types.OrbitID{
		ProtocolID:     types.PROTOCOL_CCTP,
		CounterpartyID: "0",
	}
	denom := "uusdc"

	// ACT
	result := dispatcher.HasDispatchedAmount(ctx, sourceInfo, destinationInfo, denom)

	// ASSERT
	require.False(t, result)

	// ARRANGE: Set a dispatch record
	amount := types.AmountDispatched{
		Incoming: math.NewInt(100),
		Outgoing: math.NewInt(50),
	}
	err := dispatcher.SetDispatchedAmount(ctx, sourceInfo, destinationInfo, denom, amount)
	require.NoError(t, err)

	// ACT: Test existing dispatch record
	result = dispatcher.HasDispatchedAmount(ctx, sourceInfo, destinationInfo, denom)

	// ASSERT
	require.True(t, result)
}

func TestSetDispatched(t *testing.T) {
	// ARRANGE
	dispatcher, deps := newDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceOrbitID := types.OrbitID{
		ProtocolID:     types.PROTOCOL_IBC,
		CounterpartyID: "channel-1",
	}
	destinationOrbitID := types.OrbitID{
		ProtocolID:     types.PROTOCOL_CCTP,
		CounterpartyID: "0",
	}
	denom := "uusdc"

	amount := types.AmountDispatched{
		Incoming: math.NewInt(200),
		Outgoing: math.NewInt(100),
	}

	// ACT: Test setting dispatch record
	err := dispatcher.SetDispatchedAmount(ctx, sourceOrbitID, destinationOrbitID, denom, amount)

	// ASSERT
	require.NoError(t, err)

	result := dispatcher.GetDispatchedAmount(ctx, sourceOrbitID, destinationOrbitID, denom)
	require.Equal(t, amount, result)

	// ARRANGE: Test updating existing dispatch record
	updatedAmount := types.AmountDispatched{
		Incoming: math.NewInt(300),
		Outgoing: math.NewInt(150),
	}

	// ACT
	err = dispatcher.SetDispatchedAmount(
		ctx,
		sourceOrbitID,
		destinationOrbitID,
		denom,
		updatedAmount,
	)

	// ASSERT
	require.NoError(t, err)

	// Verify the record was updated
	result = dispatcher.GetDispatchedAmount(ctx, sourceOrbitID, destinationOrbitID, denom)
	require.Equal(t, updatedAmount, result)
}

func TestGetDispatchedByProtocolID(t *testing.T) {
	// ARRANGE
	dispatcher, deps := newDispatcherComponent(t)
	ctx := deps.SdkCtx

	protocolID := types.PROTOCOL_IBC

	// ACT: Test empty protocol (no dispatch records)
	result := dispatcher.GetDispatchedAmountsByProtocolID(ctx, protocolID)

	// ASSERT
	require.NotNil(t, result.ChainsAmount())
	require.Empty(t, result.ChainsAmount())

	// ARRANGE: Set up test data
	sourceInfo1 := types.OrbitID{
		ProtocolID:     protocolID,
		CounterpartyID: "channel-1",
	}
	sourceInfo2 := types.OrbitID{
		ProtocolID:     protocolID,
		CounterpartyID: "channel-2",
	}
	destinationInfo := types.OrbitID{
		ProtocolID:     types.PROTOCOL_CCTP,
		CounterpartyID: "0",
	}
	denom := "uusdc"

	// Set dispatch records
	amount1 := types.AmountDispatched{
		Incoming: math.NewInt(100),
		Outgoing: math.NewInt(50),
	}
	amount2 := types.AmountDispatched{
		Incoming: math.NewInt(200),
		Outgoing: math.NewInt(100),
	}

	err := dispatcher.SetDispatchedAmount(ctx, sourceInfo1, destinationInfo, denom, amount1)
	require.NoError(t, err)
	err = dispatcher.SetDispatchedAmount(ctx, sourceInfo2, destinationInfo, denom, amount2)
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

func TestDispatched_EmptyStates(t *testing.T) {
	dispatcher, deps := newDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceInfo := types.OrbitID{
		ProtocolID:     types.PROTOCOL_IBC,
		CounterpartyID: "channel-1",
	}
	destinationInfo := types.OrbitID{
		ProtocolID:     types.PROTOCOL_CCTP,
		CounterpartyID: "0",
	}
	denom := "uusdc"

	// Test all methods with empty state
	result := dispatcher.GetDispatchedAmount(ctx, sourceInfo, destinationInfo, denom)
	require.Equal(t, types.AmountDispatched{
		Incoming: math.ZeroInt(),
		Outgoing: math.ZeroInt(),
	}, result)

	hasDispatched := dispatcher.HasDispatchedAmount(ctx, sourceInfo, destinationInfo, denom)
	require.False(t, hasDispatched)

	totalDispatched := dispatcher.GetDispatchedAmountsByProtocolID(ctx, types.PROTOCOL_IBC)
	require.NotNil(t, totalDispatched.ChainsAmount())
	require.Empty(t, totalDispatched.ChainsAmount())

	// Test iteration with empty state
	called := false
	dispatcher.IterateDispatchedAmountsByProtocolID(
		ctx,
		types.PROTOCOL_IBC,
		func(sourceCounterpartyId string, dispatchedInfo types.ChainAmountDispatched) bool {
			called = true
			return false
		},
	)
	require.False(t, called)
}

func TestDispatched_MultipleProtocolsAndChains(t *testing.T) {
	dispatcher, deps := newDispatcherComponent(t)
	ctx := deps.SdkCtx

	// Test with multiple protocols and chains
	testCases := []struct {
		sourceInfo      types.OrbitID
		destinationInfo types.OrbitID
		denom           string
		amount          types.AmountDispatched
	}{
		{
			sourceInfo: types.OrbitID{
				ProtocolID:     types.PROTOCOL_IBC,
				CounterpartyID: "channel-1",
			},
			destinationInfo: types.OrbitID{
				ProtocolID:     types.PROTOCOL_CCTP,
				CounterpartyID: "0",
			},
			denom: "uusdc",
			amount: types.AmountDispatched{
				Incoming: math.NewInt(100),
				Outgoing: math.NewInt(50),
			},
		},
		{
			sourceInfo: types.OrbitID{
				ProtocolID:     types.PROTOCOL_IBC,
				CounterpartyID: "channel-2",
			},
			destinationInfo: types.OrbitID{
				ProtocolID:     types.PROTOCOL_CCTP,
				CounterpartyID: "1",
			},
			denom: "uusdc",
			amount: types.AmountDispatched{
				Incoming: math.NewInt(200),
				Outgoing: math.NewInt(100),
			},
		},
		{
			sourceInfo: types.OrbitID{
				ProtocolID:     types.PROTOCOL_CCTP,
				CounterpartyID: "0",
			},
			destinationInfo: types.OrbitID{
				ProtocolID:     types.PROTOCOL_IBC,
				CounterpartyID: "channel-3",
			},
			denom: "uusdc",
			amount: types.AmountDispatched{
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
	ibcTotal := dispatcher.GetDispatchedAmountsByProtocolID(ctx, types.PROTOCOL_IBC)
	require.Len(t, ibcTotal.ChainsAmount(), 2) // channel-1 and channel-2

	cctpTotal := dispatcher.GetDispatchedAmountsByProtocolID(ctx, types.PROTOCOL_CCTP)
	require.Len(t, cctpTotal.ChainsAmount(), 1) // only counterparty "0"
}

// ====================================================================================================
// Indexes
// ====================================================================================================

func TestGetDispatchedByDestinationProtocolID(t *testing.T) {
	// ARRANGE
	dispatcher, deps := newDispatcherComponent(t)
	ctx := deps.SdkCtx

	protocolSource1 := types.PROTOCOL_IBC
	protocolSource2 := types.PROTOCOL_HYPERLANE
	protocolDestination := types.PROTOCOL_CCTP

	// ACT: Test empty protocol (no dispatch records)
	result := dispatcher.GetDispatchedAmountsByProtocolID(ctx, protocolSource1)

	// ASSERT
	require.NotNil(t, result.ChainsAmount())
	require.Empty(t, result.ChainsAmount())

	// ARRANGE: Set up test data
	sourceInfo1 := types.OrbitID{
		ProtocolID:     protocolSource1,
		CounterpartyID: "channel-1",
	}
	sourceInfo2 := types.OrbitID{
		ProtocolID:     protocolSource1,
		CounterpartyID: "channel-2",
	}
	sourceInfo3 := types.OrbitID{
		ProtocolID:     protocolSource2,
		CounterpartyID: "ethereum",
	}
	destinationInfo := types.OrbitID{
		ProtocolID:     protocolDestination,
		CounterpartyID: "0",
	}
	denom := "uusdc"

	// Set dispatch records
	amount1 := types.AmountDispatched{
		Incoming: math.NewInt(100),
		Outgoing: math.NewInt(50),
	}
	amount2 := types.AmountDispatched{
		Incoming: math.NewInt(200),
		Outgoing: math.NewInt(100),
	}

	err := dispatcher.SetDispatchedAmount(ctx, sourceInfo1, destinationInfo, denom, amount1)
	require.NoError(t, err)
	err = dispatcher.SetDispatchedAmount(ctx, sourceInfo2, destinationInfo, denom, amount2)
	require.NoError(t, err)
	err = dispatcher.SetDispatchedAmount(ctx, sourceInfo3, destinationInfo, denom, amount2)
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
	require.Equal(t, channelOne.OrbitID(), destinationInfo)
	require.Equal(t, channelTwo.OrbitID(), destinationInfo)
	require.Equal(t, ethereum.OrbitID(), destinationInfo)
	require.Equal(t, channelOne.AmountDispatched(), amount1)
	require.Equal(t, channelTwo.AmountDispatched(), amount2)
	require.Equal(t, ethereum.AmountDispatched(), amount2)
}
