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
	"github.com/noble-assets/orbiter/types/hyperlane"
)

func TestPayloadConversion(t *testing.T) {
	testAddr := ethcommon.BytesToAddress(testutil.AddressBytes())
	testAmount := big.NewInt(1000)

	expWarpPayload, err := warptypes.NewWarpPayload(testAddr.Bytes(), *testAmount)
	require.NoError(t, err, "failed to build warp payload")

	bz := expWarpPayload.Bytes()

	testPayloadHash := ethcommon.LeftPadBytes(
		[]byte{1, 0, 0, 1, 1, 0, 0, 0},
		hyperlane.PAYLOAD_HASH_LENGTH,
	)

	fullPayload := make([]byte, hyperlane.ORBITER_PAYLOAD_SIZE)
	copy(fullPayload[:len(bz)], bz)
	copy(fullPayload[len(bz):], testPayloadHash)

	gotHash, err := hyperlane.GetPayloadHashFromWarpMessageBody(fullPayload)
	require.NoError(t, err, "failed to get orbiter hash")
	require.Equal(t, testPayloadHash, gotHash, "expected different payload hash")

	message := hyperlaneutil.HyperlaneMessage{Body: fullPayload}

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
