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

	examplePayload := createTestPendingPayloadWithSequence(t, 0)
	exampleHash, err := examplePayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		setup       func(*testing.T, context.Context, codec.Codec, orbitertypes.MsgServer)
		hash        []byte
		expPayload  *core.PendingPayload
		errContains string
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
			hash:       exampleHash.Bytes(),
		},
		{
			name:        "error - hash not found",
			setup:       nil,
			hash:        exampleHash.Bytes(),
			expPayload:  examplePayload,
			errContains: codes.NotFound.String(),
		},
		{
			name:        "error - nil hash",
			setup:       nil,
			hash:        nil,
			errContains: codes.InvalidArgument.String(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			ms := orbiterkeeper.NewMsgServer(k)
			qs := orbiterkeeper.NewQueryServer(k)

			if tc.setup != nil {
				tc.setup(t, ctx, k.Codec(), ms)
			}

			got, err := qs.PendingPayload(
				ctx,
				&orbitertypes.QueryPendingPayloadRequest{
					Hash: tc.hash,
				})

			if tc.errContains == "" {
				require.NoError(t, err, "failed to get pending payload")
				require.Equal(t, tc.expPayload.String(), got.String(), "expected different payload")
			} else {
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}
