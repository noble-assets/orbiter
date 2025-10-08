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
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/noble-assets/orbiter/types"
	dispatchertypes "github.com/noble-assets/orbiter/types/component/dispatcher"
	"github.com/noble-assets/orbiter/types/core"
)

var _ types.PayloadDispatcher = &Dispatcher{}

// Dispatcher is a component used to orchestrate the dispatch of an incoming orbiter
// packet. The dispatcher keeps track of the statistics associated with the handled dispatches.
type Dispatcher struct {
	logger log.Logger

	// Packet elements handlers
	ForwardingHandler types.PacketHandler[*types.ForwardingPacket]
	ActionHandler     types.PacketHandler[*types.ActionPacket]
	// Stats
	dispatchedAmounts *collections.IndexedMap[DispatchedAmountsKey, dispatchertypes.AmountDispatched, DispatchedAmountsIndexes]
	dispatchedCounts  *collections.IndexedMap[DispatchedCountsKey, uint64, DispatchedCountsIndexes]
}

// New creates a new validated instance of the dispatcher component.
func New(
	cdc codec.BinaryCodec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	forwardingHandler types.PacketHandler[*types.ForwardingPacket],
	actionHandler types.PacketHandler[*types.ActionPacket],
) (*Dispatcher, error) {
	if cdc == nil {
		return nil, core.ErrNilPointer.Wrap("codec cannot be nil")
	}
	if sb == nil {
		return nil, core.ErrNilPointer.Wrap("schema builder cannot be nil")
	}
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	d := Dispatcher{
		logger:            logger.With(core.ComponentPrefix, core.DispatcherName),
		ForwardingHandler: forwardingHandler,
		ActionHandler:     actionHandler,
		dispatchedAmounts: collections.NewIndexedMap(
			sb,
			core.DispatchedAmountsPrefix,
			core.DispatchedAmountsName,
			collections.QuadKeyCodec(
				collections.Int32Key,
				collections.StringKey,
				collections.StringKey,
				collections.StringKey,
			),
			codec.CollValue[dispatchertypes.AmountDispatched](cdc),
			newDispatchedAmountsIndexes(sb),
		),
		dispatchedCounts: collections.NewIndexedMap(
			sb,
			core.DispatchedCountsPrefix,
			core.DispatchedCountsName,
			collections.QuadKeyCodec(
				collections.Int32Key,
				collections.StringKey,
				collections.Int32Key,
				collections.StringKey,
			),
			collections.Uint64Value,
			newDispatchedCountsIndexes(sb),
		),
	}

	return &d, d.Validate()
}

// Validate checks that the fields of the dispatcher component are valid.
func (d *Dispatcher) Validate() error {
	if d.logger == nil {
		return core.ErrNilPointer.Wrap("logger is not set")
	}
	if d.ForwardingHandler == nil {
		return core.ErrNilPointer.Wrap("forwarding handler is not set")
	}
	if d.ActionHandler == nil {
		return core.ErrNilPointer.Wrap("action handler is not set")
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
	payload *core.Payload,
) error {
	if err := d.ValidatePayload(payload); err != nil {
		return core.ErrValidation.Wrap(err.Error())
	}

	if err := d.dispatchActions(ctx, transferAttr, payload.PreActions); err != nil {
		return errorsmod.Wrap(err, "actions dispatch failed")
	}

	if err := d.dispatchForwarding(ctx, transferAttr, payload.Forwarding); err != nil {
		return errorsmod.Wrap(err, "forwarding dispatch failed")
	}

	if err := d.UpdateStats(ctx, transferAttr, payload.Forwarding); err != nil {
		// NOTE: we don't want to interrupt a dispatch in case the stats are not updated.
		d.logger.Error("Error updating Orbiter statistics", "error", err)
	}

	return nil
}

// ValidatePayload checks if the payload is valid.
func (d *Dispatcher) ValidatePayload(payload *core.Payload) error {
	if err := payload.Validate(); err != nil {
		return err
	}

	return nil
}

// dispatchActions iterates through all the actions, creates
// the action packets and dispatch them for execution.
func (d *Dispatcher) dispatchActions(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	actions []*core.Action,
) error {
	d.logger.Debug("started actions dispatching", "num_actions", len(actions))

	for _, action := range actions {
		packet, err := types.NewActionPacket(transferAttr, action)
		if err != nil {
			return errorsmod.Wrapf(err, "error creating action %s packet", action.ID())
		}

		d.logger.Debug(
			"dispatching action",
			"id",
			action.ID(),
			"dest_denom",
			transferAttr.DestinationDenom(),
			"dest_amount",
			transferAttr.DestinationAmount().String(),
		)
		err = d.dispatchActionPacket(ctx, packet)
		if err != nil {
			return errorsmod.Wrapf(err, "error dispatching action %s packet", action.ID())
		}
	}

	d.logger.Debug("completed actions dispatching")

	return nil
}

// dispatchForwarding creates the forwarding packet and dispatch
// it for execution.
func (d *Dispatcher) dispatchForwarding(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	forwarding *core.Forwarding,
) error {
	d.logger.Debug("started forwarding dispatching")
	packet, err := types.NewForwardingPacket(transferAttr, forwarding)
	if err != nil {
		errDescription := fmt.Sprintf(
			"error creating forwarding packet for protocol ID %s",
			forwarding.ProtocolID(),
		)
		d.logger.Debug(errDescription, "error", err.Error())

		return errorsmod.Wrap(
			err,
			errDescription,
		)
	}

	d.logger.Debug(
		"dispatching forwarding",
		"id",
		forwarding.ProtocolID(),
		"dest_denom",
		transferAttr.DestinationDenom(),
		"dest_amount",
		transferAttr.DestinationAmount().String(),
	)
	err = d.dispatchForwardingPacket(ctx, packet)
	if err != nil {
		errDescription := fmt.Sprintf(
			"error dispatching forwarding packet for protocol ID %s",
			forwarding.ProtocolID(),
		)
		d.logger.Debug(errDescription, "error", err.Error())

		return errorsmod.Wrap(
			err,
			errDescription,
		)
	}

	d.logger.Debug("completed forwarding dispatching")

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
