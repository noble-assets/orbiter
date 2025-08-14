package dispatcher_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/testutil/mocks"
	dispatchertypes "orbiter.dev/types/component/dispatcher"
)

func TestInitGenesis(t *testing.T) {
	d, deps := mocks.NewDispatcherComponent(t)
	ctx := deps.SdkCtx

	var g *dispatchertypes.GenesisState

	err := d.InitGenesis(ctx, g)
	require.ErrorContains(t, err, "nil")
}
