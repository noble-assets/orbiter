package components_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/testutil/mocks"
	"orbiter.dev/types"
)

func TestCheckPassthroughPayloadSize(t *testing.T) {
	// ARRANGE
	adapter, deps := mocks.NewAdapterComponent(t)
	ctx := deps.SdkCtx
	payload := []byte{}

	// ACT: No error when params is not set
	err := adapter.CheckPassthroughPayloadSize(ctx, payload)

	// ASSERT
	require.NoError(t, err)

	// ARRANGE
	err = adapter.SetParams(ctx, types.AdapterParams{
		MaxPassthroughPayloadSize: 0,
	})
	require.NoError(t, err)

	// ACT
	err = adapter.CheckPassthroughPayloadSize(ctx, payload)

	// ASSERT
	require.NoError(t, err)

	// ARRANGE
	payload = []byte("i like you")

	// ACT: Payload exceeds maximum size
	err = adapter.CheckPassthroughPayloadSize(ctx, payload)

	// ASSERT
	require.Error(t, err)

	// ARRANGE
	err = adapter.SetParams(ctx, types.AdapterParams{
		MaxPassthroughPayloadSize: 10,
	})
	require.NoError(t, err)

	// ACT: Payload exceeds maximum size
	err = adapter.CheckPassthroughPayloadSize(ctx, payload)

	// ASSERT: Works with equal bytes size
	require.NoError(t, err)
}
