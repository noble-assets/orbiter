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

package executor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"orbiter.dev/keeper/component/executor"
	mockorbiter "orbiter.dev/testutil/mocks/orbiter"
	executortypes "orbiter.dev/types/component/executor"
	"orbiter.dev/types/core"
)

func TestInitGenesis(t *testing.T) {
	tests := []struct {
		name       string
		setupState func(ctx context.Context, k *executor.Executor)
		genState   *executortypes.GenesisState
		expErr     string
	}{
		{
			name:       "success - default genesis state",
			setupState: func(ctx context.Context, k *executor.Executor) {},
			genState:   executortypes.DefaultGenesisState(),
			expErr:     "",
		},
		{
			name:       "success - genesis state with paused action IDs",
			setupState: func(ctx context.Context, k *executor.Executor) {},
			genState: &executortypes.GenesisState{
				PausedActionIds: []core.ActionID{core.ACTION_FEE},
			},
			expErr: "",
		},
		{
			name: "success - init genesis overwrites existing paused actions",
			setupState: func(ctx context.Context, k *executor.Executor) {
				require.NoError(t, k.SetPausedAction(ctx, core.ACTION_FEE))
			},
			genState: &executortypes.GenesisState{
				PausedActionIds: []core.ActionID{},
			},
			expErr: "",
		},
		{
			name:       "error - nil genesis state",
			setupState: func(ctx context.Context, k *executor.Executor) {},
			genState:   nil,
			expErr:     "executor genesis: invalid nil pointer",
		},
		{
			name:       "error - invalid action ID",
			setupState: func(ctx context.Context, k *executor.Executor) {},
			genState: &executortypes.GenesisState{
				PausedActionIds: []core.ActionID{core.ACTION_UNSUPPORTED},
			},
			expErr: "action ID is not supported",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			ex := k.Executor()

			tc.setupState(ctx, ex)

			err := ex.InitGenesis(ctx, tc.genState)
			if tc.expErr != "" {
				require.ErrorContains(t, err, tc.expErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	tests := []struct {
		name             string
		setupState       func(ctx context.Context, k *executor.Executor)
		expPausedActions []core.ActionID
	}{
		{
			name:             "success - export default genesis state",
			setupState:       func(ctx context.Context, k *executor.Executor) {},
			expPausedActions: []core.ActionID{},
		},
		{
			name: "success - export genesis state with paused actions",
			setupState: func(ctx context.Context, k *executor.Executor) {
				require.NoError(t, k.SetPausedAction(ctx, core.ACTION_FEE))
			},
			expPausedActions: []core.ActionID{core.ACTION_FEE},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := mockorbiter.OrbiterKeeper(t)
			ex := k.Executor()

			tc.setupState(ctx, ex)

			genState := ex.ExportGenesis(ctx)

			require.NotNil(t, genState)
			require.ElementsMatch(t, tc.expPausedActions, genState.PausedActionIds)
		})
	}
}
