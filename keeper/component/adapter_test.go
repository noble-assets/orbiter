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

package component_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/testutil/mocks"
	"orbiter.dev/types"
)

func TestCheckPassthroughPayloadSize(t *testing.T) {
	// ARRANGE
	adapter, deps := mocks.NewAdapterComponent(t)
	ctx := deps.SdkCtx
	payload := []byte{}

	// ACT: No error when params is not set
	err := adapter.CheckPassthroughPayloadSize(ctx, payload)

	// ASSERT
	require.NoError(t, err)

	// ARRANGE
	err = adapter.SetParams(ctx, types.AdapterParams{
		MaxPassthroughPayloadSize: 0,
	})
	require.NoError(t, err)

	// ACT
	err = adapter.CheckPassthroughPayloadSize(ctx, payload)

	// ASSERT
	require.NoError(t, err)

	// ARRANGE
	payload = []byte("i like you")

	// ACT: Payload exceeds maximum size
	err = adapter.CheckPassthroughPayloadSize(ctx, payload)

	// ASSERT
	require.Error(t, err)

	// ARRANGE
	err = adapter.SetParams(ctx, types.AdapterParams{
		MaxPassthroughPayloadSize: 10,
	})
	require.NoError(t, err)

	// ACT: Payload equals maximum size
	err = adapter.CheckPassthroughPayloadSize(ctx, payload)

	// ASSERT: Works with equal bytes size
	require.NoError(t, err)
}
