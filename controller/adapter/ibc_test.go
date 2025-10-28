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

package adapter_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	adapterctrl "github.com/noble-assets/orbiter/controller/adapter"
	"github.com/noble-assets/orbiter/testutil"
	"github.com/noble-assets/orbiter/testutil/testdata"
	"github.com/noble-assets/orbiter/types"
	adaptertypes "github.com/noble-assets/orbiter/types/component/adapter"
	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

func TestNewIBCParser(t *testing.T) {
	parser, err := adapterctrl.NewIBCParser(nil)
	require.ErrorContains(t, err, "cannot be nil")
	require.Nil(t, parser)
}

func TestParsePayload(t *testing.T) {
	testCases := []struct {
		name          string
		setup         func(reg codectypes.InterfaceRegistry)
		payloadBz     []byte
		expectPayload *core.Payload
		expErr        string
	}{
		{
			name:      "error - when memo is not a valid json",
			payloadBz: []byte("not json memo"),
			expErr:    "not a valid json",
		},
		{
			name: "error - orbiter payload with nil forwarding attributes",
			payloadBz: []byte(
				`{"orbiter": {"forwarding": {"protocol_id": 1, "attributes": null}}}`,
			),
			expErr: "not set",
		},
		{
			name: "success - valid orbiter payload",
			setup: func(reg codectypes.InterfaceRegistry) {
				// Payload Any types must be registered in the interface registry
				// to be valid.
				reg.RegisterImplementations(
					(*core.ForwardingAttributes)(nil),
					&testdata.TestForwardingAttr{},
				)
			},
			payloadBz: []byte(testutil.CreateValidOrbiterPayload()),
			expectPayload: &core.Payload{
				Forwarding: &core.Forwarding{
					ProtocolId: core.PROTOCOL_CCTP,
					Attributes: &codectypes.Any{
						TypeUrl: "/testpb.TestForwardingAttr",
					},
				},
			},
		},
		{
			name: "success - valid orbiter payload with actions",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*core.ForwardingAttributes)(nil),
					&testdata.TestForwardingAttr{},
				)
				reg.RegisterImplementations(
					(*core.ActionAttributes)(nil),
					&testdata.TestActionAttr{},
				)
			},
			payloadBz: []byte(testutil.CreateValidOrbiterPayloadWithActions()),
			expectPayload: &core.Payload{
				Forwarding: &core.Forwarding{
					ProtocolId: core.PROTOCOL_CCTP,
					Attributes: &codectypes.Any{TypeUrl: "/testpb.TestForwardingAttr"},
				},
				PreActions: []*core.Action{
					{
						Id:         core.ACTION_FEE,
						Attributes: &codectypes.Any{TypeUrl: "/testpb.TestActionAttr"},
					},
				},
			},
		},
		// NOTE: the following test case comes from an external audit report.
		{
			name: "success - orbiter payload with incomplete CCTP attributes",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*core.ForwardingAttributes)(nil),
					&forwardingtypes.CCTPAttributes{},
				)
			},
			payloadBz: func() []byte {
				memo := map[string]any{
					"orbiter": map[string]any{
						"forwarding": map[string]any{
							"protocol_id": 2,
							"attributes": map[string]any{
								"@type":          "/noble.orbiter.controller.forwarding.v1.CCTPAttributes",
								"mint_recipient": "PNWAxASH2RPmgMV+/Tb4e78ON1WL8SoFGnwbWWHxfuA=",
							},
						},
					},
				}

				memoBz, err := json.MarshalIndent(memo, "", "  ")
				require.NoError(t, err)

				return memoBz
			}(),
			expectPayload: &core.Payload{
				Forwarding: &core.Forwarding{
					ProtocolId: core.PROTOCOL_CCTP,
					Attributes: &codectypes.Any{
						TypeUrl: "/noble.orbiter.controller.forwarding.v1.CCTPAttributes",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encCfg := testutil.MakeTestEncodingConfig("noble")

			if tc.setup != nil {
				tc.setup(encCfg.InterfaceRegistry)
			}

			parser, err := adapterctrl.NewIBCParser(encCfg.Codec)
			require.NoError(t, err, "expected no error creating parser")
			payload, err := parser.ParsePayload(tc.payloadBz)

			if tc.expErr != "" {
				require.ErrorContains(t, err, tc.expErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, payload.Forwarding)
				require.Equal(t, tc.expectPayload.Forwarding.ProtocolId, payload.Forwarding.ProtocolId, "expected different id")
				require.Equal(t, tc.expectPayload.Forwarding.Attributes.TypeUrl, payload.Forwarding.Attributes.TypeUrl, "expected different forwarding attributes type url")

				require.Len(t, payload.PreActions, len(tc.expectPayload.PreActions))
				if len(tc.expectPayload.PreActions) != 0 {
					require.Equal(t, tc.expectPayload.PreActions[0].Id, payload.PreActions[0].Id)
					require.Equal(t, tc.expectPayload.PreActions[0].Attributes.TypeUrl, payload.PreActions[0].Attributes.TypeUrl)
				}
			}
		})
	}
}

func TestParsePacket(t *testing.T) {
	sender := testutil.NewNobleAddress()

	testCases := []struct {
		name          string
		setup         func(reg codectypes.InterfaceRegistry)
		ccPacket      func() adaptertypes.CrossChainPacket
		expParsedData *types.ParsedData
		expErr        string
	}{
		{
			name: "skip - not ics20 packet",
			ccPacket: func() adaptertypes.CrossChainPacket {
				return adaptertypes.NewIBCCrossChainPacket(
					"nontransfer",
					"channel-1",
					[]byte(`{"some": "other packet type"}`),
				)
			},
			expErr: "not for orbiter",
		},
		{
			name: "skip - receiver is not orbiter module",
			ccPacket: func() adaptertypes.CrossChainPacket {
				data := testutil.CreateValidIBCPacketData(
					sender,
					testutil.NewNobleAddress(),
					testutil.CreateValidOrbiterPayload(),
				)

				return adaptertypes.NewIBCCrossChainPacket(
					"nontransfer",
					"channel-1",
					data,
				)
			},
			expErr: "not for orbiter",
		},
		{
			name: "error - when memo is not a valid json",
			ccPacket: func() adaptertypes.CrossChainPacket {
				data := testutil.CreateValidIBCPacketData(
					sender,
					core.ModuleAddress.String(),
					"not json memo",
				)

				return adaptertypes.NewIBCCrossChainPacket(
					"nontransfer",
					"channel-1",
					data,
				)
			},
			expErr: "not a valid json",
		},
		{
			name: "error - denom is not native (multi hop)",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*core.ForwardingAttributes)(nil),
					&testdata.TestForwardingAttr{},
				)
				reg.RegisterImplementations(
					(*core.ActionAttributes)(nil),
					&testdata.TestActionAttr{},
				)
			},
			ccPacket: func() adaptertypes.CrossChainPacket {
				data := transfertypes.NewFungibleTokenPacketData(
					"transfer/channel-2/uosmo",
					"1000000",
					sender,
					core.ModuleAddress.String(),
					testutil.CreateValidOrbiterPayloadWithActions(),
				)

				return adaptertypes.NewIBCCrossChainPacket(
					"transfer",
					"channel-1",
					data.GetBytes(),
				)
			},
			expErr: "coin is native of source",
		},
		{
			name: "error - denom is not native (native on source)",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*core.ForwardingAttributes)(nil),
					&testdata.TestForwardingAttr{},
				)
				reg.RegisterImplementations(
					(*core.ActionAttributes)(nil),
					&testdata.TestActionAttr{},
				)
			},
			ccPacket: func() adaptertypes.CrossChainPacket {
				data := transfertypes.NewFungibleTokenPacketData(
					"uosmo",
					"1000000",
					sender,
					core.ModuleAddress.String(),
					testutil.CreateValidOrbiterPayloadWithActions(),
				)

				return adaptertypes.NewIBCCrossChainPacket(
					"transfer",
					"channel-1",
					data.GetBytes(),
				)
			},
			expErr: "coin is native of source",
		},
		{
			name: "success - valid orbiter payload with actions",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*core.ForwardingAttributes)(nil),
					&testdata.TestForwardingAttr{},
				)
				reg.RegisterImplementations(
					(*core.ActionAttributes)(nil),
					&testdata.TestActionAttr{},
				)
			},
			ccPacket: func() adaptertypes.CrossChainPacket {
				data := transfertypes.NewFungibleTokenPacketData(
					"transfer/channel-1/uusdc",
					"1000000",
					sender,
					core.ModuleAddress.String(),
					testutil.CreateValidOrbiterPayloadWithActions(),
				)

				return adaptertypes.NewIBCCrossChainPacket(
					"transfer",
					"channel-1",
					data.GetBytes(),
				)
			},
			expParsedData: &types.ParsedData{
				Coin: sdk.Coin{
					Denom:  "uusdc",
					Amount: sdkmath.NewIntFromUint64(1_000_000),
				},
				Payload: core.Payload{
					Forwarding: &core.Forwarding{
						ProtocolId: core.PROTOCOL_CCTP,
						Attributes: &codectypes.Any{TypeUrl: "/testpb.TestForwardingAttr"},
					},
					PreActions: []*core.Action{
						{
							Id:         core.ACTION_FEE,
							Attributes: &codectypes.Any{TypeUrl: "/testpb.TestActionAttr"},
						},
					},
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			encCfg := testutil.MakeTestEncodingConfig("noble")

			if tC.setup != nil {
				tC.setup(encCfg.InterfaceRegistry)
			}

			adapter, err := adapterctrl.NewIBCAdapter(encCfg.Codec, log.NewNopLogger())
			require.NoError(t, err)

			parsedData, err := adapter.ParsePacket(tC.ccPacket())

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
				require.Nil(t, parsedData)
			} else {
				require.NoError(t, err)

				if tC.expParsedData == nil {
					require.Nil(t, parsedData)
				} else {
					require.NotNil(t, parsedData)
					require.Equal(t, tC.expParsedData.Coin, parsedData.Coin)

					expPayload := tC.expParsedData.Payload
					payload := parsedData.Payload

					require.NotNil(t, payload.Forwarding)
					require.Equal(t, expPayload.Forwarding.ProtocolId, payload.Forwarding.ProtocolId, "expected different id")
					require.Equal(t, expPayload.Forwarding.Attributes.TypeUrl, payload.Forwarding.Attributes.TypeUrl, "expected different forwarding attributes type url")

					require.Len(t, payload.PreActions, len(expPayload.PreActions))
					if len(expPayload.PreActions) != 0 {
						require.Equal(t, expPayload.PreActions[0].Id, payload.PreActions[0].Id)
						require.Equal(t, expPayload.PreActions[0].Attributes.TypeUrl, payload.PreActions[0].Attributes.TypeUrl)
					}

					expCoin := tC.expParsedData.Coin
					coin := parsedData.Coin
					require.Equal(t, expCoin.String(), coin.String())
				}
			}
		})
	}
}
