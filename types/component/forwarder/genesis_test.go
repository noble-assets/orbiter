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

package forwarder

import (
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/types/core"
)

func TestValidate(t *testing.T) {
	testcases := []struct {
		name     string
		genState *GenesisState
		expErr   string
	}{
		{
			name:     "success - default genesis state",
			genState: DefaultGenesisState(),
		},
		{
			name: "success - valid genesis state with protocol and cross-chain ids",
			genState: &GenesisState{
				PausedProtocolIds: []core.ProtocolID{core.PROTOCOL_IBC, core.PROTOCOL_CCTP},
				PausedCrossChainIds: []*core.CrossChainID{
					{ProtocolId: core.PROTOCOL_IBC, CounterpartyId: "ibc-1"},
					{ProtocolId: core.PROTOCOL_CCTP, CounterpartyId: "cctp-2"},
				},
			},
		},
		{
			name:     "error - nil genesis state",
			genState: nil,
			expErr:   core.ErrNilPointer.Error(),
		},
		{
			name: "error - invalid genesis state (unsupported protocol)",
			genState: &GenesisState{
				PausedProtocolIds: []core.ProtocolID{core.PROTOCOL_UNSUPPORTED},
			},
			expErr: "invalid paused protocol ID",
		},
		{
			name: "error - invalid genesis state (nil cross-chain id)",
			genState: &GenesisState{
				PausedCrossChainIds: []*core.CrossChainID{nil},
			},
			expErr: "invalid paused cross-chain ID",
		},
		{
			name: "error - invalid genesis state (unsupported cross chain ids)",
			genState: &GenesisState{
				PausedCrossChainIds: []*core.CrossChainID{
					{ProtocolId: core.PROTOCOL_UNSUPPORTED, CounterpartyId: "unsupported-2"},
				},
			},
			expErr: "invalid paused cross-chain ID",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.expErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.expErr)
			}
		})
	}
}
