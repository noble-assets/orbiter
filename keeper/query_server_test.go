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

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/query"

	orbiterkeeper "github.com/noble-assets/orbiter/keeper"
	"github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
)

func TestPendingPayloads(t *testing.T) {
	examplePayload := createTestPendingPayloadWithSequence(t, 0).Payload

	testcases := []struct {
		name       string
		nPayloads  int
		pagination *query.PageRequest
		expLen     int
	}{
		{
			name:      "success - no hashes stored",
			nPayloads: 0,
			expLen:    0,
		},
		{
			name:      "success - 1 hashes stored",
			nPayloads: 1,
			expLen:    1,
		},
		{
			name:      "success - 5 hashes stored",
			nPayloads: 5,
			expLen:    5,
		},
		{
			name:       "success - 5 hashes stored with 2 pagination",
			nPayloads:  5,
			pagination: &query.PageRequest{Offset: 1, Limit: 2},
			expLen:     2,
		},
		{
			name:       "success - 5 hashes stored with offset out of range so no results returned",
			nPayloads:  5,
			pagination: &query.PageRequest{Offset: 6, Limit: 2},
			expLen:     0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := orbiter.OrbiterKeeper(t)
			ms := orbiterkeeper.NewQueryServer(k)

			for range tc.nPayloads {
				_, err := k.Submit(ctx, examplePayload)
				require.NoError(t, err, "failed to setup payloads")
			}

			res, err := ms.PendingPayloads(
				ctx,
				&orbitertypes.QueryPendingPayloadsRequest{
					Pagination: tc.pagination,
				},
			)
			require.NoError(t, err, "failed to query pending payloads")
			require.Equal(
				t,
				tc.expLen,
				len(res.Hashes),
				"expected different number of hashes returned",
			)
		})
	}
}
