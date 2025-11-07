// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, NASD Inc. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN "AS IS" BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

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
				Recipient: "",
				FeeType: &actiontypes.FeeInfo_BasisPoints_{
					BasisPoints: &actiontypes.FeeInfo_BasisPoints{
						Value: 0,
					},
				},
			},
			expErr: "fee basis point must be > 0 and < 10000",
		},
		{
			name: "error - over maximum basis points",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "",
				FeeType: &actiontypes.FeeInfo_BasisPoints_{
					BasisPoints: &actiontypes.FeeInfo_BasisPoints{
						Value: actiontypes.BPSNormalizer + 1,
					},
				},
			},
			expErr: "fee basis point must be > 0 and < 10000",
		},
		{
			name: "error - negative amount",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
				FeeType: &actiontypes.FeeInfo_Amount_{
					Amount: &actiontypes.FeeInfo_Amount{
						Value: "-1",
					},
				},
			},
			expErr: "fee amount must be positive",
		},
		{
			name: "error - not a number amount",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
				FeeType: &actiontypes.FeeInfo_Amount_{
					Amount: &actiontypes.FeeInfo_Amount{
						Value: "pizza",
					},
				},
			},
			expErr: "cannot convert",
		},
		{
			name: "error - zero amount",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
				FeeType: &actiontypes.FeeInfo_Amount_{
					Amount: &actiontypes.FeeInfo_Amount{
						Value: "0",
					},
				},
			},
			expErr: "fee amount must be positive",
		},
		{
			name: "error - recipient is empty",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "",
				FeeType: &actiontypes.FeeInfo_BasisPoints_{
					BasisPoints: &actiontypes.FeeInfo_BasisPoints{
						Value: 1,
					},
				},
			},
			expErr: "empty address",
		},
		{
			name: "error - recipient is not valid address",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "a",
				FeeType: &actiontypes.FeeInfo_BasisPoints_{
					BasisPoints: &actiontypes.FeeInfo_BasisPoints{
						Value: 1,
					},
				},
			},
			expErr: "invalid bech32",
		},
		{
			name: "error - basis point content is nil",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "a",
				FeeType: &actiontypes.FeeInfo_BasisPoints_{
					BasisPoints: nil,
				},
			},
			expErr: "fee info bps: invalid nil pointer",
		},
		{
			name: "error - basis point is typed nil",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "a",
				FeeType:   (*actiontypes.FeeInfo_BasisPoints_)(nil),
			},
			expErr: "fee info bps wrapper: invalid nil pointer",
		},
		{
			name: "success - basis point",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
				FeeType: &actiontypes.FeeInfo_BasisPoints_{
					BasisPoints: &actiontypes.FeeInfo_BasisPoints{
						Value: 1,
					},
				},
			},
			expErr: "",
		},
		{
			name: "success - amount",
			feeInfo: &actiontypes.FeeInfo{
				Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
				FeeType: &actiontypes.FeeInfo_Amount_{
					Amount: &actiontypes.FeeInfo_Amount{
						Value: "1",
					},
				},
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

func TestValidateFeeAttributes(t *testing.T) {
	testutil.SetSDKConfig()

	testCases := []struct {
		name    string
		feeInfo *actiontypes.FeeAttributes
		expErr  string
	}{
		{
			name:   "error - nil fee attributes",
			expErr: core.ErrNilPointer.Error(),
		},
		{
			name: "error - over maximum fee recipient",
			feeInfo: &actiontypes.FeeAttributes{
				FeesInfo: func() []*actiontypes.FeeInfo {
					fees := make([]*actiontypes.FeeInfo, actiontypes.MaxFeeRecipients+1)
					for i := range fees {
						fees[i] = &actiontypes.FeeInfo{
							Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
							FeeType: &actiontypes.FeeInfo_BasisPoints_{
								BasisPoints: &actiontypes.FeeInfo_BasisPoints{
									Value: 1,
								},
							},
						}
					}

					return fees
				}(),
			},
			expErr: "maximum fee recipients",
		},
		{
			name: "success - maximum fee recipient",
			feeInfo: &actiontypes.FeeAttributes{
				FeesInfo: func() []*actiontypes.FeeInfo {
					fees := make([]*actiontypes.FeeInfo, actiontypes.MaxFeeRecipients)
					for i := range fees {
						fees[i] = &actiontypes.FeeInfo{
							Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
							FeeType: &actiontypes.FeeInfo_BasisPoints_{
								BasisPoints: &actiontypes.FeeInfo_BasisPoints{
									Value: 1,
								},
							},
						}
					}

					return fees
				}(),
			},
			expErr: "",
		},
		{
			name: "success - mixed fee types",
			feeInfo: &actiontypes.FeeAttributes{
				FeesInfo: func() []*actiontypes.FeeInfo {
					fees := make([]*actiontypes.FeeInfo, 0)
					fees = append(fees,
						&actiontypes.FeeInfo{
							Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
							FeeType: &actiontypes.FeeInfo_BasisPoints_{
								BasisPoints: &actiontypes.FeeInfo_BasisPoints{
									Value: 1,
								},
							},
						},
						&actiontypes.FeeInfo{
							Recipient: "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
							FeeType: &actiontypes.FeeInfo_Amount_{
								Amount: &actiontypes.FeeInfo_Amount{
									Value: "1",
								},
							},
						},
					)

					return fees
				}(),
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
