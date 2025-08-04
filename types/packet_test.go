package types_test

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/types"
)

func TestOrbitID(t *testing.T) {
	testCases := []struct {
		name       string
		orbitID    types.OrbitID
		expectedID string
	}{
		{
			name: "IBC orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_IBC,
				CounterpartyID: "channel-1",
			},
			expectedID: fmt.Sprintf("%d:channel-1", types.PROTOCOL_IBC),
		},
		{
			name: "CCTP orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_CCTP,
				CounterpartyID: "0",
			},
			expectedID: fmt.Sprintf("%d:0", types.PROTOCOL_CCTP),
		},
		{
			name: "Hyperlane orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_HYPERLANE,
				CounterpartyID: "ethereum",
			},
			expectedID: fmt.Sprintf("%d:ethereum", types.PROTOCOL_HYPERLANE),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			id := tC.orbitID.ID()
			require.Equal(t, tC.expectedID, id)
		})
	}
}

func TestParseOrbitID(t *testing.T) {
	testCases := []struct {
		name                   string
		id                     string
		expectedProtocolID     types.ProtocolID
		expectedCounterpartyID string
		expErr                 string
	}{
		{
			name:   "error - when invalid format (no colon)",
			id:     "1channel-1",
			expErr: "invalid orbit",
		},
		{
			name:   "error - when non numeric protocol ID",
			id:     "invalid:channel-1",
			expErr: "invalid protocol",
		},
		{
			name:   "error - when empty string",
			id:     "",
			expErr: "invalid orbit",
		},
		{
			name:                   "error - when the format is not valid (multiple colons)",
			id:                     "1:channel:1",
			expectedProtocolID:     types.PROTOCOL_IBC,
			expectedCounterpartyID: "channel:1",
		},
		{
			name:                   "success - with valid IBC ID",
			id:                     "1:channel-1",
			expectedProtocolID:     types.PROTOCOL_IBC,
			expectedCounterpartyID: "channel-1",
		},
		{
			name:                   "success - with valid CCTP ID",
			id:                     "2:0",
			expectedProtocolID:     types.PROTOCOL_CCTP,
			expectedCounterpartyID: "0",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			orbitID, err := types.ParseOrbitID(tC.id)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.expectedProtocolID, orbitID.ProtocolID)
				require.Equal(t, tC.expectedCounterpartyID, orbitID.CounterpartyID)
			}
		})
	}
}
