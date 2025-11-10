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

package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	orbiterkeeper "github.com/noble-assets/orbiter/keeper"
	mockorbiter "github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
)

func TestBeginBlock(t *testing.T) {
	ctx, _, k := mockorbiter.OrbiterKeeper(t)
	ms := orbiterkeeper.NewMsgServer(k)
	qs := orbiterkeeper.NewQueryServer(k)

	pendingPayload, err := createTestPendingPayloadWithSequence(0, ctx.BlockTime())
	require.NoError(t, err, "failed to create pending payload")

	hash, err := pendingPayload.SHA256Hash()
	require.NoError(t, err, "failed to create payload hash")

	payloadBz, err := orbitertypes.MarshalJSON(k.Codec(), pendingPayload.Payload)
	require.NoError(t, err, "failed to marshal payload to json")

	// ASSERT: should run fine with no pending payloads.
	require.NoError(t, k.BeginBlock(ctx), "failed to run begin block")

	// ACT: set up a pending payload.
	res, err := ms.SubmitPayload(ctx, &orbitertypes.MsgSubmitPayload{Payload: string(payloadBz)})
	require.NoError(t, err, "failed to submit payload")
	require.Equal(t, hash.String(), res.Hash, "expected different hash")

	// ASSERT: payload should be registered.
	_, err = qs.PendingPayload(ctx, &orbitertypes.QueryPendingPayloadRequest{Hash: hash.String()})
	require.NoError(t, err, "failed to query pending payload")

	// ASSERT: running begin block should NOT remove the just submitted payload.
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Hour))
	require.NoError(t, k.BeginBlock(ctx), "failed to run begin block with payload registered")

	_, err = qs.PendingPayload(ctx, &orbitertypes.QueryPendingPayloadRequest{Hash: hash.String()})
	require.NoError(t, err, "expected payload to still be registered")

	// ASSERT: after passing the lifespan, the pending payload should be removed.
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(orbiterkeeper.PendingPayloadLifespan))
	require.NoError(t, k.BeginBlock(ctx), "failed to run begin block with expired payload")

	_, err = qs.PendingPayload(ctx, &orbitertypes.QueryPendingPayloadRequest{Hash: hash.String()})
	require.ErrorContains(
		t,
		err,
		fmt.Sprintf("payload with hash %s: not found", hash.String()),
		"expected payload to not be registered",
	)
}
