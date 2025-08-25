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

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"

	"github.com/noble-assets/orbiter/types/core"
)

// ====================================================================================================
// Paused protocols
// ====================================================================================================

func (f *Forwarder) IsProtocolPaused(
	ctx context.Context,
	protocolID core.ProtocolID,
) (bool, error) {
	return f.pausedProtocols.Has(ctx, int32(protocolID))
}

func (f *Forwarder) SetPausedProtocol(ctx context.Context, protocolID core.ProtocolID) error {
	if err := protocolID.Validate(); err != nil {
		return err
	}

	paused, err := f.IsProtocolPaused(ctx, protocolID)
	if err != nil {
		return err
	}

	if paused {
		return core.ErrAlreadySet.Wrapf("paused=%v", paused)
	}

	return f.pausedProtocols.Set(ctx, int32(protocolID))
}

func (f *Forwarder) SetUnpausedProtocol(
	ctx context.Context,
	protocolID core.ProtocolID,
) error {
	if err := protocolID.Validate(); err != nil {
		return err
	}

	paused, err := f.IsProtocolPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if !paused {
		return core.ErrAlreadySet.Wrapf("paused=%v", paused)
	}

	return f.pausedProtocols.Remove(ctx, int32(protocolID))
}

func (f *Forwarder) GetPausedProtocols(
	ctx context.Context,
) ([]core.ProtocolID, error) {
	paused := make([]core.ProtocolID, 0)

	err := f.pausedProtocols.Walk(ctx, nil, func(key int32) (stop bool, err error) {
		paused = append(paused, core.ProtocolID(key))

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return paused, nil
}

// ====================================================================================================
// Paused cross-chain id
// ====================================================================================================

func (f *Forwarder) IsCrossChainPaused(
	ctx context.Context,
	ccID core.CrossChainID,
) (bool, error) {
	return f.pausedCrossChains.Has(
		ctx,
		collections.Join(int32(ccID.GetProtocolId()), ccID.GetCounterpartyId()),
	)
}

func (f *Forwarder) SetPausedCrossChain(
	ctx context.Context,
	ccID core.CrossChainID,
) error {
	if err := ccID.Validate(); err != nil {
		return err
	}

	paused, err := f.IsCrossChainPaused(ctx, ccID)
	if err != nil {
		return err
	}
	if paused {
		return core.ErrAlreadySet.Wrapf("paused=%v", paused)
	}

	return f.pausedCrossChains.Set(
		ctx,
		collections.Join(int32(ccID.GetProtocolId()), ccID.GetCounterpartyId()),
	)
}

func (f *Forwarder) SetUnpausedCrossChain(
	ctx context.Context,
	ccID core.CrossChainID,
) error {
	if err := ccID.Validate(); err != nil {
		return err
	}

	paused, err := f.IsCrossChainPaused(ctx, ccID)
	if err != nil {
		return err
	}
	if !paused {
		return core.ErrAlreadySet.Wrapf("paused=%v", paused)
	}

	return f.pausedCrossChains.Remove(
		ctx,
		collections.Join(int32(ccID.GetProtocolId()), ccID.GetCounterpartyId()),
	)
}

// GetAllPausedCrossChainIDs returns a slice of all paused cross-chain IDs.
//
// CONTRACT: this assumes that all cross-chain ids in state are VALID!
func (f *Forwarder) GetAllPausedCrossChainIDs(
	ctx context.Context,
) ([]*core.CrossChainID, error) {
	crossChainIDs := make([]*core.CrossChainID, 0)

	err := f.pausedCrossChains.Walk(
		ctx,
		nil,
		func(key collections.Pair[int32, string]) (stop bool, err error) {
			ccid := core.CrossChainID{
				ProtocolId:     core.ProtocolID(key.K1()),
				CounterpartyId: key.K2(),
			}

			crossChainIDs = append(crossChainIDs, &ccid)

			return false, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return crossChainIDs, nil
}

// GetPausedCrossChainsMap returns all the paused cross-chain IDs in a map for easier display in
// query results.
//
// NOTE: this method is intended to ONLY be used for queries.
func (f *Forwarder) GetPausedCrossChainsMap(
	ctx context.Context,
	protocolID *core.ProtocolID,
) (map[int32][]string, error) {
	var err error
	var iter collections.KeySetIterator[collections.Pair[int32, string]]
	if protocolID != nil {
		rng := collections.NewPrefixedPairRange[int32, string](int32(*protocolID))
		iter, err = f.pausedCrossChains.Iterate(ctx, rng)
	} else {
		iter, err = f.pausedCrossChains.Iterate(ctx, nil)
	}

	if err != nil {
		return nil, err
	}
	defer func() {
		_ = iter.Close()
	}()

	paused := make(map[int32][]string)

	// Process each key-value pair
	for ; iter.Valid(); iter.Next() {
		key, err := iter.Key()
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to get key from iterator")
		}

		pID := key.K1()
		cID := key.K2()

		paused[pID] = append(paused[pID], cID)
	}

	return paused, nil
}
