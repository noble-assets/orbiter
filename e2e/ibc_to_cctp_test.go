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
	"strconv"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"orbiter.dev"
	"orbiter.dev/testutil"
	"orbiter.dev/types"
	"orbiter.dev/types/controllers/actions"
	"orbiter.dev/types/controllers/orbits"
)

const OrbiterModuleAddr = "noble15xt7kx5mles58vkkfxvf0lq78sw04jajvfgd4d"

func TestIbc(t *testing.T) {
	t.Parallel()

	// ARRANGE
	ctx, s := NewSuite(t, true, true)

	orbiterToCounterpartyChannelID, counterpartyToOrbiterChannelID := s.GetChannels(t, ctx)

	amountToSend := math.NewInt(OneE6)

	transfer := ibc.WalletAmount{
		Address: s.IBC.Account.FormattedAddress(),
		Denom:   "uusdc",
		Amount:  amountToSend,
	}

	_, err := s.Chain.SendIBCTransfer(
		ctx,
		orbiterToCounterpartyChannelID,
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
			orbiterToCounterpartyChannelID,
		),
		"expected no error relaying MsgRecvPacket & MsgAcknowledgement",
	)

	srcDenomTrace := transfertypes.ParseDenomTrace(
		transfertypes.GetPrefixedDenom("transfer", counterpartyToOrbiterChannelID, "uusdc"),
	)
	dstIbcDenom := srcDenomTrace.IBCDenom()

	counterpartyWalletBal, err := s.IBC.CounterpartyChain.GetBalance(
		ctx,
		s.IBC.Account.FormattedAddress(),
		dstIbcDenom,
	)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, counterpartyWalletBal)

	// Generate orbiter payload
	destinationDomain := uint32(0)
	mintRecipient := testutil.RandomBytes(32)
	destinationCaller := testutil.RandomBytes(32)
	passthroughPayload := []byte("")

	orbit, err := orbits.NewCCTPOrbit(
		destinationDomain,
		mintRecipient,
		destinationCaller,
		passthroughPayload,
	)
	require.NoError(t, err)

	feeRecipientAddr := testutil.NewNobleAddress()
	feeAttr := actions.FeeAttributes{
		FeesInfo: []*actions.FeeInfo{
			{
				Recipient:   feeRecipientAddr,
				BasisPoints: 100,
			},
		},
	}

	action := types.Action{
		Id: types.ACTION_FEE,
	}
	err = action.SetAttributes(&feeAttr)
	require.NoError(t, err)

	payload, err := types.NewPayloadWrapper(orbit, []*types.Action{&action})
	require.NoError(t, err)

	encCfg := testutil.MakeTestEncodingConfig("noble")
	orbiter.RegisterInterfaces(encCfg.InterfaceRegistry)
	payloadStr, err := types.MarshalJSON(encCfg.Codec, payload)
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
		counterpartyToOrbiterChannelID,
		s.IBC.Account.KeyName(),
		transfer,
		ibc.TransferOptions{
			Memo: string(payloadStr),
		},
	)
	require.NoError(t, err)
	require.NoError(
		t,
		s.IBC.Relayer.Flush(
			ctx,
			s.IBC.RelayerReporter,
			s.IBC.PathName,
			counterpartyToOrbiterChannelID,
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
