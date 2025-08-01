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
	"orbiter.dev/types/router"
)

type ActionRouter = interfaces.Router[types.ActionID, interfaces.ControllerAction]

var _ interfaces.ActionComponent = &ActionComponent{}

type ActionComponent struct {
	logger log.Logger
	// router is an action controllers router.
	router ActionRouter
	// PausedControllers keeps track of the ids of paused actions.
	PausedControllers collections.KeySet[int32]
}

// NewActionComponent returns a validated instance of an action component.
func NewActionComponent(
	cdc codec.Codec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
) (*ActionComponent, error) {
	actionComponent := ActionComponent{
		logger: logger.With(types.ComponentPrefix, types.ActionComponentName),
		router: router.New[types.ActionID, interfaces.ControllerAction](),
		PausedControllers: collections.NewKeySet(
			sb,
			types.PausedActionControllersPrefix,
			types.PausedActionControllersName,
			collections.Int32Key,
		),
	}

	return &actionComponent, actionComponent.Validate()
}

// Validate returns an error if the component instance is not valid.
func (c *ActionComponent) Validate() error {
	if c.logger == nil {
		return types.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if c.router == nil {
		return types.ErrNilPointer.Wrap("router cannot be nil")
	}

	return nil
}

func (c *ActionComponent) Logger() log.Logger {
	return c.logger
}

func (c *ActionComponent) Router() ActionRouter {
	return c.router
}

func (c *ActionComponent) SetRouter(acr ActionRouter) error {
	if c.router != nil && c.router.Sealed() {
		return errors.New("cannot reset a sealed router")
	}

	c.router = acr
	c.router.Seal()

	return nil
}

// Pause allows to pause an action controller.
func (c *ActionComponent) Pause(ctx context.Context, actionID types.ActionID) error {
	if err := c.SetPausedController(ctx, actionID); err != nil {
		return fmt.Errorf(
			"error pausing action %s: %w",
			actionID,
			err,
		)
	}

	return nil
}

// Unpause allows to unpause an action controller.
func (c *ActionComponent) Unpause(ctx context.Context, actionID types.ActionID) error {
	if err := c.SetUnpausedController(ctx, actionID); err != nil {
		return fmt.Errorf(
			"error unpausing action %s: %w",
			actionID,
			err,
		)
	}

	return nil
}

func (c *ActionComponent) HandlePacket(
	ctx context.Context,
	packet *types.ActionPacket,
) error {
	if err := c.validatePacket(ctx, packet); err != nil {
		return types.ErrValidation.Wrap(err.Error())
	}

	controller, found := c.router.Route(packet.Action.ID())
	if !found {
		return fmt.Errorf("controller not found for action ID: %s", packet.Action.ID())
	}

	return controller.HandlePacket(ctx, packet)
}

func (c *ActionComponent) validatePacket(ctx context.Context, packet *types.ActionPacket) error {
	err := packet.Validate()
	if err != nil {
		return fmt.Errorf("error validating action packet: %w", err)
	}

	err = c.validateController(ctx, packet.Action.ID())
	if err != nil {
		return fmt.Errorf("error validating action controller: %w", err)
	}

	return nil
}

// validateController returns an error if the controller associated with
// the action ID is not valid.
func (c *ActionComponent) validateController(
	ctx context.Context,
	id types.ActionID,
) error {
	isPaused, err := c.IsControllerPaused(ctx, id)
	if err != nil {
		return err
	}
	if isPaused {
		return fmt.Errorf("action id %s is paused", id)
	}

	return nil
}
