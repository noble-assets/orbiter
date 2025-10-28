package adapter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/controller/adapter"
)

func TestRecoverNativeDenom(t *testing.T) {
	testCases := []struct {
		name          string
		denom         string
		sourcePort    string
		sourceChannel string
		expErr        string
		expDenom      string
	}{
		{
			name:          "error - non native denom (multi-hop)",
			denom:         "transfer/channel-2/transfer/channel-3/uusdc",
			sourcePort:    "transfer",
			sourceChannel: "channel-1",
			expErr:        "coin is native of source chain",
		},
		{
			name:          "error - non native denom (multi-hop)",
			denom:         "transfer/channel-1/transfer/channel-2/uusdc",
			sourcePort:    "transfer",
			sourceChannel: "channel-1",
			expErr:        "orbiter supports only native tokens",
		},
		{
			name:          "success - native denom",
			denom:         "transfer/channel-1/uusdc",
			sourcePort:    "transfer",
			sourceChannel: "channel-1",
			expDenom:      "uusdc",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			denom, err := adapter.RecoverNativeDenom(tC.denom, tC.sourcePort, tC.sourceChannel)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.expDenom, denom)
			}
		})
	}
}
