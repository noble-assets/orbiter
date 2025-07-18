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

package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper wraps the bank behaviors expected by the orbiter
// keeper and its components.
type BankKeeper interface {
	BankKeeperOrbit
	BankKeeperAdapter
}

// BankKeeperOrbit represents the bank behavior expected
// by the orbit component.
type BankKeeperOrbit interface {
	// Queries
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// BankKeeperAdapter represents the bank behavior expected
// by the adapter component.
type BankKeeperAdapter interface {
	// Queries
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	// Txs
	SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error
}
