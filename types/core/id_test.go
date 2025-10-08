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

package core_test

import (
	fmt "fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/types/core"
)

func TestCrossChainID(t *testing.T) {
	testCases := []struct {
		name       string
		id         core.CrossChainID
		expectedID string
	}{
		{
			name: "IBC cross-chain ID",
			id: core.CrossChainID{
				ProtocolId:     core.PROTOCOL_IBC,
				CounterpartyId: "channel-1",
			},
			expectedID: fmt.Sprintf("%d:channel-1", core.PROTOCOL_IBC),
		},
		{
			name: "CCTP cross-chain ID",
			id: core.CrossChainID{
				ProtocolId:     core.PROTOCOL_CCTP,
				CounterpartyId: "0",
			},
			expectedID: fmt.Sprintf("%d:0", core.PROTOCOL_CCTP),
		},
		{
			name: "Hyperlane cross-chain ID",
			id: core.CrossChainID{
				ProtocolId:     core.PROTOCOL_HYPERLANE,
				CounterpartyId: "ethereum",
			},
			expectedID: fmt.Sprintf("%d:ethereum", core.PROTOCOL_HYPERLANE),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			id := tC.id.ID()
			require.Equal(t, tC.expectedID, id)
		})
	}
}

func TestParseCrossChainID(t *testing.T) {
	testCases := []struct {
		name              string
		id                string
		expProtocolID     core.ProtocolID
		expCounterpartyID string
		expErr            string
	}{
		{
			name:   "error - when invalid format (no colon)",
			id:     "1channel-1",
			expErr: "invalid cross-chain ID",
		},
		{
			name:   "error - when non numeric protocol ID",
			id:     "invalid:channel-1",
			expErr: "invalid protocol",
		},
		{
			name:   "error - when empty string",
			id:     "",
			expErr: "invalid cross-chain ID",
		},
		{
			name:   "error - when the format is not valid (multiple colons)",
			id:     "1:channel:1",
			expErr: "invalid cross-chain ID",
		},
		{
			name:              "success - with valid IBC ID",
			id:                "1:channel-1",
			expProtocolID:     core.PROTOCOL_IBC,
			expCounterpartyID: "channel-1",
		},
		{
			name:              "success - with valid CCTP ID",
			id:                "2:0",
			expProtocolID:     core.PROTOCOL_CCTP,
			expCounterpartyID: "0",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ccID, err := core.ParseCrossChainID(tC.id)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tC.expProtocolID, ccID.GetProtocolId())
				require.Equal(t, tC.expCounterpartyID, ccID.GetCounterpartyId())
			}
		})
	}
}

func TestValidateCounterpartyID(t *testing.T) {
	testCases := []struct {
		name       string
		id         string
		protocolID core.ProtocolID
		expError   string
	}{
		{
			name:       "error - empty id",
			id:         "",
			protocolID: core.PROTOCOL_CCTP,
			expError:   "cannot be empty",
		},
		{
			name:       "error - too long",
			id:         strings.Repeat("a", core.MaxCounterpartyIDLength+1),
			protocolID: core.PROTOCOL_CCTP,
			expError:   "cannot contains more",
		},
		{
			name:       "error - not a number with CCTP",
			id:         " counterparty ",
			protocolID: core.PROTOCOL_CCTP,
			expError:   "invalid counterparty ID",
		},
		{
			name:       "error - wrong string format with IBC",
			id:         " counterparty ",
			protocolID: core.PROTOCOL_IBC,
			expError:   "invalid counterparty ID",
		},
		{
			name:     "error - with default protocol ID",
			id:       "12345",
			expError: "invalid counterparty ID",
		},
		{
			name:       "success - valid id with CCTP",
			id:         "12345",
			protocolID: core.PROTOCOL_CCTP,
			expError:   "",
		},
		{
			name:       "success - valid id with IBC",
			id:         "channel-1",
			protocolID: core.PROTOCOL_IBC,
			expError:   "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := core.ValidateCounterpartyID(tC.id, tC.protocolID)
			if tC.expError != "" {
				require.ErrorContains(t, err, tC.expError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
