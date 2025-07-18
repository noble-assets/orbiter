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
	fmt "fmt"
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
		expErr             string
	}{
		{
			name:               "success - with valid attributes",
			id:                 types.PROTOCOL_IBC,
			attributes:         &testdata.TestOrbitAttr{Planet: "earth"},
			passthroughPayload: []byte("test payload"),
			expErr:             "",
		},
		{
			name:               "success - with nil passthrough payload",
			id:                 types.PROTOCOL_CCTP,
			attributes:         &testdata.TestOrbitAttr{Planet: "earth"},
			passthroughPayload: nil,
			expErr:             "",
		},
		{
			name:               "success - with default attributes",
			id:                 types.PROTOCOL_HYPERLANE,
			attributes:         &testdata.TestOrbitAttr{},
			passthroughPayload: []byte("test"),
			expErr:             "",
		},
		{
			name:               "error - with nil attributes",
			id:                 types.PROTOCOL_HYPERLANE,
			attributes:         nil,
			passthroughPayload: []byte("test"),
			expErr:             "can't proto marshal",
		},
		{
			name:               "error - with unsupported id",
			id:                 types.PROTOCOL_UNSUPPORTED,
			attributes:         &testdata.TestOrbitAttr{Planet: "earth"},
			passthroughPayload: []byte("test"),
			expErr:             types.ErrIDNotSupported.Error(),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			orbit, err := types.NewOrbit(tC.id, tC.attributes, tC.passthroughPayload)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.id, orbit.ProtocolId)
				require.Equal(t, tC.passthroughPayload, orbit.PassthroughPayload)

				// Verify attributes can be retrieved
				attrs, err := orbit.CachedAttributes()
				require.NoError(t, err)
				require.Equal(t, tC.attributes, attrs)
			}
		})
	}
}

func TestValidateOrbit(t *testing.T) {
	testCases := []struct {
		name   string
		orbit  *types.Orbit
		expErr string
	}{
		{
			name: "error - with unsupported orbit",
			orbit: &types.Orbit{
				ProtocolId: types.PROTOCOL_UNSUPPORTED,
			},
			expErr: types.ErrIDNotSupported.Error(),
		},
		{
			name:   "error - with nil orbit",
			orbit:  nil,
			expErr: "orbit is a nil pointer",
		},
		{
			name: "error - when attributes are nil",
			orbit: &types.Orbit{
				ProtocolId: types.PROTOCOL_IBC,
			},
			expErr: "not set",
		},
		{
			name: "success - with supported orbit an non nil attributes",
			orbit: &types.Orbit{
				ProtocolId: types.PROTOCOL_IBC,
				Attributes: &codectypes.Any{},
			},
			expErr: "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := tC.orbit.Validate()

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProtocolIDOrbit(t *testing.T) {
	testCases := []struct {
		name       string
		orbit      *types.Orbit
		expectedID types.ProtocolID
	}{
		{
			name: "return orbit ID when orbit is valid",
			orbit: &types.Orbit{
				ProtocolId: types.PROTOCOL_IBC,
			},
			expectedID: types.PROTOCOL_IBC,
		},
		{
			name:       "return PROTOCOL_UNSUPPORTED when orbit is nil",
			orbit:      nil,
			expectedID: types.PROTOCOL_UNSUPPORTED,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			id := tC.orbit.ProtocolID()
			require.Equal(t, tC.expectedID, id)
		})
	}
}

func TestOrbitID(t *testing.T) {
	testCases := []struct {
		name       string
		orbitID    types.OrbitID
		expectedID string
	}{
		{
			name: "IBC orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_IBC,
				CounterpartyID: "channel-1",
			},
			expectedID: fmt.Sprintf("%d:channel-1", types.PROTOCOL_IBC),
		},
		{
			name: "CCTP orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_CCTP,
				CounterpartyID: "0",
			},
			expectedID: fmt.Sprintf("%d:0", types.PROTOCOL_CCTP),
		},
		{
			name: "Hyperlane orbit ID",
			orbitID: types.OrbitID{
				ProtocolID:     types.PROTOCOL_HYPERLANE,
				CounterpartyID: "ethereum",
			},
			expectedID: fmt.Sprintf("%d:ethereum", types.PROTOCOL_HYPERLANE),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			id := tC.orbitID.ID()
			require.Equal(t, tC.expectedID, id)
		})
	}
}

func TestParseOrbitID(t *testing.T) {
	testCases := []struct {
		name                   string
		id                     string
		expectedProtocolID     types.ProtocolID
		expectedCounterpartyID string
		expErr                 string
	}{
		{
			name:   "error - when invalid format (no colon)",
			id:     "1channel-1",
			expErr: "invalid orbit",
		},
		{
			name:   "error - when non numeric protocol ID",
			id:     "invalid:channel-1",
			expErr: "invalid protocol",
		},
		{
			name:   "error - when empty string",
			id:     "",
			expErr: "invalid orbit",
		},
		{
			name:                   "error - when the format is not valid (multiple colons)",
			id:                     "1:channel:1",
			expectedProtocolID:     types.PROTOCOL_IBC,
			expectedCounterpartyID: "channel:1",
		},
		{
			name:                   "success - with valid IBC ID",
			id:                     "1:channel-1",
			expectedProtocolID:     types.PROTOCOL_IBC,
			expectedCounterpartyID: "channel-1",
		},
		{
			name:                   "success - with valid CCTP ID",
			id:                     "2:0",
			expectedProtocolID:     types.PROTOCOL_CCTP,
			expectedCounterpartyID: "0",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			orbitID, err := types.ParseOrbitID(tC.id)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.expectedProtocolID, orbitID.ProtocolID)
				require.Equal(t, tC.expectedCounterpartyID, orbitID.CounterpartyID)
			}
		})
	}
}
