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

package forwarding_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/controller/forwarding"
	"github.com/noble-assets/orbiter/testutil"
	"github.com/noble-assets/orbiter/testutil/mocks"
	"github.com/noble-assets/orbiter/types"
	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

func TestNewInternalController(t *testing.T) {
	testCases := []struct {
		name     string
		logger   log.Logger
		handler  forwardingtypes.InternalHandler
		expError string
	}{
		{
			name:    "success - valid controller creation",
			logger:  log.NewNopLogger(),
			handler: mocks.NewInternalHandler(),
		},
		{
			name:     "error - nil logger",
			expError: "logger cannot be nil",
		},
		{
			name:     "error - when no internal handler is provided",
			logger:   log.NewNopLogger(),
			expError: "handler",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			controller, err := forwarding.NewInternalController(
				tC.logger,
				tC.handler,
			)

			if tC.expError != "" {
				require.ErrorContains(t, err, tC.expError)
				require.Nil(t, controller)
			} else {
				require.NoError(t, err)
				require.NotNil(t, controller)
			}
		})
	}
}

func TestExtractAttribuetes_Internal(t *testing.T) {
	nobleAddr := testutil.NewNobleAddress()

	testCases := []struct {
		name       string
		forwarding func() *core.Forwarding
		expAttr    forwardingtypes.InternalAttributes
		expError   string
	}{
		{
			name:       "error - nil forwarding",
			forwarding: func() *core.Forwarding { return nil },
			expError:   "forwarding is not set",
		},
		{
			name: "error - CCTP forwarding",
			forwarding: func() *core.Forwarding {
				f, err := forwardingtypes.NewCCTPForwarding(
					0,
					testutil.RandomBytes(32),
					testutil.RandomBytes(32),
					[]byte{},
				)

				require.NoError(t, err)

				return f
			},
			expError: "invalid type",
		},
		{
			name: "success - valid forwarding",
			forwarding: func() *core.Forwarding {
				f, err := forwardingtypes.NewInternalForwarding(nobleAddr)
				require.NoError(t, err)

				return f
			},
			expAttr: forwardingtypes.InternalAttributes{
				Recipient: nobleAddr,
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			controller, err := forwarding.NewInternalController(
				log.NewNopLogger(),
				mocks.NewInternalHandler(),
			)
			require.NoError(t, err)

			intAttr, err := controller.ExtractAttributes(tC.forwarding())

			if tC.expError != "" {
				require.ErrorContains(t, err, tC.expError)
				require.Nil(t, intAttr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, intAttr)
			}
		})
	}
}

func TestHandlePacket_Internal(t *testing.T) {
	nobleAddr := testutil.NewNobleAddress()

	validForwarding, err := forwardingtypes.NewInternalForwarding(nobleAddr)
	require.NoError(t, err)

	validTransfer, err := types.NewTransferAttributes(
		core.PROTOCOL_IBC,
		"channel-0",
		"usdn",
		math.NewInt(1),
	)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		setup    func(*mocks.InternalHandler) context.Context
		packet   func() *types.ForwardingPacket
		expError string
	}{
		{
			name:     "error - when the forwarding packet is nil",
			packet:   func() *types.ForwardingPacket { return nil },
			expError: "internal controller received nil packet",
		},
		{
			name: "error - when forwarding is default values",
			packet: func() *types.ForwardingPacket {
				return &types.ForwardingPacket{
					TransferAttributes: validTransfer,
					Forwarding:         &core.Forwarding{},
				}
			},
			expError: "error extracting internal forwarding",
		},
		{
			name: "error - when transfer attributes are default values",
			packet: func() *types.ForwardingPacket {
				return &types.ForwardingPacket{
					TransferAttributes: &types.TransferAttributes{},
					Forwarding:         validForwarding,
				}
			},
			expError: "error validating internal forwarding",
		},
		{
			name: "error - when the handler fails",
			packet: func() *types.ForwardingPacket {
				return &types.ForwardingPacket{
					TransferAttributes: validTransfer,
					Forwarding:         validForwarding,
				}
			},
			expError: "not enough balance",
		},
		{
			name: "success",
			setup: func(m *mocks.InternalHandler) context.Context {
				m.Balances[core.ModuleAddress.String()] = sdk.Coins{
					sdk.NewCoin(
						validTransfer.DestinationDenom(),
						validTransfer.DestinationAmount(),
					),
				}
				return context.Background()
			},
			packet: func() *types.ForwardingPacket {
				return &types.ForwardingPacket{
					TransferAttributes: validTransfer,
					Forwarding:         validForwarding,
				}
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			handler := mocks.NewInternalHandler()

			ctx := context.Background()
			if tC.setup != nil {
				ctx = tC.setup(handler)
			}

			controller, err := forwarding.NewInternalController(log.NewNopLogger(), handler)
			require.NoError(t, err)

			err = controller.HandlePacket(ctx, tC.packet())

			if tC.expError != "" {
				require.ErrorContains(t, err, tC.expError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
