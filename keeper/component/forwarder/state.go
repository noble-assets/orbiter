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

	"cosmossdk.io/collections"

	"orbiter.dev/types/core"
)

// ====================================================================================================
// Paused protocols
// ====================================================================================================

func (f *Forwarder) IsProtocolPaused(
	ctx context.Context,
	protocolID core.ProtocolID,
) (bool, error) {
	paused, err := f.PausedProtocols.Has(ctx, int32(protocolID))

	return paused, err
}

func (f *Forwarder) SetPausedProtocol(ctx context.Context, protocolID core.ProtocolID) error {
	paused, err := f.IsProtocolPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if paused {
		return nil
	}

	return f.PausedProtocols.Set(ctx, int32(protocolID))
}

func (f *Forwarder) SetUnpausedProtocol(
	ctx context.Context,
	protocolID core.ProtocolID,
) error {
	paused, err := f.IsProtocolPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if !paused {
		return nil
	}

	return f.PausedProtocols.Remove(ctx, int32(protocolID))
}

func (f *Forwarder) GetPausedProtocols(
	ctx context.Context,
) ([]core.ProtocolID, error) {
	iter, err := f.PausedProtocols.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var paused []core.ProtocolID
	for ; iter.Valid(); iter.Next() {
		k, err := iter.Key()
		if err != nil {
			return nil, err
		}

		id, err := core.NewProtocolID(k)
		if err != nil {
			return nil, fmt.Errorf("cannot create protocol ID from iterator key: %w", err)
		}
		paused = append(paused, id)
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
	return f.PausedCrossChains.Has(
		ctx,
		collections.Join(int32(ccID.GetProtocolId()), ccID.GetCounterpartyId()),
	)
}

func (f *Forwarder) SetPausedCrossChain(
	ctx context.Context,
	ccID core.CrossChainID,
) error {
	paused, err := f.IsCrossChainPaused(ctx, ccID)
	if err != nil {
		return err
	}
	if paused {
		return nil
	}

	return f.PausedCrossChains.Set(
		ctx,
		collections.Join(int32(ccID.GetProtocolId()), ccID.GetCounterpartyId()),
	)
}

func (f *Forwarder) SetUnpausedCrossChain(
	ctx context.Context,
	ccID core.CrossChainID,
) error {
	paused, err := f.IsCrossChainPaused(ctx, ccID)
	if err != nil {
		return err
	}
	if !paused {
		return nil
	}

	return f.PausedCrossChains.Remove(
		ctx,
		collections.Join(int32(ccID.GetProtocolId()), ccID.GetCounterpartyId()),
	)
}

func (f *Forwarder) GetPausedCounterparties(
	ctx context.Context,
	protocolID core.ProtocolID,
) ([]string, error) {
	rng := collections.NewPrefixedPairRange[int32, string](int32(protocolID))

	iter, err := f.PausedCrossChains.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var paused []string
	for ; iter.Valid(); iter.Next() {
		k, err := iter.Key()
		if err != nil {
			return nil, err
		}

		paused = append(paused, k.K2())
	}

	return paused, nil
}
