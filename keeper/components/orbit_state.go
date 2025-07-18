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

package components

import (
	"context"

	"cosmossdk.io/collections"

	"orbiter.dev/types"
)

// ====================================================================================================
// PausedControllers
// ====================================================================================================

func (c *OrbitComponent) IsControllerPaused(
	ctx context.Context,
	protocolID types.ProtocolID,
) (bool, error) {
	paused, err := c.PausedControllers.Has(ctx, int32(protocolID))
	return paused, err
}

func (c *OrbitComponent) SetPausedController(ctx context.Context,
	protocolID types.ProtocolID,
) error {
	paused, err := c.IsControllerPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if paused {
		return nil
	}

	return c.PausedControllers.Set(ctx, int32(protocolID))
}

func (c *OrbitComponent) SetUnpausedController(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	paused, err := c.IsControllerPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if !paused {
		return nil
	}

	return c.PausedControllers.Remove(ctx, int32(protocolID))
}

// ====================================================================================================
// PausedOrbits
// ====================================================================================================

func (c *OrbitComponent) IsOrbitPaused(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) (bool, error) {
	paused, err := c.PausedOrbits.Has(ctx, collections.Join(int32(protocolID), counterpartyID))
	return paused, err
}

func (c *OrbitComponent) SetPausedOrbit(ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	paused, err := c.IsOrbitPaused(ctx, protocolID, counterpartyID)
	if err != nil {
		return err
	}
	if paused {
		return nil
	}

	return c.PausedOrbits.Set(ctx, collections.Join(int32(protocolID), counterpartyID))
}

func (c *OrbitComponent) SetUnpausedOrbit(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	paused, err := c.IsOrbitPaused(ctx, protocolID, counterpartyID)
	if err != nil {
		return err
	}
	if !paused {
		return nil
	}

	return c.PausedOrbits.Remove(ctx, collections.Join(int32(protocolID), counterpartyID))
}
