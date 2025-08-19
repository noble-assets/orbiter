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
	"context"
	"strconv"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	interchainutil "github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"orbiter.dev"
	"orbiter.dev/testutil"
	"orbiter.dev/types"
	actiontypes "orbiter.dev/types/controller/action"
	forwardingtypes "orbiter.dev/types/controller/forwarding"
	"orbiter.dev/types/core"
)

func TestIbcFailing(t *testing.T) {
	t.Parallel()

	testutil.SetSDKConfig()
	ctx, s := NewSuite(t, true, true)

	fromOrbiterChannelID, toOrbiterChannelID := s.GetChannels(t, ctx)

	amountToSend := math.NewInt(OneE6)
	transfer := ibc.WalletAmount{
		Address: s.IBC.CounterpartySender.FormattedAddress(),
		Denom:   Usdc,
		Amount:  amountToSend,
	}
	_, err := s.Chain.SendIBCTransfer(
		ctx,
		fromOrbiterChannelID,
		s.sender.KeyName(),
		transfer,
		ibc.TransferOptions{Memo: "pls send them back"},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, fromOrbiterChannelID)

	srcUsdcTrace := transfertypes.ParseDenomTrace(
		transfertypes.GetPrefixedDenom("transfer", toOrbiterChannelID, Usdc),
	)
	dstUsdcDenom := srcUsdcTrace.IBCDenom()
	dstSenderBal, err := s.IBC.CounterpartyChain.GetBalance(
		ctx,
		s.IBC.CounterpartySender.FormattedAddress(),
		dstUsdcDenom,
	)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, dstSenderBal)

	// Register the interfaces to enable marhsal/unmarshal
	encCfg := testutil.MakeTestEncodingConfig("noble")
	orbiter.RegisterInterfaces(encCfg.InterfaceRegistry)

	t.Run("FailingWithoutForwarding", func(t *testing.T) {
		// Your current test logic here
		testIbcFailingWithoutForwarding(
			t,
			ctx,
			encCfg.Codec,
			&s,
			dstUsdcDenom,
			dstSenderBal,
			toOrbiterChannelID,
		)
	})
}

func testIbcFailingWithoutForwarding(
	t *testing.T,
	ctx context.Context,
	cdc codec.Codec,
	s *Suite,
	dstUsdcDenom string,
	dstSenderBal math.Int,
	toOrbiterChannelID string,
) {
	amountToSend := math.NewInt(OneE6)

	// Create a wrapped payload for the IBC memo without the required forwarding info.
	feeRecipientAddr := testutil.NewNobleAddress()
	action, err := actiontypes.NewFeeAction(&actiontypes.FeeInfo{
		Recipient:   feeRecipientAddr,
		BasisPoints: 100,
	},
	)
	require.NoError(t, err)

	p := core.PayloadWrapper{Orbiter: &core.Payload{
		PreActions: []*core.Action{action},
		Forwarding: nil,
	}}
	payloadBz, err := types.MarshalJSON(cdc, &p)
	require.NoError(t, err)

	height, err := s.IBC.CounterpartyChain.Height(ctx)
	require.NoError(t, err)

	transfer := ibc.WalletAmount{
		Address: OrbiterModuleAddr,
		Denom:   dstUsdcDenom,
		Amount:  amountToSend,
	}
	ibcTx, err := s.IBC.CounterpartyChain.SendIBCTransfer(
		ctx,
		toOrbiterChannelID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, toOrbiterChannelID)

	msg, err := interchainutil.PollForAck(
		ctx,
		s.IBC.CounterpartyChain,
		height,
		height+10,
		ibcTx.Packet,
	)
	require.NoError(t, err)

	expectedAck := &channeltypes.Acknowledgement{}
	err = cdc.UnmarshalJSON(msg.Acknowledgement, expectedAck)
	require.NoError(t, err)
	require.Contains(
		t,
		expectedAck.GetError(),
		"orbiter-middleware error",
		"expected the error in the ack to contains the orbiter middleware error",
	)

	resp, err := s.IBC.CounterpartyChain.GetBalance(
		ctx,
		s.IBC.CounterpartySender.FormattedAddress(),
		dstUsdcDenom,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		dstSenderBal,
		resp,
		"expected the address on the counterparty chain to have funds unlocked",
	)
}

func TestIbc(t *testing.T) {
	t.Parallel()

	ctx, s := NewSuite(t, true, true)

	fromOrbiterChannelID, toOrbiterChannelID := s.GetChannels(t, ctx)

	amountToSend := math.NewInt(OneE6)

	transfer := ibc.WalletAmount{
		Address: s.IBC.CounterpartySender.FormattedAddress(),
		Denom:   "uusdc",
		Amount:  amountToSend,
	}

	_, err := s.Chain.SendIBCTransfer(
		ctx,
		fromOrbiterChannelID,
		s.sender.KeyName(),
		transfer,
		ibc.TransferOptions{},
	)
	require.NoError(t, err)
	require.NoError(
		t,
		s.IBC.Relayer.Flush(
			ctx,
			s.IBC.RelayerReporter,
			s.IBC.PathName,
			fromOrbiterChannelID,
		),
		"expected no error relaying MsgRecvPacket & MsgAcknowledgement",
	)

	srcDenomTrace := transfertypes.ParseDenomTrace(
		transfertypes.GetPrefixedDenom("transfer", toOrbiterChannelID, "uusdc"),
	)
	dstIbcDenom := srcDenomTrace.IBCDenom()

	counterpartyWalletBal, err := s.IBC.CounterpartyChain.GetBalance(
		ctx,
		s.IBC.CounterpartySender.FormattedAddress(),
		dstIbcDenom,
	)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, counterpartyWalletBal)

	// Generate orbiter payload
	destinationDomain := uint32(0)
	mintRecipient := testutil.RandomBytes(32)
	destinationCaller := testutil.RandomBytes(32)
	passthroughPayload := []byte("")

	forwarding, err := forwardingtypes.NewCCTPForwarding(
		destinationDomain,
		mintRecipient,
		destinationCaller,
		passthroughPayload,
	)
	require.NoError(t, err)

	feeRecipientAddr := testutil.NewNobleAddress()
	feeAttr := actiontypes.FeeAttributes{
		FeesInfo: []*actiontypes.FeeInfo{
			{
				Recipient:   feeRecipientAddr,
				BasisPoints: 100,
			},
		},
	}

	action, err := core.NewAction(core.ACTION_FEE, &feeAttr)
	require.NoError(t, err)

	payload, err := core.NewPayloadWrapper(forwarding, action)
	require.NoError(t, err)

	encCfg := testutil.MakeTestEncodingConfig("noble")
	orbiter.RegisterInterfaces(encCfg.InterfaceRegistry)
	payloadBz, err := types.MarshalJSON(encCfg.Codec, payload)
	require.NoError(t, err)

	height, err := s.IBC.CounterpartyChain.Height(ctx)
	require.NoError(t, err)

	// Transfer funds to the orbiter module
	transfer = ibc.WalletAmount{
		Address: OrbiterModuleAddr,
		Denom:   dstIbcDenom,
		Amount:  amountToSend,
	}
	_, err = s.IBC.CounterpartyChain.SendIBCTransfer(
		ctx,
		toOrbiterChannelID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	require.NoError(
		t,
		s.IBC.Relayer.Flush(
			ctx,
			s.IBC.RelayerReporter,
			s.IBC.PathName,
			toOrbiterChannelID,
		),
		"expected no error relaying MsgRecvPacket & MsgAcknowledgement",
	)

	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)
	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], strconv.Itoa(int(ibcHeight)))
	require.Equal(t, txsResult.TotalCount, uint64(1), "expected only one tx")

	found := SearchEvents(txsResult.Txs[0].Events, []string{
		"circle.cctp.v1.DepositForBurn",
	})
	require.True(t, found)

	feeAmt, err := s.Chain.BankQueryBalance(ctx, feeRecipientAddr, "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(10_000).String(), feeAmt.String())
}
