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
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
	router "orbiter.dev/types/router"
)

type OrbitRouter = interfaces.Router[types.ProtocolID, interfaces.ControllerOrbit]

var _ interfaces.OrbitComponent = &OrbitComponent{}

type OrbitComponent struct {
	logger     log.Logger
	bankKeeper types.BankKeeperOrbit
	// router is an orbit controllers router.
	router OrbitRouter
	// PausedOrbits keeps track of the paused protocol id and counterparty id combinations.
	PausedOrbits collections.KeySet[collections.Pair[int32, string]]
	// PausedController keeps track of the paused protocol ids.
	PausedControllers collections.KeySet[int32]
}

// NewOrbitComponent returns a validated instance of an orbit component.
func NewOrbitComponent(
	cdc codec.Codec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	bankKeeper types.BankKeeperOrbit,
) (*OrbitComponent, error) {
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	orbitComponent := OrbitComponent{
		logger:     logger.With(types.ComponentPrefix, types.OrbitComponentName),
		bankKeeper: bankKeeper,

		router: router.New[types.ProtocolID, interfaces.ControllerOrbit](),
		PausedOrbits: collections.NewKeySet(
			sb,
			types.PausedOrbitPrefix,
			types.PausedOrbitsName,
			collections.PairKeyCodec(collections.Int32Key, collections.StringKey),
		),
		PausedControllers: collections.NewKeySet(
			sb,
			types.PausedOrbitControllersPrefix,
			types.PausedOrbitControllersName,
			collections.Int32Key,
		),
	}

	return &orbitComponent, orbitComponent.Validate()
}

func (c *OrbitComponent) Validate() error {
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

func (c *OrbitComponent) Logger() log.Logger {
	return c.logger
}

func (c *OrbitComponent) Router() OrbitRouter {
	return c.router
}

func (c *OrbitComponent) SetRouter(ocr OrbitRouter) error {
	if c.router != nil && c.router.Sealed() {
		return errors.New("cannot reset a sealed router")
	}

	c.router = ocr
	c.router.Seal()
	return nil
}

func (c *OrbitComponent) HandlePacket(ctx context.Context, packet *types.OrbitPacket) error {
	if err := c.validatePacket(ctx, packet); err != nil {
		return types.ErrValidation.Wrap(err.Error())
	}

	controller, found := c.router.Route(packet.Orbit.ProtocolID())
	if !found {
		return fmt.Errorf(
			"controller not found for orbit with protocol ID: %s",
			packet.Orbit.ProtocolID(),
		)
	}

	return controller.HandlePacket(ctx, packet)
}

func (c *OrbitComponent) validatePacket(ctx context.Context, packet *types.OrbitPacket) error {
	err := packet.Validate()
	if err != nil {
		return fmt.Errorf("error validating orbit packet: %w", err)
	}

	attr, err := packet.Orbit.CachedAttributes()
	if err != nil {
		return fmt.Errorf("error getting attributes from orbit packet: %w", err)
	}

	err = c.ValidateOrbit(ctx, packet.Orbit.ProtocolID(), attr.CounterpartyID())
	if err != nil {
		return fmt.Errorf(
			"error validating orbit controller for protocol ID %s and counterparty ID %s: %w",
			packet.Orbit.ProtocolID(), attr.CounterpartyID(), err,
		)
	}

	return c.validateInitialConditions(ctx, packet)
}

func (c *OrbitComponent) ValidateOrbit(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	if err := c.validateController(ctx, protocolID); err != nil {
		return err
	}
	return c.validateOrbit(ctx, protocolID, counterpartyID)
}

func (c *OrbitComponent) validateController(
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

func (c *OrbitComponent) validateOrbit(
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
			"orbit is paused for protocol %v and counterparty %s",
			protocolID,
			counterpartyID,
		)
	}
	return nil
}

func (c *OrbitComponent) validateInitialConditions(
	ctx context.Context,
	packet *types.OrbitPacket,
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

func (c *OrbitComponent) Pause(
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

func (c *OrbitComponent) Unpause(
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

func (c *OrbitComponent) pauseProtocol(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	if err := c.SetPausedController(ctx, protocolID); err != nil {
		return fmt.Errorf(
			"error pausing all orbits for protocol %s: %w",
			protocolID,
			err,
		)
	}
	return nil
}

func (c *OrbitComponent) pauseProtocolDestinations(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		if err := c.SetPausedOrbit(ctx, protocolID, ID); err != nil {
			return fmt.Errorf(
				"error pausing orbit for protocol %s and counterparty %s: %w",
				protocolID,
				ID,
				err,
			)
		}
	}
	return nil
}

func (c *OrbitComponent) unpauseProtocol(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	if err := c.SetUnpausedController(ctx, protocolID); err != nil {
		return fmt.Errorf(
			"error unpausing all orbits for protocol %s: %w",
			protocolID,
			err,
		)
	}
	return nil
}

func (c *OrbitComponent) unpauseProtocolDestinations(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		if err := c.SetUnpausedOrbit(ctx, protocolID, ID); err != nil {
			return fmt.Errorf(
				"error unpausing orbit for protocol %s and counterparty %s: %w",
				protocolID,
				ID,
				err,
			)
		}
	}
	return nil
}
