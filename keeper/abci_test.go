package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	orbiterkeeper "github.com/noble-assets/orbiter/keeper"
	mockorbiter "github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
)

func TestBeginBlock(t *testing.T) {
	ctx, _, k := mockorbiter.OrbiterKeeper(t)
	ms := orbiterkeeper.NewMsgServer(k)
	qs := orbiterkeeper.NewQueryServer(k)

	pendingPayload := createTestPendingPayloadWithSequence(t, 0, ctx.BlockTime())
	hash, err := pendingPayload.SHA256Hash()
	require.NoError(t, err, "failed to create payload hash")

	payloadBz, err := orbitertypes.MarshalJSON(k.Codec(), pendingPayload.Payload)
	require.NoError(t, err, "failed to marshal payload to json")

	// ASSERT: should run fine with no pending payloads.
	require.NoError(t, k.BeginBlock(ctx), "failed to run begin block")

	// ACT: set up a pending payload.
	res, err := ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{Payload: string(payloadBz)})
	require.NoError(t, err, "failed to submit payload")
	require.Equal(t, hash.String(), res.Hash, "expected different hash")

	// ASSERT: payload should be registered.
	_, err = qs.PendingPayload(ctx, &orbitertypes.QueryPendingPayloadRequest{Hash: hash.String()})
	require.NoError(t, err, "failed to query pending payload")

	// ASSERT: running begin block should NOT remove the just submitted payload.
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour))
	require.NoError(t, k.BeginBlock(ctx), "failed to run begin block with payload registered")

	_, err = qs.PendingPayload(ctx, &orbitertypes.QueryPendingPayloadRequest{Hash: hash.String()})
	require.NoError(t, err, "expected payload to still be registered")

	// ASSERT: after passing the lifespan, the pending payload should be removed.
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(orbiterkeeper.PendingPayloadLifespan))
	require.NoError(t, k.BeginBlock(ctx), "failed to run begin block with expired payload")

	_, err = qs.PendingPayload(ctx, &orbitertypes.QueryPendingPayloadRequest{Hash: hash.String()})
	require.ErrorContains(
		t,
		err,
		fmt.Sprintf("payload with hash %s: not found", hash.String()),
		"expected payload to not be registered",
	)
}
