package executor

import (
	"github.com/stretchr/testify/require"
	"orbiter.dev/types/core"
	"testing"
)

func TestValidate(t *testing.T) {
	testcases := []struct {
		name     string
		genState *GenesisState
		expErr   string
	}{
		{
			name:     "success - default genesis state",
			genState: DefaultGenesisState(),
		},
		{
			name: "success - genesis state with paused action ids",
			genState: &GenesisState{
				PausedActionIds: []core.ActionID{core.ACTION_FEE},
			},
		},
		{
			name:     "error - genesis state with paused action ids",
			genState: nil,
			expErr:   core.ErrNilPointer.Error(),
		},
		{
			name: "error - genesis state with invalid action id",
			genState: &GenesisState{
				PausedActionIds: []core.ActionID{core.ACTION_UNSUPPORTED},
			},
			expErr: "action ID is not supported",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.expErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expErr)
			}
		})
	}
}
