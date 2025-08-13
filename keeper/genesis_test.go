package keeper_test

import (
	adaptertypes "orbiter.dev/types/component/adapter"
	executortypes "orbiter.dev/types/component/executor"
	forwardertypes "orbiter.dev/types/component/forwarder"
	"testing"

	"github.com/stretchr/testify/require"
	mockorbiter "orbiter.dev/testutil/mocks/orbiter"
	orbitertypes "orbiter.dev/types"
)

func TestInitGenesis(t *testing.T) {
	testcases := []struct {
		name     string
		genState orbitertypes.GenesisState
		expPanic bool
	}{
		{
			name:     "success - default genesis state",
			genState: *orbitertypes.DefaultGenesisState(),
			expPanic: false,
		},
		{
			name: "error - nil adapter genesis state",
			genState: orbitertypes.GenesisState{
				AdapterGenesis:   nil,
				ExecutorGenesis:  executortypes.DefaultGenesisState(),
				ForwarderGenesis: forwardertypes.DefaultGenesisState(),
			},
			expPanic: true,
		},
		{
			name: "error - nil executor genesis state",
			genState: orbitertypes.GenesisState{
				AdapterGenesis:   adaptertypes.DefaultGenesisState(),
				ExecutorGenesis:  nil,
				ForwarderGenesis: forwardertypes.DefaultGenesisState(),
			},
			expPanic: true,
		},
		{
			name: "error - nil forwarder genesis state",
			genState: orbitertypes.GenesisState{
				AdapterGenesis:   adaptertypes.DefaultGenesisState(),
				ExecutorGenesis:  executortypes.DefaultGenesisState(),
				ForwarderGenesis: nil,
			},
			expPanic: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)

			if tc.expPanic {
				require.Panics(t, func() {
					k.InitGenesis(ctx, tc.genState)
				})
			} else {
				require.NotPanics(t, func() {
					k.InitGenesis(ctx, tc.genState)
				})
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	ctx, _, k := mockorbiter.OrbiterKeeper(t)
	defaultGenState := orbitertypes.DefaultGenesisState()
	k.InitGenesis(ctx, *defaultGenState)

	genState := k.ExportGenesis(ctx)
	// NOTE: we're comparing the `String()` here to avoid the difference in the way that Go gets the nil vs. empty slice here
	require.Equal(t, defaultGenState.String(), genState.String())
}
