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

package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"orbiter.dev/testutil"
	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
)

func TestMarshalUnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name    string
		setup   func(codectypes.InterfaceRegistry)
		payload func() *types.Payload
		expErr  string
	}{
		{
			name: "success - payload with default values (resulting types are nil)",
			payload: func() *types.Payload {
				return &types.Payload{}
			},
			expErr: "",
		},
		{
			name: "success - payload with one orbit and no actions",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*types.OrbitAttributes)(nil),
					&testdata.TestOrbitAttr{},
				)
			},
			payload: func() *types.Payload {
				attr := testdata.TestOrbitAttr{
					Planet: "saturn",
				}
				orbit, err := types.NewOrbit(types.PROTOCOL_IBC, &attr, []byte{})
				require.NoError(t, err)
				return &types.Payload{
					Orbit: orbit,
				}
			},
			expErr: "",
		},
		{
			name:  "error - payload with action not registered",
			setup: func(reg codectypes.InterfaceRegistry) {},
			payload: func() *types.Payload {
				attr := testdata.TestActionAttr{
					Whatever: "doesn't kill you makes you stronger",
				}
				action, err := types.NewAction(types.ACTION_FEE, &attr)
				require.NoError(t, err)

				return &types.Payload{
					PreActions: []*types.Action{action},
				}
			},
			expErr: "unable to resolve",
		},
		{
			name: "error - payload with orbit not registered",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*types.ActionAttributes)(nil),
					&testdata.TestActionAttr{},
				)
			}, payload: func() *types.Payload {
				attrOrbit := testdata.TestOrbitAttr{
					Planet: "saturn",
				}
				orbit, err := types.NewOrbit(types.PROTOCOL_IBC, &attrOrbit, []byte{})
				require.NoError(t, err)

				return &types.Payload{
					Orbit: orbit,
				}
			},
			expErr: "unable to resolve",
		},
		{
			name: "success - payload with orbit and actions",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*types.OrbitAttributes)(nil),
					&testdata.TestOrbitAttr{},
				)
				reg.RegisterImplementations(
					(*types.ActionAttributes)(nil),
					&testdata.TestActionAttr{},
				)
			}, payload: func() *types.Payload {
				attrOrbit := testdata.TestOrbitAttr{
					Planet: "saturn",
				}
				orbit, err := types.NewOrbit(types.PROTOCOL_IBC, &attrOrbit, []byte{})
				require.NoError(t, err)

				attrActions := testdata.TestActionAttr{
					Whatever: "doesn't kill you makes you stronger",
				}
				action, err := types.NewAction(types.ACTION_FEE, &attrActions)
				require.NoError(t, err)

				return &types.Payload{
					Orbit:      orbit,
					PreActions: []*types.Action{action},
				}
			}, expErr: "",
		},
	}

	for _, tC := range testCases {
		encCfg := testutil.MakeTestEncodingConfig("noble")
		if tC.setup != nil {
			tC.setup(encCfg.InterfaceRegistry)
		}

		t.Run("Marshal/"+tC.name, func(t *testing.T) {
			payloadBz, err := types.MarshalJSON(encCfg.Codec, tC.payload())
			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)

				t.Run("Unmarshal", func(t *testing.T) {
					payload := types.Payload{}
					err = types.UnmarshalJSON(encCfg.Codec, payloadBz, &payload)
					require.NoError(t, err)
					require.Equal(t, tC.payload().Orbit, payload.Orbit)
					require.Equal(t, len(tC.payload().PreActions), len(payload.PreActions))
				})
			}
		})

		t.Run("Wrapper/"+tC.name, func(t *testing.T) {
			payload := tC.payload()
			wrapper := types.PayloadWrapper{
				Orbiter: payload,
			}
			payloadWrapperBz, err := types.MarshalJSON(encCfg.Codec, &wrapper)
			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)

				t.Run("Unmarshal", func(t *testing.T) {
					payloadWrapper := types.PayloadWrapper{}
					err = types.UnmarshalJSON(encCfg.Codec, payloadWrapperBz, &payloadWrapper)
					require.NoError(t, err)
					require.Equal(t, tC.payload().Orbit, payloadWrapper.Orbiter.Orbit)
					require.Equal(t, len(tC.payload().PreActions), len(payloadWrapper.Orbiter.PreActions))
				})
			}
		})
	}
}
