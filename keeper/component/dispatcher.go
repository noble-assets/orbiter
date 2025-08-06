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
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

var _ interfaces.PayloadDispatcher = &Dispatcher{}

// Dispatcher is a component used to orchestrate the
// dispatch of an incoming orbiter packet. The dispatcher
// keeps track of the statistics associated with the handled dispatches.
type Dispatcher struct {
	logger log.Logger
	// Packet elements handlers
	ForwardingHandler interfaces.PacketHandler[*types.ForwardingPacket]
	ActionHandler     interfaces.PacketHandler[*types.ActionPacket]
	// Stats
	DispatchedAmounts *collections.IndexedMap[DispatchedAmountsKey, types.AmountDispatched, DispatchedAmountsIndexes]
	DispatchCounts    *collections.IndexedMap[DispatchedCountsKey, uint32, DispatchedCountsIndexes]
}

// NewDispatcher creates a new validated instance of a the dispatcher
// component.
func NewDispatcher(
	cdc codec.BinaryCodec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	forwardingHandler interfaces.PacketHandler[*types.ForwardingPacket],
	actionHandler interfaces.PacketHandler[*types.ActionPacket],
) (*Dispatcher, error) {
	if cdc == nil {
		return nil, types.ErrNilPointer.Wrap("codec cannot be nil")
	}
	if sb == nil {
		return nil, types.ErrNilPointer.Wrap("schema builder cannot be nil")
	}
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	dispatcherComponent := Dispatcher{
		logger:            logger.With(types.ComponentPrefix, types.DispatcherComponentName),
		ForwardingHandler: forwardingHandler,
		ActionHandler:     actionHandler,
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
func (d *Dispatcher) Validate() error {
	if d.ForwardingHandler == nil {
		return types.ErrNilPointer.Wrap("forwarding handler cannot be nil")
	}
	if d.ActionHandler == nil {
		return types.ErrNilPointer.Wrap("actions handler cannot be nil")
	}

	return nil
}

func (d *Dispatcher) Logger() log.Logger {
	return d.logger
}

// DispatchPayload is the entry point to initiate the dispatching
// of an orbiter payload.
func (d *Dispatcher) DispatchPayload(
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

	if err := d.dispatchForwarding(ctx, transferAttr, payload.Forwarding); err != nil {
		return fmt.Errorf("forwarding dispatch failed: %w", err)
	}

	if err := d.UpdateStats(ctx, transferAttr, payload.Forwarding); err != nil {
		d.logger.Error("Error updating Orbiter statistics", "error", err.Error())
	}

	return nil
}

// validatePayload checks if the payload is not nil, and calls
// its validation method.
func (d *Dispatcher) validatePayload(payload *types.Payload) error {
	if payload == nil {
		return types.ErrNilPointer.Wrap("payload cannot be nil")
	}

	return payload.Validate()
}

// dispatchActions iterates through all the actions, creates
// the action packets and dispatch them for execution.
func (d *Dispatcher) dispatchActions(
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

// dispatchForwarding creates the forwarding packet and dispatch
// it for execution.
func (d *Dispatcher) dispatchForwarding(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	forwarding *types.Forwarding,
) error {
	packet, err := types.NewForwardingPacket(transferAttr, forwarding)
	if err != nil {
		return fmt.Errorf(
			"error creating forwarding packet for protocol ID %s: %w",
			packet.Forwarding.ProtocolID(),
			err,
		)
	}

	err = d.dispatchForwardingPacket(ctx, packet)
	if err != nil {
		return fmt.Errorf(
			"error dispatching forwarding packet for protocol %s: %w",
			packet.Forwarding.ProtocolID(),
			err,
		)
	}

	return nil
}

// dispatchActionPacket dispatch the action packet to the
// action handler.
func (d *Dispatcher) dispatchActionPacket(
	ctx context.Context,
	packet *types.ActionPacket,
) error {
	return d.ActionHandler.HandlePacket(ctx, packet)
}

// dispatchForwardingPacket dispatch the forwarding packet to the
// forwarding handler.
func (d *Dispatcher) dispatchForwardingPacket(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	return d.ForwardingHandler.HandlePacket(ctx, packet)
}
