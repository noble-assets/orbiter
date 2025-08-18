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
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/testutil/mocks"
	executortypes "orbiter.dev/types/component/executor"
	"orbiter.dev/types/core"
)

func TestInitGenesis(t *testing.T) {
	e, deps := mocks.NewExecutorComponent(t)

	defaultPausedActionIDs, err := e.GetPausedActions(deps.SdkCtx)
	require.NoError(t, err, "failed to get paused actions")

	// ACT: fail for invalid genesis state
	invalidGenState := executortypes.GenesisState{
		PausedActionIds: []core.ActionID{core.ACTION_UNSUPPORTED},
	}
	err = e.InitGenesis(deps.SdkCtx, &invalidGenState)
	require.ErrorContains(
		t,
		err,
		"ID is not supported",
		"expected error initializing genesis",
	)

	actionIDs, err := e.GetPausedActions(deps.SdkCtx)
	require.NoError(t, err, "failed to get paused actions")
	require.ElementsMatch(
		t,
		defaultPausedActionIDs,
		actionIDs,
		"expected paused actions to not have changed",
	)

	// ACT: update correctly for valid genesis state
	updatedActionIDs := []core.ActionID{core.ACTION_FEE, core.ACTION_SWAP}
	require.NotElementsMatch(
		t,
		defaultPausedActionIDs,
		updatedActionIDs,
		"updated action IDs should be different",
	)

	validGenState := executortypes.GenesisState{PausedActionIds: updatedActionIDs}
	err = e.InitGenesis(deps.SdkCtx, &validGenState)
	require.NoError(t, err, "failed to init genesis state")

	actionIDs, err = e.GetPausedActions(deps.SdkCtx)
	require.NoError(t, err, "failed to get paused actions")
	require.ElementsMatch(t, updatedActionIDs, actionIDs, "action IDs should have been updated")
}

func TestExportGenesis(t *testing.T) {
	e, deps := mocks.NewExecutorComponent(t)

	expPausedActions := []core.ActionID{core.ACTION_FEE}
	expGenState := executortypes.GenesisState{PausedActionIds: expPausedActions}

	err := e.SetPausedAction(deps.SdkCtx, expPausedActions[0])
	require.NoError(t, err, "failed to set paused action")

	genState := e.ExportGenesis(deps.SdkCtx)
	require.Equal(t, expGenState.String(), genState.String())
}
