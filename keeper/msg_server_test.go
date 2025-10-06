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

			if tc.setup != nil {
				tc.setup(t, ctx, k)
			}

			res, err := k.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{
				Payload: *tc.payload,
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
