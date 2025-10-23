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

package hyperlane_test

import (
	"math/big"
	"testing"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/testutil"
	"github.com/noble-assets/orbiter/types/controller/action"
	"github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
	"github.com/noble-assets/orbiter/types/hyperlane"
)

func TestPayloadConversion(t *testing.T) {
	testutil.SetSDKConfig()

	testAddr := ethcommon.BytesToAddress(testutil.AddressBytes())
	testAmount := big.NewInt(1000)

	expWarpPayload, err := warptypes.NewWarpPayload(testAddr.Bytes(), *testAmount)
	require.NoError(t, err, "failed to build warp payload")

	bz := expWarpPayload.Bytes()

	testPayload := createTestPayload(t)

	testPayloadBz, err := testPayload.Marshal()
	require.NoError(t, err, "failed to marshal test payload")

	fullPayloadBz := make([]byte, len(bz)+len(testPayloadBz))
	copy(fullPayloadBz[:len(bz)], bz)
	copy(fullPayloadBz[len(bz):], testPayloadBz)

	gotPayload, err := hyperlane.GetPayloadFromWarpMessageBody(fullPayloadBz)
	require.NoError(t, err, "failed to get orbiter payload")
	require.Equal(t, testPayload.String(), gotPayload.String(), "expected different payload")

	message := hyperlaneutil.HyperlaneMessage{Body: fullPayloadBz}

	warpMessage, err := hyperlane.GetReducedWarpMessageFromOrbiterMessage(message)
	require.NoError(t, err, "failed to get reduced warp message")
	require.Len(t, warpMessage.Body, 64, "expected different message body length")

	warpPayload, err := warptypes.ParseWarpPayload(warpMessage.Body)
	require.NoError(t, err, "failed to parse warp payload")
	require.Equal(t, testAmount, warpPayload.Amount(), "expected different warp amount")
	require.Equal(t,
		ethcommon.LeftPadBytes(testAddr.Bytes(), hyperlaneutil.HEX_ADDRESS_LENGTH),
		warpPayload.Recipient(),
		"expected different warp recipient",
	)
}

func createTestPayload(t *testing.T) *core.Payload {
	t.Helper()

	feeAction, err := action.NewFeeAction(&action.FeeInfo{
		Recipient:   "noble1vpq5ecul8v3ecq5hl4dhneekwu08sjwkugm979",
		BasisPoints: 123,
	})
	require.NoError(t, err, "failed to build fee action")

	fw, err := forwarding.NewCCTPForwarding(
		0,
		testutil.RandomBytes(32),
		testutil.RandomBytes(32),
		nil,
	)
	require.NoError(t, err, "failed to build forwarding")

	payload, err := core.NewPayload(fw, feeAction)
	require.NoError(t, err, "failed to build payload")

	return payload
}
