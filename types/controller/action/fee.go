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
	fmt "fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/types/core"
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

func (f *FeeAttributes) Validate() error {
	if f == nil {
		return core.ErrNilPointer.Wrap("fee attributes")
	}

	if len(f.FeesInfo) > core.MaxFeeRecipients {
		return fmt.Errorf(
			"maximum fee recipients %d, received %d",
			core.MaxFeeRecipients,
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

	if f.BasisPoints == 0 || f.BasisPoints > core.BPSNormalizer {
		return fmt.Errorf(
			"fee basis point must be > 0 and < %d, received %d",
			core.BPSNormalizer,
			f.BasisPoints,
		)
	}

	_, err := sdk.AccAddressFromBech32(f.Recipient)

	return err
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
