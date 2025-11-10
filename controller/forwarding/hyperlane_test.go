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

	"github.com/noble-assets/orbiter/v2/controller/forwarding"
	"github.com/noble-assets/orbiter/v2/testutil"
	"github.com/noble-assets/orbiter/v2/testutil/mocks"
	"github.com/noble-assets/orbiter/v2/types"
	forwardingtypes "github.com/noble-assets/orbiter/v2/types/controller/forwarding"
	"github.com/noble-assets/orbiter/v2/types/core"
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
				TokenId:           usdnID,
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
				require.Equal(t, tC.expHypAttr.TokenId, hypAttr.TokenId)
				require.Equal(t, tC.expHypAttr.DestinationDomain, hypAttr.DestinationDomain)
				require.Equal(t, len(tC.expHypAttr.Recipient), len(hypAttr.Recipient))
				require.Equal(t, len(tC.expHypAttr.CustomHookId), len(hypAttr.CustomHookId))
				require.True(t, tC.expHypAttr.GasLimit.Equal(hypAttr.GasLimit))
				require.Equal(t, tC.expHypAttr.MaxFee, hypAttr.MaxFee)
			}
		})
	}
}

func TestValidateForwarding_Hyperlane(t *testing.T) {
	transferAttr, err := core.NewTransferAttributes(
		core.PROTOCOL_IBC,
		"channel-0",
		"usdn",
		math.NewInt(1),
	)
	require.NoError(t, err)

	usdnID := make([]byte, 32)
	copy(usdnID, "usdn id")

	validHypAttr := forwardingtypes.HypAttributes{
		TokenId:           usdnID,
		DestinationDomain: 0,
		Recipient:         make([]byte, 32), // we don't check the content but only the length
		CustomHookId:      make([]byte, 32), // we don't check the content but only the length
		GasLimit:          math.NewInt(1),
		MaxFee:            sdk.NewInt64Coin("usdn", 1),
	}

	tokenID := hyperlaneutil.HexAddress(usdnID)
	hypToken := warptypes.WrappedHypToken{
		Id:            tokenID.String(),
		Owner:         "noble",
		TokenType:     0,
		OriginMailbox: "",
		OriginDenom:   "usdn",
		IsmId:         &hyperlaneutil.HexAddress{},
	}

	testCases := []struct {
		name         string
		setup        func(*mocks.HyperlaneHandler)
		hypAttr      *forwardingtypes.HypAttributes
		transferAttr *core.TransferAttributes
		expError     string
	}{
		{
			name: "success - when the attributes are valid",
			setup: func(m *mocks.HyperlaneHandler) {
				m.Tokens[hypToken.Id] = hypToken
			},
			transferAttr: transferAttr,
			hypAttr:      &validHypAttr,
		},
		{
			name: "error - when the hyperlane denom is not destination denom",
			setup: func(m *mocks.HyperlaneHandler) {
				invalidHypToken := hypToken
				invalidHypToken.OriginDenom = "btc"

				m.Tokens[hypToken.Id] = invalidHypToken
			},
			transferAttr: transferAttr,
			hypAttr:      &validHypAttr,
			expError:     "invalid forwarding token",
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
			hypAttr:      &validHypAttr,
			expError:     "invalid transfer attributes",
		},
		{
			name:         "error - when the token does not exist",
			setup:        func(m *mocks.HyperlaneHandler) {},
			transferAttr: transferAttr,
			hypAttr:      &validHypAttr,
			expError:     "invalid Hyperlane forwarding",
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

	tokenID := hyperlaneutil.HexAddress(usdnID)
	hypToken := warptypes.WrappedHypToken{
		Id:            tokenID.String(),
		Owner:         "noble",
		TokenType:     0,
		OriginMailbox: "",
		OriginDenom:   "usdn",
		IsmId:         &hyperlaneutil.HexAddress{},
	}

	validForwarding, err := forwardingtypes.NewHyperlaneForwarding(
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

	validTransfer, err := core.NewTransferAttributes(
		core.PROTOCOL_IBC,
		"channel-0",
		"usdn",
		math.NewInt(1),
	)
	require.NoError(t, err)

	testCases := []struct {
		name     string
		setup    func(*mocks.HyperlaneHandler) context.Context
		packet   func() *types.ForwardingPacket
		expError string
	}{
		{
			name:     "error - when forwarding packet is nil",
			packet:   func() *types.ForwardingPacket { return nil },
			expError: "Hyperlane controller received nil packet",
		},
		{
			name: "error - when forwarding is default values",
			packet: func() *types.ForwardingPacket {
				return &types.ForwardingPacket{
					TransferAttributes: validTransfer,
					Forwarding:         &core.Forwarding{},
				}
			},
			expError: "error extracting Hyperlane forwarding",
		},
		{
			name: "error - when transfer attributes are default values",
			packet: func() *types.ForwardingPacket {
				return &types.ForwardingPacket{
					TransferAttributes: &core.TransferAttributes{},
					Forwarding:         validForwarding,
				}
			},
			expError: "invalid Hyperlane forwarding",
		},
		{
			name: "error - when the remote transfer fails",
			setup: func(m *mocks.HyperlaneHandler) context.Context {
				m.Tokens[hypToken.Id] = hypToken

				return context.WithValue(context.Background(), mocks.FailingContextKey, true)
			},
			packet: func() *types.ForwardingPacket {
				return &types.ForwardingPacket{
					TransferAttributes: validTransfer,
					Forwarding:         validForwarding,
				}
			},
			expError: "error executing Hyperlane forwarding",
		},
		{
			name: "success - when the packet is valid",
			setup: func(m *mocks.HyperlaneHandler) context.Context {
				m.Tokens[hypToken.Id] = hypToken

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
