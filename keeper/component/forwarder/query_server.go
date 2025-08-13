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
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	forwardertypes "orbiter.dev/types/component/forwarder"
	"orbiter.dev/types/core"
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
	if err := req.ProtocolId.Validate(); err != nil {
		return nil, fmt.Errorf("invalid protocol ID: %w", err)
	}

	paused, err := s.Forwarder.IsProtocolPaused(ctx, req.ProtocolId)
	if err != nil {
		return nil, fmt.Errorf("unable to query protocol paused status: %w", err)
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
		return nil, fmt.Errorf("unable to query paused protocols: %w", err)
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

	ccID, err := core.NewCrossChainID(req.ProtocolId, req.CounterpartyId)
	if err != nil {
		return nil, fmt.Errorf("unable to query cross-chain paused status: %w", err)
	}

	paused, err := s.Forwarder.IsCrossChainPaused(ctx, ccID)
	if err != nil {
		return nil, fmt.Errorf("unable to query cross-chain paused status: %w", err)
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

	id := req.ProtocolId
	if err := id.Validate(); err != nil {
		return nil, fmt.Errorf("invalid protocol ID: %w", err)
	}
	paused, err := s.GetPausedCrossChainsMap(ctx, &id)
	if err != nil {
		return nil, fmt.Errorf("unable to query paused counterparty: %w", err)
	}

	return &forwardertypes.QueryPausedCrossChainsResponse{
		CounterpartyIds: paused[int32(id)],
	}, nil
}
