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

	"cosmossdk.io/collections"

	"orbiter.dev/keeper/component"
	"orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

var _ dispatcher.QueryServer = &queryServerDispatcher{}

type queryServerDispatcher struct {
	*Keeper
}

func NewQueryServerDispatcher(keeper *Keeper) queryServerDispatcher {
	return queryServerDispatcher{Keeper: keeper}
}

// DispatchedAmount implements dispatcher.QueryServer.
func (q *queryServerDispatcher) DispatchedAmount(
	ctx context.Context,
	req *dispatcher.QueryDispatchedAmountRequest,
) (*dispatcher.QueryDispatchedAmountResponse, error) {
	d := q.Dispatcher()

	sourceOrbitID, err := core.NewOrbitID(req.SourceProtocolId, req.SourceCounterpartyId)
	if err != nil {
		return nil, fmt.Errorf("invalid source orbit ID: %w", err)
	}

	destOrbitID, err := core.ParseOrbitID(req.DestinationOrbitId)
	if err != nil {
		return nil, fmt.Errorf("invalid destination orbit ID: %w", err)
	}

	amountDispatched := d.GetDispatchedAmount(ctx, sourceOrbitID, destOrbitID, req.Denom)

	return &dispatcher.QueryDispatchedAmountResponse{
		AmountDispatched: amountDispatched,
	}, nil
}

// DispatchedAmountsByDestinationProtocol implements dispatcher.QueryServer.
func (q *queryServerDispatcher) DispatchedAmountsByDestinationProtocol(
	ctx context.Context,
	req *dispatcher.QueryDispatchedAmountsByDestinationProtocolRequest,
) (*dispatcher.QueryDispatchedAmountsByDestinationProtocolResponse, error) {
	d := q.Dispatcher()

	var entries []dispatcher.DispatchedAmountEntry

	// Walk through the index to get all entries for the destination protocol
	err := d.DispatchedAmounts.Indexes.ByDestinationProtocolID.Walk(
		ctx,
		collections.NewPrefixedPairRange[uint32, component.DispatchedAmountsKey](
			req.ProtocolId.Uint32(),
		),
		func(indexingKey uint32, indexedKey component.DispatchedAmountsKey) (stop bool, err error) {
			// Get the actual value from the main collection
			value, err := d.DispatchedAmounts.Get(ctx, indexedKey)
			if err != nil {
				return true, err
			}

			// Parse the source and destination orbit IDs from the key
			sourceOrbitID, err := core.NewOrbitID(core.ProtocolID(indexedKey.K1()), indexedKey.K2())
			if err != nil {
				return true, err
			}

			destOrbitID, err := core.ParseOrbitID(indexedKey.K3())
			if err != nil {
				return true, err
			}

			entry := dispatcher.DispatchedAmountEntry{
				SourceId:         &sourceOrbitID,
				DestinationId:    &destOrbitID,
				Denom:            indexedKey.K4(),
				AmountDispatched: value,
			}
			entries = append(entries, entry)
			return false, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"unable to query dispatched amounts by destination protocol: %w",
			err,
		)
	}

	return &dispatcher.QueryDispatchedAmountsByDestinationProtocolResponse{
		Entries: entries,
	}, nil
}

// DispatchedAmountsBySourceProtocol implements dispatcher.QueryServer.
func (q *queryServerDispatcher) DispatchedAmountsBySourceProtocol(
	ctx context.Context,
	req *dispatcher.QueryDispatchedAmountsBySourceProtocolRequest,
) (*dispatcher.QueryDispatchedAmountsBySourceProtocolResponse, error) {
	d := q.Dispatcher()

	var entries []dispatcher.DispatchedAmountEntry

	// Walk through the main collection with source protocol prefix
	prefix := collections.NewPrefixedQuadRange[uint32, string, string, string](
		req.ProtocolId.Uint32(),
	)

	err := d.DispatchedAmounts.Walk(
		ctx,
		prefix,
		func(key component.DispatchedAmountsKey, value dispatcher.AmountDispatched) (stop bool, err error) {
			// Parse the source and destination orbit IDs from the key
			sourceOrbitID, err := core.NewOrbitID(core.ProtocolID(key.K1()), key.K2())
			if err != nil {
				return true, err
			}

			destOrbitID, err := core.ParseOrbitID(key.K3())
			if err != nil {
				return true, err
			}

			entry := dispatcher.DispatchedAmountEntry{
				SourceId:         &sourceOrbitID,
				DestinationId:    &destOrbitID,
				Denom:            key.K4(),
				AmountDispatched: value,
			}
			entries = append(entries, entry)
			return false, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to query dispatched amounts by source protocol: %w", err)
	}

	return &dispatcher.QueryDispatchedAmountsBySourceProtocolResponse{
		Entries: entries,
	}, nil
}

// DispatchedCounts implements dispatcher.QueryServer.
func (q *queryServerDispatcher) DispatchedCounts(
	ctx context.Context,
	req *dispatcher.QueryDispatchedCountsRequest,
) (*dispatcher.QueryDispatchedCountsResponse, error) {
	d := q.Dispatcher()

	sourceOrbitID, err := core.NewOrbitID(req.SourceProtocolId, req.SourceCounterpartyId)
	if err != nil {
		return nil, fmt.Errorf("invalid source orbit ID: %w", err)
	}

	destOrbitID, err := core.ParseOrbitID(req.DestinationOrbitId)
	if err != nil {
		return nil, fmt.Errorf("invalid destination orbit ID: %w", err)
	}

	count := d.GetDispatchedCounts(ctx, sourceOrbitID, destOrbitID)

	return &dispatcher.QueryDispatchedCountsResponse{
		Count: count,
	}, nil
}
