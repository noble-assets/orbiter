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

package mocks

import (
	"context"
	"errors"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"

	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/controller/forwarding"
)

var _ forwarding.HyperlaneHandler = HyperlaneHandler{}

type HyperlaneHandler struct {
	Tokens map[string]warptypes.WrappedHypToken
}

// RemoteTransfer implements forwarding.HyperlaneHandler.
func (h HyperlaneHandler) RemoteTransfer(
	ctx context.Context,
	msg *warptypes.MsgRemoteTransfer,
) (*warptypes.MsgRemoteTransferResponse, error) {
	if CheckIfFailing(ctx) {
		return nil, errors.New("error execuring remote transfer")
	}

	return &warptypes.MsgRemoteTransferResponse{
		MessageId: hyperlaneutil.HexAddress{},
	}, nil
}

// Token implements forwarding.HyperlaneHandler.
func (h HyperlaneHandler) Token(
	ctx context.Context,
	request *warptypes.QueryTokenRequest,
) (*warptypes.QueryTokenResponse, error) {
	t, found := h.Tokens[request.Id]
	if !found {
		return nil, errors.New("token does not exist")
	}

	return &warptypes.QueryTokenResponse{
		Token: &t,
	}, nil
}

var _ types.HyperlaneCoreKeeper = HyperlaneCoreKeeper{}

type HyperlaneCoreKeeper struct {
	appRouter *hyperlaneutil.Router[hyperlaneutil.HyperlaneApp]
}

func (hck HyperlaneCoreKeeper) AppRouter() *hyperlaneutil.Router[hyperlaneutil.HyperlaneApp] {
	return hck.appRouter
}
