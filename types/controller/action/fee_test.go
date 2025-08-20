package action_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/testutil"
	actiontypes "github.com/noble-assets/orbiter/types/controller/action"
	"github.com/noble-assets/orbiter/types/core"
)

func TestValidateFee(t *testing.T) {
	testutil.SetSDKConfig()

	testCases := []struct {
		name    string
		feeInfo *actiontypes.FeeInfo
		expErr  string
	}{
		{
			name:   "error - nil fee info",
			expErr: core.ErrNilPointer.Error(),
		},
		{
			name: "error - zero basis points",
			feeInfo: &actiontypes.FeeInfo{
				Recipient:   "",
				BasisPoints: 0,
			},
			expErr: "fee basis point must be > 0 and < 10000",
		},
		{
			name: "error - over maximum basis points",
			feeInfo: &actiontypes.FeeInfo{
				Recipient:   "",
				BasisPoints: core.BPSNormalizer + 1,
			},
			expErr: "fee basis point must be > 0 and < 10000",
		},
		{
			name: "error - recipient is empty",
			feeInfo: &actiontypes.FeeInfo{
				Recipient:   "",
				BasisPoints: 1,
			},
			expErr: "empty address",
		},
		{
			name: "error - recipient is not valid address",
			feeInfo: &actiontypes.FeeInfo{
				Recipient:   "a",
				BasisPoints: 1,
			},
			expErr: "invalid bech32",
		},
		{
			name: "success",
			feeInfo: &actiontypes.FeeInfo{
				Recipient:   "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
				BasisPoints: 1,
			},
			expErr: "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := tC.feeInfo.Validate()

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
