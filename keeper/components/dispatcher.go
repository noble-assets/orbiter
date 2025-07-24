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
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

var _ interfaces.PayloadDispatcher = &DispatcherComponent{}

// DispatcherComponent is a component used to orchestrate the
// dispatch of an incoming orbiter packet. The dispatcher
// keeps track of the statistics associated with the handled dispatches.
type DispatcherComponent struct {
	logger log.Logger
	// Packet elements handlers
	OrbitHandler  interfaces.PacketHandler[*types.OrbitPacket]
	ActionHandler interfaces.PacketHandler[*types.ActionPacket]
	// Stats
	DispatchedAmounts *collections.IndexedMap[DispatchedAmountsKey, types.AmountDispatched, DispatchedAmountsIndexes]
	DispatchCounts    *collections.IndexedMap[DispatchedCountsKey, uint32, DispatchedCountsIndexes]
}

// NewDispatcherComponent creates a new validated instance of a the dispatcher
// component.
func NewDispatcherComponent(
	cdc codec.BinaryCodec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	orbitHandler interfaces.PacketHandler[*types.OrbitPacket],
	actionHandler interfaces.PacketHandler[*types.ActionPacket],
) (*DispatcherComponent, error) {
	if cdc == nil {
		return nil, types.ErrNilPointer.Wrap("codec cannot be nil")
	}
	if sb == nil {
		return nil, types.ErrNilPointer.Wrap("schema builder cannot be nil")
	}
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	dispatcherComponent := DispatcherComponent{
		logger:        logger.With(types.ComponentPrefix, types.DispatcherComponentName),
		OrbitHandler:  orbitHandler,
		ActionHandler: actionHandler,
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

	return &dispatcherComponent, dispatcherComponent.Validate()
}

// Validate checks that the fields of the dispatcher component are valid.
func (d *DispatcherComponent) Validate() error {
	if d.OrbitHandler == nil {
		return types.ErrNilPointer.Wrap("orbits handler cannot be nil")
	}
	if d.ActionHandler == nil {
		return types.ErrNilPointer.Wrap("actions handler cannot be nil")
	}
	return nil
}

func (d *DispatcherComponent) Logger() log.Logger {
	return d.logger
}

// DispatchPayload is the entry point to initiate the dispatching
// of an orbiter payload.
func (d *DispatcherComponent) DispatchPayload(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	payload *types.Payload,
) error {
	if err := d.validatePayload(payload); err != nil {
		return types.ErrValidation.Wrap(err.Error())
	}

	if err := d.dispatchActions(ctx, transferAttr, payload.PreActions); err != nil {
		return fmt.Errorf("actions dispatch failed: %w", err)
	}

	if err := d.dispatchOrbit(ctx, transferAttr, payload.Orbit); err != nil {
		return fmt.Errorf("orbit dispatch failed: %w", err)
	}

	if err := d.updateStats(ctx, transferAttr, payload.Orbit); err != nil {
		d.logger.Error("Error ypdating Orbiter statistics", "error", err.Error())
	}

	return nil
}

// validatePayload checks if the payload is not nil, and calls
// its validation method.
func (d *DispatcherComponent) validatePayload(payload *types.Payload) error {
	if payload == nil {
		return types.ErrNilPointer.Wrap("payload cannot be nil")
	}
	return payload.Validate()
}

// dispatchActions iterates through all the actions, creates
// the action packets and dispatch them for execution.
func (d *DispatcherComponent) dispatchActions(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	actions []*types.Action,
) error {
	for _, action := range actions {

		packet, err := types.NewActionPacket(transferAttr, action)
		if err != nil {
			return fmt.Errorf("error creating action %s packet: %w", action.ID(), err)
		}

		err = d.dispatchActionPacket(ctx, packet)
		if err != nil {
			return fmt.Errorf("error dispatching action %s packet: %w", action.ID(), err)
		}
	}
	return nil
}

// dispatchOrbit creates the orbit packet and dispatch
// it for execution.
func (d *DispatcherComponent) dispatchOrbit(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	orbit *types.Orbit,
) error {
	packet, err := types.NewOrbitPacket(transferAttr, orbit)
	if err != nil {
		return fmt.Errorf(
			"error creating orbit packet for protocol ID %s: %w",
			packet.Orbit.ProtocolID(),
			err,
		)
	}

	err = d.dispatchOrbitPacket(ctx, packet)
	if err != nil {
		return fmt.Errorf(
			"error dispatching orbit packet for protocol %s: %w",
			packet.Orbit.ProtocolID(),
			err,
		)
	}
	return nil
}

// dispatchActionPacket dispatch the action packet to the
// action handler.
func (d *DispatcherComponent) dispatchActionPacket(
	ctx context.Context,
	packet *types.ActionPacket,
) error {
	return d.ActionHandler.HandlePacket(ctx, packet)
}

// dispatchOrbitPacket dispatch the orbit packet to the
// orbit handler.
func (d *DispatcherComponent) dispatchOrbitPacket(
	ctx context.Context,
	packet *types.OrbitPacket,
) error {
	return d.OrbitHandler.HandlePacket(ctx, packet)
}
