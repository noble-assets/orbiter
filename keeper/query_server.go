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
)

var _ orbitertypes.QueryServer = &queryServer{}

type queryServer struct {
	*Keeper
}

func NewQueryServer(k *Keeper) orbitertypes.QueryServer {
	return &queryServer{Keeper: k}
}

func (s queryServer) PendingPayload(
	ctx context.Context,
	req *orbitertypes.QueryPendingPayloadRequest,
) (*orbitertypes.QueryPendingPayloadResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.Hash) != orbitertypes.PayloadHashLength {
		return nil, status.Error(codes.InvalidArgument, "malformed hash")
	}

	payload, err := s.pendingPayload(ctx, req.Hash)
	if err != nil {
		return nil, status.Error(codes.NotFound, "payload not found")
	}

	return &orbitertypes.QueryPendingPayloadResponse{
		Payload: payload.Payload,
	}, nil
}
