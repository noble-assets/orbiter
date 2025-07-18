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

package mocks

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"orbiter.dev/types"
)

var _ types.BankKeeper = BankKeeper{}

type BankKeeper struct {
	Balances map[string]sdk.Coins
}

func (k BankKeeper) GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins {
	coins, ok := k.Balances[addr.String()]
	if !ok {
		return sdk.Coins{}
	}

	return coins
}

func (k BankKeeper) SendCoins(
	ctx context.Context,
	fromAddr sdk.AccAddress,
	toAddr sdk.AccAddress,
	amt sdk.Coins,
) error {
	if CheckIfFailing(ctx) {
		return errors.New("error sending coins")
	}

	fromCoins, found := k.Balances[fromAddr.String()]
	if !found {
		return errors.New("from account not found")
	}

	fromFinalCoins, negativeAmt := fromCoins.SafeSub(amt...)
	if negativeAmt {
		return errors.New("error during coins deduction")
	}

	toCoins, found := k.Balances[toAddr.String()]
	if !found {
		toCoins = sdk.Coins{}
	}

	toFinalCoins := toCoins.Add(amt...)

	k.Balances[fromAddr.String()] = fromFinalCoins
	k.Balances[toAddr.String()] = toFinalCoins

	return nil
}
