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

func (f *Forwarder) Validate() error {
	if f.logger == nil {
		return types.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if f.bankKeeper == nil {
		return types.ErrNilPointer.Wrap("bank keeper cannot be nil")
	}
	if f.router == nil {
		return types.ErrNilPointer.Wrap("controllers router cannot be nil")
	}

	return nil
}

func (f *Forwarder) Logger() log.Logger {
	return f.logger
}

func (f *Forwarder) Router() ForwardingRouter {
	return f.router
}

func (f *Forwarder) SetRouter(r ForwardingRouter) error {
	if r == nil {
		return types.ErrNilPointer.Wrap("router cannot be nil")
	}

	if f.router != nil && f.router.Sealed() {
		return errors.New("cannot reset a sealed router")
	}

	f.router = r
	f.router.Seal()

	return nil
}

func (f *Forwarder) Pause(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	switch {
	case len(counterpartyIDs) == 0:
		return f.pauseProtocol(ctx, protocolID)
	default:
		return f.pauseProtocolDestinations(ctx, protocolID, counterpartyIDs)
	}
}

func (f *Forwarder) Unpause(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	if len(counterpartyIDs) == 0 {
		return f.unpauseProtocol(ctx, protocolID)
	} else {
		return f.unpauseProtocolDestinations(ctx, protocolID, counterpartyIDs)
	}
}

func (f *Forwarder) HandlePacket(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	if err := f.validatePacket(ctx, packet); err != nil {
		return types.ErrValidation.Wrap(err.Error())
	}

	controller, found := f.router.Route(packet.Forwarding.ProtocolID())
	if !found {
		return fmt.Errorf(
			"controller not found for forwarding with protocol ID: %s",
			packet.Forwarding.ProtocolID(),
		)
	}

	return controller.HandlePacket(ctx, packet)
}

func (f *Forwarder) ValidateForwarding(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	if err := f.validateController(ctx, protocolID); err != nil {
		return err
	}

	return f.validateForwarding(ctx, protocolID, counterpartyID)
}

func (f *Forwarder) validatePacket(
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

	err = f.ValidateForwarding(ctx, packet.Forwarding.ProtocolID(), attr.CounterpartyID())
	if err != nil {
		return fmt.Errorf(
			"error validating forwarding controller for protocol ID %s and counterparty ID %s: %w",
			packet.Forwarding.ProtocolID(), attr.CounterpartyID(), err,
		)
	}

	return f.validateInitialConditions(ctx, packet)
}

func (f *Forwarder) validateController(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	isPaused, err := f.IsControllerPaused(ctx, protocolID)
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

func (f *Forwarder) validateForwarding(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyID string,
) error {
	isPaused, err := f.IsOrbitPaused(ctx, protocolID, counterpartyID)
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

func (f *Forwarder) validateInitialConditions(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	balances := f.bankKeeper.GetAllBalances(ctx, types.ModuleAddress)

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

func (f *Forwarder) pauseProtocol(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	if err := f.SetPausedController(ctx, protocolID); err != nil {
		return fmt.Errorf(
			"error pausing all forwardings for protocol %s: %w",
			protocolID,
			err,
		)
	}

	return nil
}

func (f *Forwarder) pauseProtocolDestinations(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		if err := f.SetPausedOrbit(ctx, protocolID, ID); err != nil {
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

func (f *Forwarder) unpauseProtocol(
	ctx context.Context,
	protocolID types.ProtocolID,
) error {
	if err := f.SetUnpausedController(ctx, protocolID); err != nil {
		return fmt.Errorf(
			"error unpausing all forwardings for protocol %s: %w",
			protocolID,
			err,
		)
	}

	return nil
}

func (f *Forwarder) unpauseProtocolDestinations(
	ctx context.Context,
	protocolID types.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		if err := f.SetUnpausedOrbit(ctx, protocolID, ID); err != nil {
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
