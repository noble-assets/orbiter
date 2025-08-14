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

package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	mockorbiter "orbiter.dev/testutil/mocks/orbiter"
	orbitertypes "orbiter.dev/types"
	adaptertypes "orbiter.dev/types/component/adapter"
	executortypes "orbiter.dev/types/component/executor"
	forwardertypes "orbiter.dev/types/component/forwarder"
)

func TestInitGenesis(t *testing.T) {
	testcases := []struct {
		name     string
		genState orbitertypes.GenesisState
		expPanic bool
	}{
		{
			name:     "success - default genesis state",
			genState: *orbitertypes.DefaultGenesisState(),
			expPanic: false,
		},
		{
			name: "error - nil adapter genesis state",
			genState: orbitertypes.GenesisState{
				AdapterGenesis:   nil,
				ExecutorGenesis:  executortypes.DefaultGenesisState(),
				ForwarderGenesis: forwardertypes.DefaultGenesisState(),
			},
			expPanic: true,
		},
		{
			name: "error - nil executor genesis state",
			genState: orbitertypes.GenesisState{
				AdapterGenesis:   adaptertypes.DefaultGenesisState(),
				ExecutorGenesis:  nil,
				ForwarderGenesis: forwardertypes.DefaultGenesisState(),
			},
			expPanic: true,
		},
		{
			name: "error - nil forwarder genesis state",
			genState: orbitertypes.GenesisState{
				AdapterGenesis:   adaptertypes.DefaultGenesisState(),
				ExecutorGenesis:  executortypes.DefaultGenesisState(),
				ForwarderGenesis: nil,
			},
			expPanic: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)

			if tc.expPanic {
				require.Panics(t, func() {
					k.InitGenesis(ctx, tc.genState)
				})
			} else {
				require.NotPanics(t, func() {
					k.InitGenesis(ctx, tc.genState)
				})
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	ctx, _, k := mockorbiter.OrbiterKeeper(t)
	defaultGenState := orbitertypes.DefaultGenesisState()
	k.InitGenesis(ctx, *defaultGenState)

	genState := k.ExportGenesis(ctx)
	// NOTE: we're comparing the `String()` here to avoid the difference in the way that Go gets the
	// nil vs. empty slice here
	require.Equal(t, defaultGenState.String(), genState.String())
}
