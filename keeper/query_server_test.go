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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"

	"github.com/cosmos/cosmos-sdk/codec"

	orbiterkeeper "github.com/noble-assets/orbiter/keeper"
	mockorbiter "github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

func TestPendingPayload(t *testing.T) {
	t.Parallel()

	nowUTC := time.Now().UTC()

	examplePayload, err := createTestPendingPayloadWithSequence(0, nowUTC)
	require.NoError(t, err, "failed to create test payload")

	exampleHash, err := examplePayload.SHA256Hash()
	require.NoError(t, err, "failed to hash payload")

	testCases := []struct {
		name       string
		setup      func(*testing.T, context.Context, codec.Codec, orbitertypes.MsgServer)
		expPayload *core.PendingPayload
		req        *orbitertypes.QueryPendingPayloadRequest
		expError   string
	}{
		{
			name: "success - hash found",
			setup: func(t *testing.T, ctx context.Context, cdc codec.Codec, ms orbitertypes.MsgServer) {
				t.Helper()

				bz, err := cdc.MarshalJSON(examplePayload.Payload)
				require.NoError(t, err, "failed to marshal payload")

				_, err = ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
					Payload: string(bz),
				})
				require.NoError(t, err)
			},
			expPayload: examplePayload,
			req: &orbitertypes.QueryPendingPayloadRequest{
				Hash: exampleHash.String(),
			},
		},
		{
			name:  "error - hash not found",
			setup: nil,
			req: &orbitertypes.QueryPendingPayloadRequest{
				Hash: exampleHash.String(),
			},
			expError: codes.NotFound.String(),
		},
		{
			name:     "error - nil request",
			setup:    nil,
			req:      nil,
			expError: codes.InvalidArgument.String(),
		},
		{
			name:  "error - empty hash",
			setup: nil,
			req: &orbitertypes.QueryPendingPayloadRequest{
				Hash: "",
			},
			expError: codes.InvalidArgument.String(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			ms := orbiterkeeper.NewMsgServer(k)
			qs := orbiterkeeper.NewQueryServer(k)

			// NOTE: we have to set this because the block time is included in the hash value.
			ctx = ctx.WithBlockTime(nowUTC)

			if tc.setup != nil {
				tc.setup(t, ctx, k.Codec(), ms)
			}

			got, err := qs.PendingPayload(
				ctx,
				tc.req,
			)

			if tc.expError == "" {
				require.NoError(t, err, "failed to get pending payload")
				require.Equal(
					t,
					tc.expPayload.Payload.String(),
					got.Payload.String(),
					"expected different payload",
				)
			} else {
				require.ErrorContains(t, err, tc.expError, "expected different error")
			}
		})
	}
}
