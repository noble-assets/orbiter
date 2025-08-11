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

package component

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/math"

	dispatchertypes "orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

type (
	// DispatchedAmountsKey is defined as:
	// (source protocol ID, source counterparty ID, destination cross-chain ID, denom).
	DispatchedAmountsKey = collections.Quad[uint32, string, string, string]
	// DispatchedCountsKey is defined as:
	// (source protocol ID, source counterparty ID, destination cross-chain ID).
	DispatchedCountsKey = collections.Triple[uint32, string, string]
)

type DispatchedAmountsIndexes struct {
	// ByDestinationProtocolID keeps track of entries indexes associated
	// with a single destination protocol ID.
	ByDestinationProtocolID *indexes.Multi[uint32, DispatchedAmountsKey, dispatchertypes.AmountDispatched]
	// ByDestinationOrbitID keeps track of entries indexes associated with a tuple:
	// (destination protocol Id, destination chain Id, denom).
	ByDestinationOrbitID *indexes.Multi[collections.Triple[uint32, string, string], DispatchedAmountsKey, dispatchertypes.AmountDispatched]
}

func newDispatchedAmountsIndexes(sb *collections.SchemaBuilder) DispatchedAmountsIndexes {
	primaryKeyCodec := collections.QuadKeyCodec(
		collections.Uint32Key,
		collections.StringKey,
		collections.StringKey,
		collections.StringKey,
	)

	return DispatchedAmountsIndexes{
		ByDestinationProtocolID: indexes.NewMulti(
			sb,
			core.DispatchedAmountsPrefixByDestinationProtocolID,
			core.DispatchedAmountsName+"_by_destination_protocol_id",
			collections.Uint32Key,
			primaryKeyCodec,
			func(pk DispatchedAmountsKey, value dispatchertypes.AmountDispatched) (uint32, error) {
				orbitID, err := core.ParseOrbitID(pk.K3())
				if err != nil {
					return 0, fmt.Errorf("error parsing destination orbit ID: %w", err)
				}

				return orbitID.GetProtocolId().Uint32(), nil
			},
		),
		ByDestinationOrbitID: indexes.NewMulti(
			sb,
			core.DispatchedAmountsPrefixByDestinationOrbitID,
			core.DispatchedAmountsName+"_by_destination_orbit_id",
			collections.TripleKeyCodec(
				collections.Uint32Key,
				collections.StringKey,
				collections.StringKey,
			),
			primaryKeyCodec,
			func(pk DispatchedAmountsKey, value dispatchertypes.AmountDispatched) (collections.Triple[uint32, string, string], error) {
				orbitID, err := core.ParseOrbitID(pk.K3())
				if err != nil {
					return collections.Triple[uint32, string, string]{}, fmt.Errorf(
						"error parsing destination orbit ID: %w",
						err,
					)
				}

				return collections.Join3(
					orbitID.GetProtocolId().Uint32(),
					orbitID.GetCounterpartyId(),
					pk.K4(),
				), nil
			},
		),
	}
}

type DispatchedCountsIndexes struct {
	// ByDestinationProtocolID keeps track of entries indexes associated
	// with a single destination protocol ID.
	ByDestinationProtocolID *indexes.Multi[uint32, DispatchedCountsKey, uint32]
}

func newDispatchedCountsIndexes(sb *collections.SchemaBuilder) DispatchedCountsIndexes {
	primaryKeyCodec := collections.TripleKeyCodec(
		collections.Uint32Key,
		collections.StringKey,
		collections.StringKey,
	)

	return DispatchedCountsIndexes{
		ByDestinationProtocolID: indexes.NewMulti(
			sb,
			core.DispatchedCountsPrefixByDestinationProtocolID,
			core.DispatchedCountsName+"_by_destination_protocol_id",
			collections.Uint32Key,
			primaryKeyCodec,
			func(pk DispatchedCountsKey, _ uint32) (uint32, error) {
				orbitID, err := core.ParseOrbitID(pk.K3())
				if err != nil {
					return 0, fmt.Errorf(
						"error parsing destination orbit ID: %w",
						err,
					)
				}

				return orbitID.GetProtocolId().Uint32(), nil
			},
		),
	}
}

// ====================================================================================================
// Dispatched Amount
// ====================================================================================================

func (d *Dispatcher) GetDispatchedAmount(
	ctx context.Context,
	sourceOrbitID core.OrbitID,
	destinationOrbitID core.OrbitID,
	denom string,
) dispatchertypes.AmountDispatched {
	key := collections.Join4(
		sourceOrbitID.GetProtocolId().Uint32(),
		sourceOrbitID.GetCounterpartyId(),
		destinationOrbitID.ID(),
		denom,
	)

	amountDispatched, err := d.DispatchedAmounts.Get(ctx, key)
	if err != nil {
		amountDispatched = dispatchertypes.AmountDispatched{
			Incoming: math.ZeroInt(),
			Outgoing: math.ZeroInt(),
		}
	}

	return amountDispatched
}

func (d *Dispatcher) HasDispatchedAmount(
	ctx context.Context,
	sourceInfo core.OrbitID,
	destinationInfo core.OrbitID,
	denom string,
) bool {
	amountDispatched := d.GetDispatchedAmount(ctx, sourceInfo, destinationInfo, denom)
	if amountDispatched.Incoming.IsZero() && amountDispatched.Outgoing.IsZero() {
		return false
	}

	return true
}

func (d *Dispatcher) SetDispatchedAmount(
	ctx context.Context,
	sourceOrbitID core.OrbitID,
	destOrbitID core.OrbitID,
	denom string,
	amountDispatched dispatchertypes.AmountDispatched,
) error {
	key := collections.Join4(
		sourceOrbitID.GetProtocolId().Uint32(),
		sourceOrbitID.GetCounterpartyId(),
		destOrbitID.ID(),
		denom,
	)

	return d.DispatchedAmounts.Set(ctx, key, amountDispatched)
}

func (d *Dispatcher) GetDispatchedAmountsByProtocolID(
	ctx context.Context,
	protocolID core.ProtocolID,
) dispatchertypes.TotalDispatched {
	totalDispatched := dispatchertypes.NewTotalDispatched()

	callback := func(sourceCounterpartyId string, amountDispatched dispatchertypes.ChainAmountDispatched) bool {
		totalDispatched.SetAmountDispatched(sourceCounterpartyId, amountDispatched)

		return false
	}

	d.IterateDispatchedAmountsByProtocolID(
		ctx,
		protocolID,
		callback,
	)

	return *totalDispatched
}

func (d *Dispatcher) IterateDispatchedAmountsByProtocolID(
	ctx context.Context,
	protocolID core.ProtocolID,
	callback func(string, dispatchertypes.ChainAmountDispatched) bool,
) {
	prefix := collections.NewPrefixedQuadRange[uint32, string, string, string](
		protocolID.Uint32(),
	)

	err := d.DispatchedAmounts.Walk(
		ctx,
		prefix,
		func(key DispatchedAmountsKey, value dispatchertypes.AmountDispatched) (stop bool, err error) {
			orbitID, err := core.ParseOrbitID(key.K3())
			if err != nil {
				return true, err
			}
			dispatchedInfo := dispatchertypes.NewChainAmountDispatched(orbitID, value)

			return callback(key.K2(), *dispatchedInfo), nil
		},
	)
	if err != nil {
		d.logger.Error("error in IterateDispatchedByProtocolID walking Dispatched")
	}
}

func (d *Dispatcher) GetDispatchedAmountsByDestinationProtocolID(
	ctx context.Context,
	protocolID core.ProtocolID,
) dispatchertypes.TotalDispatched {
	totalDispatched := dispatchertypes.NewTotalDispatched()

	callback := func(sourceCounterpartyId string, amountDispatched dispatchertypes.ChainAmountDispatched) bool {
		totalDispatched.SetAmountDispatched(sourceCounterpartyId, amountDispatched)

		return false
	}

	d.IterateDispatchedAmountsByDestinationProtocolID(
		ctx,
		protocolID,
		callback,
	)

	return *totalDispatched
}

func (d *Dispatcher) IterateDispatchedAmountsByDestinationProtocolID(
	ctx context.Context,
	protocolID core.ProtocolID,
	callback func(string, dispatchertypes.ChainAmountDispatched) bool,
) {
	rng := collections.NewPrefixedPairRange[uint32, DispatchedAmountsKey](protocolID.Uint32())

	err := d.DispatchedAmounts.Indexes.ByDestinationProtocolID.Walk(
		ctx,
		rng,
		func(
			indexingKey uint32,
			indexedKey DispatchedAmountsKey,
		) (stop bool, err error) {
			// Get the actual value from the main collection using the indexed key
			value, err := d.DispatchedAmounts.Get(ctx, indexedKey)
			if err != nil {
				return true, err
			}

			orbitID, err := core.ParseOrbitID(indexedKey.K3())
			if err != nil {
				return true, err
			}
			dispatchedInfo := dispatchertypes.NewChainAmountDispatched(orbitID, value)

			return callback(indexedKey.K2(), *dispatchedInfo), nil
		},
	)
	if err != nil {
		d.logger.Error(
			"error in IterateDispatchedByDestinationProtocolID walking ByDestinationProtocolID index",
		)
	}
}

// ====================================================================================================
// Dispatched Counts
// ====================================================================================================

func (d *Dispatcher) GetDispatchedCounts(
	ctx context.Context,
	sourceOrbitID core.OrbitID,
	destinationOrbitID core.OrbitID,
) uint32 {
	key := collections.Join3(
		sourceOrbitID.GetProtocolId().Uint32(),
		sourceOrbitID.GetCounterpartyId(),
		destinationOrbitID.ID(),
	)

	countDispatches, err := d.DispatchCounts.Get(ctx, key)
	if err != nil {
		countDispatches = 0
	}

	return countDispatches
}

func (d *Dispatcher) HasDispatchedCounts(
	ctx context.Context,
	sourceInfo core.OrbitID,
	destinationInfo core.OrbitID,
) bool {
	countDispatches := d.GetDispatchedCounts(ctx, sourceInfo, destinationInfo)

	return countDispatches != 0
}

func (d *Dispatcher) SetDispatchedCounts(
	ctx context.Context,
	sourceOrbitID core.OrbitID,
	destinationOrbitID core.OrbitID,
	countDispatches uint32,
) error {
	key := collections.Join3(
		sourceOrbitID.GetProtocolId().Uint32(),
		sourceOrbitID.GetCounterpartyId(),
		destinationOrbitID.ID(),
	)

	return d.DispatchCounts.Set(ctx, key, countDispatches)
}
