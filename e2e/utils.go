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
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/jsonpb"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

const (
	OneCoin         = 1_000_000
	MaxSearchBlocks = 30
)

// GetChannels returns the channel IDs of the IBC connection.
// The first ID returned is from the orbiter chain to the counterparty,
// the second one is the ID from the counterparty to the orbiter chain.
func (s Suite) GetChannels(t *testing.T, ctx context.Context) (string, string) {
	orbiterToCounterpartyChannelInfo, err := s.IBC.Relayer.GetChannels(
		ctx,
		s.IBC.RelayerReporter,
		s.Chain.Config().ChainID,
	)
	require.NoError(t, err)
	orbiterToCounterpartyChannelID := orbiterToCounterpartyChannelInfo[0].ChannelID

	counterpartyToOrbiterChannelInfo, err := s.IBC.Relayer.GetChannels(
		ctx,
		s.IBC.RelayerReporter,
		s.IBC.CounterpartyChain.Config().ChainID,
	)
	require.NoError(t, err)
	counterpartyToOrbiterChannelID := counterpartyToOrbiterChannelInfo[0].ChannelID

	return orbiterToCounterpartyChannelID, counterpartyToOrbiterChannelID
}

// GetIbcTransferBlockExecution finds the first block at or after the given height
// that contains an IBC transfer (MsgRecvPacket) and returns that block height.
func (s *Suite) GetIbcTransferBlockExecution(
	t *testing.T,
	ctx context.Context,
	startHeight int64,
) int64 {
	reg := s.Chain.Config().EncodingConfig.InterfaceRegistry

	maxHeight := startHeight + MaxSearchBlocks

	for height := startHeight; height <= maxHeight; height++ {

		_, err := cosmos.PollForMessage[*channeltypes.MsgRecvPacket](
			ctx,
			s.Chain,
			reg,
			height,
			height,
			nil,
		)

		if err == nil {
			return height
		}
	}
	require.True(t, false, "expected MsgRecvPacket to be found")

	return 0
}

func GetTxsResult(
	t *testing.T,
	ctx context.Context,
	validator *cosmos.ChainNode,
	height string,
) *sdk.SearchTxsResult {
	t.Helper()

	raw, _, err := validator.ExecQuery(ctx, "txs", "--query", fmt.Sprintf("tx.height = %s", height))
	require.NoError(t, err, "expected no error querying block results")

	var res sdk.SearchTxsResult
	require.NoError(
		t,
		jsonpb.Unmarshal(bytes.NewReader(raw), &res),
		"expected no error parsing txs search",
	)

	return &res
}

// SearchEvents returns true if the slice of ABCI events contains all the event
// types provided. Returns false otherwise.
func SearchEvents(events []abci.Event, eventTypes []string) bool {
	if len(eventTypes) == 0 {
		return true
	}

	needed := make(map[string]bool)
	for _, t := range eventTypes {
		needed[t] = true
	}

	for _, event := range events {
		if needed[event.Type] {
			delete(needed, event.Type)
			if len(needed) == 0 {
				return true
			}
		}
	}

	return false
}
