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

package forwarder

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	forwardertypes "github.com/noble-assets/orbiter/v2/types/component/forwarder"
	"github.com/noble-assets/orbiter/v2/types/core"
)

var _ forwardertypes.QueryServer = &queryServer{}

type queryServer struct {
	*Forwarder
}

func NewQueryServer(f *Forwarder) queryServer {
	return queryServer{Forwarder: f}
}

// IsProtocolPaused implements forwarder.QueryServer.
func (s queryServer) IsProtocolPaused(
	ctx context.Context,
	req *forwardertypes.QueryIsProtocolPausedRequest,
) (*forwardertypes.QueryIsProtocolPausedResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	protocolID, err := core.NewProtocolIDFromString(req.ProtocolId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	paused, err := s.Forwarder.IsProtocolPaused(ctx, protocolID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forwardertypes.QueryIsProtocolPausedResponse{
		IsPaused: paused,
	}, nil
}

// PausedProtocols implements forwarder.QueryServer.
func (s queryServer) PausedProtocols(
	ctx context.Context,
	req *forwardertypes.QueryPausedProtocolsRequest,
) (*forwardertypes.QueryPausedProtocolsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	paused, err := s.GetPausedProtocols(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forwardertypes.QueryPausedProtocolsResponse{
		ProtocolIds: paused,
	}, nil
}

func (s queryServer) IsCrossChainPaused(
	ctx context.Context,
	req *forwardertypes.QueryIsCrossChainPausedRequest,
) (*forwardertypes.QueryIsCrossChainPausedResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	protocolID, err := core.NewProtocolIDFromString(req.ProtocolId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ccID, err := core.NewCrossChainID(protocolID, req.CounterpartyId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	paused, err := s.Forwarder.IsCrossChainPaused(ctx, ccID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forwardertypes.QueryIsCrossChainPausedResponse{
		IsPaused: paused,
	}, nil
}

// PausedCrossChains implements forwarder.QueryServer.
func (s queryServer) PausedCrossChains(
	ctx context.Context,
	req *forwardertypes.QueryPausedCrossChainsRequest,
) (*forwardertypes.QueryPausedCrossChainsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	protocolID, err := core.NewProtocolIDFromString(req.ProtocolId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	counterparties, pageResp, err := s.GetPaginatedPausedCrossChains(
		ctx,
		protocolID,
		req.Pagination,
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &forwardertypes.QueryPausedCrossChainsResponse{
		CounterpartyIds: counterparties,
		Pagination:      pageResp,
	}, nil
}
