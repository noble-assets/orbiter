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
	"fmt"

	"orbiter.dev/types/component/forwarder"
	"orbiter.dev/types/core"
)

var _ forwarder.QueryServer = &queryServerForwarder{}

type queryServerForwarder struct {
	*Keeper
}

func NewQueryServerForwarder(keeper *Keeper) queryServerForwarder {
	return queryServerForwarder{Keeper: keeper}
}

// IsProtocolPaused implements forwarder.QueryServer.
func (s queryServerForwarder) IsProtocolPaused(
	ctx context.Context,
	req *forwarder.QueryIsProtocolPausedRequest,
) (*forwarder.QueryIsProtocolPausedResponse, error) {
	f := s.Forwarder()

	paused, err := f.IsProtocolPaused(ctx, req.ProtocolId)
	if err != nil {
		return nil, fmt.Errorf("unable to query protocol paused status: %w", err)
	}

	return &forwarder.QueryIsProtocolPausedResponse{
		IsPaused: paused,
	}, nil
}

// PausedProtocols implements forwarder.QueryServer.
func (s queryServerForwarder) PausedProtocols(
	ctx context.Context,
	req *forwarder.QueryPausedProtocolsRequest,
) (*forwarder.QueryPausedProtocolsResponse, error) {
	f := s.Forwarder()

	paused, err := f.GetPausedProtocols(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to query paused protocols: %w", err)
	}

	return &forwarder.QueryPausedProtocolsResponse{
		ProtocolId: paused,
	}, nil
}

// IsCounterpartyPaused implements forwarder.QueryServer.
func (s queryServerForwarder) IsCounterpartyPaused(
	ctx context.Context,
	req *forwarder.QueryIsCounterpartyPausedRequest,
) (*forwarder.QueryIsCounterpartyPausedResponse, error) {
	f := s.Forwarder()

	orbitID, err := core.NewOrbitID(req.ProtocolId, req.CounterpartyId)
	if err != nil {
		return nil, fmt.Errorf("unable to query counterparty paused status: %w", err)
	}

	paused, err := f.IsProtocolCounterpartyPaused(ctx, orbitID)
	if err != nil {
		return nil, fmt.Errorf("unable to query counterparty paused status: %w", err)
	}

	return &forwarder.QueryIsCounterpartyPausedResponse{
		IsPaused: paused,
	}, nil
}

// PausedCounterparties implements forwarder.QueryServer.
func (s queryServerForwarder) PausedCounterparties(
	ctx context.Context,
	req *forwarder.QueryPausedCounterpartiesRequest,
) (*forwarder.QueryPausedCounterpartiesResponse, error) {
	f := s.Forwarder()

	paused, err := f.GetPausedOrbits(ctx, req.ProtocolId)
	if err != nil {
		return nil, fmt.Errorf("unable to query paused counterparty: %w", err)
	}

	return &forwarder.QueryPausedCounterpartiesResponse{
		CounterpartyIds: paused,
	}, nil
}
