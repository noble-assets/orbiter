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
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	"github.com/noble-assets/orbiter/keeper/component/forwarder"
)

func NewForwarderComponent(tb testing.TB) (*forwarder.Forwarder, *Dependencies) {
	tb.Helper()

	deps := NewDependencies(tb)

	sb := collections.NewSchemaBuilder(deps.StoreService)

	f, err := forwarder.New(
		deps.EncCfg.Codec,
		sb,
		deps.Logger,
		NewBankKeeper(),
	)
	require.NoError(tb, err)
	_, err = sb.Build()
	require.NoError(tb, err)

	return f, &deps
}
