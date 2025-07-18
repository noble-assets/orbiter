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
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

var _ interfaces.PayloadDispatcher = &DispatcherKeeper{}

// DispatcherKeeper is a sub-keeper used to orchestrate the
// components dispatch of an incoming orbiter packet. The DispatcherKeeper
// keeps track of the statistics associated with the handled dispatches.
type DispatcherKeeper struct {
	logger log.Logger

	// Packet component handlers
	OrbitsHandler  interfaces.PacketHandler[*types.OrbitPacket]
	ActionsHandler interfaces.PacketHandler[*types.ActionPacket]

	// Stats
	DispatchedAmounts *collections.IndexedMap[DispatchedAmountsKey, types.AmountDispatched, DispatchedAmountsIndexes]
	DispatchCounts    *collections.IndexedMap[DispatchedCountsKey, uint32, DispatchedCountsIndexes]
}

// NewDispatcherKeeper creates a new instance of a DispatcherKeeper.
func NewDispatcherKeeper(
	cdc codec.BinaryCodec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	orbitHandler types.PacketHandler[*types.OrbitPacket],
	actionHandler types.PacketHandler[*types.ActionPacket],
) (*DispatcherKeeper, error) {
	if cdc == nil {
		return nil, errors.New("codec cannot be nil")
	}
	if sb == nil {
		return nil, errors.New("schema builder cannot be nil")
	}
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	dispatcherKeeper := &DispatcherKeeper{
		logger:         logger.With(types.SubKeeperPrefix, types.DispatcherKeeperName),
		OrbitsHandler:  orbitHandler,
		ActionsHandler: actionHandler,
		DispatchedAmounts: collections.NewIndexedMap(
			sb,
			types.DispatchedAmountsPrefix,
			types.DispatchedAmountsName,
			collections.QuadKeyCodec(
				collections.Uint32Key,
				collections.StringKey,
				collections.StringKey,
				collections.StringKey,
			),
			codec.CollValue[types.AmountDispatched](cdc),
			newDispatchedAmountsIndexes(sb),
		),
		DispatchCounts: collections.NewIndexedMap(
			sb,
			types.DispatchedCountsPrefix,
			types.DispatchedCountsName,
			collections.TripleKeyCodec(
				collections.Uint32Key,
				collections.StringKey,
				collections.StringKey,
			),
			collections.Uint32Value,
			newDispatchedCountsIndexes(sb),
		),
	}

	return dispatcherKeeper, dispatcherKeeper.Validate()
}

// Validate checks that the field of the DispatcherKeeper
// are valid.
func (d *DispatcherKeeper) Validate() error {
	if d.OrbitsHandler == nil {
		return errors.New("orbits handler cannot be nil")
	}
	if d.ActionsHandler == nil {
		return errors.New("actions handler cannot be nil")
	}
	return nil
}

// DispatchPayload is the entry point to initiate the dispatching
// of an orbiter payload.
func (d *DispatcherKeeper) DispatchPayload(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	payload *types.Payload,
) error {
	if err := d.validatePayload(payload); err != nil {
		return fmt.Errorf("payload validation failed: %w", err)
	}

	if err := d.dispatchActions(ctx, transferAttr, payload.PreActions); err != nil {
		return fmt.Errorf("actions processing failed: %w", err)
	}

	if err := d.dispatchOrbit(ctx, transferAttr, payload.Orbit); err != nil {
		return fmt.Errorf("orbit processing failed: %w", err)
	}

	if err := d.updateStats(ctx, transferAttr, payload.Orbit); err != nil {
		d.logger.Error("Error Updating Orbiter Statistics", "error", err.Error())
	}

	return nil
}

// validatePayload checks if the payload is not nil, and calls
// its validation method.
func (d *DispatcherKeeper) validatePayload(payload *types.Payload) error {
	if payload == nil {
		return errors.New("payload cannot be nil")
	}
	return payload.Validate()
}

// dispatchActions iterates through all the actions, creates
// the action packets and dispatch them for execution.
func (d *DispatcherKeeper) dispatchActions(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	actions []*types.Action,
) error {
	for _, action := range actions {

		packet, err := types.NewActionPacket(transferAttr, action)
		if err != nil {
			return err
		}

		err = d.dispatchActionPacket(ctx, packet)
		if err != nil {
			return err
		}
	}
	return nil
}

// dispatchOrbit creates the orbit packet and dispatch
// it for execution.
func (d *DispatcherKeeper) dispatchOrbit(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	orbit *types.Orbit,
) error {
	packet, err := types.NewOrbitPacket(transferAttr, orbit)
	if err != nil {
		return err
	}

	return d.dispatchOrbitPacket(ctx, packet)
}

// dispatchActionPacket dispatch the action packet to the
// actions handler.
func (d *DispatcherKeeper) dispatchActionPacket(
	ctx context.Context,
	packet *types.ActionPacket,
) error {
	return d.ActionsHandler.HandlePacket(ctx, packet)
}

// dispatchOrbitPacket dispatch the orbit packet to the
// orbits handler.
func (d *DispatcherKeeper) dispatchOrbitPacket(
	ctx context.Context,
	packet *types.OrbitPacket,
) error {
	return d.OrbitsHandler.HandlePacket(ctx, packet)
}
