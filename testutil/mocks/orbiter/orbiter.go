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

package orbiter

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/codec"

	orbiter "orbiter.dev"
	"orbiter.dev/keeper"
	"orbiter.dev/testutil"
	"orbiter.dev/testutil/mocks"
)

func OrbiterKeeper(tb testing.TB) (sdk.Context, *mocks.Mocks, *keeper.Keeper) {
	tb.Helper()

	deps := mocks.NewDependencies(tb)
	mocks := mocks.NewMocks()
	k, ctx := orbiterKeeperWithMocks(&deps, &mocks)

	return ctx, &mocks, k
}

func orbiterKeeperWithMocks(
	deps *mocks.Dependencies,
	m *mocks.Mocks,
) (*keeper.Keeper, sdk.Context) {
	orbiter.RegisterInterfaces(deps.EncCfg.InterfaceRegistry)

	addressCodec := codec.NewBech32Codec("noble")

	k := keeper.NewKeeper(
		deps.EncCfg.Codec,
		addressCodec,
		deps.Logger,
		deps.StoreService,
		testutil.Authority,
		m.BankKeeper,
	)

	return k, deps.SdkCtx
}
