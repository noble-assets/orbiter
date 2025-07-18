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

package subkeepers

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/controllers"
	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

type ActionControllersRouter = interfaces.Router[types.ActionID, interfaces.ActionController]

var _ interfaces.ActionSubkeeper = &ActionKeeper{}

type ActionKeeper struct {
	logger log.Logger

	controllersRouter ActionControllersRouter

	// PausedControllers maps an action id to a boolean indicating
	// whether the action controller is paused or not.
	PausedControllers collections.Map[int32, bool]
}

func NewActionKeeper(
	cdc codec.Codec,
	sb *collections.SchemaBuilder,
	logger log.Logger,
) (*ActionKeeper, error) {
	actionsKeeper := ActionKeeper{
		logger: logger.With(types.SubKeeperPrefix, types.ActionsKeeperName),

		controllersRouter: controllers.NewRouter[types.ActionID, interfaces.ActionController](),

		PausedControllers: collections.NewMap(
			sb,
			types.PausedActionControllersPrefix,
			types.PausedActionControllersName,
			collections.Int32Key,
			collections.BoolValue,
		),
	}

	return &actionsKeeper, actionsKeeper.Validate()
}

func (k *ActionKeeper) Validate() error {
	if k.logger == nil {
		return errors.New("logger cannot be nil")
	}
	if k.controllersRouter == nil {
		return errors.New("controllers router cannot be nil")
	}
	return nil
}

func (k *ActionKeeper) Logger() log.Logger {
	return k.logger
}

func (k *ActionKeeper) Router() ActionControllersRouter {
	return k.controllersRouter
}

func (k *ActionKeeper) SetRouter(acr ActionControllersRouter) {
	if k.controllersRouter != nil && k.controllersRouter.Sealed() {
		panic(errors.New("cannot reset a sealed controller router"))
	}

	k.controllersRouter = acr
	k.controllersRouter.Seal()
}

func (k *ActionKeeper) HandlePacket(
	ctx context.Context,
	packet *types.ActionPacket,
) error {
	if err := k.ValidatePacket(ctx, packet); err != nil {
		return err
	}

	c, found := k.controllersRouter.Route(packet.Action.ID())
	if !found {
		return errors.New("controller is not registered")
	}

	return c.HandlePacket(ctx, packet)
}

func (k *ActionKeeper) ValidatePacket(ctx context.Context, packet *types.ActionPacket) error {
	err := packet.Validate()
	if err != nil {
		return err
	}

	err = k.validateController(ctx, packet.Action.ID())
	if err != nil {
		return err
	}
	return nil
}

func (k *ActionKeeper) validateController(
	ctx context.Context,
	id types.ActionID,
) error {
	isPaused, err := k.IsControllerPaused(ctx, id)
	if err != nil {
		return err
	}
	if isPaused {
		return errors.New("action is paused")
	}

	return nil
}

// Pause implements types.ActionsSubkeeper.
func (k *ActionKeeper) Pause(ctx context.Context, actionID types.ActionID) error {
	if err := k.SetPausedController(ctx, actionID); err != nil {
		return fmt.Errorf(
			"error pausing action %s: %w",
			actionID,
			err,
		)
	}
	return nil
}

// Unpause implements types.ActionsSubkeeper.
func (k *ActionKeeper) Unpause(ctx context.Context, actionID types.ActionID) error {
	if err := k.SetUnpausedController(ctx, actionID); err != nil {
		return fmt.Errorf(
			"error unpausing action %s: %w",
			actionID,
			err,
		)
	}
	return nil
}
