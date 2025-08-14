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

package dispatcher

import (
	"context"

	"cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	dispatchertypes "orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

var _ dispatchertypes.QueryServer = &queryServer{}

type queryServer struct {
	*Dispatcher
}

func NewQueryServer(d *Dispatcher) dispatchertypes.QueryServer {
	return queryServer{Dispatcher: d}
}

func (q queryServer) DispatchedCounts(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedCountsRequest,
) (*dispatchertypes.QueryDispatchedCountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	sourceID, err := core.NewCrossChainID(req.SourceProtocolId, req.SourceCounterpartyId)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating source cross-chain ID")
	}

	destID, err := core.NewCrossChainID(req.DestinationProtocolId, req.DestinationCounterpartyId)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating destination cross-chain ID")
	}

	counts := q.GetDispatchedCounts(ctx, &sourceID, &destID)

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts: []*dispatchertypes.DispatchCountEntry{counts},
	}, nil
}

func (q queryServer) DispatchedCountsByDestinationProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedCountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedCountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	if err := req.ProtocolId.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid protocol ID")
	}

	counts := q.GetDispatchedCountsByDestinationProtocolID(ctx, req.ProtocolId)

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts: counts,
	}, nil
}

func (q queryServer) DispatchedCountsBySourceProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedCountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedCountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	if err := req.ProtocolId.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid protocol ID")
	}

	counts := q.GetDispatchedCountsBySourceProtocolID(ctx, req.ProtocolId)

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts: counts,
	}, nil
}

// DispatchedAmounts implements dispatcher.QueryServer.
func (q queryServer) DispatchedAmounts(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedAmountsRequest,
) (*dispatchertypes.QueryDispatchedAmountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	sourceID, err := core.NewCrossChainID(req.SourceProtocolId, req.SourceCounterpartyId)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating source cross-chain ID")
	}

	destID, err := core.NewCrossChainID(req.DestinationProtocolId, req.DestinationCounterpartyId)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating destination cross-chain ID")
	}

	amounts := q.GetDispatchedAmount(ctx, &sourceID, &destID, req.Denom)

	return &dispatchertypes.QueryDispatchedAmountsResponse{
		Amounts: []*dispatchertypes.DispatchedAmountEntry{amounts},
	}, nil
}

// DispatchedAmountsByDestinationProtocolID implements dispatcher.QueryServer.
func (q queryServer) DispatchedAmountsByDestinationProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedAmountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedAmountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	if err := req.ProtocolId.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid protocol ID")
	}

	amounts := q.GetDispatchedAmountsBySourceProtocolID(ctx, req.ProtocolId)

	return &dispatchertypes.QueryDispatchedAmountsResponse{
		Amounts: amounts,
	}, nil
}

// DispatchedAmountsBySourceProtocolID implements dispatcher.QueryServer.
func (q queryServer) DispatchedAmountsBySourceProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedAmountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedAmountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	if err := req.ProtocolId.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid protocol ID")
	}

	amounts := q.GetDispatchedAmountsByDestinationProtocolID(ctx, req.ProtocolId)

	return &dispatchertypes.QueryDispatchedAmountsResponse{
		Amounts: amounts,
	}, nil
}
