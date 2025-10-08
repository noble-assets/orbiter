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

package adapter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/testutil/mocks"
	adaptertypes "github.com/noble-assets/orbiter/types/component/adapter"
	"github.com/noble-assets/orbiter/types/core"
)

func TestInitGenesis(t *testing.T) {
	a, deps := mocks.NewAdapterComponent(t)

	defaultParams, err := a.GetParams(deps.SdkCtx)
	require.Error(t, err)

	// ACT: fail for invalid genesis state (nil)
	err = a.InitGenesis(deps.SdkCtx, nil)
	require.ErrorIs(t, err, core.ErrNilPointer)

	params, _ := a.GetParams(deps.SdkCtx)
	require.Equal(t, defaultParams, params, "params should not have changed")

	// ACT: update params for valid genesis state
	validParams := adaptertypes.Params{MaxPassthroughPayloadSize: 1024}
	require.NotEqual(
		t,
		defaultParams,
		validParams,
		"new params should be different from current params",
	)
	validGenState := adaptertypes.GenesisState{Params: validParams}

	err = a.InitGenesis(deps.SdkCtx, &validGenState)
	require.NoError(t, err, "failed to init genesis state")

	params, err = a.GetParams(deps.SdkCtx)
	require.NoError(t, err)
	require.Equal(
		t,
		validGenState.Params,
		params,
		"expected params to have been updated",
	)
}

func TestExportGenesis(t *testing.T) {
	a, deps := mocks.NewAdapterComponent(t)
	params, _ := a.GetParams(deps.SdkCtx)
	expGenState := adaptertypes.GenesisState{Params: params}

	genState := a.ExportGenesis(deps.SdkCtx)
	require.Equal(t, expGenState.String(), genState.String(), "expected different gen state")
}
