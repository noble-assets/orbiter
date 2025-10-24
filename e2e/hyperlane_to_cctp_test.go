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

package e2e

import (
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter"
	"github.com/noble-assets/orbiter/testutil"
)

func TestHyperlaneToCCTP(t *testing.T) {
	testutil.SetSDKConfig()

	ctx, s := NewSuite(t, true, false, true)

	orbiter.RegisterInterfaces(s.Chain.GetCodec().InterfaceRegistry())

	transferAmount := sdkmath.NewInt(2 * OneE6)
	fundAmount := transferAmount.MulRaw(2)
	recipient := testutil.NewNobleAddress()

	relayer := interchaintest.GetAndFundTestUsers(
		t,
		ctx,
		"relayer",
		fundAmount,
		s.Chain,
	)[0]

	node := s.Chain.GetFullNode()

	testPayload, err := createTestPayload()
	require.NoError(t, err, "failed to create test payload")

	inputs := &orbHypTransferInputs{
		amount:            transferAmount,
		destinationDomain: s.hyperlaneDestinationDomain,
		nonce:             0,
		originDomain:      s.hyperlaneOriginDomain,
		payload:           testPayload,
		recipient:         sdktypes.MustAccAddressFromBech32(recipient),
	}

	mailbox, err := getHyperlaneMailbox(ctx, node)
	require.NoError(t, err, "failed to get hyperlane mailbox")

	metadata := ""

	message, err := buildHyperlaneOrbiterMessage(inputs)
	require.NoError(t, err, "failed to build orbiter message")

	// ACT: try to handle `MsgProcessMessage` with payload registered
	_, err = node.ExecTx(
		ctx,
		relayer.KeyName(),
		"hyperlane",
		"mailbox",
		"process",
		mailbox.Id.String(),
		metadata,
		message.String(),
	)
	require.NoError(t, err, "failed to handle orbiter message with hyperlane")
}
