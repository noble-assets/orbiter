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
	"strings"
	"testing"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/noble-assets/orbiter/v2"
	"github.com/noble-assets/orbiter/v2/testutil"
	orbitertypes "github.com/noble-assets/orbiter/v2/types"
	forwardingtypes "github.com/noble-assets/orbiter/v2/types/controller/forwarding"
	"github.com/noble-assets/orbiter/v2/types/core"
)

// TestIBCToHyperlane tests the "auto-lane" flow, which forwards an incoming IBC packet,
// which contains a valid orbiter payload through the Hyperlane bridge.
//
// NOTE: here we are not testing any actions or general failure cases as those
// are sufficiently covered in the IBC-to-CCTP case.
func TestIBCToHyperlane(t *testing.T) {
	testutil.SetSDKConfig()
	// NOTE: this has to also include an IBC connected chain which is used to send the orbiter
	// payload to the module.
	ctx, s := NewSuite(t, true, true, true)

	orbiter.RegisterInterfaces(s.Chain.GetCodec().InterfaceRegistry())

	fromOrbiterChannelID, toOrbiterChannelID := s.GetChannels(t, ctx)

	// This creates the corresponding IBC coin as it is available on the connected IBC chain
	// when USDC is sent to it from the orbiter.
	fundAmount := sdkmath.NewInt(2 * OneE6)
	fundedIBCCoin := transfertypes.GetTransferCoin(
		"transfer",
		toOrbiterChannelID, // need to use this because we need to use the channel ID from the perspective of the receiving chain
		uusdcDenom,
		fundAmount,
	)

	// Fund the account on the counterparty chain such that it can send USDC
	// via IBC to the orbiter.
	ibcRecipient := s.IBC.CounterpartySender.FormattedAddress()
	s.FundIBCRecipient(
		t,
		ctx,
		fundedIBCCoin.Amount,
		ibcRecipient,
		fromOrbiterChannelID,
		fundedIBCCoin.Denom,
	)

	unwrappedTokenID, err := hyperlaneutil.DecodeHexAddress(s.hyperlaneToken.Id)
	require.NoError(t, err, "failed to decode token id")

	customHookID := []byte{}
	customHookMetadata := ""
	passthroughPayload := []byte{}

	forwarding, err := forwardingtypes.NewHyperlaneForwarding(
		unwrappedTokenID.Bytes(),
		s.hyperlaneDestinationDomain,
		s.mintRecipient,
		customHookID,
		customHookMetadata,
		sdkmath.ZeroInt(),
		sdk.NewInt64Coin(uusdcDenom, 1e4),
		passthroughPayload,
	)
	require.NoError(t, err, "failed to create hyperlane forwarding")

	p, err := core.NewPayloadWrapper(forwarding)
	require.NoError(t, err, "failed to create payload wrapper")

	payloadBytes, err := orbitertypes.MarshalJSON(s.Chain.GetCodec(), p)
	require.NoError(t, err, "failed to marshal payload")

	height, err := s.Chain.Height(ctx)
	require.NoError(t, err, "failed to query height")

	amountToSend := fundAmount.SubRaw(1e5)
	transferAmount := ibc.WalletAmount{
		Address: core.ModuleAddress.String(),
		Denom:   fundedIBCCoin.Denom,
		Amount:  amountToSend,
	}

	_, err = s.IBC.CounterpartyChain.SendIBCTransfer(
		ctx,
		toOrbiterChannelID,
		s.IBC.CounterpartySender.KeyName(),
		transferAmount,
		ibc.TransferOptions{
			Memo: string(payloadBytes),
		},
	)
	require.NoError(t, err, "failed to send IBC transfer")

	s.FlushRelayer(t, ctx, toOrbiterChannelID)

	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)
	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], ibcHeight)
	require.Equal(t, txsResult.TotalCount, uint64(1), "expected only one tx")

	found, events := SearchEvents(txsResult.Txs[0].Events, []string{
		"hyperlane.core.v1.EventDispatch",
		"hyperlane.warp.v1.EventSendRemoteTransfer",
		"noble.orbiter.component.adapter.v1.EventPayloadProcessed",
	})

	// NOTE: log missing events here
	missingEvents := make([]string, 0, len(events))
	if !found {
		for _, event := range events {
			missingEvents = append(missingEvents, event.Type)
		}
	}
	require.True(t, found, "some expected events are missing: "+strings.Join(missingEvents, ", "))
}
