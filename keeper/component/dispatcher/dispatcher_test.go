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

package dispatcher_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/keeper/component/dispatcher"
	"orbiter.dev/testutil/mocks"
	"orbiter.dev/types"
)

func TestNew(t *testing.T) {
	deps := mocks.NewDependencies(t)

	testCases := []struct {
		name              string
		codec             codec.Codec
		logger            log.Logger
		sb                *collections.SchemaBuilder
		ForwardingHandler types.PacketHandler[*types.ForwardingPacket]
		ActionHandler     types.PacketHandler[*types.ActionPacket]
		expError          string
	}{
		{
			name:              "success - passing all correct inputs",
			codec:             deps.EncCfg.Codec,
			sb:                collections.NewSchemaBuilder(deps.StoreService),
			logger:            deps.Logger,
			ForwardingHandler: &mocks.ForwardingHandler{},
			ActionHandler:     &mocks.ActionsHandler{},
			expError:          "",
		},
		{
			name:              "error - nil codec",
			sb:                collections.NewSchemaBuilder(deps.StoreService),
			logger:            deps.Logger,
			ForwardingHandler: &mocks.ForwardingHandler{},
			ActionHandler:     &mocks.ActionsHandler{},
			expError:          "codec cannot be nil",
		},
		{
			name:              "error - nil schema builder",
			codec:             deps.EncCfg.Codec,
			logger:            deps.Logger,
			ForwardingHandler: &mocks.ForwardingHandler{},
			ActionHandler:     &mocks.ActionsHandler{},
			expError:          "schema builder cannot be nil",
		},
		{
			name:              "error - nil logger",
			codec:             deps.EncCfg.Codec,
			sb:                collections.NewSchemaBuilder(deps.StoreService),
			ForwardingHandler: &mocks.ForwardingHandler{},
			ActionHandler:     &mocks.ActionsHandler{},
			expError:          "logger cannot be nil",
		},
		{
			name:              "error - nil forwarding handler",
			codec:             deps.EncCfg.Codec,
			sb:                collections.NewSchemaBuilder(deps.StoreService),
			logger:            deps.Logger,
			ForwardingHandler: nil,
			ActionHandler:     &mocks.ActionsHandler{},
			expError:          "forwarding handler is not set",
		},
		{
			name:              "error - nil actions handler",
			codec:             deps.EncCfg.Codec,
			sb:                collections.NewSchemaBuilder(deps.StoreService),
			logger:            deps.Logger,
			ForwardingHandler: &mocks.ForwardingHandler{},
			ActionHandler:     nil,
			expError:          "action handler is not set",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			_, err := dispatcher.New(
				tC.codec,
				tC.sb,
				tC.logger,
				tC.ForwardingHandler,
				tC.ActionHandler,
			)

			if tC.expError != "" {
				require.ErrorContains(t, err, tC.expError)
			} else {
				require.NoError(t, err)
				_, err = tC.sb.Build()
				require.NoError(t, err)
			}
		})
	}
}
