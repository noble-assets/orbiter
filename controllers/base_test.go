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

package controllers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"orbiter.dev/controllers"
	"orbiter.dev/types"
)

func TestNewBaseController(t *testing.T) {
	testCases := []struct {
		name       string
		protocolID types.ProtocolID
		expErr     string
	}{
		{
			name:       "error - when the ID is not valid",
			protocolID: types.PROTOCOL_UNSUPPORTED,
			expErr:     "id is not supported",
		},
		{
			name:       "success - with valid ID",
			protocolID: types.PROTOCOL_IBC,
			expErr:     "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			controller, err := controllers.NewBaseController(tC.protocolID)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				id := controller.ID()
				require.Equal(t, tC.protocolID, id)

				name := controller.Name()
				require.Equal(t, tC.protocolID.String(), name)
			}
		})
	}
}
