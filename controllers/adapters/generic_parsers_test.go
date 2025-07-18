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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"orbiter.dev/controllers/adapters"
	"orbiter.dev/testutil"
	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
)

func TestJSONParser_Parse(t *testing.T) {
	validPayload, validPayloadStr := testutil.CreatePayloadWrapperJSON(t)
	validPayloadWithActions, validPayloadWithActionsStr := testutil.CreatePayloadWrapperWithActionJSON(
		t,
	)

	testCases := []struct {
		name           string
		setup          func(reg codectypes.InterfaceRegistry)
		orbiterPayload func() string
		expPayload     *types.Payload
		expErr         string
	}{
		{
			name:           "fail when string is empty",
			orbiterPayload: func() string { return "" },
			expErr:         "not a valid json",
		},
		{
			name:           "fail when string is not valid JSON",
			orbiterPayload: func() string { return "invalid json string" },
			expErr:         "not a valid json",
		},
		{
			name:           "fail when string does not contain orbiter prefix",
			orbiterPayload: func() string { return `{"other_field": "value"}` },
			expErr:         "json does not contain orbiter prefix",
		},
		{
			name: "fail when orbiter prefix exists but is null",
			orbiterPayload: func() string {
				return fmt.Sprintf(`{"%s": null}`, types.OrbiterPrefix)
			},
			expErr: "json does not contain orbiter prefix",
		},
		{
			name: "fail when orbiter prefix is not a map",
			orbiterPayload: func() string {
				return fmt.Sprintf(`{"%s": "string_value"}`, types.OrbiterPrefix)
			},
			expErr: "failed to cast json string into Payload",
		},
		{
			name: "fail when orbiter prefix is an array",
			orbiterPayload: func() string {
				return fmt.Sprintf(`{"%s": ["array", "value"]}`, types.OrbiterPrefix)
			},
			expErr: "failed to cast json string into Payload",
		},
		{
			name: "fail when orbiter prefix is a number",
			orbiterPayload: func() string {
				return fmt.Sprintf(`{"%s": 123}`, types.OrbiterPrefix)
			},
			expErr: "failed to cast json string into Payload",
		},
		{
			name: "fail when orbiter prefix is a boolean",
			orbiterPayload: func() string {
				return fmt.Sprintf(`{"%s": true}`, types.OrbiterPrefix)
			},
			expErr: "failed to cast json string into Payload",
		},
		{
			name: "fail when orbiter prefix contains invalid data for NewPayloadFromString",
			orbiterPayload: func() string {
				return fmt.Sprintf(
					`{"%s": {"invalid_field": "invalid_value"}}`,
					types.OrbiterPrefix,
				)
			},
			expErr: "failed to cast json string into Payload",
		},
		{
			name: "fail - when payload is valid but attributes are not registered",
			orbiterPayload: func() string {
				_, str := testutil.CreatePayloadWrapperJSON(t)
				return str
			},
			expErr: "unable to resolve type",
		},
		{
			name: "success - valid payload",
			setup: func(reg codectypes.InterfaceRegistry) {
				reg.RegisterImplementations(
					(*types.OrbitAttributes)(nil),
					&testdata.TestOrbitAttr{},
				)
			},
			orbiterPayload: func() string {
				return validPayloadStr
			},
			expPayload: validPayload,
			expErr:     "",
		},
		{
			name: "success - valid payload with actions",
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
			orbiterPayload: func() string {
				return validPayloadWithActionsStr
			},
			expPayload: validPayloadWithActions,
			expErr:     "",
		},
	}

	for _, tC := range testCases {
		encCfg := testutil.MakeTestEncodingConfig("noble")

		parser, err := adapters.NewJSONParser(encCfg.Codec)
		require.NoError(t, err)

		t.Run(tC.name, func(t *testing.T) {
			if tC.setup != nil {
				tC.setup(encCfg.InterfaceRegistry)
			}

			payload, err := parser.Parse(tC.orbiterPayload())

			if tC.expErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tC.expErr)
				require.Equal(t, tC.expPayload, payload)
			} else {
				require.NoError(t, err)
				expAttr, err := tC.expPayload.Orbit.CachedAttributes()
				require.NoError(t, err)
				attr, err := payload.Orbit.CachedAttributes()
				require.NoError(t, err)
				require.Equal(t, expAttr, attr)

				require.Equal(t, len(tC.expPayload.PreActions), len(payload.PreActions))
				for idx, action := range tC.expPayload.PreActions {
					expAttr, err := action.CachedAttributes()
					require.NoError(t, err)
					attr, err := payload.PreActions[idx].CachedAttributes()
					require.NoError(t, err)
					require.Equal(t, expAttr, attr)
				}
			}
		})
	}
}
