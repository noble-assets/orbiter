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
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	dispatchertypes "github.com/noble-assets/orbiter/types/component/dispatcher"
	"github.com/noble-assets/orbiter/types/core"
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

	val, exists := core.ProtocolID_value[req.SourceProtocolId]
	if !exists {
		return nil, fmt.Errorf(
			"source protocol with ID %s does not exist",
			req.SourceProtocolId,
		)
	}
	sourceProtocolID := core.ProtocolID(val)

	sourceID, err := core.NewCrossChainID(sourceProtocolID, req.SourceCounterpartyId)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error creating source cross-chain ID")
	}

	val, exists = core.ProtocolID_value[req.DestinationProtocolId]
	if !exists {
		return nil, fmt.Errorf(
			"destination protocol with ID %s does not exist",
			req.DestinationProtocolId,
		)
	}
	destProtocolID := core.ProtocolID(val)

	destID, err := core.NewCrossChainID(destProtocolID, req.DestinationCounterpartyId)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error creating destination cross-chain ID")
	}

	if !q.HasDispatchedCounts(ctx, &sourceID, &destID) {
		return nil, fmt.Errorf(
			"dispatched counts do not exist for source ID %s and destination ID %s",
			sourceID.String(),
			destID.String(),
		)
	}

	counts := q.GetDispatchedCounts(ctx, &sourceID, &destID)

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts:     []*dispatchertypes.DispatchCountEntry{counts},
		Pagination: nil,
	}, nil
}

func (q queryServer) DispatchedCountsByDestinationProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedCountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedCountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	val, exists := core.ProtocolID_value[req.ProtocolId]
	if !exists {
		return nil, core.ErrUnableToPause.Wrapf(
			"protocol with ID %s does not exist",
			req.ProtocolId,
		)
	}
	protocolID := core.ProtocolID(val)

	counts, pageRes, err := q.GetDispatchedCountsByDestinationProtocolID(
		ctx,
		protocolID,
		req.Pagination,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error retrieving paginated results")
	}

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts:     counts,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) DispatchedCountsBySourceProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedCountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedCountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	val, exists := core.ProtocolID_value[req.ProtocolId]
	if !exists {
		return nil, core.ErrUnableToPause.Wrapf(
			"protocol with ID %s does not exist",
			req.ProtocolId,
		)
	}
	protocolID := core.ProtocolID(val)

	counts, pageRes, err := q.GetDispatchedCountsBySourceProtocolID(ctx, protocolID, req.Pagination)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error retrieving paginated results")
	}

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts:     counts,
		Pagination: pageRes,
	}, nil
}

func (q queryServer) DispatchedAmounts(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedAmountsRequest,
) (*dispatchertypes.QueryDispatchedAmountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	if req.Denom == "" {
		return nil, errorsmod.Wrapf(
			core.ErrEmptyString,
			"error querying an empty string token denom",
		)
	}

	val, exists := core.ProtocolID_value[req.SourceProtocolId]
	if !exists {
		return nil, fmt.Errorf(
			"source protocol with ID %s does not exist",
			req.SourceProtocolId,
		)
	}
	sourceProtocolID := core.ProtocolID(val)

	sourceID, err := core.NewCrossChainID(sourceProtocolID, req.SourceCounterpartyId)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error creating source cross-chain ID")
	}

	val, exists = core.ProtocolID_value[req.DestinationProtocolId]
	if !exists {
		return nil, fmt.Errorf(
			"destination protocol with ID %s does not exist",
			req.DestinationProtocolId,
		)
	}
	destProtocolID := core.ProtocolID(val)

	destID, err := core.NewCrossChainID(destProtocolID, req.DestinationCounterpartyId)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "error creating destination cross-chain ID")
	}

	amounts := q.GetDispatchedAmount(ctx, &sourceID, &destID, req.Denom)
	if !amounts.AmountDispatched.IsPositive() {
		return nil, fmt.Errorf(
			"dispatched amount does not exist for source ID %s, destination ID %s, and denom %s",
			sourceID.String(),
			destID.String(),
			req.Denom,
		)
	}

	return &dispatchertypes.QueryDispatchedAmountsResponse{
		Amounts: []*dispatchertypes.DispatchedAmountEntry{amounts},
	}, nil
}

func (q queryServer) DispatchedAmountsByDestinationProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedAmountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedAmountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	val, exists := core.ProtocolID_value[req.ProtocolId]
	if !exists {
		return nil, core.ErrUnableToPause.Wrapf(
			"protocol with ID %s does not exist",
			req.ProtocolId,
		)
	}
	protocolID := core.ProtocolID(val)

	amounts, pageResp, err := q.GetDispatchedAmountsByDestinationProtocolID(
		ctx,
		protocolID,
		req.Pagination,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error retrieving paginated results")
	}

	return &dispatchertypes.QueryDispatchedAmountsResponse{
		Amounts:    amounts,
		Pagination: pageResp,
	}, nil
}

func (q queryServer) DispatchedAmountsBySourceProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedAmountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedAmountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	val, exists := core.ProtocolID_value[req.ProtocolId]
	if !exists {
		return nil, core.ErrUnableToPause.Wrapf(
			"protocol with ID %s does not exist",
			req.ProtocolId,
		)
	}
	protocolID := core.ProtocolID(val)

	amounts, pageResp, err := q.GetDispatchedAmountsBySourceProtocolID(
		ctx,
		protocolID,
		req.Pagination,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error retrieving paginated results")
	}

	return &dispatchertypes.QueryDispatchedAmountsResponse{
		Amounts:    amounts,
		Pagination: pageResp,
	}, nil
}
