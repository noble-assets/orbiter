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

package actions_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	controllers "orbiter.dev/controllers/actions"
	"orbiter.dev/testutil"
	"orbiter.dev/testutil/mocks"
	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
	"orbiter.dev/types/controllers/actions"
)

func TestGetAttributesFeeController(t *testing.T) {
	recipient := sdk.AccAddress(testutil.AddressBytes())

	testCases := []struct {
		name          string
		action        func() *types.Action
		expAttributes actions.FeeAttributes
		expErr        string
	}{
		{
			name: "error - nil action",
			action: func() *types.Action {
				return nil
			},
			expErr: "received nil fee attributes",
		},
		{
			name: "error - invalid attributes type",
			action: func() *types.Action {
				action, err := types.NewAction(
					types.ACTION_FEE,
					&testdata.TestActionAttr{Whatever: "works"},
				)
				require.NoError(t, err)
				return action
			},
			expErr: "expected *actions.FeeAttributes",
		},
		{
			name: "error - nil attributes",
			action: func() *types.Action {
				action := types.Action{
					Id:         types.ACTION_FEE,
					Attributes: nil,
				}
				return &action
			},
			expErr: "action attributes are not set",
		},
		{
			name: "success - valid attributes",
			action: func() *types.Action {
				action, err := types.NewAction(
					types.ACTION_FEE,
					&actions.FeeAttributes{
						FeesInfo: []*actions.FeeInfo{
							{
								Recipient:   recipient.String(),
								BasisPoints: 100,
							},
						},
					},
				)
				require.NoError(t, err)
				return action
			},
			expAttributes: actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   recipient.String(),
						BasisPoints: 100,
					},
				},
			},
			expErr: "",
		},
	}

	deps := mocks.NewDependencies(t)
	m := mocks.NewMocks()
	controller, err := controllers.NewFeeController(deps.Logger, m.BankKeeper)
	require.NoError(t, err)

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			attr, err := controller.GetAttributes(tC.action())

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
				require.Nil(t, attr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.expAttributes, *attr)
			}
		})
	}
}

func TestComputeFeesToDistribute(t *testing.T) {
	denom := "uusdc"
	bigNumber, ok := sdkmath.NewIntFromString(
		"115792089237316195423570985008687907853269984665640564039457584007913129639935",
	)
	require.True(t, ok)

	recipient1 := sdk.AccAddress(testutil.AddressBytes())
	recipient2 := sdk.AccAddress(testutil.AddressBytes())

	testCases := []struct {
		name               string
		amount             sdkmath.Int
		feesInfo           []*actions.FeeInfo
		expFeeToDistribute *actions.FeesToDistribute
		expErr             string
	}{
		{
			name:   "success - single fee recipient",
			amount: sdkmath.NewInt(1_000_000),
			feesInfo: []*actions.FeeInfo{
				{
					Recipient:   recipient1.String(),
					BasisPoints: 100,
				},
			},
			expFeeToDistribute: &actions.FeesToDistribute{
				Total: sdkmath.NewInt(10_000), // 1% of 1,000,000 = 10,000
				Values: []actions.RecipientAmount{
					{
						Recipient: recipient1,
						Amount:    sdk.NewCoins(sdk.NewInt64Coin(denom, 10_000)),
					},
				},
			},
		},
		{
			name:   "success - multiple fee recipients",
			amount: sdkmath.NewInt(1_000_000),
			feesInfo: []*actions.FeeInfo{
				{
					Recipient:   recipient1.String(),
					BasisPoints: 100, // 1%
				},
				{
					Recipient:   recipient2.String(),
					BasisPoints: 200, // 2%
				},
			},
			expFeeToDistribute: &actions.FeesToDistribute{
				Total: sdkmath.NewInt(30_000), // 1% + 2% = 3% of 1,000,000 = 30,000
				Values: []actions.RecipientAmount{
					{
						Recipient: recipient1,
						Amount:    sdk.NewCoins(sdk.NewInt64Coin(denom, 10_000)),
					},
					{
						Recipient: recipient2,
						Amount:    sdk.NewCoins(sdk.NewInt64Coin(denom, 20_000)),
					},
				},
			},
		},
		{
			name:   "success - zero amount input",
			amount: sdkmath.ZeroInt(),
			feesInfo: []*actions.FeeInfo{
				{
					Recipient:   recipient1.String(),
					BasisPoints: 100,
				},
			},
			expFeeToDistribute: &actions.FeesToDistribute{
				Total:  sdkmath.ZeroInt(),
				Values: []actions.RecipientAmount{},
			},
		},
		{
			name:     "success - empty fees info",
			amount:   sdkmath.NewInt(1_000_000),
			feesInfo: []*actions.FeeInfo{},
			expFeeToDistribute: &actions.FeesToDistribute{
				Total:  sdkmath.ZeroInt(),
				Values: []actions.RecipientAmount{},
			},
		},
		{
			name:   "success - maximum basis points",
			amount: sdkmath.NewInt(1_000_000),
			feesInfo: []*actions.FeeInfo{
				{
					Recipient:   recipient1.String(),
					BasisPoints: types.BPSNormalizer, // 100%
				},
			},
			expFeeToDistribute: &actions.FeesToDistribute{
				Total: sdkmath.NewInt(1_000_000), // 100%
				Values: []actions.RecipientAmount{
					{
						Recipient: recipient1,
						Amount:    sdk.NewCoins(sdk.NewInt64Coin(denom, 1_000_000)),
					},
				},
			},
		},
		{
			name:   "success - mixed calculations",
			amount: sdkmath.NewInt(1_000_000),
			feesInfo: []*actions.FeeInfo{
				{
					Recipient:   recipient1.String(),
					BasisPoints: 100, // Normal calculation
				},
				{
					Recipient:   recipient2.String(),
					BasisPoints: 1, // Very small basis points
				},
			},
			expFeeToDistribute: &actions.FeesToDistribute{
				Total: sdkmath.NewInt(10_100), // 10000 + 100 = 10100
				Values: []actions.RecipientAmount{
					{
						Recipient: recipient1,
						Amount:    sdk.NewCoins(sdk.NewInt64Coin(denom, 10_000)),
					},
					{
						Recipient: recipient2,
						Amount:    sdk.NewCoins(sdk.NewInt64Coin(denom, 100)),
					},
				},
			},
		},
		{
			name:   "success - basis points resulting in zero fee",
			amount: sdkmath.NewInt(50),
			feesInfo: []*actions.FeeInfo{
				{
					Recipient:   recipient1.String(),
					BasisPoints: 1, // 0.01% of 50 = 0.005, which rounds to 0
				},
			},
			expFeeToDistribute: &actions.FeesToDistribute{
				Total:  sdkmath.ZeroInt(),
				Values: []actions.RecipientAmount{},
			},
		},
		{
			name:   "error - overflow handling returns zero fee",
			amount: bigNumber,
			feesInfo: []*actions.FeeInfo{
				{
					Recipient:   recipient1.String(),
					BasisPoints: 100,
				},
			},
			expFeeToDistribute: &actions.FeesToDistribute{
				Total:  sdkmath.ZeroInt(),
				Values: []actions.RecipientAmount{},
			},
			expErr: "something",
		},
	}

	deps := mocks.NewDependencies(t)
	m := mocks.NewMocks()
	controller, err := controllers.NewFeeController(deps.Logger, m.BankKeeper)
	require.NoError(t, err)

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			result, err := controller.ComputeFeesToDistribute(tC.amount, denom, tC.feesInfo)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NotNil(t, result)
				require.Equal(t, tC.expFeeToDistribute.Total, result.Total)
				require.Equal(t, tC.expFeeToDistribute.Values, result.Values)
			}
		})
	}
}

func TestValidateAttributesFeeController(t *testing.T) {
	recipient := sdk.AccAddress(testutil.AddressBytes())

	testCases := []struct {
		name       string
		attributes *actions.FeeAttributes
		expErr     string
	}{
		{
			name:       "error - nil attributes",
			attributes: nil,
			expErr:     types.ErrNilPointer.Error(),
		},
		{
			name: "success - empty fee info slice",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{},
			},
			expErr: "",
		},
		{
			name: "success - valid single fee",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   recipient.String(),
						BasisPoints: 100,
					},
				},
			},
			expErr: "",
		},
		{
			name: "success - multiple valid fees",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   recipient.String(),
						BasisPoints: 100,
					},
					{
						Recipient:   sdk.AccAddress(testutil.AddressBytes()).String(),
						BasisPoints: 200,
					},
				},
			},
			expErr: "",
		},
		{
			name: "error - nil fee info in slice",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   recipient.String(),
						BasisPoints: 100,
					},
					nil,
				},
			},
			expErr: types.ErrNilPointer.Error(),
		},
		{
			name: "error - zero basis points",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   recipient.String(),
						BasisPoints: 0,
					},
				},
			},
			expErr: "must be greater than zero",
		},
		{
			name: "error - basis points over maximum",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   recipient.String(),
						BasisPoints: types.BPSNormalizer + 1,
					},
				},
			},
			expErr: "cannot be higher",
		},
		{
			name: "error - empty recipient address",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   "",
						BasisPoints: 100,
					},
				},
			},
			expErr: "empty address",
		},
		{
			name: "error - invalid bech32 recipient",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   "invalid_address",
						BasisPoints: 100,
					},
				},
			},
			expErr: "decoding bech32 failed",
		},
		{
			name: "error - fails on first invalid fee in multiple fees",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   recipient.String(),
						BasisPoints: 100,
					},
					{
						Recipient:   "",
						BasisPoints: 200,
					},
					{
						Recipient:   recipient.String(),
						BasisPoints: 300,
					},
				},
			},
			expErr: "empty address",
		},
		{
			name: "success - maximum basis points",
			attributes: &actions.FeeAttributes{
				FeesInfo: []*actions.FeeInfo{
					{
						Recipient:   recipient.String(),
						BasisPoints: types.BPSNormalizer,
					},
				},
			},
			expErr: "",
		},
	}

	deps := mocks.NewDependencies(t)
	m := mocks.NewMocks()
	controller, err := controllers.NewFeeController(deps.Logger, m.BankKeeper)
	require.NoError(t, err)

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := controller.ValidateAttributes(tC.attributes)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateFee(t *testing.T) {
	testCases := []struct {
		name    string
		feeInfo *actions.FeeInfo
		expErr  string
	}{
		{
			name:   "error - nil fee info",
			expErr: types.ErrNilPointer.Error(),
		},
		{
			name: "error - zero basis points",
			feeInfo: &actions.FeeInfo{
				Recipient:   "",
				BasisPoints: 0,
			},
			expErr: "must be greater than zero",
		},
		{
			name: "error - over maximum basis points",
			feeInfo: &actions.FeeInfo{
				Recipient:   "",
				BasisPoints: types.BPSNormalizer + 1,
			},
			expErr: "cannot be higher",
		},
		{
			name: "error - recipient is empty",
			feeInfo: &actions.FeeInfo{
				Recipient:   "",
				BasisPoints: 1,
			},
			expErr: "empty address",
		},
		{
			name: "error - recipient is not valid address",
			feeInfo: &actions.FeeInfo{
				Recipient:   "a",
				BasisPoints: 1,
			},
			expErr: "invalid bech32",
		},
		{
			name: "success",
			feeInfo: &actions.FeeInfo{
				Recipient:   "noble1h8tqx833l3t2s45mwxjz29r85dcevy93wk63za",
				BasisPoints: 1,
			},
			expErr: "",
		},
	}

	deps := mocks.NewDependencies(t)
	m := mocks.NewMocks()
	controller, err := controllers.NewFeeController(deps.Logger, m.BankKeeper)
	require.NoError(t, err)

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := controller.ValidateFee(tC.feeInfo)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandlePacketFeeController(t *testing.T) {
	recipient := sdk.AccAddress(testutil.AddressBytes())
	validAction, err := types.NewAction(
		types.ACTION_FEE,
		&actions.FeeAttributes{
			FeesInfo: []*actions.FeeInfo{
				{
					Recipient:   recipient.String(),
					BasisPoints: 10,
				},
			},
		},
	)
	require.NoError(t, err)
	transferAttr, err := types.NewTransferAttributes(
		types.PROTOCOL_CCTP,
		"1",
		"uusdc",
		sdkmath.NewInt(1_000_000),
	)
	require.NoError(t, err)

	testCases := []struct {
		name            string
		setup           func(*mocks.Mocks)
		action          func() *types.Action
		transferAttr    func() *types.TransferAttributes
		expTransferAttr func() *types.TransferAttributes
		postCheck       func(*mocks.Mocks)
		expErr          string
	}{
		{
			name: "error - invalid attributes",
			action: func() *types.Action {
				action, err := types.NewAction(
					types.ACTION_FEE,
					&testdata.TestActionAttr{Whatever: "works"},
				)
				require.NoError(t, err)
				return action
			},
			transferAttr: func() *types.TransferAttributes {
				return transferAttr
			},
			expErr: "expected *actions.FeeAttributes",
		},
		{
			name: "success - valid packet",
			setup: func(m *mocks.Mocks) {
				m.BankKeeper.Balances[types.ModuleAddress.String()] = sdk.NewCoins(
					sdk.NewInt64Coin("uusdc", 1_000_000_000),
				)
			},
			action: func() *types.Action {
				return validAction
			},
			transferAttr: func() *types.TransferAttributes {
				return transferAttr
			},
			postCheck: func(m *mocks.Mocks) {
				recipientBalance := m.BankKeeper.Balances[recipient.String()]
				require.Len(t, recipientBalance, 1)
				require.Equal(t, sdkmath.NewInt(1_000), recipientBalance[0].Amount)
			},
			expTransferAttr: func() *types.TransferAttributes {
				expTransferAttr := transferAttr
				expTransferAttr.SetDestinationAmount(
					transferAttr.DestinationAmount().SubRaw(1_000),
				)
				return transferAttr
			},
			expErr: "",
		},
	}

	deps := mocks.NewDependencies(t)
	m := mocks.NewMocks()
	controller, err := controllers.NewFeeController(deps.Logger, m.BankKeeper)
	require.NoError(t, err)

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			if tC.setup != nil {
				tC.setup(&m)
			}

			transferAttr := tC.transferAttr()
			packet, err := types.NewActionPacket(transferAttr, tC.action())
			require.NoError(t, err)
			err = controller.HandlePacket(deps.SdkCtx, packet)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				tC.postCheck(&m)
				require.Equal(t, tC.expTransferAttr(), transferAttr)
			}
		})
	}
}
