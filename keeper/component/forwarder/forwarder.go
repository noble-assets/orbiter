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

package forwarder

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/event"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
	router "github.com/noble-assets/orbiter/types/router"
)

type ForwardingRouter = router.Router[core.ProtocolID, types.ControllerForwarding]

var _ types.Forwarder = &Forwarder{}

// Forwarder is an Orbiter module component that handles the forwarding
// logic of an Orbiter packet. This component manages its own storage, and
// orcherstrate the controllers of the supported outgoing cross-chain
// transfer bridges.
type Forwarder struct {
	logger       log.Logger
	eventService event.Service
	bankKeeper   types.BankKeeperForwarder
	// router is a forwarding controllers router.
	router *ForwardingRouter
	// PausedController keeps track of the paused protocol ids.
	pausedProtocols collections.KeySet[int32]
	// pausedCrossChains keeps track of the paused protocol id and counterparty id combinations.
	pausedCrossChains collections.KeySet[collections.Pair[int32, string]]
}

// New returns a validated instance of a forwarding component.
func New(
	cdc codec.Codec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	eventService event.Service,
	bankKeeper types.BankKeeperForwarder,
) (*Forwarder, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	f := Forwarder{
		logger:       logger.With(core.ComponentPrefix, core.ForwarderName),
		eventService: eventService,
		bankKeeper:   bankKeeper,
		router:       router.New[core.ProtocolID, types.ControllerForwarding](),
		pausedCrossChains: collections.NewKeySet(
			sb,
			core.PausedCrossChainsPrefix,
			core.PausedCrossChainsName,
			collections.PairKeyCodec(collections.Int32Key, collections.StringKey),
		),
		pausedProtocols: collections.NewKeySet(
			sb,
			core.PausedProtocolsPrefix,
			core.PausedProtocolsName,
			collections.Int32Key,
		),
	}

	return &f, f.Validate()
}

func (f *Forwarder) Validate() error {
	if f.logger == nil {
		return core.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if f.eventService == nil {
		return core.ErrNilPointer.Wrap("event service cannot be nil")
	}
	if f.bankKeeper == nil {
		return core.ErrNilPointer.Wrap("bank keeper cannot be nil")
	}
	if f.router == nil {
		return core.ErrNilPointer.Wrap("controllers router cannot be nil")
	}

	return nil
}

func (f *Forwarder) Logger() log.Logger {
	return f.logger
}

func (f *Forwarder) Router() *ForwardingRouter {
	return f.router
}

func (f *Forwarder) SetRouter(r *ForwardingRouter) error {
	if r == nil {
		return core.ErrNilPointer.Wrap("router cannot be nil")
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
	protocolID core.ProtocolID,
	counterpartyIDs []string,
) error {
	if err := ValidateCrossChains(protocolID, counterpartyIDs); err != nil {
		return core.ErrUnableToPause.Wrap(err.Error())
	}

	switch {
	case len(counterpartyIDs) == 0:
		return f.pauseProtocol(ctx, protocolID)
	default:
		return f.pauseCrossChains(ctx, protocolID, counterpartyIDs)
	}
}

func (f *Forwarder) Unpause(
	ctx context.Context,
	protocolID core.ProtocolID,
	counterpartyIDs []string,
) error {
	if err := ValidateCrossChains(protocolID, counterpartyIDs); err != nil {
		return core.ErrUnableToUnpause.Wrap(err.Error())
	}

	if len(counterpartyIDs) == 0 {
		return f.unpauseProtocol(ctx, protocolID)
	} else {
		return f.unpauseCrossChains(ctx, protocolID, counterpartyIDs)
	}
}

func ValidateCrossChains(
	protocolID core.ProtocolID,
	counterpartyIDs []string,
) error {
	if err := protocolID.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid protocol ID")
	}
	for _, id := range counterpartyIDs {
		if err := core.ValidateCounterpartyID(id); err != nil {
			return errorsmod.Wrap(err, "invalid counterparty ID")
		}
	}

	return nil
}

func (f *Forwarder) HandlePacket(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	if err := f.validatePacket(ctx, packet); err != nil {
		return core.ErrValidation.Wrap(err.Error())
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
	protocolID core.ProtocolID,
	counterpartyID string,
) error {
	if err := f.validateProtocol(ctx, protocolID); err != nil {
		return err
	}

	return f.validateCrossChain(ctx, protocolID, counterpartyID)
}

func (f *Forwarder) validatePacket(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	err := packet.Validate()
	if err != nil {
		return errorsmod.Wrap(err, "error validating forwarding packet")
	}

	attr, err := packet.Forwarding.CachedAttributes()
	if err != nil {
		return errorsmod.Wrap(err, "error getting attributes from forwarding packet")
	}

	err = f.ValidateForwarding(ctx, packet.Forwarding.ProtocolID(), attr.CounterpartyID())
	if err != nil {
		return errorsmod.Wrapf(
			err,
			"error validating forwarding controller for protocol ID %s and counterparty ID %s",
			packet.Forwarding.ProtocolID(),
			attr.CounterpartyID(),
		)
	}

	return f.validateInitialConditions(ctx, packet)
}

func (f *Forwarder) validateProtocol(
	ctx context.Context,
	protocolID core.ProtocolID,
) error {
	isPaused, err := f.IsProtocolPaused(ctx, protocolID)
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

func (f *Forwarder) validateCrossChain(
	ctx context.Context,
	protocolID core.ProtocolID,
	counterpartyID string,
) error {
	ccID, err := core.NewCrossChainID(protocolID, counterpartyID)
	if err != nil {
		return errorsmod.Wrapf(
			err,
			"invalid cross-chain ID for protocol %v and counterparty %s",
			protocolID,
			counterpartyID,
		)
	}
	isPaused, err := f.IsCrossChainPaused(ctx, ccID)
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
	balances := f.bankKeeper.GetAllBalances(ctx, core.ModuleAddress)

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
	protocolID core.ProtocolID,
) error {
	if err := f.SetPausedProtocol(ctx, protocolID); err != nil {
		return errorsmod.Wrapf(err, "error pausing all forwardings for protocol %s", protocolID)
	}

	return nil
}

func (f *Forwarder) pauseCrossChains(
	ctx context.Context,
	protocolID core.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		ccID := core.CrossChainID{ProtocolId: protocolID, CounterpartyId: ID}
		if err := f.SetPausedCrossChain(ctx, ccID); err != nil {
			return errorsmod.Wrapf(
				err,
				"error pausing forwarding for protocol %s and counterparty %s",
				protocolID,
				ID,
			)
		}
	}

	return nil
}

func (f *Forwarder) unpauseProtocol(
	ctx context.Context,
	protocolID core.ProtocolID,
) error {
	if err := f.SetUnpausedProtocol(ctx, protocolID); err != nil {
		return errorsmod.Wrapf(err, "error unpausing all forwardings for protocol %s", protocolID)
	}

	return nil
}

func (f *Forwarder) unpauseCrossChains(
	ctx context.Context,
	protocolID core.ProtocolID,
	counterpartyIDs []string,
) error {
	for _, ID := range counterpartyIDs {
		ccID := core.CrossChainID{ProtocolId: protocolID, CounterpartyId: ID}
		if err := f.SetUnpausedCrossChain(ctx, ccID); err != nil {
			return errorsmod.Wrapf(
				err,
				"error unpausing forwarding for protocol %s and counterparty %s",
				protocolID,
				ID,
			)
		}
	}

	return nil
}
