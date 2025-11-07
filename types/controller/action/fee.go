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

package action

import (
	"errors"
	fmt "fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/types/core"
)

const (
	// BPSNormalizer is used to normalize the basis points
	// defined in a fee action execution.
	BPSNormalizer = 10_000
	// MaxFeeRecipients is the maximum number of addresses that can
	// be specified for a fee payment.
	MaxFeeRecipients = 5
)

func NewFeeAction(feesInfo ...*FeeInfo) (*core.Action, error) {
	attr, err := NewFeeAttributes(feesInfo...)
	if err != nil {
		return nil, err
	}

	return core.NewAction(core.ACTION_FEE, attr)
}

func NewFeeAttributes(feesInfo ...*FeeInfo) (*FeeAttributes, error) {
	attr := FeeAttributes{
		FeesInfo: feesInfo,
	}

	return &attr, attr.Validate()
}

func NewFeeAmount(value string) (*FeeInfo_Amount_, error) {
	amount := &FeeInfo_Amount{
		Value: value,
	}

	if err := validateAmount(amount); err != nil {
		return nil, err
	}

	return &FeeInfo_Amount_{
		Amount: amount,
	}, nil
}

func NewFeeBasisPoints(value uint32) (*FeeInfo_BasisPoints_, error) {
	bps := &FeeInfo_BasisPoints{
		Value: value,
	}

	if err := validateBasisPoints(bps); err != nil {
		return nil, err
	}

	return &FeeInfo_BasisPoints_{
		BasisPoints: bps,
	}, nil
}

func (f *FeeAttributes) Validate() error {
	if f == nil {
		return core.ErrNilPointer.Wrap("fee attributes")
	}

	if len(f.FeesInfo) > MaxFeeRecipients {
		return fmt.Errorf(
			"maximum fee recipients %d, received %d",
			MaxFeeRecipients,
			len(f.FeesInfo),
		)
	}

	for _, i := range f.FeesInfo {
		if err := i.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (f *FeeInfo) Validate() error {
	if f == nil {
		return core.ErrNilPointer.Wrap("fee info")
	}

	if f.GetFeeType() == nil {
		return core.ErrNilPointer.Wrap("fee type")
	}

	switch feeType := f.FeeType.(type) {
	case *FeeInfo_Amount_:
		if feeType == nil {
			return core.ErrNilPointer.Wrap("fee info amount wrapper")
		}
		if feeType.Amount == nil {
			return core.ErrNilPointer.Wrap("fee info amount")
		}
		if err := validateAmount(feeType.Amount); err != nil {
			return err
		}
	case *FeeInfo_BasisPoints_:
		if feeType == nil {
			return core.ErrNilPointer.Wrap("fee info bps wrapper")
		}
		if feeType.BasisPoints == nil {
			return core.ErrNilPointer.Wrap("fee info bps")
		}
		if err := validateBasisPoints(feeType.BasisPoints); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown fee type %T", feeType)
	}

	_, err := sdk.AccAddressFromBech32(f.Recipient)

	return err
}

func validateAmount(amt *FeeInfo_Amount) error {
	val, ok := math.NewIntFromString(amt.GetValue())
	if !ok {
		return fmt.Errorf("cannot convert %s into a number", amt.GetValue())
	}
	if !val.IsPositive() {
		return errors.New("fee amount must be positive")
	}

	return nil
}

func validateBasisPoints(bps *FeeInfo_BasisPoints) error {
	value := bps.GetValue()
	if value == 0 || value > BPSNormalizer {
		return fmt.Errorf(
			"fee basis point must be > 0 and < %d, received %d",
			BPSNormalizer,
			value,
		)
	}

	return nil
}

type RecipientAmount struct {
	Recipient sdk.AccAddress
	Amount    sdk.Coins
}

type FeesToDistribute struct {
	Total  math.Int
	Values []RecipientAmount
}

func NewFeesToDistribute() FeesToDistribute {
	return FeesToDistribute{
		Total:  math.ZeroInt(),
		Values: make([]RecipientAmount, 0),
	}
}
