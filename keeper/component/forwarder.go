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
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
	router "orbiter.dev/types/router"
)

type ForwardingRouter = interfaces.Router[types.ProtocolID, interfaces.ControllerForwarding]

var _ interfaces.Forwarder = &Forwarder{}

type Forwarder struct {
	logger     log.Logger
	bankKeeper types.BankKeeperForwarder
	// router is a forwarding controllers router.
	router ForwardingRouter
	// PausedForwardings keeps track of the paused protocol id and counterparty id combinations.
	PausedForwardings collections.KeySet[collections.Pair[int32, string]]
	// PausedController keeps track of the paused protocol ids.
	PausedControllers collections.KeySet[int32]
}

// NewForwarder returns a validated instance of an forwarding component.
func NewForwarder(
	cdc codec.Codec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	bankKeeper types.BankKeeperForwarder,
) (*Forwarder, error) {
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	forwarder := Forwarder{
		logger:     logger.With(types.ComponentPrefix, types.ForwardingComponentName),
		bankKeeper: bankKeeper,

		router: router.New[types.ProtocolID, interfaces.ControllerForwarding](),
		PausedForwardings: collections.NewKeySet(
			sb,
			types.PausedForwardingPrefix,
			types.PausedForwardingName,
			collections.PairKeyCodec(collections.Int32Key, collections.StringKey),
		),
		PausedControllers: collections.NewKeySet(
			sb,
			types.PausedForwardingControllersPrefix,
			types.PausedForwardingControllersName,
			collections.Int32Key,
		),
	}

	return &forwarder, forwarder.Validate()
}

func (c *Forwarder) Validate() error {
	if c.logger == nil {
		return types.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if c.bankKeeper == nil {
		return types.ErrNilPointer.Wrap("bank keeper cannot be nil")
	}
	if c.router == nil {
		return types.ErrNilPointer.Wrap("controllers router cannot be nil")
	}

	return nil
}

func (c *Forwarder) Logger() log.Logger {
	return c.logger
}

func (c *Forwarder) Router() ForwardingRouter {
	return c.router
}

func (c *Forwarder) SetRouter(ocr ForwardingRouter) error {
	if c.router != nil && c.router.Sealed() {
		return errors.New("cannot reset a sealed router")
	}

	c.router = ocr
	c.router.Seal()

	return nil
}

func (c *Forwarder) Pause(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	switch {
	case len(counterpartyIDs) == 0:
		return c.pauseProtocol(ctx, protocolID)
	default:
		return c.pauseProtocolDestinations(ctx, protocolID, counterpartyIDs)
	}
}

func (c *Forwarder) Unpause(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	if len(counterpartyIDs) == 0 {
		return c.unpauseProtocol(ctx, protocolID)
	} else {
		return c.unpauseProtocolDestinations(ctx, protocolID, counterpartyIDs)
	}
}

func (c *Forwarder) HandlePacket(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	if err := c.validatePacket(ctx, packet); err != nil {
		return types.ErrValidation.Wrap(err.Error())
	}

	controller, found := c.router.Route(packet.Forwarding.ProtocolID())
	if !found {
		return fmt.Errorf(
			"controller not found for forwarding with protocol ID: %s",
			packet.Forwarding.ProtocolID(),
		)
	}

	return controller.HandlePacket(ctx, packet)
}

func (c *Forwarder) ValidateForwarding(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	if err := c.validateController(ctx, protocolID); err != nil {
		return err
	}

	return c.validateForwarding(ctx, protocolID, counterpartyID)
}

func (c *Forwarder) validatePacket(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	err := packet.Validate()
	if err != nil {
		return fmt.Errorf("error validating forwarding packet: %w", err)
	}

	attr, err := packet.Forwarding.CachedAttributes()
	if err != nil {
		return fmt.Errorf("error getting attributes from forwarding packet: %w", err)
	}

	err = c.ValidateForwarding(ctx, packet.Forwarding.ProtocolID(), attr.CounterpartyID())
	if err != nil {
		return fmt.Errorf(
			"error validating forwarding controller for protocol ID %s and counterparty ID %s: %w",
			packet.Forwarding.ProtocolID(), attr.CounterpartyID(), err,
		)
	}

	return c.validateInitialConditions(ctx, packet)
}

func (c *Forwarder) validateController(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	isPaused, err := c.IsControllerPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if isPaused {
		return fmt.Errorf(
			"controller is paused for protocol %v",
			protocolID,
		)
	}

	return nil
}

func (c *Forwarder) validateForwarding(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	isPaused, err := c.IsOrbitPaused(ctx, protocolID, counterpartyID)
	if err != nil {
		return err
	}
	if isPaused {
		return fmt.Errorf(
			"forwarding is paused for protocol %v and counterparty %s",
			protocolID,
			counterpartyID,
		)
	}

	return nil
}

func (c *Forwarder) validateInitialConditions(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	balances := c.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)

	if balances.Len() != 1 {
		return fmt.Errorf("expected exactly 1 balance, got %d", balances.Len())
	}

	if balances[0].Denom != packet.TransferAttributes.DestinationDenom() {
		return fmt.Errorf("denom mismatch: expected %s, got %s",
			packet.TransferAttributes.DestinationDenom(), balances[0].Denom)
	}
	if !balances[0].Amount.Equal(packet.TransferAttributes.DestinationAmount()) {
		return fmt.Errorf("amount mismatch: expected %s, got %s",
			packet.TransferAttributes.DestinationAmount(), balances[0].Amount)
	}

	return nil
}

func (c *Forwarder) pauseProtocol(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	if err := c.SetPausedController(ctx, protocolID); err != nil {
		return fmt.Errorf(
			"error pausing all forwardings for protocol %s: %w",
			protocolID,
			err,
		)
	}

	return nil
}

func (c *Forwarder) pauseProtocolDestinations(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		if err := c.SetPausedOrbit(ctx, protocolID, ID); err != nil {
			return fmt.Errorf(
				"error pausing forwarding for protocol %s and counterparty %s: %w",
				protocolID,
				ID,
				err,
			)
		}
	}

	return nil
}

func (c *Forwarder) unpauseProtocol(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	if err := c.SetUnpausedController(ctx, protocolID); err != nil {
		return fmt.Errorf(
			"error unpausing all forwardings for protocol %s: %w",
			protocolID,
			err,
		)
	}

	return nil
}

func (c *Forwarder) unpauseProtocolDestinations(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		if err := c.SetUnpausedOrbit(ctx, protocolID, ID); err != nil {
			return fmt.Errorf(
				"error unpausing forwarding for protocol %s and counterparty %s: %w",
				protocolID,
				ID,
				err,
			)
		}
	}

	return nil
}
