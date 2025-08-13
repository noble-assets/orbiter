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
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/keeper/component/adapter"
	"orbiter.dev/testutil/mocks"
	adaptertypes "orbiter.dev/types/component/adapter"
)

func TestInitGenesis(t *testing.T) {
	testcases := []struct {
		name     string
		genState *adaptertypes.GenesisState
		expErr   string
	}{
		{
			name:     "success - default genesis state",
			genState: adaptertypes.DefaultGenesisState(),
			expErr:   "",
		},
		{
			name: "success - genesis state with custom params",
			genState: &adaptertypes.GenesisState{
				Params: adaptertypes.Params{
					MaxPassthroughPayloadSize: 1024,
				},
			},
			expErr: "",
		},
		{
			name:     "error - nil genesis state",
			genState: nil,
			expErr:   "adapter genesis: invalid nil pointer",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			a, deps := mocks.NewAdapterComponent(t)
			ctx := deps.SdkCtx

			err := a.InitGenesis(ctx, tc.genState)
			if tc.expErr != "" {
				require.ErrorContains(t, err, tc.expErr)
			} else {
				require.NoError(t, err)

				params := a.GetParams(ctx)
				require.Equal(t, tc.genState.Params.MaxPassthroughPayloadSize, params.MaxPassthroughPayloadSize)
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	testcases := []struct {
		name      string
		setup     func(ctx context.Context, k *adapter.Adapter)
		expParams adaptertypes.Params
	}{
		{
			name: "success - export default genesis state",
			expParams: adaptertypes.Params{
				MaxPassthroughPayloadSize: 0,
			},
		},
		{
			name: "success - export genesis state with custom params",
			setup: func(ctx context.Context, k *adapter.Adapter) {
				params := adaptertypes.Params{MaxPassthroughPayloadSize: 1024}
				require.NoError(t, k.SetParams(ctx, params))
			},
			expParams: adaptertypes.Params{
				MaxPassthroughPayloadSize: 1024,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			a, deps := mocks.NewAdapterComponent(t)
			ctx := deps.SdkCtx

			if tc.setup != nil {
				tc.setup(ctx, a)
			}

			genState := a.ExportGenesis(ctx)

			require.Equal(
				t,
				tc.expParams.MaxPassthroughPayloadSize,
				genState.Params.MaxPassthroughPayloadSize,
			)
		})
	}
}
