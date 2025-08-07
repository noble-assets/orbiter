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

package keeper

import (
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"orbiter.dev/keeper/component"
	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

// Keeper is the main module keeper.
type Keeper struct {
	cdc    codec.Codec
	logger log.Logger

	// authority represents the module manager.
	authority string

	// Components.
	executor   interfaces.Executor
	forwarder  interfaces.Forwarder
	dispatcher interfaces.Dispatcher
	adapter    interfaces.Adapter
}

// NewKeeper returns a reference to a validated instance of the keeper.
// Panic if the keeper initialization fails.
func NewKeeper(
	cdc codec.Codec,
	addressCdc address.Codec,
	logger log.Logger,
	storeService store.KVStoreService,
	authority string,
	bankKeeper types.BankKeeper,
) *Keeper {
	if err := validateKeeperInputs(cdc, addressCdc, logger, storeService, authority); err != nil {
		panic(err)
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:       cdc,
		logger:    logger.With("module", fmt.Sprintf("x/%s", types.ModuleName)),
		authority: authority,
	}

	if err := k.setComponents(k.cdc, k.logger, sb, bankKeeper); err != nil {
		panic(err)
	}

	if _, err := sb.Build(); err != nil {
		panic(err)
	}

	if err := k.Validate(); err != nil {
		panic(err)
	}

	return &k
}

// validateKeeperInputs check that all Keeper inputs
// are valid or panic.
func validateKeeperInputs(
	cdc codec.Codec,
	addressCdc address.Codec,
	logger log.Logger,
	storeService store.KVStoreService,
	authority string,
) error {
	if cdc == nil {
		return errors.New("codec cannot be nil")
	}
	if logger == nil {
		return errors.New("logger cannot be nil")
	}
	if storeService == nil {
		return errors.New("store service cannot be nil")
	}
	if addressCdc == nil {
		return errors.New("address codec cannot be nil")
	}
	_, err := addressCdc.StringToBytes(authority)
	if err != nil {
		return errors.New("authority for x/orbiter module is not valid")
	}

	return nil
}

// Validate returns an error if any of the keeper fields is not valid.
func (k *Keeper) Validate() error {
	if k.logger == nil {
		return errors.New("logger cannot be nil")
	}
	if k.cdc == nil {
		return errors.New("codec cannot be nil")
	}

	return nil
}

func (k *Keeper) Codec() codec.Codec {
	return k.cdc
}

func (k *Keeper) Logger() log.Logger {
	return k.logger
}

func (k *Keeper) Authority() string {
	return k.authority
}

func (k *Keeper) Executor() interfaces.Executor {
	return k.executor
}

func (k *Keeper) Forwarder() interfaces.Forwarder {
	return k.forwarder
}

func (k *Keeper) Dispatcher() interfaces.PayloadDispatcher {
	return k.dispatcher
}

func (k *Keeper) Adapter() interfaces.Adapter {
	return k.adapter
}

func (k *Keeper) SetForwardingControllers(controllers ...interfaces.ControllerForwarding) {
	router := k.forwarder.Router()
	for _, c := range controllers {
		if err := router.AddRoute(c); err != nil {
			panic(err)
		}
	}
	if err := k.forwarder.SetRouter(router); err != nil {
		panic(err)
	}
}

func (k *Keeper) SetActionControllers(controllers ...interfaces.ControllerAction) {
	router := k.executor.Router()
	for _, c := range controllers {
		if err := router.AddRoute(c); err != nil {
			panic(err)
		}
	}
	if err := k.executor.SetRouter(router); err != nil {
		panic(err)
	}
}

func (k *Keeper) SetAdapterControllers(controllers ...interfaces.ControllerAdapter) {
	router := k.adapter.Router()
	for _, c := range controllers {
		if err := router.AddRoute(c); err != nil {
			panic(err)
		}
	}
	if err := k.adapter.SetRouter(router); err != nil {
		panic(err)
	}
}

// CheckIsAuthority returns an error is the signer is not the
// keeper authority.
func (k *Keeper) CheckIsAuthority(signer string) error {
	if k.Authority() != signer {
		return types.ErrUnauthorized
	}

	return nil
}

// setComponents registers all required components in the orbiter keeper.
func (k *Keeper) setComponents(
	cdc codec.Codec,
	logger log.Logger,
	sb *collections.SchemaBuilder,
	bankKeeper types.BankKeeper,
) error {
	executor, err := component.NewExecutor(cdc, sb, logger)
	if err != nil {
		return fmt.Errorf("error creating a new action component: %w", err)
	}

	forwarder, err := component.NewForwarder(cdc, sb, logger, bankKeeper)
	if err != nil {
		return fmt.Errorf("error creating a new forwarding component: %w", err)
	}

	dispatcher, err := component.NewDispatcher(
		cdc,
		sb,
		logger,
		forwarder,
		executor,
	)
	if err != nil {
		return fmt.Errorf("error creating a new dispatcher component: %w", err)
	}

	adapter, err := component.NewAdapter(logger, bankKeeper, dispatcher)
	if err != nil {
		return fmt.Errorf("error creating a new adapter component: %w", err)
	}

	k.executor = executor
	k.forwarder = forwarder
	k.dispatcher = dispatcher
	k.adapter = adapter

	return nil
}
