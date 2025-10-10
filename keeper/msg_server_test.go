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
	"encoding/hex"
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

func TestSubmitPayload(t *testing.T) {
	seq := uint64(0)
	examplePayload := createTestPendingPayloadWithSequence(t, seq)

	exampleHash, err := examplePayload.Keccak256Hash()
	require.NoError(t, err, "failed to hash payload")

	destDomain := uint32(1)
	recipient := testutil.RandomBytes(32)

	testcases := []struct {
		name        string
		setup       func(*testing.T, context.Context, *orbiterkeeper.Keeper)
		payload     func() *core.Payload
		errContains string
		expHash     string
	}{
		{
			name:    "success - valid payload",
			payload: func() *core.Payload { return examplePayload.Payload },
			expHash: exampleHash.String(),
		},
		{
			name: "error - payload contains paused action",
			setup: func(t *testing.T, ctx context.Context, k *orbiterkeeper.Keeper) {
				t.Helper()

				err := k.Executor().Pause(ctx, core.ACTION_FEE)
				require.NoError(t, err, "failed to pause fee action")
			},
			payload: func() *core.Payload {
				p := *examplePayload.Payload

				preAction, err := core.NewAction(core.ACTION_FEE, &action.FeeAttributes{})
				require.NoError(t, err, "failed to construct fee action")

				p.PreActions = append(
					p.PreActions,
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
			payload: func() *core.Payload {
				p := *examplePayload.Payload

				fw, err := forwarding.NewCCTPForwarding(destDomain, recipient, nil, nil)
				require.NoError(t, err, "failed to construct forwarding")

				p.Forwarding = fw

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
			payload: func() *core.Payload {
				p := *examplePayload.Payload

				fw, err := forwarding.NewCCTPForwarding(destDomain, recipient, nil, nil)
				require.NoError(t, err, "failed to construct forwarding")

				p.Forwarding = fw

				return &p
			},
			errContains: "cross-chain 2:1 is paused",
		},
		{
			name:        "error - invalid (empty) payload",
			payload:     func() *core.Payload { return &core.Payload{} },
			errContains: "forwarding is not set: invalid nil pointer",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			ms := orbiterkeeper.NewMsgServer(k)

			if tc.setup != nil {
				tc.setup(t, ctx, k)
			}

			payloadJSON, err := orbitertypes.MarshalJSON(k.Codec(), tc.payload())
			require.NoError(t, err, "failed to marshal payload")

			res, err := ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
				Payload: string(payloadJSON),
			})
			if tc.errContains == "" {
				require.NoError(t, err, "failed to submit payload")
				require.Equal(t, tc.expHash, ethcommon.BytesToHash(res.Hash).String())
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
	gotHash := hex.EncodeToString(res.Hash)
	require.Equal(t, hex.EncodeToString(expHash.Bytes()), gotHash, "expected different hash")

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
	gotHash2 := hex.EncodeToString(res2.Hash)
	require.Equal(t, hex.EncodeToString(expHash2.Bytes()), gotHash2, "expected different hash")

	// ASSERT: hashes of subsequent submissions of the same payload are different
	require.NotEqual(t, gotHash, gotHash2, "expected different hashes")
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
	require.NotEqual(
		t,
		hex.EncodeToString(expHash.Bytes()),
		hex.EncodeToString(expHash2.Bytes()),
		"expected different hash",
	)
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
