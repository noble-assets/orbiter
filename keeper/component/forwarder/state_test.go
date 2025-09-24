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

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/noble-assets/orbiter/keeper/component/forwarder"
	"github.com/noble-assets/orbiter/testutil/mocks"
	"github.com/noble-assets/orbiter/types/core"
)

func TestGetPaginatedPausedCrossChains(t *testing.T) {
	f, deps := mocks.NewForwarderComponent(t)
	ctx := deps.SdkCtx

	createPausedCrossChainsEntires(t, ctx, f)

	testCases := []struct {
		name       string
		protocolID core.ProtocolID
		pagination *query.PageRequest
		expLen     int
		postChecks func(counterparties []string, pageResp *query.PageResponse)
	}{
		{
			name:       "no results for protocols not stored",
			protocolID: core.ProtocolID(100),
			expLen:     0,
		},
		{
			name:       "all paused cross-chain for portocol ID 3 with no pagination",
			protocolID: core.ProtocolID(3),
			expLen:     5,
		},
		{
			name:       "all paused cross-chain for portocol ID 3 with pagination",
			protocolID: core.ProtocolID(3),
			pagination: &query.PageRequest{
				Offset:     1,
				Limit:      2,
				CountTotal: true,
			},
			expLen: 2,
			postChecks: func(counterparties []string, pageResp *query.PageResponse) {
				require.Equal(t, "1", counterparties[0])
				require.Equal(t, "2", counterparties[1])

				require.Equal(t, uint64(5), pageResp.Total)
			},
		},
		{
			name:       "all paused cross-chain for portocol ID 3 with pagination reversed",
			protocolID: core.ProtocolID(3),
			pagination: &query.PageRequest{
				Offset:  1,
				Limit:   2,
				Reverse: true,
			},
			expLen: 2,
			postChecks: func(counterparties []string, pageResp *query.PageResponse) {
				require.Equal(t, "3", counterparties[0])
				require.Equal(t, "2", counterparties[1])
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			counterparties, pageResp, err := f.GetPaginatedPausedCrossChains(
				ctx,
				tC.protocolID,
				tC.pagination,
			)
			require.NoError(t, err)
			require.Len(t, counterparties, tC.expLen)

			if tC.postChecks != nil {
				tC.postChecks(counterparties, pageResp)
			}
		})
	}
}

func createPausedCrossChainsEntires(t *testing.T, ctx context.Context, f *forwarder.Forwarder) {
	t.Helper()

	ccIDs := []core.CrossChainID{
		{
			ProtocolId:     3,
			CounterpartyId: "0",
		},
		{
			ProtocolId:     3,
			CounterpartyId: "1",
		},
		{
			ProtocolId:     3,
			CounterpartyId: "2",
		},
		{
			ProtocolId:     3,
			CounterpartyId: "3",
		},
		{
			ProtocolId:     3,
			CounterpartyId: "4",
		},
		{
			ProtocolId:     2,
			CounterpartyId: "0",
		},
		{
			ProtocolId:     2,
			CounterpartyId: "1",
		},
	}

	for _, ccID := range ccIDs {
		err := f.SetPausedCrossChain(ctx, ccID)
		require.NoError(t, err)
	}
}
