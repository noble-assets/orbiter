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

package executor

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
			name: "success - genesis state with paused action ids",
			genState: &GenesisState{
				PausedActionIds: []core.ActionID{core.ACTION_FEE},
			},
		},
		{
			name:     "error - genesis state with paused action ids",
			genState: nil,
			expErr:   core.ErrNilPointer.Error(),
		},
		{
			name: "error - genesis state with invalid action id",
			genState: &GenesisState{
				PausedActionIds: []core.ActionID{core.ACTION_UNSUPPORTED},
			},
			expErr: "ID is not supported",
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
