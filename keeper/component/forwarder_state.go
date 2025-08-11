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

	"cosmossdk.io/collections"

	"orbiter.dev/types/core"
)

// ====================================================================================================
// PausedControllers
// ====================================================================================================

func (f *Forwarder) IsControllerPaused(
	ctx context.Context,
	protocolID core.ProtocolID,
) (bool, error) {
	paused, err := f.PausedProtocols.Has(ctx, int32(protocolID))

	return paused, err
}

func (f *Forwarder) SetPausedController(ctx context.Context,
	protocolID core.ProtocolID,
) error {
	paused, err := f.IsControllerPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if paused {
		return nil
	}

	return f.PausedProtocols.Set(ctx, int32(protocolID))
}

func (f *Forwarder) SetUnpausedController(
	ctx context.Context,
	protocolID core.ProtocolID,
) error {
	paused, err := f.IsControllerPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if !paused {
		return nil
	}

	return f.PausedProtocols.Remove(ctx, int32(protocolID))
}

// ====================================================================================================
// PausedOrbits
// ====================================================================================================

func (f *Forwarder) IsOrbitPaused(
	ctx context.Context,
	protocolID core.ProtocolID,
	counterpartyID string,
) (bool, error) {
	paused, err := f.PausedProtocolCounterparties.Has(
		ctx,
		collections.Join(int32(protocolID), counterpartyID),
	)

	return paused, err
}

func (f *Forwarder) SetPausedOrbit(ctx context.Context,
	protocolID core.ProtocolID,
	counterpartyID string,
) error {
	paused, err := f.IsOrbitPaused(ctx, protocolID, counterpartyID)
	if err != nil {
		return err
	}
	if paused {
		return nil
	}

	return f.PausedProtocolCounterparties.Set(
		ctx,
		collections.Join(int32(protocolID), counterpartyID),
	)
}

func (f *Forwarder) SetUnpausedOrbit(
	ctx context.Context,
	protocolID core.ProtocolID,
	counterpartyID string,
) error {
	paused, err := f.IsOrbitPaused(ctx, protocolID, counterpartyID)
	if err != nil {
		return err
	}
	if !paused {
		return nil
	}

	return f.PausedProtocolCounterparties.Remove(
		ctx,
		collections.Join(int32(protocolID), counterpartyID),
	)
}
