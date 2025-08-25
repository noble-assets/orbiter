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
	"encoding/base64"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	interchainutil "github.com/strangelove-ventures/interchaintest/v8/testutil"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/noble-assets/orbiter"
	"github.com/noble-assets/orbiter/testutil"
	"github.com/noble-assets/orbiter/types"
	actiontypes "github.com/noble-assets/orbiter/types/controller/action"
	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

func TestIBCToCCTP(t *testing.T) {
	t.Parallel()

	testutil.SetSDKConfig()
	ctx, s := NewSuite(t, true, true)

	orbiter.RegisterInterfaces(s.Chain.GetCodec().InterfaceRegistry())

	fromOrbiterChanID, toOrbiterChanID := s.GetChannels(t, ctx)

	srcUsdcTrace := transfertypes.ParseDenomTrace(
		transfertypes.GetPrefixedDenom("transfer", toOrbiterChanID, Usdc),
	)
	dstUsdcDenom := srcUsdcTrace.IBCDenom()

	amountToSend := math.NewInt(2 * OneE6)

	// Fund the account on the counterparty chain such that it can
	// send USDC via IBC to the Orbiter.
	ibcRecipient := s.IBC.CounterpartySender.FormattedAddress()
	s.FundIBCRecipient(t, ctx, amountToSend, ibcRecipient, fromOrbiterChanID, dstUsdcDenom)

	t.Run("FailingWithoutForwarding", func(t *testing.T) {
		testIbcFailingWithoutForwarding(t, ctx, &s, dstUsdcDenom, toOrbiterChanID)
	})

	t.Run("PassingWithFeeAction", func(t *testing.T) {
		testIbcPassingWithFeeAction(t, ctx, &s, dstUsdcDenom, toOrbiterChanID)
	})

	t.Run("PassingWithoutActions", func(t *testing.T) {
		testIbcPassingWithoutActions(t, ctx, &s, dstUsdcDenom, toOrbiterChanID)
	})
}

func testIbcFailingWithoutForwarding(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	dstUsdcDenom string,
	toOrbiterChanID string,
) {
	cdc := s.Chain.GetCodec()
	amountToSend := math.NewInt(OneE6)
	initAmount, err := s.IBC.CounterpartyChain.GetBalance(
		ctx,
		s.IBC.CounterpartySender.FormattedAddress(),
		dstUsdcDenom,
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
		toOrbiterChanID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, toOrbiterChanID)

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
		initAmount,
		resp,
		"expected the address on the counterparty chain to have funds unlocked",
	)
}

func testIbcPassingWithFeeAction(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	dstUsdcDenom string,
	toOrbiterChanID string,
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
		toOrbiterChanID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, toOrbiterChanID)

	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)
	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], strconv.Itoa(int(ibcHeight)))
	require.Equal(t, txsResult.TotalCount, uint64(1), "expected only one tx")

	found, _ := SearchEvents(txsResult.Txs[0].Events, []string{
		"circle.cctp.v1.DepositForBurn",
		"noble.orbiter.controller.action.v1.EventFeeAction",
	})
	require.True(t, found, "expected events not found")

	feeAmt, err := s.Chain.BankQueryBalance(ctx, feeRecipientAddr, "uusdc")
	require.NoError(t, err)
	require.Equal(t, math.NewInt(10_000).String(), feeAmt.String())
}

func testIbcPassingWithoutActions(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	dstUsdcDenom string,
	toOrbiterChanID string,
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
		toOrbiterChanID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, toOrbiterChanID)

	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)
	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], strconv.Itoa(int(ibcHeight)))
	require.Equal(t, txsResult.TotalCount, uint64(1), "expected only one tx")

	found, events := SearchEvents(txsResult.Txs[0].Events, []string{DepositForBurnEvent})
	require.True(t, found, "expected the DepositForBurn event to be emitted")
	require.Len(t, events, 1)

	for _, attribute := range events[DepositForBurnEvent].Attributes {
		switch attribute.Key {
		case "amount":
			var v string
			require.NoError(t, json.Unmarshal([]byte(attribute.Value), &v))
			require.Equal(
				t,
				strconv.Itoa(int(amountToSend.Int64())),
				v,
				"expected a different amount in the DepositForBurn event",
			)
		case "destination_domain":
			v, _ := strconv.ParseUint(attribute.Value, 10, 32)
			require.Equal(
				t,
				s.destinationDomain,
				uint32(v),
				"expected a different destination domain in the DepositForBurn event",
			)
		case "destination_caller":
			expectedBase64 := base64.StdEncoding.EncodeToString(s.destinationCaller)
			var v string
			require.NoError(t, json.Unmarshal([]byte(attribute.Value), &v))
			require.Equal(
				t,
				expectedBase64,
				v,
				"expected a different destination caller in the DepositForBurn event",
			)
		case "mint_recipient":
			expectedBase64 := base64.StdEncoding.EncodeToString(s.mintRecipient)
			var v string
			require.NoError(t, json.Unmarshal([]byte(attribute.Value), &v))
			require.Equal(
				t,
				expectedBase64,
				v,
				"expected a different mint recipient in the DestinationForBurn event",
			)
		case "depositor":
			var v string
			json.Unmarshal([]byte(attribute.Value), &v)
			require.Equal(
				t,
				core.ModuleAddress.String(),
				v,
				"expected a different depositor in the DestinationForBurn event",
			)
		}
	}

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
