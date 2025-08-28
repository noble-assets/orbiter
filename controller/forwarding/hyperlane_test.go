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

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
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

func TestNewHyperlaneController(t *testing.T) {
	testCases := []struct {
		name     string
		logger   log.Logger
		handler  forwardingtypes.HyperlaneHandler
		expError string
	}{
		{
			name:    "success - valid controller creation",
			logger:  log.NewNopLogger(),
			handler: &mocks.HyperlaneHandler{},
		},
		{
			name:     "error - nil logger",
			expError: "logger cannot be nil",
		},
		{
			name:     "error - when no Hyperlane handler is provided",
			logger:   log.NewNopLogger(),
			expError: "handler",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			controller, err := forwarding.NewHyperlaneController(
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

func TestExtractAttributes_Hyperlane(t *testing.T) {
	usdnID := make([]byte, 32)
	copy(usdnID, "usdn id")

	testCases := []struct {
		name       string
		forwarding func() *core.Forwarding
		expHypAttr forwardingtypes.HypAttributes
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
				f, err := forwardingtypes.NewHyperlaneForwarding(
					usdnID,
					0,
					make([]byte, 32),
					make([]byte, 32),
					"",
					math.NewInt(1),
					sdk.NewInt64Coin("usdn", 1),
					[]byte{},
				)
				require.NoError(t, err)

				return f
			},
			expHypAttr: forwardingtypes.HypAttributes{
				TokenId:           []byte("token id"),
				DestinationDomain: 0,
				Recipient:         make([]byte, 32),
				CustomHookId:      make([]byte, 32),
				GasLimit:          math.NewInt(1),
				MaxFee:            sdk.NewInt64Coin("usdn", 1),
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			controller, err := forwarding.NewHyperlaneController(
				log.NewNopLogger(),
				mocks.HyperlaneHandler{},
			)
			require.NoError(t, err)

			hypAttr, err := controller.ExtractAttributes(tC.forwarding())

			if tC.expError != "" {
				require.ErrorContains(t, err, tC.expError)
				require.Nil(t, hypAttr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, hypAttr)
			}
		})
	}
}

func TestValidateForwarding_Hyperlane(t *testing.T) {
	transferAttr, err := types.NewTransferAttributes(
		core.PROTOCOL_IBC,
		"channel-0",
		"usdn",
		math.NewInt(1),
	)
	require.NoError(t, err)

	usdnID := make([]byte, 32)
	copy(usdnID, "usdn id")

	testCases := []struct {
		name         string
		setup        func(*mocks.HyperlaneHandler)
		hypAttr      *forwardingtypes.HypAttributes
		transferAttr *types.TransferAttributes
		expError     string
	}{
		{
			name: "success - when the attributes are valid",
			setup: func(m *mocks.HyperlaneHandler) {
				tokenID := hyperlaneutil.HexAddress(usdnID)
				hypToken := &warptypes.WrappedHypToken{
					Id:            tokenID.String(),
					Owner:         "noble",
					TokenType:     0,
					OriginMailbox: "",
					OriginDenom:   "usdn",
					IsmId:         &hyperlaneutil.HexAddress{},
				}

				m.Tokens[hypToken.Id] = *hypToken
			},
			transferAttr: transferAttr,
			hypAttr: &forwardingtypes.HypAttributes{
				TokenId:           usdnID,
				DestinationDomain: 0,
				Recipient:         make([]byte, 32),
				CustomHookId:      make([]byte, 32),
				GasLimit:          math.NewInt(1),
				MaxFee:            sdk.NewInt64Coin("usdn", 1),
			},
		},
		{
			name: "error - when the hyperlane denom is not destination denom",
			setup: func(m *mocks.HyperlaneHandler) {
				tokenID := hyperlaneutil.HexAddress(usdnID)
				hypToken := &warptypes.WrappedHypToken{
					Id:            tokenID.String(),
					Owner:         "noble",
					TokenType:     0,
					OriginMailbox: "",
					OriginDenom:   "btc",
					IsmId:         &hyperlaneutil.HexAddress{},
				}

				m.Tokens[hypToken.Id] = *hypToken
			},
			transferAttr: transferAttr,
			hypAttr: &forwardingtypes.HypAttributes{
				TokenId:           usdnID,
				DestinationDomain: 0,
				Recipient:         make([]byte, 32),
				CustomHookId:      make([]byte, 32),
				GasLimit:          math.NewInt(1),
				MaxFee:            sdk.NewInt64Coin("usdn", 1),
			},
			expError: "invalid forwarding token",
		},
		{
			name:         "error - when hyperlane attributes are nil",
			setup:        func(m *mocks.HyperlaneHandler) {},
			transferAttr: transferAttr,
			hypAttr:      nil,
			expError:     "invalid Hyperlane attributes",
		},
		{
			name:         "error - when transfer attributes are nil",
			setup:        func(m *mocks.HyperlaneHandler) {},
			transferAttr: nil,
			hypAttr: &forwardingtypes.HypAttributes{
				TokenId:           usdnID,
				DestinationDomain: 0,
				Recipient:         make([]byte, 32),
				CustomHookId:      make([]byte, 32),
				GasLimit:          math.NewInt(1),
				MaxFee:            sdk.NewInt64Coin("usdn", 1),
			},
			expError: "invalid transfer attributes",
		},
		{
			name:         "error - when the token does not exist",
			setup:        func(m *mocks.HyperlaneHandler) {},
			transferAttr: transferAttr,
			hypAttr: &forwardingtypes.HypAttributes{
				TokenId:           usdnID,
				DestinationDomain: 0,
				Recipient:         make([]byte, 32),
				CustomHookId:      make([]byte, 32),
				GasLimit:          math.NewInt(1),
				MaxFee:            sdk.NewInt64Coin("usdn", 1),
			},
			expError: "invalid Hyperlane forwarding",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			handler := mocks.HyperlaneHandler{Tokens: make(map[string]warptypes.WrappedHypToken)}

			if tC.setup != nil {
				tC.setup(&handler)
			}

			controller, err := forwarding.NewHyperlaneController(
				log.NewNopLogger(),
				handler,
			)
			require.NoError(t, err)

			err = controller.ValidateForwarding(
				context.Background(),
				tC.transferAttr,
				tC.hypAttr,
			)

			if tC.expError != "" {
				require.ErrorContains(t, err, tC.expError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandlePacket_Hyperlane(t *testing.T) {
	usdnID := make([]byte, 32)
	copy(usdnID, "usdn id")

	testCases := []struct {
		name     string
		setup    func(*mocks.HyperlaneHandler) context.Context
		packet   func() *types.ForwardingPacket
		expError string
	}{
		{
			name:     "error - when the attributes are valid",
			packet:   func() *types.ForwardingPacket { return nil },
			expError: "Hyperlane controller received nil packet",
		},
		{
			name: "error - when forwarding is default values",
			packet: func() *types.ForwardingPacket {
				a, err := types.NewTransferAttributes(
					core.PROTOCOL_IBC,
					"channel-0",
					"usdn",
					math.NewInt(1),
				)
				require.NoError(t, err)

				return &types.ForwardingPacket{
					TransferAttributes: a,
					Forwarding:         &core.Forwarding{},
				}
			},
			expError: "error extracting Hyperlane forwarding",
		},
		{
			name: "error - when transfer attributes are default values",
			packet: func() *types.ForwardingPacket {
				f, err := forwardingtypes.NewHyperlaneForwarding(
					usdnID,
					0,
					make([]byte, 32),
					make([]byte, 32),
					"",
					math.ZeroInt(),
					sdk.NewCoin("usdn", math.ZeroInt()),
					[]byte{},
				)
				require.NoError(t, err)

				return &types.ForwardingPacket{
					TransferAttributes: &types.TransferAttributes{},
					Forwarding:         f,
				}
			},
			expError: "error validating Hyperlane forwarding",
		},
		{
			name: "error - when the remote transfer fails",
			setup: func(m *mocks.HyperlaneHandler) context.Context {
				tokenID := hyperlaneutil.HexAddress(usdnID)
				hypToken := &warptypes.WrappedHypToken{
					Id:            tokenID.String(),
					Owner:         "noble",
					TokenType:     0,
					OriginMailbox: "",
					OriginDenom:   "usdn",
					IsmId:         &hyperlaneutil.HexAddress{},
				}

				m.Tokens[hypToken.Id] = *hypToken

				return context.WithValue(context.Background(), mocks.FailingContextKey, true)
			},
			packet: func() *types.ForwardingPacket {
				a, err := types.NewTransferAttributes(
					core.PROTOCOL_IBC,
					"channel-0",
					"usdn",
					math.NewInt(1),
				)
				require.NoError(t, err)

				f, err := forwardingtypes.NewHyperlaneForwarding(
					usdnID,
					0,
					make([]byte, 32),
					make([]byte, 32),
					"",
					math.ZeroInt(),
					sdk.NewCoin("usdn", math.ZeroInt()),
					[]byte{},
				)
				require.NoError(t, err)

				return &types.ForwardingPacket{
					TransferAttributes: a,
					Forwarding:         f,
				}
			},
			expError: "error executing Hyperlane forwarding",
		},
		{
			name: "success - when the packet is valid",
			setup: func(m *mocks.HyperlaneHandler) context.Context {
				tokenID := hyperlaneutil.HexAddress(usdnID)
				hypToken := &warptypes.WrappedHypToken{
					Id:            tokenID.String(),
					Owner:         "noble",
					TokenType:     0,
					OriginMailbox: "",
					OriginDenom:   "usdn",
					IsmId:         &hyperlaneutil.HexAddress{},
				}

				m.Tokens[hypToken.Id] = *hypToken

				return context.Background()
			},
			packet: func() *types.ForwardingPacket {
				a, err := types.NewTransferAttributes(
					core.PROTOCOL_IBC,
					"channel-0",
					"usdn",
					math.NewInt(1),
				)
				require.NoError(t, err)

				f, err := forwardingtypes.NewHyperlaneForwarding(
					usdnID,
					0,
					make([]byte, 32),
					make([]byte, 32),
					"",
					math.ZeroInt(),
					sdk.NewCoin("usdn", math.ZeroInt()),
					[]byte{},
				)
				require.NoError(t, err)

				return &types.ForwardingPacket{
					TransferAttributes: a,
					Forwarding:         f,
				}
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			handler := mocks.HyperlaneHandler{Tokens: make(map[string]warptypes.WrappedHypToken)}

			ctx := context.Background()
			if tC.setup != nil {
				ctx = tC.setup(&handler)
			}

			controller, err := forwarding.NewHyperlaneController(
				log.NewNopLogger(),
				handler,
			)
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
