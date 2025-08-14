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

	"cosmossdk.io/collections"
	"cosmossdk.io/collections/indexes"
	"cosmossdk.io/errors"
	"cosmossdk.io/math"

	dispatchertypes "orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

// DispatchedAmountsKey is defined as:
// (source protocol ID, source counterparty ID, destination cross-chain ID, denom).
type DispatchedAmountsKey = collections.Quad[int32, string, string, string]

type DispatchedAmountsIndexes struct {
	// ByDestinationProtocolID keeps track of entries indexes associated
	// with a single destination protocol ID.
	ByDestinationProtocolID *indexes.Multi[int32, DispatchedAmountsKey, dispatchertypes.AmountDispatched]
	// ByDestinationCrossChainID keeps track of entries indexes associated with a tuple:
	// (destination protocol Id, destination chain Id, denom).
	ByDestinationCrossChainID *indexes.Multi[collections.Triple[int32, string, string], DispatchedAmountsKey, dispatchertypes.AmountDispatched]
}

func newDispatchedAmountsIndexes(sb *collections.SchemaBuilder) DispatchedAmountsIndexes {
	primaryKeyCodec := collections.QuadKeyCodec(
		collections.Int32Key,  // source protocol ID
		collections.StringKey, // source counterparty ID
		collections.StringKey, // destination cross-chain ID
		collections.StringKey, // denom
	)

	return DispatchedAmountsIndexes{
		ByDestinationProtocolID: indexes.NewMulti(
			sb,
			core.DispatchedAmountsPrefixByDestinationProtocolID,
			core.DispatchedAmountsName+"_by_destination_protocol_id",
			collections.Int32Key,
			primaryKeyCodec,
			func(pk DispatchedAmountsKey, value dispatchertypes.AmountDispatched) (int32, error) {
				ccID, err := core.ParseCrossChainID(pk.K3())
				if err != nil {
					return 0, errors.Wrap(err, "error parsing destination cross-chain ID")
				}

				return int32(ccID.GetProtocolId()), nil
			},
		),
		ByDestinationCrossChainID: indexes.NewMulti(
			sb,
			core.DispatchedAmountsPrefixByDestinationCrossChainID,
			core.DispatchedAmountsName+"_by_destination_orbit_id",
			collections.TripleKeyCodec(
				collections.Int32Key,
				collections.StringKey,
				collections.StringKey,
			),
			primaryKeyCodec,
			func(pk DispatchedAmountsKey, value dispatchertypes.AmountDispatched) (collections.Triple[int32, string, string], error) {
				ccID, err := core.ParseCrossChainID(pk.K3())
				if err != nil {
					return collections.Triple[int32, string, string]{}, errors.Wrap(
						err,
						"error parsing destination cross-chain ID",
					)
				}

				return collections.Join3(
					int32(ccID.GetProtocolId()),
					ccID.GetCounterpartyId(),
					pk.K4(),
				), nil
			},
		),
	}
}

// ====================================================================================================
// Dispatched Amount
// ====================================================================================================

func (d *Dispatcher) GetDispatchedAmount(
	ctx context.Context,
	sourceID core.CrossChainID,
	destID core.CrossChainID,
	denom string,
) dispatchertypes.AmountDispatched {
	key := collections.Join4(
		int32(sourceID.GetProtocolId()),
		sourceID.GetCounterpartyId(),
		destID.ID(),
		denom,
	)

	amountDispatched, err := d.dispatchedAmounts.Get(ctx, key)
	if err != nil {
		d.logger.Error("received an error getting dispatched amounts",
			"error", err,
			"source_id", sourceID.ID(),
			"dest_id", destID.ID(),
			"denom", denom,
		)
		amountDispatched = dispatchertypes.AmountDispatched{
			Incoming: math.ZeroInt(),
			Outgoing: math.ZeroInt(),
		}
	}

	return amountDispatched
}

func (d *Dispatcher) HasDispatchedAmount(
	ctx context.Context,
	sourceID core.CrossChainID,
	destID core.CrossChainID,
	denom string,
) bool {
	amountDispatched := d.GetDispatchedAmount(ctx, sourceID, destID, denom)
	return amountDispatched.IsPositive()
}

func (d *Dispatcher) SetDispatchedAmount(
	ctx context.Context,
	sourceID *core.CrossChainID,
	destID *core.CrossChainID,
	denom string,
	amountDispatched dispatchertypes.AmountDispatched,
) error {
	key := collections.Join4(
		int32(sourceID.GetProtocolId()),
		sourceID.GetCounterpartyId(),
		destID.ID(),
		denom,
	)

	return d.dispatchedAmounts.Set(ctx, key, amountDispatched)
}

func (d *Dispatcher) GetDispatchedAmountsByProtocolID(
	ctx context.Context,
	protocolID core.ProtocolID,
) []dispatchertypes.DispatchedAmountEntry {
	amounts := []dispatchertypes.DispatchedAmountEntry{}

	rng := collections.NewPrefixedQuadRange[int32, string, string, string](int32(protocolID))

	err := d.dispatchedAmounts.Walk(
		ctx,
		rng,
		func(k DispatchedAmountsKey, v dispatchertypes.AmountDispatched) (stop bool, err error) {
			entry, err := d.GetDispatchedAmountEntryFromKey(ctx, k)
			if err != nil {
				return true, err
			}

			amounts = append(amounts, entry)

			return false, nil
		},
	)
	if err != nil {
		d.logger.Error(
			"error in dispatched amounts walking by source protocol ID",
			"error",
			err,
		)

		return []dispatchertypes.DispatchedAmountEntry{}
	}

	return amounts
}

func (d *Dispatcher) GetDispatchedAmountsByDestinationProtocolID(
	ctx context.Context,
	protocolID core.ProtocolID,
) []dispatchertypes.DispatchedAmountEntry {
	amounts := []dispatchertypes.DispatchedAmountEntry{}

	rng := collections.NewPrefixedPairRange[int32, DispatchedAmountsKey](int32(protocolID))

	err := d.dispatchedAmounts.Indexes.ByDestinationProtocolID.Walk(
		ctx,
		rng,
		func(_ int32, k DispatchedAmountsKey) (stop bool, err error) {
			entry, err := d.GetDispatchedAmountEntryFromKey(ctx, k)
			if err != nil {
				return true, err
			}

			amounts = append(amounts, entry)

			return false, nil
		},
	)
	if err != nil {
		d.logger.Error(
			"error in dispatched amounts walking by destination protocol ID index",
			"error",
			err,
		)

		return []dispatchertypes.DispatchedAmountEntry{}
	}

	return amounts
}

func (d *Dispatcher) GetAllDispatchedAmounts(
	ctx context.Context,
) []dispatchertypes.DispatchedAmountEntry {
	amounts := []dispatchertypes.DispatchedAmountEntry{}

	err := d.dispatchedAmounts.Walk(
		ctx,
		nil,
		func(k DispatchedAmountsKey, v dispatchertypes.AmountDispatched) (stop bool, err error) {
			entry, err := d.GetDispatchedAmountEntryFromKey(ctx, k)
			if err != nil {
				return true, err
			}

			amounts = append(amounts, entry)

			return false, nil
		},
	)
	if err != nil {
		d.logger.Error("error in dispatched amounts walking all values")

		return []dispatchertypes.DispatchedAmountEntry{}
	}

	return amounts
}

func (d *Dispatcher) GetDispatchedAmountEntryFromKey(
	ctx context.Context,
	k DispatchedAmountsKey,
) (dispatchertypes.DispatchedAmountEntry, error) {
	var entry dispatchertypes.DispatchedAmountEntry

	value, err := d.dispatchedAmounts.Get(ctx, k)
	if err != nil {
		return entry, errors.Wrap(err, "failed to get disptched amount")
	}

	sourceID, err := core.NewCrossChainID(core.ProtocolID(k.K1()), k.K2())
	if err != nil {
		return entry, errors.Wrap(err, "failed to create source cross-chain ID")
	}

	destID, err := core.ParseCrossChainID(k.K3())
	if err != nil {
		return entry, errors.Wrap(err, "failed to create destination cross-chain ID")
	}

	entry.SourceId = &sourceID
	entry.DestinationId = &destID
	entry.Denom = k.K4()
	entry.AmountDispatched = value

	return entry, nil
}

// ====================================================================================================
// Dispatched Counts
// ====================================================================================================

// DispatchedCountsKey is defined as:
// (source protocol ID, source counterparty ID, destination cross-chain ID).
type DispatchedCountsKey = collections.Quad[int32, string, int32, string]

type DispatchedCountsIndexes struct {
	// ByDestinationProtocolID keeps track of entries indexes associated
	// with a single destination protocol ID.
	ByDestinationProtocolID *indexes.Multi[int32, DispatchedCountsKey, uint64]
}

func newDispatchedCountsIndexes(sb *collections.SchemaBuilder) DispatchedCountsIndexes {
	primaryKeyCodec := collections.QuadKeyCodec(
		collections.Int32Key,  // source protocol ID
		collections.StringKey, // source counterparty ID
		collections.Int32Key,  // destination protocol ID
		collections.StringKey, // destination counterparty ID
	)

	return DispatchedCountsIndexes{
		ByDestinationProtocolID: indexes.NewMulti(
			sb,
			core.DispatchedCountsPrefixByDestinationProtocolID,
			core.DispatchedCountsName+"_by_destination_protocol_id",
			collections.Int32Key,
			primaryKeyCodec,
			func(pk DispatchedCountsKey, _ uint64) (int32, error) {
				return pk.K3(), nil
			},
		),
	}
}

// GetDispatchedCounts returns the number of dispatches between the
// two cross-chain IDs. Return 0 in case of an error.
func (d *Dispatcher) GetDispatchedCounts(
	ctx context.Context,
	sourceID *core.CrossChainID,
	destID *core.CrossChainID,
) *dispatchertypes.DispatchCountEntry {
	key := collections.Join4(
		int32(sourceID.GetProtocolId()),
		sourceID.GetCounterpartyId(),
		int32(destID.GetProtocolId()),
		destID.GetCounterpartyId(),
	)

	counts, err := d.dispatchCounts.Get(ctx, key)
	if err != nil {
		d.logger.Error("received an error getting dispatches count",
			"error", err,
			"source_id", sourceID.ID(),
			"dest_id", destID.ID(),
		)
		counts = 0
	}

	return &dispatchertypes.DispatchCountEntry{
		SourceId:      sourceID,
		DestinationId: destID,
		Count:         counts,
	}
}

func (d *Dispatcher) HasDispatchedCounts(
	ctx context.Context,
	sourceID *core.CrossChainID,
	destID *core.CrossChainID,
) bool {
	dc := d.GetDispatchedCounts(ctx, sourceID, destID)

	return dc.Count != 0
}

func (d *Dispatcher) SetDispatchedCounts(
	ctx context.Context,
	sourceID *core.CrossChainID,
	destID *core.CrossChainID,
	counts uint64,
) error {
	key := collections.Join4(
		int32(sourceID.GetProtocolId()),
		sourceID.GetCounterpartyId(),
		int32(destID.GetProtocolId()),
		destID.GetCounterpartyId(),
	)

	return d.dispatchCounts.Set(ctx, key, counts)
}

func (d *Dispatcher) GetAllDispatchedCounts(
	ctx context.Context,
) []dispatchertypes.DispatchCountEntry {
	counts := []dispatchertypes.DispatchCountEntry{}

	err := d.dispatchCounts.Walk(
		ctx,
		nil,
		func(k DispatchedCountsKey, v uint64) (stop bool, err error) {
			entry, err := d.getDispatchCountEntryFromKey(ctx, k)
			if err != nil {
				return true, err
			}

			counts = append(counts, entry)

			return false, nil
		},
	)
	if err != nil {
		d.logger.Error("error in dispatchedCounts walking all values")

		return []dispatchertypes.DispatchCountEntry{}
	}

	return counts
}

func (d *Dispatcher) GetDispatchedCountsBySourceProtocolID(
	ctx context.Context,
	id core.ProtocolID,
) []*dispatchertypes.DispatchCountEntry {
	counts := []*dispatchertypes.DispatchCountEntry{}

	rng := collections.NewPrefixedQuadRange[int32, string, int32, string](int32(id))
	err := d.dispatchCounts.Walk(
		ctx,
		rng,
		func(k DispatchedCountsKey, v uint64) (stop bool, err error) {
			entry, err := d.getDispatchCountEntryFromKey(ctx, k)
			if err != nil {
				return true, err
			}

			counts = append(counts, &entry)

			return false, nil
		},
	)
	if err != nil {
		d.logger.Error("error in dispatched counts walking by source protocol ID")

		return []*dispatchertypes.DispatchCountEntry{}
	}

	return counts
}

func (d *Dispatcher) GetDispatchedCountsByDestinationProtocolID(
	ctx context.Context,
	id core.ProtocolID,
) []*dispatchertypes.DispatchCountEntry {
	counts := []*dispatchertypes.DispatchCountEntry{}

	rng := collections.NewPrefixedPairRange[int32, DispatchedCountsKey](int32(id))
	err := d.dispatchCounts.Indexes.ByDestinationProtocolID.Walk(
		ctx,
		rng,
		func(_ int32, k DispatchedCountsKey) (stop bool, err error) {
			entry, err := d.getDispatchCountEntryFromKey(ctx, k)
			if err != nil {
				return true, err
			}

			counts = append(counts, &entry)

			return false, nil
		},
	)
	if err != nil {
		d.logger.Error("error in dispatched counts walking by destination protocol ID index")

		return []*dispatchertypes.DispatchCountEntry{}
	}

	return counts
}

func (d *Dispatcher) getDispatchCountEntryFromKey(
	ctx context.Context,
	k DispatchedCountsKey,
) (dispatchertypes.DispatchCountEntry, error) {
	var entry dispatchertypes.DispatchCountEntry

	value, err := d.dispatchCounts.Get(ctx, k)
	if err != nil {
		return entry, errors.Wrap(err, "failed to get disptched counts")
	}

	sourceID, err := core.NewCrossChainID(core.ProtocolID(k.K1()), k.K2())
	if err != nil {
		return entry, errors.Wrap(err, "failed to create source cross-chain ID")
	}

	destID, err := core.NewCrossChainID(core.ProtocolID(k.K3()), k.K4())
	if err != nil {
		return entry, errors.Wrap(err, "failed to create destination cross-chain ID")
	}

	entry.SourceId = &sourceID
	entry.DestinationId = &destID
	entry.Count = value

	return entry, nil
}
