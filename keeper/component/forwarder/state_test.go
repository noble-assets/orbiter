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

package forwarder_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/testutil/mocks"
	"orbiter.dev/types/core"
)

func TestGetPausedCrossChains(t *testing.T) {
	ccIDs := []core.CrossChainID{
		{
			ProtocolId:     1,
			CounterpartyId: "one",
		},
		{
			ProtocolId:     1,
			CounterpartyId: "two",
		},
		{
			ProtocolId:     1,
			CounterpartyId: "three",
		},
		{
			ProtocolId:     2,
			CounterpartyId: "one",
		},
		{
			ProtocolId:     2,
			CounterpartyId: "two",
		},
	}

	f, deps := mocks.NewForwarderComponent(t)
	ctx := deps.SdkCtx

	for _, ccID := range ccIDs {
		err := f.SetPausedCrossChain(ctx, ccID)
		require.NoError(t, err)
	}

	testCases := []struct {
		name       string
		protocolID *core.ProtocolID
		expMapLen  int
		wantKey    core.ProtocolID
		expFound   bool
		expLen     int
	}{
		{
			name:       "no results for protocols not stored",
			protocolID: ptr(core.ProtocolID(100)),
			expMapLen:  0,
			wantKey:    0,
			expFound:   false,
			expLen:     0,
		},
		{
			name:       "paused cross-chains for protocol ID 1",
			protocolID: ptr(core.ProtocolID(1)),
			expMapLen:  1,
			wantKey:    1,
			expFound:   true,
			expLen:     3,
		},
		{
			name:       "paused cross-chains for protocol ID 2",
			protocolID: ptr(core.ProtocolID(2)),
			expMapLen:  1,
			wantKey:    2,
			expFound:   true,
			expLen:     2,
		},
		{
			name:       "all paused cross chains",
			protocolID: nil,
			expMapLen:  2,
			wantKey:    0,     // not checking specific key for nil case
			expFound:   false, // not checking found for nil case
			expLen:     0,     // not checking chains length for nil case
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			paused, err := f.GetPausedCrossChainsMap(ctx, tC.protocolID)
			require.NoError(t, err)
			require.Len(t, paused, tC.expMapLen)

			// Skip key checks for nil protocol ID case
			if tC.protocolID != nil {
				idPaused, found := paused[int32(tC.wantKey)]
				require.Equal(t, tC.expFound, found)
				if found {
					require.Len(t, idPaused, tC.expLen)
				}
			}
		})
	}
}

// Helper function.
func ptr[T any](v T) *T {
	return &v
}
