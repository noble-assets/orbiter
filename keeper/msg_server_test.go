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

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	orbiterkeeper "github.com/noble-assets/orbiter/keeper"
	mockorbiter "github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

func TestSubmitPayload(t *testing.T) {
	seq := uint64(0)
	validPayload := createTestPendingPayloadWithSequence(t, seq)

	expHash, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		setup       func(*testing.T, context.Context, *orbiterkeeper.Keeper)
		payload     *core.Payload
		errContains string
		expHash     string
	}{
		{
			name:    "success - valid payload",
			payload: validPayload.Payload,
			expHash: expHash.String(),
		},
		{
			name:        "error - invalid (empty) payload",
			payload:     &core.Payload{},
			errContains: "forwarding is not set: invalid nil pointer",
		},
		{
			name: "error - hash already set",
			setup: func(t *testing.T, ctx context.Context, k *orbiterkeeper.Keeper) {
				t.Helper()

				preSeq, err := k.PendingPayloadsSequence.Peek(ctx)
				require.NoError(t, err, "failed to get current payloads sequence")

				_, err = k.AcceptPayload(ctx, validPayload.Payload)
				require.NoError(t, err, "failed to accept payload")

				// NOTE: we're resetting the nonce here to get the exact same hash bytes
				require.NoError(t, k.PendingPayloadsSequence.Set(ctx, preSeq))
			},
			payload:     validPayload.Payload,
			errContains: core.ErrSubmitPayload.Error(),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			ms := orbiterkeeper.NewMsgServer(k)

			if tc.setup != nil {
				tc.setup(t, ctx, k)
			}

			payloadJSON, err := orbitertypes.MarshalJSON(k.Codec(), tc.payload)
			require.NoError(t, err, "failed to marshal payload")

			res, err := ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
				Payload: string(payloadJSON),
			})
			if tc.errContains == "" {
				require.NoError(t, err, "failed to accept payload")
				require.Equal(t, tc.expHash, ethcommon.BytesToHash(res.Hash).String())
			} else {
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}
