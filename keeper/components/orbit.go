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

type OrbitRouter = interfaces.Router[types.ProtocolID, interfaces.OrbitController]

var _ interfaces.OrbitComponent = &OrbitComponent{}

type OrbitComponent struct {
	logger log.Logger

	bankKeeper types.BankKeeperOrbit

	router OrbitRouter

	// PausedOrbits maps a protocol id and counterparty id to a boolean indicating
	// whether the orbit is paused or not.
	PausedOrbits collections.Map[collections.Pair[int32, string], bool]
	// PausedController maps a protocol id to a boolean indicating
	// whether the protocol controller is paused or not.
	PausedControllers collections.Map[int32, bool]
}

func NewOrbitComponent(
	cdc codec.Codec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	bankKeeper types.BankKeeperOrbit,
) (*OrbitComponent, error) {
	if logger == nil {
		return nil, errors.New("logger cannot be nil")
	}

	orbitsKeeper := OrbitComponent{
		logger:     logger.With(types.ComponentPrefix, types.OrbitsKeeperName),
		bankKeeper: bankKeeper,

		router: router.New[types.ProtocolID, interfaces.OrbitController](),

		PausedOrbits: collections.NewMap(
			sb,
			types.PausedOrbitPrefix,
			types.PausedOrbitsName,
			collections.PairKeyCodec(collections.Int32Key, collections.StringKey),
			collections.BoolValue,
		),
		PausedControllers: collections.NewMap(
			sb,
			types.PausedOrbitControllersPrefix,
			types.PausedOrbitControllersName,
			collections.Int32Key,
			collections.BoolValue,
		),
	}

	return &orbitsKeeper, orbitsKeeper.Validate()
}

func (k *OrbitComponent) Validate() error {
	if k.logger == nil {
		return errors.New("logger cannot be nil")
	}
	if k.bankKeeper == nil {
		return errors.New("bank keeper cannot be nil")
	}
	if k.router == nil {
		return errors.New("controllers router cannot be nil")
	}
	return nil
}

func (k *OrbitComponent) Logger() log.Logger {
	return k.logger
}

func (k *OrbitComponent) Router() OrbitRouter {
	return k.router
}

func (k *OrbitComponent) SetRouter(ocr OrbitRouter) {
	if k.router != nil && k.router.Sealed() {
		panic(errors.New("cannot reset a sealed controller router"))
	}

	k.router = ocr
	k.router.Seal()
}

func (k *OrbitComponent) HandlePacket(ctx context.Context, packet *types.OrbitPacket) error {
	if err := k.ValidatePacket(ctx, packet); err != nil {
		return err
	}

	c, found := k.router.Route(packet.Orbit.ProtocolID())
	if !found {
		return errors.New("controller is not registered")
	}

	return c.HandlePacket(ctx, packet)
}

func (k *OrbitComponent) ValidatePacket(ctx context.Context, packet *types.OrbitPacket) error {
	err := packet.Validate()
	if err != nil {
		return err
	}

	attr, err := packet.Orbit.CachedAttributes()
	if err != nil {
		return err
	}

	err = k.ValidateOrbit(ctx, packet.Orbit.ProtocolID(), attr.CounterpartyID())
	if err != nil {
		return err
	}

	return k.validateInitialConditions(ctx, packet)
}

func (k *OrbitComponent) ValidateOrbit(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	if err := k.validateController(ctx, protocolID); err != nil {
		return err
	}
	return k.validateOrbit(ctx, protocolID, counterpartyID)
}

func (k *OrbitComponent) validateController(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	isPaused, err := k.IsControllerPaused(ctx, protocolID)
	if err != nil {
		return err
	}
	if isPaused {
		return errors.New("controller is paused")
	}

	return nil
}

func (k *OrbitComponent) validateOrbit(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	return nil
}

func (k *OrbitComponent) validateInitialConditions(
	ctx context.Context,
	packet *types.OrbitPacket,
) error {
	balances := k.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)

	if balances.Len() != 1 {
		return errors.New("wrong balance")
	}

	if balances[0].Denom != packet.TransferAttributes.DestinationDenom() {
		return errors.New("wrong denom")
	}
	if !balances[0].Amount.Equal(packet.TransferAttributes.DestinationAmount()) {
		return errors.New("wrong amount")
	}

	return nil
}

func (k *OrbitComponent) Pause(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	if len(counterpartyIDs) == 0 {
		return k.pauseProtocol(ctx, protocolID)
	} else {
		return k.pauseProtocolDestinations(ctx, protocolID, counterpartyIDs)
	}
}

func (k *OrbitComponent) Unpause(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	if len(counterpartyIDs) == 0 {
		return k.unpauseProtocol(ctx, protocolID)
	} else {
		return k.unpauseProtocolDestinations(ctx, protocolID, counterpartyIDs)
	}
}

func (k *OrbitComponent) pauseProtocol(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	if err := k.SetPausedController(ctx, protocolID); err != nil {
		return fmt.Errorf(
			"error pausing all orbits for protocol %s: %w",
			protocolID,
			err,
		)
	}
	return nil
}

func (k *OrbitComponent) pauseProtocolDestinations(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		if err := k.SetPausedOrbit(ctx, protocolID, ID); err != nil {
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

func (k *OrbitComponent) unpauseProtocol(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	if err := k.SetUnpausedController(ctx, protocolID); err != nil {
		return fmt.Errorf(
			"error unpausing all orbits for protocol %s: %w",
			protocolID,
			err,
		)
	}
	return nil
}

func (k *OrbitComponent) unpauseProtocolDestinations(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		if err := k.SetUnpausedOrbit(ctx, protocolID, ID); err != nil {
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
