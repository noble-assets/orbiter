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

package adapters_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"orbiter.dev/controllers/adapters"
	"orbiter.dev/testutil"
	"orbiter.dev/testutil/mocks"
	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
)

func TestHooks(t *testing.T) {
	deps := mocks.NewDependencies(t)
	adapter, err := adapters.NewIBCAdapter(deps.EncCfg.Codec, deps.Logger)
	require.NoError(t, err)

	err = adapter.AfterTransferHook(context.Background(), &types.Payload{})
	require.NoError(t, err)

	err = adapter.BeforeTransferHook(context.Background(), &types.Payload{})
	require.NoError(t, err)
}

func TestNewIBCParser(t *testing.T) {
	parser, err := adapters.NewIBCParser(nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "cannot be nil")
	require.Nil(t, parser)
}

func TestParsePayload(t *testing.T) {
	sender := testutil.NobleAddress()

	testCases := []struct {
		name            string
		setup           func(reg codectypes.InterfaceRegistry)
		payloadBz       []byte
		expectIsOrbiter bool
		expectPayload   *types.Payload
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
				testutil.NobleAddress(),
				testutil.CreateValidOrbiterPayload(),
			),
			expectIsOrbiter: false,
			expectError:     false,
		},
		{
			name: "error - when memo is not a valid json",
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				types.ModuleAddress.String(),
				"not json memo",
			),
			expectIsOrbiter: true,
			expectError:     true,
			errorContains:   "not a valid json",
		},
		{
			name: "fail - orbiter payload with nil orbit attributes",
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				types.ModuleAddress.String(),
				`{"orbiter": {"orbit": {"protocol_id": 1, "attributes": null}}}`,
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
					(*types.OrbitAttributes)(nil),
					&testdata.TestOrbitAttr{},
				)
			},
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				types.ModuleAddress.String(),
				testutil.CreateValidOrbiterPayload(),
			),
			expectIsOrbiter: true,
			expectPayload: &types.Payload{
				Orbit: &types.Orbit{
					ProtocolId: types.PROTOCOL_CCTP,
					Attributes: &codectypes.Any{
						TypeUrl: "/testpb.TestOrbitAttr",
					},
				},
			},
			expectError: false,
		},
		{
			name: "success - valid orbiter payload with actions",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*types.OrbitAttributes)(nil),
					&testdata.TestOrbitAttr{},
				)
				reg.RegisterImplementations(
					(*types.ActionAttributes)(nil),
					&testdata.TestActionAttr{},
				)
			},
			payloadBz: testutil.CreateValidIBCPacketData(
				sender,
				types.ModuleAddress.String(),
				testutil.CreateValidOrbiterPayloadWithActions(),
			),
			expectIsOrbiter: true,
			expectPayload: &types.Payload{
				Orbit: &types.Orbit{
					ProtocolId: types.PROTOCOL_CCTP,
					Attributes: &codectypes.Any{TypeUrl: "/testpb.TestOrbitAttr"},
				},
				PreActions: []*types.Action{
					{
						Id:         types.ACTION_FEE,
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

			parser, err := adapters.NewIBCParser(encCfg.Codec)
			require.NoError(t, err)

			isOrbiterPayload, payload, err := parser.ParsePayload(tc.payloadBz)

			require.Equal(t, tc.expectIsOrbiter, isOrbiterPayload)
			if tc.expectError {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errorContains)
			} else {
				require.NoError(t, err)
				if tc.expectIsOrbiter {
					require.NotNil(t, payload.Orbit, "expected orbit to be not nil")
					require.Equal(t, tc.expectPayload.Orbit.ProtocolId, payload.Orbit.ProtocolId, "expected different id")
					require.Equal(t, tc.expectPayload.Orbit.Attributes.TypeUrl, payload.Orbit.Attributes.TypeUrl, "expected different orbit attributes type url")

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
