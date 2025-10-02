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

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper wraps the bank behaviors expected by the orbiter
// keeper and its components.
type BankKeeper interface {
	BankKeeperForwarder
	BankKeeperAdapter
}

// BankKeeperForwarder represents the bank behavior expected
// by the forwarder component.
type BankKeeperForwarder interface {
	// Queries
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
}

// BankKeeperAdapter represents the bank behavior expected
// by the adapter component.
type BankKeeperAdapter interface {
	// Queries
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	// Txs
	SendCoinsFromModuleToModule(
		ctx context.Context,
		senderModule, recipientModule string,
		amt sdk.Coins,
	) error
}

// HyperlaneCoreKeeper specifies the expected interface of Hyperlane
// core functionality that is required for the Orbiter execution.
type HyperlaneCoreKeeper interface {
	AppRouter() *hyperlaneutil.Router[hyperlaneutil.HyperlaneApp]
}

// HyperlaneWarpKeeper specifies the expected interface of Hyperlane
// warp functionality that is required for the Orbiter execution.
type HyperlaneWarpKeeper interface {
	Handle(
		ctx context.Context,
		mailboxId hyperlaneutil.HexAddress,
		message hyperlaneutil.HyperlaneMessage,
	) error
}

// PendingPayloadsHandler defines the interface to adjust and query the Orbiter module
// state as it relates to the bookkeeping of pending payloads.
type PendingPayloadsHandler interface {
	AcceptPayload(ctx context.Context, payload *PendingPayload) ([]byte, error)
	RemovePendingPayload(ctx context.Context, hash []byte) error
	PendingPayload(ctx context.Context, hash []byte) (*PendingPayload, error)
}
