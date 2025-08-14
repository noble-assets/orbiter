package adapter

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
			name: "success - valid genesis state with new params",
			genState: &GenesisState{
				Params: Params{MaxPassthroughPayloadSize: 1024},
			},
		},
		{
			name:     "error - nil genesis state",
			genState: nil,
			expErr:   core.ErrNilPointer.Error(),
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
