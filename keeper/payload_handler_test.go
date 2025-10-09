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
	"fmt"
	"strings"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	orbiterkeeper "github.com/noble-assets/orbiter/keeper"
	"github.com/noble-assets/orbiter/testutil"
	mockorbiter "github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/controller/action"
	"github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

func TestSubmit(t *testing.T) {
	seq := uint64(0)
	destDomain := uint32(1)
	recipient := testutil.RandomBytes(32)
	validPayload := createTestPendingPayloadWithSequence(t, seq)

	expHash, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		setup       func(*testing.T, context.Context, *orbiterkeeper.Keeper)
		payload     func() *core.PendingPayload
		errContains string
		expHash     string
	}{
		{
			name:    "success - valid payload",
			payload: func() *core.PendingPayload { return validPayload },
			expHash: expHash.String(),
		},
		{
			name: "error - payload contains paused action",
			setup: func(t *testing.T, ctx context.Context, k *orbiterkeeper.Keeper) {
				t.Helper()

				err := k.Executor().Pause(ctx, core.ACTION_FEE)
				require.NoError(t, err, "failed to pause fee action")
			},
			payload: func() *core.PendingPayload {
				p := *validPayload

				preAction, err := core.NewAction(core.ACTION_FEE, &action.FeeAttributes{})
				require.NoError(t, err, "failed to construct fee action")

				p.Payload.PreActions = append(
					p.Payload.PreActions,
					preAction,
				)

				return &p
			},
			errContains: "action ACTION_FEE is paused",
		},
		{
			name: "error - payload contains paused protocol",
			setup: func(t *testing.T, ctx context.Context, k *orbiterkeeper.Keeper) {
				t.Helper()

				err := k.Forwarder().Pause(ctx, core.PROTOCOL_CCTP, nil)
				require.NoError(t, err, "failed to unpause fee action")
			},
			payload: func() *core.PendingPayload {
				p := *validPayload

				fw, err := forwarding.NewCCTPForwarding(destDomain, recipient, nil, nil)
				require.NoError(t, err, "failed to construct forwarding")

				p.Payload.Forwarding = fw

				return &p
			},
			errContains: "protocol PROTOCOL_CCTP is paused",
		},
		{
			name: "error - payload contains paused cross-chain",
			setup: func(t *testing.T, ctx context.Context, k *orbiterkeeper.Keeper) {
				t.Helper()

				cID := (&forwarding.CCTPAttributes{
					DestinationDomain: destDomain,
					MintRecipient:     recipient,
				}).CounterpartyID()

				err = k.Forwarder().Pause(ctx, core.PROTOCOL_CCTP, []string{cID})
				require.NoError(t, err, "failed to unpause cross-chain forwarding")
			},
			payload: func() *core.PendingPayload {
				p := *validPayload

				fw, err := forwarding.NewCCTPForwarding(destDomain, recipient, nil, nil)
				require.NoError(t, err, "failed to construct forwarding")

				p.Payload.Forwarding = fw

				return &p
			},
			errContains: "cross-chain 2:1 is paused",
		},
		{
			name: "error - invalid (empty) payload",
			payload: func() *core.PendingPayload {
				return &core.PendingPayload{Payload: &core.Payload{}}
			},
			errContains: "forwarding is not set: invalid nil pointer",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)

			if tc.setup != nil {
				tc.setup(t, ctx, k)
			}

			pendingPayload := tc.payload()

			gotHash, err := k.Submit(ctx, pendingPayload.Payload)
			if tc.errContains == "" {
				require.NoError(t, err, "failed to accept payload")

				// ASSERT: expected hash returned
				require.Equal(t, tc.expHash, ethcommon.BytesToHash(gotHash).String())

				// ASSERT: expected event emitted
				events := ctx.EventManager().Events()
				require.Len(t, events, 1, "expected 1 event, got %d", len(events))

				found := false
				for _, e := range events {
					if strings.Contains(e.Type, "EventPayloadSubmitted") {
						require.False(t, found, "expected event to be emitted just once")
						found = true
					}
				}
				require.True(t, found, "expected event payload submitted to be found")
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
		setup       func(*testing.T, context.Context, orbitertypes.PendingPayloadsHandler)
		hash        []byte
		expPayload  *core.PendingPayload
		errContains string
	}{
		{
			name: "success - hash found",
			setup: func(t *testing.T, ctx context.Context, handler orbitertypes.PendingPayloadsHandler) {
				t.Helper()

				_, err := handler.Submit(ctx, validPayload.Payload)
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

			got, err := k.PendingPayload(ctx, tc.hash)

			if tc.errContains == "" {
				require.NoError(t, err, "failed to get pending payload")
				require.Equal(t, tc.expPayload.String(), got.String(), "expected different payload")
			} else {
				require.ErrorContains(t, err, tc.errContains, "expected different error")
			}
		})
	}
}

func TestRemovePayload(t *testing.T) {
	t.Parallel()

	validPayload := createTestPendingPayloadWithSequence(t, 0)
	expHash, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	testcases := []struct {
		name        string
		setup       func(*testing.T, context.Context, orbitertypes.PendingPayloadsHandler)
		hash        []byte
		errContains string
	}{
		{
			name: "success - valid payload",
			setup: func(t *testing.T, ctx context.Context, handler orbitertypes.PendingPayloadsHandler) {
				t.Helper()
				_, err := handler.Submit(ctx, validPayload.Payload)
				require.NoError(t, err, "failed to setup testcase; accepting payload")

				gotPayload, err := handler.PendingPayload(ctx, expHash.Bytes())
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
			name:        "error - valid payload but not found in store",
			setup:       nil,
			hash:        expHash.Bytes(),
			errContains: fmt.Sprintf("payload with hash %s not found", expHash.Hex()),
		},
		{
			name:        "error - nil hash",
			setup:       nil,
			hash:        nil,
			errContains: fmt.Sprintf("payload with hash %s not found", ethcommon.Hash{}.Hex()),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, _, k := mockorbiter.OrbiterKeeper(t)

			if tc.setup != nil {
				tc.setup(t, ctx, k)
			}

			err := k.RemovePendingPayload(ctx, tc.hash)
			if tc.errContains == "" {
				require.NoError(t, err, "failed to remove payload")

				// ASSERT: value with hash was removed.
				gotPayload, err := k.PendingPayload(ctx, tc.hash)
				require.Error(t, err, "payload should not be present anymore")
				require.Nil(t, gotPayload, "expected nil payload")

				// ASSERT: event was emitted.
				found := false
				for _, event := range ctx.EventManager().ABCIEvents() {
					if strings.Contains(event.Type, "EventPayloadRemoved") {
						require.False(t, found, "event should only be emitted once")

						found = true
					}
				}
				require.True(t, found, "expected event to be emitted")
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
	ms := orbiterkeeper.NewMsgServer(k)

	validPayload := createTestPendingPayloadWithSequence(t, 0)
	expHash, err := validPayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	validPayloadJSON, err := orbitertypes.MarshalJSON(k.Codec(), validPayload.Payload)
	require.NoError(t, err, "failed to marshal payload into json")

	// ACT: submit first payload
	res, err := ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
		Signer:  testutil.NewNobleAddress(),
		Payload: string(validPayloadJSON),
	})
	require.NoError(t, err, "failed to submit payload")

	// ASSERT: expected hash is returned
	gotHash := ethcommon.BytesToHash(res.Hash)
	require.Equal(t, expHash.String(), gotHash.String(), "expected different hash")

	// ACT: submit identical payload again
	res2, err := ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
		Signer:  testutil.NewNobleAddress(),
		Payload: string(validPayloadJSON),
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
	seq := uint64(1)
	validForwarding := createTestPendingPayloadWithSequence(t, seq)
	expHash, err := validForwarding.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	// ACT: Generate pending payload with sequence 2
	validForwarding2 := createTestPendingPayloadWithSequence(t, seq+1)
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
) *core.PendingPayload {
	t.Helper()

	recipient := make([]byte, 32)
	copy(recipient[32-3:], []byte{1, 2, 3})

	validForwarding, err := forwarding.NewCCTPForwarding(
		0,
		recipient,
		recipient,
		nil,
	)
	require.NoError(t, err, "failed to create valid forwarding")

	validPayload := &core.Payload{
		PreActions: nil,
		Forwarding: validForwarding,
	}

	return &core.PendingPayload{
		Sequence: sequence,
		Payload:  validPayload,
	}
}
