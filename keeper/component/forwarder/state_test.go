package forwarder_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/testutil/mocks"
	"orbiter.dev/types/core"
)

func TestGetPausedCrossChains(t *testing.T) {
	ccIDs := []core.CrossChainID{
		{
			ProtocolId:     1,
			CounterpartyId: "one",
		},
		{
			ProtocolId:     1,
			CounterpartyId: "two",
		},
		{
			ProtocolId:     1,
			CounterpartyId: "three",
		},
		{
			ProtocolId:     2,
			CounterpartyId: "one",
		},
		{
			ProtocolId:     2,
			CounterpartyId: "two",
		},
	}

	f, deps := mocks.NewForwarderComponent(t)
	ctx := deps.SdkCtx

	for _, ccID := range ccIDs {
		err := f.SetPausedCrossChain(ctx, ccID)
		require.NoError(t, err)
	}

	t.Run("success - no results for protocols not stored", func(t *testing.T) {
		id := core.ProtocolID(100)
		paused, err := f.GetPausedCrossChains(ctx, &id)
		require.NoError(t, err)
		require.Len(t, paused, 0)

		idPaused, found := paused[0]
		require.False(t, found)
		require.Len(t, idPaused, 0)
	})

	t.Run("success - paused cross-chains for protocol ID 1", func(t *testing.T) {
		id := core.ProtocolID(1)
		paused, err := f.GetPausedCrossChains(ctx, &id)
		require.NoError(t, err)
		require.Len(t, paused, 1)

		idPaused, found := paused[1]
		require.True(t, found)
		require.Len(t, idPaused, 3)
	})

	t.Run("success - paused cross-chains for protocol ID 2", func(t *testing.T) {
		id := core.ProtocolID(2)
		paused, err := f.GetPausedCrossChains(ctx, &id)
		require.NoError(t, err)
		require.Len(t, paused, 1)

		idPaused, found := paused[2]
		require.True(t, found)
		require.Len(t, idPaused, 2)
	})

	t.Run("success - all paused cross chains", func(t *testing.T) {
		paused, err := f.GetPausedCrossChains(ctx, nil)
		require.NoError(t, err)
		require.Len(t, paused, 2)
	})
}
