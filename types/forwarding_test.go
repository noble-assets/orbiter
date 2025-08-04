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
		attributes         types.ForwardingAttributes
		passthroughPayload []byte
		expErr             string
	}{
		{
			name:               "success - with valid attributes",
			id:                 types.PROTOCOL_IBC,
			attributes:         &testdata.TestForwardingAttr{Planet: "earth"},
			passthroughPayload: []byte("test payload"),
			expErr:             "",
		},
		{
			name:               "success - with nil passthrough payload",
			id:                 types.PROTOCOL_CCTP,
			attributes:         &testdata.TestForwardingAttr{Planet: "earth"},
			passthroughPayload: nil,
			expErr:             "",
		},
		{
			name:               "success - with default attributes",
			id:                 types.PROTOCOL_HYPERLANE,
			attributes:         &testdata.TestForwardingAttr{},
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
			attributes:         &testdata.TestForwardingAttr{Planet: "earth"},
			passthroughPayload: []byte("test"),
			expErr:             types.ErrIDNotSupported.Error(),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			forwarding, err := types.NewForwarding(tC.id, tC.attributes, tC.passthroughPayload)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.id, forwarding.ProtocolId)
				require.Equal(t, tC.passthroughPayload, forwarding.PassthroughPayload)

				// Verify attributes can be retrieved
				attrs, err := forwarding.CachedAttributes()
				require.NoError(t, err)
				require.Equal(t, tC.attributes, attrs)
			}
		})
	}
}

func TestValidateOrbit(t *testing.T) {
	testCases := []struct {
		name       string
		forwarding *types.Forwarding
		expErr     string
	}{
		{
			name: "error - with unsupported forwarding",
			forwarding: &types.Forwarding{
				ProtocolId: types.PROTOCOL_UNSUPPORTED,
			},
			expErr: types.ErrIDNotSupported.Error(),
		},
		{
			name:       "error - with nil forwarding",
			forwarding: nil,
			expErr:     "forwarding is a nil pointer",
		},
		{
			name: "error - when attributes are nil",
			forwarding: &types.Forwarding{
				ProtocolId: types.PROTOCOL_IBC,
			},
			expErr: "not set",
		},
		{
			name: "success - with supported forwarding an non nil attributes",
			forwarding: &types.Forwarding{
				ProtocolId: types.PROTOCOL_IBC,
				Attributes: &codectypes.Any{},
			},
			expErr: "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := tC.forwarding.Validate()

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestProtocolID(t *testing.T) {
	testCases := []struct {
		name       string
		forwarding *types.Forwarding
		expectedID types.ProtocolID
	}{
		{
			name: "return orbit ID when forwarding is valid",
			forwarding: &types.Forwarding{
				ProtocolId: types.PROTOCOL_IBC,
			},
			expectedID: types.PROTOCOL_IBC,
		},
		{
			name:       "return PROTOCOL_UNSUPPORTED when forwarding is nil",
			forwarding: nil,
			expectedID: types.PROTOCOL_UNSUPPORTED,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			id := tC.forwarding.ProtocolID()
			require.Equal(t, tC.expectedID, id)
		})
	}
}
