package forwarder

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
			name: "success - valid genesis state with protocol and cross-chain ids",
			genState: &GenesisState{
				PausedProtocolIds: []core.ProtocolID{core.PROTOCOL_IBC, core.PROTOCOL_CCTP},
				PausedCrossChainIds: []*core.CrossChainID{
					{core.PROTOCOL_IBC, "ibc-1"},
					{core.PROTOCOL_CCTP, "cctp-2"},
				},
			},
		},
		{
			name:     "error - nil genesis state",
			genState: nil,
			expErr:   core.ErrNilPointer.Error(),
		},
		{
			name: "error - invalid genesis state (unsupported protocol)",
			genState: &GenesisState{
				PausedProtocolIds: []core.ProtocolID{core.PROTOCOL_UNSUPPORTED},
			},
			expErr: "invalid paused protocol ID",
		},
		{
			name: "error - invalid genesis state (nil cross-chain id)",
			genState: &GenesisState{
				PausedCrossChainIds: []*core.CrossChainID{nil},
			},
			expErr: "invalid paused cross-chain ID",
		},
		{
			name: "error - invalid genesis state (unsupported cross chain ids)",
			genState: &GenesisState{
				PausedCrossChainIds: []*core.CrossChainID{
					{core.PROTOCOL_UNSUPPORTED, "unsupported-2"},
				},
			},
			expErr: "invalid paused cross-chain ID",
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
