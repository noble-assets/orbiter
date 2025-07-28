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

package orbits

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"orbiter.dev/controllers"
	"orbiter.dev/testutil/mocks"
	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
	"orbiter.dev/types/controllers/orbits"
)

func TestNewCCTPController(t *testing.T) {
	testCases := []struct {
		name      string
		logger    log.Logger
		msgServer orbits.CCTPMsgServer
		expError  string
	}{
		{
			name:      "success - valid controller creation",
			logger:    log.NewNopLogger(),
			msgServer: &mocks.CCTPMsgServer{},
		},
		{
			name:     "error - nil logger",
			expError: "logger cannot be nil",
		},
		{
			name:     "error - when no CCTP server is provided",
			logger:   log.NewNopLogger(),
			expError: types.ErrNilPointer.Error(),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			controller, err := NewCCTPController(
				tC.logger,
				tC.msgServer,
			)

			if tC.expError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tC.expError)
				require.Nil(t, controller)
			} else {
				require.NoError(t, err)
				require.NotNil(t, controller)
			}
		})
	}
}

func TestHandlePacket(t *testing.T) {
	transferAttr, err := types.NewTransferAttributes(
		types.PROTOCOL_IBC,
		"channel-01",
		"uusdc",
		math.NewInt(1_000_000),
	)
	require.NoError(t, err)
	testCases := []struct {
		name     string
		setup    func() context.Context
		packet   func() *types.OrbitPacket
		expError string
	}{
		{
			name: "success - valid packet processing",
			packet: func() *types.OrbitPacket {
				orbit, err := orbits.NewCCTPOrbit(
					1,
					[]byte("recipient"),
					[]byte("caller"),
					[]byte(""),
				)
				require.NoError(t, err)
				return &types.OrbitPacket{
					Orbit:              orbit,
					TransferAttributes: transferAttr,
				}
			},
		},
		{
			name: "error - cctp server returns an error",
			setup: func() context.Context {
				return context.WithValue(context.Background(), "failing", true) //nolint:staticcheck
			},
			packet: func() *types.OrbitPacket {
				orbit, err := orbits.NewCCTPOrbit(
					1,
					[]byte("recipient"),
					[]byte("caller"),
					[]byte(""),
				)
				require.NoError(t, err)

				return &types.OrbitPacket{
					Orbit:              orbit,
					TransferAttributes: transferAttr,
				}
			},
			expError: types.ErrControllerExecution.Error(),
		},
	}

	logger := log.NewNopLogger()
	controller, err := NewCCTPController(
		logger,
		&mocks.CCTPMsgServer{},
	)
	require.NoError(t, err)
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			ctx := context.Background()

			if tC.setup != nil {
				ctx = tC.setup()
			}

			err = controller.HandlePacket(ctx, tC.packet())

			if tC.expError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tC.expError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExtractAttributes(t *testing.T) {
	testCases := []struct {
		name          string
		orbit         func() *types.Orbit
		expAttributes *orbits.CCTPAttributes
		expError      string
	}{
		{
			name: "success - valid attributes",
			orbit: func() *types.Orbit {
				attr, err := orbits.NewCCTPAttributes(1, []byte("recipient"), []byte("caller"))
				require.NoError(t, err)
				route := &types.Orbit{
					ProtocolId: types.PROTOCOL_CCTP,
				}
				err = route.SetAttributes(attr)
				require.NoError(t, err)

				return route
			},
		},
		{
			name: "error - wrong attributes",
			orbit: func() *types.Orbit {
				invalidAttr := testdata.TestOrbitAttr{}
				route := &types.Orbit{
					ProtocolId: types.PROTOCOL_CCTP,
				}
				err := route.SetAttributes(&invalidAttr)
				require.NoError(t, err)

				return route
			},
			expError: "expected *orbits.CCTPAttributes",
		},
		{
			name: "error - empty attributes",
			orbit: func() *types.Orbit {
				return &types.Orbit{
					ProtocolId: types.PROTOCOL_CCTP,
				}
			},
			expError: types.ErrNilPointer.Error(),
		},
	}

	id := types.PROTOCOL_CCTP
	baseController, err := controllers.NewBaseController(id)
	require.NoError(t, err)

	handler, err := newCCTPHandler(&mocks.CCTPMsgServer{})
	require.NoError(t, err)
	controller := CCTPController{
		logger:         log.NewNopLogger(),
		BaseController: baseController,
		handler:        handler,
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			attributes, err := controller.extractAttributes(tC.orbit())

			if tC.expError != "" {
				require.Error(t, err)
				require.Nil(t, attributes)
				require.Contains(t, err.Error(), tC.expError)
			} else {
				require.NoError(t, err)
				require.NotNil(t, attributes)
				require.Equal(t, uint32(1), attributes.DestinationDomain)
			}
		})
	}
}

func TestNewCCTPHandler(t *testing.T) {
	testCases := []struct {
		name      string
		msgServer orbits.CCTPMsgServer
		expErr    string
	}{
		{
			name:      "success - valid controller creation",
			msgServer: &mocks.CCTPMsgServer{},
		},
		{
			name:      "error - nil msg server",
			msgServer: nil,
			expErr:    "cannot be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler, err := newCCTPHandler(
				tc.msgServer,
			)

			if tc.expErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)
			}
		})
	}
}
