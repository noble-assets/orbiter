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
	"context"

	"orbiter.dev/types"
	executortypes "orbiter.dev/types/component/executor"
	"orbiter.dev/types/core"
)

var _ executortypes.MsgServer = &msgServer{}

// msgServer is the server used to handle messages
// for the executor component.
type msgServer struct {
	*Executor
	types.Authorizer
}

func NewMsgServer(e *Executor, a types.Authorizer) msgServer {
	return msgServer{Executor: e, Authorizer: a}
}

// Pause implements executor.MsgServer.
func (s msgServer) PauseAction(
	ctx context.Context,
	msg *executortypes.MsgPauseAction,
) (*executortypes.MsgPauseActionResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	if err := s.Pause(ctx, msg.ActionId); err != nil {
		return nil, core.ErrUnableToPause.Wrapf(
			"action: %s", err.Error(),
		)
	}

	return &executortypes.MsgPauseActionResponse{}, nil
}

// Unpause implements executor.MsgServer.
func (s msgServer) UnpauseAction(
	ctx context.Context,
	msg *executortypes.MsgUnpauseAction,
) (*executortypes.MsgUnpauseActionResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	if err := s.Unpause(ctx, msg.ActionId); err != nil {
		return nil, core.ErrUnableToUnpause.Wrapf(
			"action: %s", err.Error(),
		)
	}

	return &executortypes.MsgUnpauseActionResponse{}, nil
}
