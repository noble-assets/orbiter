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

package interchaintests

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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	"github.com/noble-assets/orbiter/v2"
	"github.com/noble-assets/orbiter/v2/testutil"
	"github.com/noble-assets/orbiter/v2/types"
	actiontypes "github.com/noble-assets/orbiter/v2/types/controller/action"
	forwardingtypes "github.com/noble-assets/orbiter/v2/types/controller/forwarding"
	"github.com/noble-assets/orbiter/v2/types/core"
)

type envIBC struct {
	CounterpartyUsdcDenom string
	ToOrbiterChanID       string
	AmountToSend          math.Int
}

func TestIBCToCCTP(t *testing.T) {
	testutil.SetSDKConfig()
	ctx, s := NewSuite(t, true, true, false)

	orbiter.RegisterInterfaces(s.Chain.GetCodec().InterfaceRegistry())

	fromOrbiterChanID, toOrbiterChanID := s.GetChannels(t, ctx)

	// Compute the denom string in the counterparty chain.
	dstUsdcTrace := transfertypes.ParseDenomTrace(
		transfertypes.GetPrefixedDenom(transfertypes.PortID, toOrbiterChanID, uusdcDenom),
	)

	env := envIBC{
		CounterpartyUsdcDenom: dstUsdcTrace.IBCDenom(),
		ToOrbiterChanID:       toOrbiterChanID,
		AmountToSend:          math.NewInt(OneE6),
	}

	// Fund the account on the counterparty chain such that it can
	// send USDC via IBC to the Orbiter.
	ibcRecipient := s.IBC.CounterpartySender.FormattedAddress()
	s.FundIBCRecipient(
		t,
		ctx,
		env.AmountToSend.MulRaw(2),
		ibcRecipient,
		fromOrbiterChanID,
		env.CounterpartyUsdcDenom,
	)

	totalEscrow := GetIBCTotalEscrow(t, ctx, s.Chain.Validators[0], uusdcDenom)
	require.Equal(
		t,
		totalEscrow.String(),
		env.AmountToSend.MulRaw(2).String(),
		"expected different usdc amount in escrow account",
	)

	testCases := []struct {
		name string
		flow func(*testing.T, context.Context, *Suite, envIBC)
	}{
		{"FailingParsingWithoutForwarding", testIbcFailingParsingWithoutForwarding},
		{"FailingAfterParsingInvalidForwarding", testIbcFailingAfterParsingInvalidForwarding},
		{"PassingWithoutActions", testIbcPassingWithoutActions},
		{"PassingWithFeeAction", testIbcPassingWithFeeAction},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) { tC.flow(t, ctx, &s, env) })
	}
}

func testIbcFailingParsingWithoutForwarding(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	env envIBC,
) {
	cdc := s.Chain.GetCodec()

	initAmount, err := s.IBC.CounterpartyChain.GetBalance(
		ctx,
		s.IBC.CounterpartySender.FormattedAddress(),
		env.CounterpartyUsdcDenom,
	)
	require.NoError(t, err)
	initialEscrow := GetIBCTotalEscrow(t, ctx, s.Chain.Validators[0], uusdcDenom)

	// Create a wrapped payload for the IBC memo without the required forwarding info.
	feeRecipientAddr := testutil.NewNobleAddress()

	bps, err := actiontypes.NewFeeBasisPoints(100)
	require.NoError(t, err)

	feeInfo, err := actiontypes.NewFeeInfo(feeRecipientAddr, bps)
	require.NoError(t, err)

	action, err := actiontypes.NewFeeAction(feeInfo)
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
		Denom:   env.CounterpartyUsdcDenom,
		Amount:  env.AmountToSend,
	}
	ibcTx, err := s.IBC.CounterpartyChain.SendIBCTransfer(
		ctx,
		env.ToOrbiterChanID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, env.ToOrbiterChanID)

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
		env.CounterpartyUsdcDenom,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		initAmount,
		resp,
		"expected the address on the counterparty chain to have funds unlocked",
	)
	finalEscrow := GetIBCTotalEscrow(t, ctx, s.Chain.Validators[0], uusdcDenom)
	require.Equal(t, initialEscrow.String(), finalEscrow.String())
}

func testIbcFailingAfterParsingInvalidForwarding(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	env envIBC,
) {
	cdc := s.Chain.GetCodec()

	initAmount, err := s.IBC.CounterpartyChain.GetBalance(
		ctx,
		s.IBC.CounterpartySender.FormattedAddress(),
		env.CounterpartyUsdcDenom,
	)
	require.NoError(t, err)
	initialEscrow := GetIBCTotalEscrow(t, ctx, s.Chain.Validators[0], uusdcDenom)

	height, err := s.IBC.CounterpartyChain.Height(ctx)
	require.NoError(t, err)

	// By using the empty byte array we can trigger an error in the CCTP handler.
	emptyByteArr := make([]byte, 32)
	payload := map[string]any{
		"orbiter": map[string]any{
			"forwarding": map[string]any{
				"protocol_id": 2,
				"attributes": map[string]any{
					"@type":          "/noble.orbiter.controller.forwarding.v1.CCTPAttributes",
					"mint_recipient": len(emptyByteArr),
				},
			},
		},
	}

	payloadBz, err := json.MarshalIndent(payload, "", "  ")
	require.NoError(t, err)

	transfer := ibc.WalletAmount{
		Address: core.ModuleAddress.String(),
		Denom:   env.CounterpartyUsdcDenom,
		Amount:  env.AmountToSend,
	}
	ibcTx, err := s.IBC.CounterpartyChain.SendIBCTransfer(
		ctx,
		env.ToOrbiterChanID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, env.ToOrbiterChanID)

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

	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)
	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], ibcHeight)
	require.Equal(t, txsResult.TotalCount, uint64(1), "expected only one tx")

	found, _ := SearchEvents(txsResult.Txs[0].Events, []string{
		"circle.cctp.v1.DepositForBurn",
	})
	require.False(t, found, "expected CCTP event not emitted")

	resp, err := s.IBC.CounterpartyChain.GetBalance(
		ctx,
		s.IBC.CounterpartySender.FormattedAddress(),
		env.CounterpartyUsdcDenom,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		initAmount,
		resp,
		"expected the address on the counterparty chain to have funds unlocked",
	)
	finalEscrow := GetIBCTotalEscrow(t, ctx, s.Chain.Validators[0], uusdcDenom)
	require.Equal(t, initialEscrow.String(), finalEscrow.String())
}

func testIbcPassingWithFeeAction(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	env envIBC,
) {
	cdc := s.Chain.GetCodec()
	forwarding, err := forwardingtypes.NewCCTPForwarding(
		uint32(0),
		testutil.RandomBytes(32),
		testutil.RandomBytes(32),
		[]byte(""),
	)
	require.NoError(t, err)

	feeRecipientAddr := testutil.NewNobleAddress()

	feeBps, err := actiontypes.NewFeeBasisPoints(100) // 1%
	require.NoError(t, err)

	feeAmount, err := actiontypes.NewFeeAmount("100")
	require.NoError(t, err)

	action, err := actiontypes.NewFeeAction(
		&actiontypes.FeeInfo{
			Recipient: feeRecipientAddr,
			FeeType:   feeBps,
		},
		&actiontypes.FeeInfo{
			Recipient: feeRecipientAddr,
			FeeType:   feeAmount,
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
		Denom:   env.CounterpartyUsdcDenom,
		Amount:  env.AmountToSend,
	}
	ibcTx, err := s.IBC.CounterpartyChain.SendIBCTransfer(
		ctx,
		env.ToOrbiterChanID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, env.ToOrbiterChanID)

	// Poll for the MsgRecvPacket on Noble.
	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)

	// Poll for the returned PacketAck on the counterparty chain.
	msg, err := interchainutil.PollForAck(
		ctx,
		s.IBC.CounterpartyChain,
		height,
		height+10,
		ibcTx.Packet,
	)
	require.NoError(t, err)

	expAck := &channeltypes.Acknowledgement{}
	require.NoError(t, cdc.UnmarshalJSON(msg.Acknowledgement, expAck))
	require.Equal(t, expAck.GetError(), "", "expected no error in the ack")

	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], ibcHeight)
	require.Equal(t, txsResult.TotalCount, uint64(1), "expected only one tx")

	found, _ := SearchEvents(txsResult.Txs[0].Events, []string{
		"circle.cctp.v1.DepositForBurn",
		"noble.orbiter.component.adapter.v1.EventPayloadProcessed",
		"noble.orbiter.controller.action.v2.EventFeeAction",
	})
	require.True(t, found, "expected events not found")

	feeAmt, err := s.Chain.BankQueryBalance(ctx, feeRecipientAddr, "uusdc")
	require.NoError(t, err)
	expFee := math.NewIntFromUint64(100). // fee from fixed fee action
						Add(math.NewInt(10_000)) // fee from basis points fee action
	require.Equal(t, expFee.String(), feeAmt.String())
}

func testIbcPassingWithoutActions(
	t *testing.T,
	ctx context.Context,
	s *Suite,
	env envIBC,
) {
	cdc := s.Chain.GetCodec()
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
		Denom:   env.CounterpartyUsdcDenom,
		Amount:  env.AmountToSend,
	}
	_, err = s.IBC.CounterpartyChain.SendIBCTransfer(
		ctx,
		env.ToOrbiterChanID,
		s.IBC.CounterpartySender.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadBz),
		},
	)
	require.NoError(t, err)
	s.FlushRelayer(t, ctx, env.ToOrbiterChanID)

	ibcHeight := s.GetIbcTransferBlockExecution(t, ctx, height)
	txsResult := GetTxsResult(t, ctx, s.Chain.Validators[0], ibcHeight)
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
				strconv.Itoa(int(env.AmountToSend.Int64())),
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
			require.NoError(t, json.Unmarshal([]byte(attribute.Value), &v))
			require.Equal(
				t,
				core.ModuleAddress.String(),
				v,
				"expected a different depositor in the DestinationForBurn event",
			)
		}
	}

	dcAddr := authtypes.NewModuleAddress(core.DustCollectorName)

	resp, err := s.Chain.GetBalance(
		ctx,
		dcAddr.String(),
		uusdcDenom,
	)
	require.NoError(t, err)
	require.Equal(
		t,
		dustAmount,
		resp,
		"expected the dust collector to have received orbiter module initial balance",
	)
}
