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

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	orbiterkeeper "github.com/noble-assets/orbiter/keeper"
	mockorbiter "github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

func TestRemovePayload(t *testing.T) {
	t.Parallel()

	validPayload := createTestPendingPayloadWithSequence(t, 0)
	expHash, err := validPayload.SHA256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		setup       func(*testing.T, context.Context, codec.Codec, orbitertypes.MsgServer)
		hash        *core.PayloadHash
		errContains string
	}{
		{
			name: "success - valid payload",
			setup: func(t *testing.T, ctx context.Context, cdc codec.Codec, ms orbitertypes.MsgServer) {
				t.Helper()

				bz, err := cdc.MarshalJSON(validPayload.Payload)
				require.NoError(t, err, "failed to marshal payload")

				_, err = ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
					Payload: string(bz),
				})
				require.NoError(t, err, "failed to setup testcase; accepting payload")
			},
			hash: expHash,
		},
		{
			name:  "error - valid payload but not found in store",
			setup: nil,
			hash:  expHash,
			errContains: sdkerrors.ErrNotFound.Wrapf(
				"payload with hash %q",
				expHash.String(),
			).Error(),
		},
		{
			name:        "error - nil hash",
			setup:       nil,
			hash:        nil,
			errContains: core.ErrNilPointer.Wrap("payload hash").Error(),
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

			err := k.RemovePendingPayload(ctx, tc.hash)
			if tc.errContains == "" {
				require.NoError(t, err, "failed to remove payload")

				// ASSERT: value with hash was removed.
				gotPayload, err := qs.PendingPayload(
					ctx,
					&orbitertypes.QueryPendingPayloadRequest{
						Hash: tc.hash.String(),
					},
				)
				require.Error(t, err, "payload should not be present anymore")
				require.Nil(t, gotPayload, "expected nil payload")
			} else {
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}

const timeBetweenBlocks = 1 * time.Second

func TestRemovePayloads(t *testing.T) {
	nowUTC := time.Now().UTC()

	testCases := []struct {
		name         string
		setup        func(*testing.T, sdk.Context, codec.Codec, orbitertypes.MsgServer)
		cutoff       time.Time
		expRemaining int
		errContains  string
	}{
		{
			name:         "success - remove only expired payloads",
			setup:        setupPayloadsInState,
			cutoff:       nowUTC.Add(4 * timeBetweenBlocks),
			expRemaining: 2,
		},
		{
			name:   "success - nothing should be removed if all are not expired",
			setup:  setupPayloadsInState,
			cutoff: nowUTC,
		},
		{
			name:   "success - no submitted payloads",
			setup:  nil,
			cutoff: nowUTC.Add(2 * timeBetweenBlocks),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			ms := orbiterkeeper.NewMsgServer(k)

			if tc.setup != nil {
				tc.setup(t, ctx, k.Codec(), ms)
			}

			err := k.RemoveExpiredPayloads(ctx, tc.cutoff)
			if tc.errContains == "" {
				require.NoError(t, err, "failed to remove expired payloads")

				// TODO: check if better assertion can be made without having to make
				// `pendingPayloads` public
			} else {
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}

func setupPayloadsInState(
	t *testing.T,
	ctx sdk.Context,
	codec codec.Codec,
	ms orbitertypes.MsgServer,
) {
	t.Helper()

	nPayloads := 4
	validPayload := createTestPendingPayloadWithSequence(t, 0)
	payloadBz, err := codec.MarshalJSON(validPayload.Payload)
	require.NoError(t, err, "failed to marshal payload")

	for range nPayloads {
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(timeBetweenBlocks))

		_, err := ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
			Payload: string(payloadBz),
		})
		require.NoError(t, err, "failed to submit payload during setup")
	}
}
