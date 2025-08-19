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
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"orbiter.dev"
	"orbiter.dev/testutil"
	"orbiter.dev/types"
	actiontypes "orbiter.dev/types/controller/action"
	forwardingtypes "orbiter.dev/types/controller/forwarding"
	"orbiter.dev/types/core"
)

func TestIBCToCCTP(t *testing.T) {
	t.Parallel()

	testutil.SetSDKConfig()
	ctx, s := NewSuite(t, true, true)

	orbiter.RegisterInterfaces(s.Chain.GetCodec().InterfaceRegistry())

	fromOrbiterChanID, toOrbiterChanlID := s.GetChannels(t, ctx)

	srcUsdcTrace := transfertypes.ParseDenomTrace(
		transfertypes.GetPrefixedDenom("transfer", toOrbiterChanlID, Usdc),
	)
	dstUsdcDenom := srcUsdcTrace.IBCDenom()

	amountToSend := math.NewInt(2 * OneE6)

	// Fund the account on the counterparty chain such that it can
	// send USDC via IBC to the Orbiter.
	ibcRecipient := s.IBC.CounterpartySender.FormattedAddress()
	s.FundIBCRecipient(t, ctx, amountToSend, ibcRecipient, fromOrbiterChanID, dstUsdcDenom)

	t.Run("FailingWithoutForwarding", func(t *testing.T) {
		testIbcFailingWithoutForwarding(t, ctx, &s, dstUsdcDenom, toOrbiterChanlID)
	})

	t.Run("FailingUnsupportedAction", func(t *testing.T) {
		testIbcFailingUnsupportedAction(t, ctx, &s, dstUsdcDenom, toOrbiterChanlID)
	})

	t.Run("PassingWithFeeAction", func(t *testing.T) {
		testIbcPassingWithFeeAction(t, ctx, &s, dstUsdcDenom, toOrbiterChanlID)
	})

	t.Run("PassingWithoutActions", func(t *testing.T) {
		testIbcPassingWithoutActions(t, ctx, &s, dstUsdcDenom, toOrbiterChanlID)
	})
}

func testIbcFailingWithoutForwarding(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	dstUsdcDenom string,
	toOrbiterChannelID string,
) {
	cdc := s.Chain.GetCodec()
	amountToSend := math.NewInt(OneE6)

	// Create a wrapped payload for the IBC memo without the required forwarding info.
	feeRecipientAddr := testutil.NewNobleAddress()
	action, err := actiontypes.NewFeeAction(
		&actiontypes.FeeInfo{
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
		Address: core.ModuleAddress.String(),
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
		amountToSend,
		resp,
		"expected the address on the counterparty chain to have funds unlocked",
	)
}

func testIbcFailingUnsupportedAction(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	dstUsdcDenom string,
	toOrbiterChannelID string,
) {
	cdc := s.Chain.GetCodec()
	amountToSend := math.NewInt(OneE6)

	forwarding, err := forwardingtypes.NewCCTPForwarding(
		uint32(0),
		testutil.RandomBytes(32),
		testutil.RandomBytes(32),
		[]byte(""),
	)
	require.NoError(t, err)

	// Create a wrapped payload for the IBC memo without the required forwarding info.
	feeRecipientAddr := testutil.NewNobleAddress()
	action, err := actiontypes.NewFeeAction(
		&actiontypes.FeeInfo{
			Recipient:   feeRecipientAddr,
			BasisPoints: 100,
		},
	)
	action.Id = core.ACTION_SWAP
	require.NoError(t, err)

	p := core.PayloadWrapper{Orbiter: &core.Payload{
		PreActions: []*core.Action{action},
		Forwarding: forwarding,
	}}
	payloadBz, err := types.MarshalJSON(cdc, &p)
	require.NoError(t, err)

	height, err := s.IBC.CounterpartyChain.Height(ctx)
	require.NoError(t, err)

	transfer := ibc.WalletAmount{
		Address: core.ModuleAddress.String(),
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
		math.NewInt(OneE6),
		resp,
		"expected the address on the counterparty chain to have funds unlocked",
	)
}

func testIbcPassingWithFeeAction(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	dstUsdcDenom string,
	toOrbiterChannelID string,
) {
	cdc := s.Chain.GetCodec()
	amountToSend := math.NewInt(OneE6)

	forwarding, err := forwardingtypes.NewCCTPForwarding(
		uint32(0),
		testutil.RandomBytes(32),
		testutil.RandomBytes(32),
		[]byte(""),
	)
	require.NoError(t, err)

	// Create a wrapped payload for the IBC memo without the required forwarding info.
	feeRecipientAddr := testutil.NewNobleAddress()
	action, err := actiontypes.NewFeeAction(
		&actiontypes.FeeInfo{
			Recipient:   feeRecipientAddr,
			BasisPoints: 100,
		},
	)
	require.NoError(t, err)

	p := core.PayloadWrapper{Orbiter: &core.Payload{
		PreActions: []*core.Action{action},
		Forwarding: forwarding,
	}}
	payloadBz, err := types.MarshalJSON(cdc, &p)
	require.NoError(t, err)

	height, err := s.Chain.Height(ctx)
	require.NoError(t, err)

	transfer := ibc.WalletAmount{
		Address: core.ModuleAddress.String(),
		Denom:   dstUsdcDenom,
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
	s.FlushRelayer(t, ctx, toOrbiterChannelID)

	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)
	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], strconv.Itoa(int(ibcHeight)))
	require.Equal(t, txsResult.TotalCount, uint64(1), "expected only one tx")

	found, _ := SearchEvents(txsResult.Txs[0].Events, []string{
		"circle.cctp.v1.DepositForBurn",
	})
	require.True(t, found)

	feeAmt, err := s.Chain.BankQueryBalance(ctx, feeRecipientAddr, "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(10_000).String(), feeAmt.String())
}

func testIbcPassingWithoutActions(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	dstUsdcDenom string,
	toOrbiterChannelID string,
) {
	cdc := s.Chain.GetCodec()
	amountToSend := math.NewInt(OneE6)
	dustAmount := math.NewInt(1)

	// We fund the orbiter module to test the initial balance transfer to the dust collector.
	s.FundRecipient(t, ctx, dustAmount, core.ModuleAddress.String())

	forwarding, err := forwardingtypes.NewCCTPForwarding(
		s.destinationDomain,
		s.mintRecipient,
		s.destinationCaller,
		[]byte(""),
	)
	require.NoError(t, err)

	p := core.PayloadWrapper{Orbiter: &core.Payload{
		PreActions: []*core.Action{},
		Forwarding: forwarding,
	}}
	payloadBz, err := types.MarshalJSON(cdc, &p)
	require.NoError(t, err)

	height, err := s.Chain.Height(ctx)
	require.NoError(t, err)

	transfer := ibc.WalletAmount{
		Address: core.ModuleAddress.String(),
		Denom:   dstUsdcDenom,
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
	s.FlushRelayer(t, ctx, toOrbiterChannelID)

	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)
	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], strconv.Itoa(int(ibcHeight)))
	require.Equal(t, txsResult.TotalCount, uint64(1), "expected only one tx")

	found, _ := SearchEvents(txsResult.Txs[0].Events, []string{
		"circle.cctp.v1.DepositForBurn",
	})
	require.True(t, found)

	resp, err := s.Chain.GetBalance(
		ctx,
		core.DustCollectorAddress.String(),
		Usdc,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		dustAmount,
		resp,
		"expected the dust collector to have received orbiter module initial balance",
	)
}
