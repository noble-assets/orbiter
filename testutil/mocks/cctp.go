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

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"

	"github.com/noble-assets/orbiter/v2/types/controller/forwarding"
	"github.com/noble-assets/orbiter/v2/types/entrypoint"
)

var (
	_ forwarding.CCTPMsgServer = CCTPServer{}
	_ entrypoint.CCTPHandler   = CCTPServer{}
)

func NewCCTPServer() *CCTPServer {
	return &CCTPServer{}
}

type CCTPServer struct{}

func (c CCTPServer) DepositForBurn(
	ctx context.Context,
	msg *cctptypes.MsgDepositForBurn,
) (*cctptypes.MsgDepositForBurnResponse, error) {
	if CheckIfFailing(ctx) {
		return nil, errors.New("error calling deposit for burn api")
	}

	return &cctptypes.MsgDepositForBurnResponse{}, nil
}

func (c CCTPServer) DepositForBurnWithCaller(
	ctx context.Context,
	msg *cctptypes.MsgDepositForBurnWithCaller,
) (*cctptypes.MsgDepositForBurnWithCallerResponse, error) {
	if CheckIfFailing(ctx) {
		return nil, errors.New("error calling deposit for burn with caller api")
	}

	return &cctptypes.MsgDepositForBurnWithCallerResponse{}, nil
}

// ReplaceDepositForBurn implements forwarding.CCTPMsgServer.
func (c CCTPServer) ReplaceDepositForBurn(
	ctx context.Context,
	msg *cctptypes.MsgReplaceDepositForBurn,
) (*cctptypes.MsgReplaceDepositForBurnResponse, error) {
	if CheckIfFailing(ctx) {
		return nil, errors.New("error calling replace deposit for burn")
	}

	return &cctptypes.MsgReplaceDepositForBurnResponse{}, nil
}

// GetTokenPair implements entrypoint.CCTPHandler.
func (c CCTPServer) GetTokenPair(
	ctx context.Context,
	_ uint32,
	_ []byte,
) (cctptypes.TokenPair, bool) {
	if CheckIfFailing(ctx) {
		return cctptypes.TokenPair{}, false
	}

	// We just need the local token for the purpose of testing the CCTP entrypoint.
	return cctptypes.TokenPair{
		RemoteDomain: 0,
		RemoteToken:  []byte("usdc"),
		LocalToken:   "usdc",
	}, true
}

// ReceiveMessage implements entrypoint.CCTPHandler.
func (c CCTPServer) ReceiveMessage(
	ctx context.Context,
	_ *cctptypes.MsgReceiveMessage,
) (*cctptypes.MsgReceiveMessageResponse, error) {
	if CheckIfFailing(ctx) {
		return nil, errors.New("error calling receive message")
	}
	return &cctptypes.MsgReceiveMessageResponse{}, nil
}
