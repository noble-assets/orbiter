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
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	adapterctrl "orbiter.dev/controller/adapter"
	"orbiter.dev/testutil"
	"orbiter.dev/testutil/mocks"
	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types/core"
)

func TestHooks(t *testing.T) {
	deps := mocks.NewDependencies(t)
	adapter, err := adapterctrl.NewIBCAdapter(deps.EncCfg.Codec, deps.Logger)
	require.NoError(t, err)

	err = adapter.AfterTransferHook(context.Background(), &core.Payload{})
	require.NoError(t, err)

	err = adapter.BeforeTransferHook(context.Background(), &core.Payload{})
	require.NoError(t, err)
}

func TestNewIBCParser(t *testing.T) {
	parser, err := adapterctrl.NewIBCParser(nil)
	require.ErrorContains(t, err, "cannot be nil")
	require.Nil(t, parser)
}

func TestParsePayload(t *testing.T) {
	sender := testutil.NewNobleAddress()

	testCases := []struct {
		name            string
		setup           func(reg codectypes.InterfaceRegistry)
		payloadBz       []byte
		expectIsOrbiter bool
		expectPayload   *core.Payload
		expectError     bool
		errorContains   string
	}{
		{
			name:            "skip - not ics20 packet",
			payloadBz:       []byte(`{"some": "other packet type"}`),
			expectIsOrbiter: false,
			expectError:     false,
			errorContains:   "",
		},
		{
			name: "skip - receiver is not orbiter module",
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				testutil.NewNobleAddress(),
				testutil.CreateValidOrbiterPayload(),
			),
			expectIsOrbiter: false,
			expectError:     false,
		},
		{
			name: "error - when memo is not a valid json",
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				core.ModuleAddress.String(),
				"not json memo",
			),
			expectIsOrbiter: true,
			expectError:     true,
			errorContains:   "not a valid json",
		},
		{
			name: "error - orbiter payload with nil forwarding attributes",
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				core.ModuleAddress.String(),
				`{"orbiter": {"forwarding": {"protocol_id": 1, "attributes": null}}}`,
			),
			expectIsOrbiter: true,
			expectError:     true,
			errorContains:   "not set",
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
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				core.ModuleAddress.String(),
				testutil.CreateValidOrbiterPayload(),
			),
			expectIsOrbiter: true,
			expectPayload: &core.Payload{
				Forwarding: &core.Forwarding{
					ProtocolId: core.PROTOCOL_CCTP,
					Attributes: &codectypes.Any{
						TypeUrl: "/testpb.TestForwardingAttr",
					},
				},
			},
			expectError: false,
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
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				core.ModuleAddress.String(),
				testutil.CreateValidOrbiterPayloadWithActions(),
			),
			expectIsOrbiter: true,
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
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encCfg := testutil.MakeTestEncodingConfig("noble")

			if tc.setup != nil {
				tc.setup(encCfg.InterfaceRegistry)
			}

			parser, err := adapterctrl.NewIBCParser(encCfg.Codec)
			require.NoError(t, err)

			isOrbiterPayload, payload, err := parser.ParsePayload(tc.payloadBz)

			require.Equal(t, tc.expectIsOrbiter, isOrbiterPayload)
			if tc.expectError {
				require.ErrorContains(t, err, tc.errorContains)
			} else {
				require.NoError(t, err)
				if tc.expectIsOrbiter {
					require.NotNil(t, payload.Forwarding)
					require.Equal(t, tc.expectPayload.Forwarding.ProtocolId, payload.Forwarding.ProtocolId, "expected different id")
					require.Equal(t, tc.expectPayload.Forwarding.Attributes.TypeUrl, payload.Forwarding.Attributes.TypeUrl, "expected different forwarding attributes type url")

					if tc.expectPayload.PreActions != nil {
						require.Len(t, payload.PreActions, len(tc.expectPayload.PreActions))
						if len(payload.PreActions) > 0 {
							require.Equal(t, tc.expectPayload.PreActions[0].Id, payload.PreActions[0].Id)
							require.Equal(t, tc.expectPayload.PreActions[0].Attributes.TypeUrl, payload.PreActions[0].Attributes.TypeUrl)
						}
					}
				} else {
					require.Nil(t, payload)
				}
			}
		})
	}
}
