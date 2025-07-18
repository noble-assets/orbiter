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

	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
)

func TestNewOrbit(t *testing.T) {
	testCases := []struct {
		name               string
		id                 types.ProtocolID
		attributes         types.OrbitAttributes
		passthroughPayload []byte
		expectedError      string
	}{
		{
			name:               "success - with valid attributes",
			id:                 types.PROTOCOL_IBC,
			attributes:         &testdata.TestOrbitAttr{Planet: "earth"},
			passthroughPayload: []byte("test payload"),
			expectedError:      "",
		},
		{
			name:               "success - with nil passthrough payload",
			id:                 types.PROTOCOL_CCTP,
			attributes:         &testdata.TestOrbitAttr{Planet: "earth"},
			passthroughPayload: nil,
			expectedError:      "",
		},
		{
			name:               "error - with empty attributes",
			id:                 types.PROTOCOL_HYPERLANE,
			attributes:         &testdata.TestOrbitAttr{},
			passthroughPayload: []byte("test"),
			expectedError:      "",
		},
		{
			name:               "error - with unsupported id",
			id:                 types.PROTOCOL_UNSUPPORTED,
			attributes:         &testdata.TestOrbitAttr{Planet: "earth"},
			passthroughPayload: []byte("test"),
			expectedError:      types.ErrIdNotSupported.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			orbit, err := types.NewOrbit(tc.id, tc.attributes, tc.passthroughPayload)

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.id, orbit.ProtocolId)
				require.Equal(t, tc.passthroughPayload, orbit.PassthroughPayload)

				// Verify attributes can be retrieved
				attrs, err := orbit.CachedAttributes()
				require.NoError(t, err)
				require.Equal(t, tc.attributes, attrs)
			}
		})
	}
}

func TestOrbit_Validate(t *testing.T) {
	testCases := []struct {
		name          string
		orbit         types.Orbit
		expectedError string
	}{
		{
			name: "fail with unsupported orbit",
			orbit: types.Orbit{
				ProtocolId: types.PROTOCOL_UNSUPPORTED,
			},
			expectedError: types.ErrIdNotSupported.Error(),
		},
		{
			name: "fail when attributes are nil",
			orbit: types.Orbit{
				ProtocolId: types.PROTOCOL_IBC,
			},
			expectedError: "not set",
		},
		{
			name: "success with supported orbit an non nil attributes",
			orbit: types.Orbit{
				ProtocolId: types.PROTOCOL_IBC,
				Attributes: &codectypes.Any{},
			},
			expectedError: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.orbit.Validate()

			if tc.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOrbit_ProtocolID(t *testing.T) {
	testCases := []struct {
		name       string
		orbit      *types.Orbit
		expectedId types.ProtocolID
	}{
		{
			name: "return orbit id when orbit is not nil",
			orbit: &types.Orbit{
				ProtocolId: types.PROTOCOL_IBC,
			},
			expectedId: types.PROTOCOL_IBC,
		},
		{
			name:       "return PROTOCOL_UNSUPPORTED when orbit is nil",
			orbit:      nil,
			expectedId: types.PROTOCOL_UNSUPPORTED,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.orbit.ProtocolID()
			require.Equal(t, tc.expectedId, id)
		})
	}
}

func TestOrbitID(t *testing.T) {
	testCases := []struct {
		name       string
		orbitID    types.OrbitID
		expectedId string
	}{
		{
			name: "IBC orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_IBC,
				CounterpartyID: "channel-1",
			},
			expectedId: "1:channel-1",
		},
		{
			name: "CCTP orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_CCTP,
				CounterpartyID: "0",
			},
			expectedId: "2:0",
		},
		{
			name: "Hyperlane orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_HYPERLANE,
				CounterpartyID: "ethereum",
			},
			expectedId: "3:ethereum",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.orbitID.ID()
			require.Equal(t, tc.expectedId, id)
		})
	}
}

func TestOrbitIDFromString(t *testing.T) {
	testCases := []struct {
		name                   string
		id                     string
		expectedProtocolID     types.ProtocolID
		expectedCounterpartyId string
		expErr                 string
	}{
		{
			name:   "invalid format - no colon",
			id:     "1channel-1",
			expErr: "invalid orbit",
		},
		{
			name:   "invalid format - multiple colons",
			id:     "1:channel:1",
			expErr: "invalid orbit",
		},
		{
			name:   "non numeric protocol Id",
			id:     "invalid:channel-1",
			expErr: "invalid protocol",
		},
		{
			name:   "empty string",
			id:     "",
			expErr: "invalid orbit",
		},
		{
			name:                   "valid IBC Id",
			id:                     "1:channel-1",
			expectedProtocolID:     types.PROTOCOL_IBC,
			expectedCounterpartyId: "channel-1",
		},
		{
			name:                   "valid CCTP Id",
			id:                     "2:0",
			expectedProtocolID:     types.PROTOCOL_CCTP,
			expectedCounterpartyId: "0",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			orbitID := types.OrbitID{}
			err := orbitID.FromString(tC.id)

			if tC.expErr != "" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.expectedProtocolID, orbitID.ProtocolID)
				require.Equal(t, tC.expectedCounterpartyId, orbitID.CounterpartyID)
			}
		})
	}
}
