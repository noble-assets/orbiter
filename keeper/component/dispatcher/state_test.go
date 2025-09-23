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
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/noble-assets/orbiter/keeper/component/dispatcher"
	"github.com/noble-assets/orbiter/testutil/mocks"
	dispatchertypes "github.com/noble-assets/orbiter/types/component/dispatcher"
	"github.com/noble-assets/orbiter/types/core"
)

const (
	UsdcDenom = "uusdc"
	channel0  = "channel-0"
	channel1  = "channel-1"
	channel2  = "channel-2"
	channel3  = "channel-3"
)

// ====================================================================================================
// Dispatched Amount
// ====================================================================================================

func TestSetHasGetDispatchedAmount(t *testing.T) {
	// ARRANGE
	d, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: channel1,
	}
	destID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// ACT: No dispatched amount records.
	found := d.HasDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom)
	da := d.GetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom)

	// ASSERT
	expAmount := dispatchertypes.AmountDispatched{
		Incoming: math.ZeroInt(),
		Outgoing: math.ZeroInt(),
	}
	require.False(t, found)
	require.Equal(t, expAmount, da.AmountDispatched)

	// ARRANGE: Set a dispatch record.
	expAmount.Incoming = math.NewInt(100)
	expAmount.Outgoing = math.NewInt(50)

	// ACT
	err := d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, expAmount)

	// ASSERT
	require.NoError(t, err)

	// ACT: Get existing dispatch record.
	found = d.HasDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom)
	da = d.GetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom)

	// ASSERT
	require.True(t, found)
	require.Equal(t, expAmount, da.AmountDispatched)

	// ARRANGE: Update a dispatch record.
	expAmount.Incoming = math.NewInt(1_000)

	// ACT: Update an existing amount.
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, expAmount)

	// ASSERT
	require.NoError(t, err)

	// ACT: Get existing dispatch record.
	found = d.HasDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom)
	da = d.GetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom)

	// ASSERT
	require.True(t, found)
	require.Equal(t, expAmount, da.AmountDispatched)

	// ARRANGE: Set an entry with only one quantity positive
	sourceID.ProtocolId = core.PROTOCOL_HYPERLANE
	expAmount.Incoming = math.ZeroInt()
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, expAmount)
	require.NoError(t, err)

	// ACT
	found = d.HasDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom)

	// ASSERT
	require.True(t, found)
}

func createDispatchedAmountEntries(t *testing.T, ctx context.Context, d *dispatcher.Dispatcher) {
	t.Helper()

	sourceID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: channel1,
	}
	destID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// Set dispatch records
	amount := dispatchertypes.AmountDispatched{
		Incoming: math.NewInt(1),
		Outgoing: math.NewInt(1),
	}

	// Set 3 entries for IBC sources.
	err := d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	sourceID.CounterpartyId = channel2
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	sourceID.CounterpartyId = channel3
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	// Set 3 entries for Hyperlane sources.
	sourceID.ProtocolId = core.PROTOCOL_HYPERLANE
	sourceID.CounterpartyId = "999"
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	sourceID.CounterpartyId = "8453"
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	sourceID.CounterpartyId = "1128614981"
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	// Set 3 entries for CCTP sources.
	sourceID.ProtocolId = core.PROTOCOL_CCTP
	sourceID.CounterpartyId = "1"
	destID.ProtocolId = core.PROTOCOL_IBC
	destID.CounterpartyId = channel0
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	sourceID.CounterpartyId = "2"
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	sourceID.CounterpartyId = "3"
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, UsdcDenom, amount)
	require.NoError(t, err)

	// Set 3 entries for CCTP sources and a different denom
	sourceID.CounterpartyId = "1"
	denom := "unoble"
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, denom, amount)
	require.NoError(t, err)

	sourceID.CounterpartyId = "2"
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, denom, amount)
	require.NoError(t, err)

	sourceID.CounterpartyId = "3"
	err = d.SetDispatchedAmount(ctx, &sourceID, &destID, denom, amount)
	require.NoError(t, err)
}

func TestGetDispatchedAmountBySourceProtocolID(t *testing.T) {
	// ARRANGE
	d, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	// ACT
	daIBC, _, err := d.GetDispatchedAmountsBySourceProtocolID(ctx, core.PROTOCOL_IBC, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daIBC, 0)

	// ACT
	daHyp, _, err := d.GetDispatchedAmountsBySourceProtocolID(ctx, core.PROTOCOL_HYPERLANE, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daHyp, 0)

	// ACT
	daCCTP, _, err := d.GetDispatchedAmountsBySourceProtocolID(ctx, core.PROTOCOL_CCTP, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daCCTP, 0)

	// ARRANGE
	createDispatchedAmountEntries(t, ctx, d)

	// ACT
	daIBC, _, err = d.GetDispatchedAmountsBySourceProtocolID(ctx, core.PROTOCOL_IBC, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daIBC, 3)

	// ACT
	daHyp, _, err = d.GetDispatchedAmountsBySourceProtocolID(ctx, core.PROTOCOL_HYPERLANE, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daHyp, 3)

	// ACT
	daCCTP, _, err = d.GetDispatchedAmountsBySourceProtocolID(ctx, core.PROTOCOL_CCTP, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daCCTP, 6)
}

func TestGetDispatchedAmountByDestinationProtocolID(t *testing.T) {
	// ARRANGE
	d, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	// ACT
	daIBC, _, err := d.GetDispatchedAmountsByDestinationProtocolID(ctx, core.PROTOCOL_IBC, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daIBC, 0)

	// ACT
	daHyp, _, err := d.GetDispatchedAmountsByDestinationProtocolID(
		ctx,
		core.PROTOCOL_HYPERLANE,
		nil,
	)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daHyp, 0)

	daCCTP, _, err := d.GetDispatchedAmountsByDestinationProtocolID(ctx, core.PROTOCOL_CCTP, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daCCTP, 0)

	// ARRANGE
	createDispatchedAmountEntries(t, ctx, d)

	// ACT
	daIBC, _, err = d.GetDispatchedAmountsByDestinationProtocolID(ctx, core.PROTOCOL_IBC, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daIBC, 6)

	// ACT
	daHyp, _, err = d.GetDispatchedAmountsByDestinationProtocolID(ctx, core.PROTOCOL_HYPERLANE, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daHyp, 0)

	// ACT
	daCCTP, _, err = d.GetDispatchedAmountsByDestinationProtocolID(ctx, core.PROTOCOL_CCTP, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, daCCTP, 6)
}

// ====================================================================================================
// Dispatched Counts
// ====================================================================================================

func TestSetHasGetDispatchedCounts(t *testing.T) {
	// ARRANGE
	d, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	sourceID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: channel1,
	}
	destID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// ACT: No dispatched counts records.
	found := d.HasDispatchedCounts(ctx, &sourceID, &destID)
	dc := d.GetDispatchedCounts(ctx, &sourceID, &destID)

	// ASSERT: Verify queries on initial conditions.
	require.False(t, found)
	require.Equal(t, uint64(0), dc.Count)

	// ACT: Set a dispatched count.
	err := d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)

	// ASSERT
	require.NoError(t, err)

	// ACT
	found = d.HasDispatchedCounts(ctx, &sourceID, &destID)
	dc = d.GetDispatchedCounts(ctx, &sourceID, &destID)

	// ASSERT: Verify state updated.
	require.True(t, found)
	require.Equal(t, uint64(1), dc.Count)

	// ACT: Update existing count.
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 10)

	// ASSERT
	require.NoError(t, err)

	// ACT
	dc = d.GetDispatchedCounts(ctx, &sourceID, &destID)

	// ASSERT: Verify the record was updated.
	require.Equal(t, uint64(10), dc.Count)
}

func TestGetAllDispatchedCounts(t *testing.T) {
	// ARRANGE
	d, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx
	createDispatchedCountEntries(t, ctx, d)

	// ACT
	dc := d.GetAllDispatchedCounts(ctx)

	// ASSERT
	require.Len(t, dc, 12)
}

func TestGetDispatchedCountsBySourceProtocolID(t *testing.T) {
	// ARRANGE
	d, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx
	createDispatchedCountEntries(t, ctx, d)

	// ACT
	dc, _, err := d.GetDispatchedCountsBySourceProtocolID(ctx, core.PROTOCOL_IBC, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 6)

	// ACT
	dc, _, err = d.GetDispatchedCountsBySourceProtocolID(ctx, core.PROTOCOL_CCTP, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 3)

	// ACT
	dc, _, err = d.GetDispatchedCountsBySourceProtocolID(ctx, core.PROTOCOL_HYPERLANE, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 3)

	// ACT
	dc, _, err = d.GetDispatchedCountsBySourceProtocolID(ctx, core.ProtocolID(100), nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 0)

	// ACT: check pagination with offset and limit. Out of 6 results, we should get the 2nd, 3th,
	// and 4th.
	dc, pagResp, err := d.GetDispatchedCountsBySourceProtocolID(
		ctx,
		core.PROTOCOL_IBC,
		&query.PageRequest{
			Offset:     1,
			Limit:      3,
			CountTotal: true,
		},
	)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 3)
	require.Equal(t, uint64(6), pagResp.Total)
	// We skip the first entry associated with source counterparty id "channel-1"
	// and destination protocol CCTP, then we have the second entry associated with
	// the same source counterparty ID.
	require.Equal(t, "1234", dc[0].DestinationId.CounterpartyId)
	// We have the entries for source counterparty id "channel-2", first for CCTP,
	// then for Hyperlane.
	require.Equal(t, "0", dc[1].DestinationId.CounterpartyId)
	require.Equal(t, "1234", dc[2].DestinationId.CounterpartyId)

	// ACT: check that passing the next key of previous response to the pagination
	// we obtain only the alst two results.
	dc, _, err = d.GetDispatchedCountsBySourceProtocolID(
		ctx,
		core.PROTOCOL_IBC,
		&query.PageRequest{
			Key: pagResp.NextKey,
		},
	)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 2)
	require.Equal(t, "0", dc[0].DestinationId.CounterpartyId)
	require.Equal(t, "1234", dc[1].DestinationId.CounterpartyId)
}

func TestGetDispatchedCountsByDestinationProtocolID(t *testing.T) {
	// ARRANGE
	d, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx
	createDispatchedCountEntries(t, ctx, d)

	// ACT
	dc, _, err := d.GetDispatchedCountsByDestinationProtocolID(ctx, core.PROTOCOL_IBC, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 3)

	// ACT
	dc, _, err = d.GetDispatchedCountsByDestinationProtocolID(ctx, core.PROTOCOL_CCTP, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 6)

	// ACT
	dc, _, err = d.GetDispatchedCountsByDestinationProtocolID(ctx, core.PROTOCOL_HYPERLANE, nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 3)

	// ACT
	dc, _, err = d.GetDispatchedCountsByDestinationProtocolID(ctx, core.ProtocolID(100), nil)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 0)

	// ACT: check pagination with offset and limit. Out of 6 results, we should get the 2nd, 3th,
	// and 4th.
	dc, pagResp, err := d.GetDispatchedCountsByDestinationProtocolID(
		ctx,
		core.PROTOCOL_CCTP,
		&query.PageRequest{
			Offset:     1,
			Limit:      3,
			CountTotal: true,
		},
	)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 3)
	require.Equal(t, uint64(6), pagResp.Total)
	require.Equal(t, "channel-2", dc[0].SourceId.CounterpartyId)
	require.Equal(t, "channel-3", dc[1].SourceId.CounterpartyId)
	require.Equal(t, "1128614981", dc[2].SourceId.CounterpartyId)

	// ACT: check that passing the next key of previous response to the pagination
	// we obtain only the alst two results.
	dc, _, err = d.GetDispatchedCountsByDestinationProtocolID(
		ctx,
		core.PROTOCOL_CCTP,
		&query.PageRequest{
			Key: pagResp.NextKey,
		},
	)

	// ASSERT
	require.NoError(t, err)
	require.Len(t, dc, 2)
	require.Equal(t, "8453", dc[0].SourceId.CounterpartyId)
	require.Equal(t, "999", dc[1].SourceId.CounterpartyId)
}

func createDispatchedCountEntries(t *testing.T, ctx context.Context, d *dispatcher.Dispatcher) {
	t.Helper()

	sourceID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_IBC,
		CounterpartyId: channel1,
	}
	destID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: "0",
	}

	// Set 6 entries for IBC sources.
	err := d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = channel2
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = channel3
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = channel1
	destID.ProtocolId = core.PROTOCOL_HYPERLANE
	destID.CounterpartyId = "1234"
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = channel2
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = channel3
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	// Set 3 entries for Hyperlane sources.
	destID.ProtocolId = core.PROTOCOL_CCTP
	destID.CounterpartyId = "0"
	sourceID.ProtocolId = core.PROTOCOL_HYPERLANE
	sourceID.CounterpartyId = "999"
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = "8453"
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = "1128614981"
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	// Set 3 entries for CCTP sources.
	sourceID.ProtocolId = core.PROTOCOL_CCTP
	sourceID.CounterpartyId = "1"
	destID.ProtocolId = core.PROTOCOL_IBC
	destID.CounterpartyId = channel0
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = "2"
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)

	sourceID.CounterpartyId = "3"
	err = d.SetDispatchedCounts(ctx, &sourceID, &destID, 1)
	require.NoError(t, err)
}
