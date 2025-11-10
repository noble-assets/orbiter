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

package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var _ orbitertypes.MsgServer = &msgServer{}

// msgServer is the main message handler for the Orbiter.
type msgServer struct {
	*Keeper
}

// NewMsgServer returns a new Orbiter message server.
func NewMsgServer(keeper *Keeper) orbitertypes.MsgServer {
	return &msgServer{keeper}
}

func (s *msgServer) SubmitPayload(
	ctx context.Context,
	req *orbitertypes.MsgSubmitPayload,
) (*orbitertypes.MsgSubmitPayloadResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var payload core.Payload
	if err := orbitertypes.UnmarshalJSON(s.cdc, []byte(req.Payload), &payload); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal payload: %s", err)
	}

	if err := payload.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid payload: %s", err)
	}

	if err := s.validatePayloadAgainstState(ctx, &payload); err != nil {
		return nil, status.Errorf(
			codes.FailedPrecondition,
			"payload failed stateful checks: %s",
			err,
		)
	}

	payloadHash, err := s.submit(
		ctx,
		&payload,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &orbitertypes.MsgSubmitPayloadResponse{
		Hash: payloadHash.String(),
	}, nil
}
