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

	"github.com/noble-assets/orbiter/testutil"
	mockorbiter "github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

func TestAcceptPayload(t *testing.T) {
	t.Parallel()

	validPayload := createTestPendingPayloadWithSequence(t, 0)

	expHash, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		payload     *orbitertypes.PendingPayload
		errContains string
		expHash     []byte
	}{
		{
			name:    "success - valid payload",
			payload: validPayload,
			expHash: expHash.Bytes(),
		},
		{
			name:        "error - nil pending payload",
			payload:     nil,
			errContains: "invalid pending payload: pending payload: invalid nil pointer",
		},
		{
			name:        "error - invalid (empty) pending payload",
			payload:     &orbitertypes.PendingPayload{},
			errContains: "invalid pending payload: payload is not set",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			hash, err := k.AcceptPayload(ctx, tc.payload)

			if tc.errContains == "" {
				require.NoError(t, err, "failed to accept payload")
				require.Equal(t, tc.expHash, hash)
			} else {
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}

func TestGetPendingPayloadWithHash(t *testing.T) {
	t.Parallel()

	validPayload := createTestPendingPayloadWithSequence(t, 0)
	expHash, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		setup       func(*testing.T, context.Context, orbitertypes.HyperlaneStateHandler)
		hash        []byte
		expPayload  *orbitertypes.PendingPayload
		errContains string
	}{
		{
			name: "success - hash found",
			setup: func(t *testing.T, ctx context.Context, handler orbitertypes.HyperlaneStateHandler) {
				t.Helper()

				_, err := handler.AcceptPayload(ctx, validPayload)
				require.NoError(t, err)
			},
			expPayload: validPayload,
			hash:       expHash.Bytes(),
		},
		{
			name:        "error - hash not found",
			setup:       nil,
			expPayload:  validPayload,
			errContains: "pending payload not found",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, _, k := mockorbiter.OrbiterKeeper(t)

			if tc.setup != nil {
				tc.setup(t, ctx, k)
			}

			got, err := k.GetPendingPayloadWithHash(ctx, tc.hash)

			if tc.errContains == "" {
				require.NoError(t, err, "failed to get pending payload")
				require.Equal(t, tc.expPayload.String(), got.String(), "expected different payload")
			} else {
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}

func TestCompletePayload(t *testing.T) {
	t.Parallel()

	validPayload := createTestPendingPayloadWithSequence(t, 0)
	expHash, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		setup       func(*testing.T, context.Context, orbitertypes.HyperlaneStateHandler)
		hash        []byte
		errContains string
	}{
		{
			name: "success - valid payload",
			setup: func(t *testing.T, ctx context.Context, handler orbitertypes.HyperlaneStateHandler) {
				t.Helper()
				_, err := handler.AcceptPayload(ctx, validPayload)
				require.NoError(t, err, "failed to setup testcase; accepting payload")

				gotPayload, err := handler.GetPendingPayloadWithHash(ctx, expHash.Bytes())
				require.NoError(t, err, "error getting pending payload")
				require.Equal(
					t,
					validPayload.String(),
					gotPayload.String(),
					"expected different payload",
				)
			},
			hash: expHash.Bytes(),
		},
		{
			name:  "no-op - valid payload but not found in store",
			setup: nil,
			hash:  expHash.Bytes(),
		},
		{
			name:  "no-op - nil hash",
			setup: nil,
			hash:  nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, _, k := mockorbiter.OrbiterKeeper(t)

			if tc.setup != nil {
				tc.setup(t, ctx, k)
			}

			err = k.CompletePayloadWithHash(ctx, tc.hash)
			if tc.errContains == "" {
				require.NoError(t, err, "failed to complete payload")

				gotPayload, err := k.GetPendingPayloadWithHash(ctx, tc.hash)
				require.Error(t, err, "payload should not be present anymore")
				require.Nil(t, gotPayload, "expected nil payload")
			} else {
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}

// TestSubsequentSubmissions asserts that two subsequently submitted
// identical orbiter payloads generate different hashes so that they can
// be uniquely identified.
func TestSubsequentSubmissions(t *testing.T) {
	ctx, _, k := mockorbiter.OrbiterKeeper(t)

	validPayload := createTestPendingPayloadWithSequence(t, 0)
	expHash, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	// ACT: submit first payload
	res, err := k.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
		Signer:  testutil.NewNobleAddress(),
		Payload: *validPayload.Payload,
	})
	require.NoError(t, err, "failed to submit payload")

	// ASSERT: expected hash is returned
	gotHash := ethcommon.BytesToHash(res.Hash)
	require.Equal(t, expHash.String(), gotHash.String(), "expected different hash")

	// ACT: submit identical payload again
	res2, err := k.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
		Signer:  testutil.NewNobleAddress(),
		Payload: *validPayload.Payload,
	})
	require.NoError(t, err, "failed to submit payload")

	validPayload.Sequence = uint64(1)
	expHash2, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	// ASSERT: expected hash is returned
	gotHash2 := ethcommon.BytesToHash(res2.Hash)
	require.Equal(t, expHash2.String(), gotHash2.String(), "expected different hash")

	// ASSERT: hashes of subsequent submissions of the same payload are different
	require.NotEqual(t, gotHash.String(), gotHash2.String(), "expected different hashes")
}

func TestDifferentSequenceGeneratesDifferentHash(t *testing.T) {
	// ACT: Generate pending payload with sequence 1
	validForwarding := createTestPendingPayloadWithSequence(t, 1)
	expHash, err := validForwarding.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	// ACT: Generate pending payload with sequence 2
	validForwarding2 := createTestPendingPayloadWithSequence(t, 1)
	expHash2, err := validForwarding2.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	// ASSERT: hash 1 and hash 2 are NOT equal
	require.NotEqual(t, expHash.String(), expHash2.String(), "expected different hash")
}

// createTestPendingPayloadWithSequence creates a new example payload that can be submitted
// to the state handler.
func createTestPendingPayloadWithSequence(
	t *testing.T,
	sequence uint64,
) *orbitertypes.PendingPayload {
	t.Helper()

	validForwarding, err := forwarding.NewCCTPForwarding(
		0,
		testutil.RandomBytes(32),
		testutil.RandomBytes(32),
		nil,
	)
	require.NoError(t, err, "failed to create valid forwarding")

	validPayload := &core.Payload{
		PreActions: nil,
		Forwarding: validForwarding,
	}

	return &orbitertypes.PendingPayload{
		Sequence: sequence,
		Payload:  validPayload,
	}
}
