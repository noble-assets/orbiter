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

package components_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/keeper/components"
	"orbiter.dev/testutil/mocks"
	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

func TestNewDispatcherComponent(t *testing.T) {
	deps := mocks.NewDependencies(t)

	testCases := []struct {
		name           string
		codec          codec.Codec
		logger         log.Logger
		OrbitsHandler  interfaces.PacketHandler[*types.OrbitPacket]
		ActionsHandler interfaces.PacketHandler[*types.ActionPacket]
		expError       string
	}{
		{
			name:           "success - passing all correct inputs",
			codec:          deps.EncCfg.Codec,
			logger:         deps.Logger,
			OrbitsHandler:  &mocks.OrbitsHandler{},
			ActionsHandler: &mocks.ActionsHandler{},
			expError:       "",
		},
	}

	for _, tc := range testCases {
		sb := collections.NewSchemaBuilder(deps.StoreService)
		_, err := components.NewDispatcherComponent(
			tc.codec,
			sb,
			tc.logger,
			tc.OrbitsHandler,
			tc.ActionsHandler,
		)

		t.Run(tc.name, func(t *testing.T) {
			if tc.expError != "" {
				require.ErrorContains(t, err, tc.expError)
			} else {
				require.NoError(t, err)
				_, err = sb.Build()
				require.NoError(t, err)
			}
		})
	}
}

func TestValidate_DispatcherComponent(t *testing.T) {
	testCases := []struct {
		name           string
		OrbitsHandler  interfaces.PacketHandler[*types.OrbitPacket]
		ActionsHandler interfaces.PacketHandler[*types.ActionPacket]
		expError       string
	}{
		{
			name:           "success - all mandatory fields are set",
			OrbitsHandler:  &mocks.OrbitsHandler{},
			ActionsHandler: &mocks.ActionsHandler{},
			expError:       "",
		},
		{
			name:           "error - nil orbits handler",
			OrbitsHandler:  nil,
			ActionsHandler: &mocks.ActionsHandler{},
			expError:       "cannot be nil",
		},
		{
			name:           "error - nil actions handler",
			OrbitsHandler:  &mocks.OrbitsHandler{},
			ActionsHandler: nil,
			expError:       "cannot be nil",
		},
	}

	for _, tc := range testCases {
		dispatcher := components.DispatcherComponent{
			OrbitHandler:  tc.OrbitsHandler,
			ActionHandler: tc.ActionsHandler,
		}
		err := dispatcher.Validate()

		t.Run(tc.name, func(t *testing.T) {
			if tc.expError != "" {
				require.ErrorContains(t, err, tc.expError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
