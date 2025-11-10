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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	executortypes "github.com/noble-assets/orbiter/v2/types/component/executor"
	"github.com/noble-assets/orbiter/v2/types/core"
)

var _ executortypes.QueryServer = &queryServer{}

type queryServer struct {
	*Executor
}

func NewQueryServer(e *Executor) queryServer {
	return queryServer{Executor: e}
}

// IsActionPaused implements executor.QueryServer.
func (s queryServer) IsActionPaused(
	ctx context.Context,
	req *executortypes.QueryIsActionPausedRequest,
) (*executortypes.QueryIsActionPausedResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	actionID, err := core.NewActionIDFromString(req.ActionId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	paused, err := s.Executor.IsActionPaused(ctx, actionID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &executortypes.QueryIsActionPausedResponse{
		IsPaused: paused,
	}, nil
}

// PausedActions implements executor.QueryServer.
func (s queryServer) PausedActions(
	ctx context.Context,
	req *executortypes.QueryPausedActionsRequest,
) (*executortypes.QueryPausedActionsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	paused, err := s.GetPausedActions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &executortypes.QueryPausedActionsResponse{
		ActionIds: paused,
	}, nil
}
