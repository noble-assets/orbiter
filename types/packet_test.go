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

	"orbiter.dev/types"
)

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
