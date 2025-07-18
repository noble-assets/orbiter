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

package subkeepers

import (
	"context"
	"errors"

	"cosmossdk.io/collections"

	"orbiter.dev/types"
)

// ====================================================================================================
// PausedControllers
// ====================================================================================================

func (k *OrbitKeeper) IsControllerPaused(
	ctx context.Context,
	protocolId types.ProtocolID,
) (bool, error) {
	paused, err := k.PausedControllers.Get(ctx, int32(protocolId))
	// Default not paused.
	if errors.Is(err, collections.ErrNotFound) {
		return false, nil
	}
	return paused, err
}

func (k *OrbitKeeper) SetPausedController(ctx context.Context,
	protocolId types.ProtocolID,
) error {
	paused, err := k.IsControllerPaused(ctx, protocolId)
	if err != nil {
		return err
	}
	if paused {
		return nil
	}

	return k.PausedControllers.Set(ctx, int32(protocolId), true)
}

func (k *OrbitKeeper) SetUnpausedController(
	ctx context.Context,
	protocolId types.ProtocolID,
) error {
	paused, err := k.IsControllerPaused(ctx, protocolId)
	if err != nil {
		return err
	}
	if !paused {
		return nil
	}

	return k.PausedControllers.Set(ctx, int32(protocolId), false)
}

// ====================================================================================================
// PausedOrbits
// ====================================================================================================

func (k *OrbitKeeper) IsOrbitPaused(
	ctx context.Context,
	protocolId types.ProtocolID,
	counterpartyId string,
) (bool, error) {
	paused, err := k.PausedOrbits.Get(ctx, collections.Join(int32(protocolId), counterpartyId))
	// default not paused.
	if errors.Is(err, collections.ErrNotFound) {
		return false, nil
	}
	return paused, err
}

func (k *OrbitKeeper) SetPausedOrbit(ctx context.Context,
	protocolId types.ProtocolID,
	counterpartyId string,
) error {
	paused, err := k.IsOrbitPaused(ctx, protocolId, counterpartyId)
	if err != nil {
		return err
	}
	if paused {
		return nil
	}

	return k.PausedOrbits.Set(ctx, collections.Join(int32(protocolId), counterpartyId), true)
}

func (k *OrbitKeeper) SetUnpausedOrbit(
	ctx context.Context,
	protocolId types.ProtocolID,
	counterpartyId string,
) error {
	paused, err := k.IsOrbitPaused(ctx, protocolId, counterpartyId)
	if err != nil {
		return err
	}
	if !paused {
		return nil
	}

	return k.PausedOrbits.Set(ctx, collections.Join(int32(protocolId), counterpartyId), false)
}
