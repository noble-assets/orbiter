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

package executor

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/event"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
	"github.com/noble-assets/orbiter/types/router"
)

type ActionRouter = *router.Router[core.ActionID, types.ActionController]

var _ types.Executor = &Executor{}

type Executor struct {
	logger       log.Logger
	eventService event.Service
	// router is an action controllers router.
	router ActionRouter
	// PausedActions keeps track of the ids of paused actions.
	PausedActions collections.KeySet[int32]
}

// New returns a validated instance of an executor component.
func New(
	cdc codec.Codec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
	eventService event.Service,
) (*Executor, error) {
	if cdc == nil {
		return nil, core.ErrNilPointer.Wrap("codec cannot be nil")
	}
	if sb == nil {
		return nil, core.ErrNilPointer.Wrap("schema builder cannot be nil")
	}
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	executor := Executor{
		logger:       logger.With(core.ComponentPrefix, core.ExecutorName),
		eventService: eventService,
		router:       router.New[core.ActionID, types.ActionController](),
		PausedActions: collections.NewKeySet(
			sb,
			core.PausedActionsPrefix,
			core.PausedActionsName,
			collections.Int32Key,
		),
	}

	return &executor, executor.Validate()
}

// Validate returns an error if the component instance is not valid.
func (e *Executor) Validate() error {
	if e.logger == nil {
		return core.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if e.eventService == nil {
		return core.ErrNilPointer.Wrap("event service cannot be nil")
	}
	if e.router == nil {
		return core.ErrNilPointer.Wrap("router cannot be nil")
	}

	return nil
}

func (e *Executor) Logger() log.Logger {
	return e.logger
}

func (e *Executor) EventService() event.Service {
	return e.eventService
}

func (e *Executor) Router() ActionRouter {
	return e.router
}

func (e *Executor) SetRouter(r ActionRouter) error {
	if r == nil {
		return core.ErrNilPointer.Wrap("router cannot be nil")
	}

	if e.router != nil && e.router.Sealed() {
		return errors.New("cannot reset a sealed router")
	}

	e.router = r
	e.router.Seal()

	return nil
}

// Pause allows to pause an action controller.
func (e *Executor) Pause(ctx context.Context, actionID core.ActionID) error {
	if err := e.SetPausedAction(ctx, actionID); err != nil {
		return errorsmod.Wrapf(err, "error pausing action %s", actionID)
	}

	return nil
}

// Unpause allows to unpause an action controller.
func (e *Executor) Unpause(ctx context.Context, actionID core.ActionID) error {
	if err := e.SetUnpausedAction(ctx, actionID); err != nil {
		return errorsmod.Wrapf(err, "error unpausing action %s", actionID)
	}

	return nil
}

func (e *Executor) HandlePacket(
	ctx context.Context,
	packet *types.ActionPacket,
) error {
	if err := e.validatePacket(ctx, packet); err != nil {
		return core.ErrValidation.Wrap(err.Error())
	}

	actionID := packet.Action.ID()
	controller, found := e.router.Route(actionID)
	if !found {
		return sdkerrors.ErrNotFound.Wrapf("controller for action ID: %s", actionID)
	}

	return controller.HandlePacket(ctx, packet)
}

func (e *Executor) validatePacket(ctx context.Context, packet *types.ActionPacket) error {
	err := packet.Validate()
	if err != nil {
		return errorsmod.Wrap(err, "error validating action packet")
	}

	err = e.validateController(ctx, packet.Action.ID())
	if err != nil {
		return errorsmod.Wrap(err, "error validating action controller")
	}

	return nil
}

// validateController returns an error if the controller associated with
// the action ID is not valid.
func (e *Executor) validateController(
	ctx context.Context,
	id core.ActionID,
) error {
	isPaused, err := e.IsActionPaused(ctx, id)
	if err != nil {
		return err
	}
	if isPaused {
		return fmt.Errorf("action ID %s is paused", id)
	}

	return nil
}
