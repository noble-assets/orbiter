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
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	adaptercomp "github.com/noble-assets/orbiter/keeper/component/adapter"
	dispatchercomp "github.com/noble-assets/orbiter/keeper/component/dispatcher"
	executorcomp "github.com/noble-assets/orbiter/keeper/component/executor"
	forwardercomp "github.com/noble-assets/orbiter/keeper/component/forwarder"
	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var _ orbitertypes.Authorizer = &Keeper{}

// Keeper is the main module keeper.
type Keeper struct {
	cdc          codec.Codec
	logger       log.Logger
	eventService event.Service

	// authority represents the module manager.
	authority string

	// Each component manages its own state.
	executor   *executorcomp.Executor
	forwarder  *forwardercomp.Forwarder
	dispatcher *dispatchercomp.Dispatcher
	adapter    *adaptercomp.Adapter

	// pendingPayloads stores the pending payloads addressed by their sha256 hash.
	pendingPayloads collections.Map[[]byte, core.PendingPayload]
	// payloadHashesByTime stores the registered payload hashes by the block time when they were processed.
	payloadHashesByTime collections.Map[collections.Pair[int64, []byte], struct{}]
	// pendingPayloadsSequence is the unique identifier of a given pending payload handled by the
	// orbiter.
	pendingPayloadsSequence collections.Sequence
}

// NewKeeper returns a reference to a validated instance of the keeper.
// Panic if the keeper initialization fails.
func NewKeeper(
	cdc codec.Codec,
	addressCdc address.Codec,
	logger log.Logger,
	eventService event.Service,
	storeService store.KVStoreService,
	authority string,
	bankKeeper orbitertypes.BankKeeper,
) *Keeper {
	if err := validateKeeperInputs(cdc, addressCdc, logger, eventService, storeService, bankKeeper, authority); err != nil {
		panic(err)
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:          cdc,
		eventService: eventService,
		logger:       logger.With("module", fmt.Sprintf("x/%s", core.ModuleName)),
		authority:    authority,

		pendingPayloads: collections.NewMap[
			[]byte,
			core.PendingPayload,
		](
			sb,
			core.PendingPayloadsPrefix,
			core.PendingPayloadsName,
			collections.BytesKey,
			codec.CollValue[core.PendingPayload](cdc),
		),
		payloadHashesByTime: collections.NewMap[collections.Pair[int64, []byte], struct{}](
			sb,
			core.PayloadHashesByTimePrefix,
			core.PayloadHashesByTimeName,
			collections.PairKeyCodec(collections.Int64Key, collections.BytesKey),
			collections.NoValue,
		),
		pendingPayloadsSequence: collections.NewSequence(
			sb,
			core.PendingPayloadsSequencePrefix,
			core.PendingPayloadsSequenceName,
		),
	}

	if err := k.setComponents(k.cdc, k.logger, k.eventService, sb, bankKeeper); err != nil {
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
	eventService event.Service,
	storeService store.KVStoreService,
	bankKeeper orbitertypes.BankKeeper,
	authority string,
) error {
	if cdc == nil {
		return core.ErrNilPointer.Wrap("codec cannot be nil")
	}
	if logger == nil {
		return core.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if eventService == nil {
		return core.ErrNilPointer.Wrap("event service cannot be nil")
	}
	if storeService == nil {
		return core.ErrNilPointer.Wrap("store service cannot be nil")
	}
	if addressCdc == nil {
		return core.ErrNilPointer.Wrap("address codec cannot be nil")
	}
	if bankKeeper == nil {
		return core.ErrNilPointer.Wrap("bank keeper cannot be nil")
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
		return core.ErrNilPointer.Wrap("logger cannot be nil")
	}
	if k.eventService == nil {
		return core.ErrNilPointer.Wrap("event service cannot be nil")
	}
	if k.cdc == nil {
		return core.ErrNilPointer.Wrap("codec cannot be nil")
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

func (k *Keeper) Executor() *executorcomp.Executor {
	return k.executor
}

func (k *Keeper) Forwarder() *forwardercomp.Forwarder {
	return k.forwarder
}

func (k *Keeper) Dispatcher() *dispatchercomp.Dispatcher {
	return k.dispatcher
}

func (k *Keeper) Adapter() *adaptercomp.Adapter {
	return k.adapter
}

func (k *Keeper) SetForwardingControllers(controllers ...orbitertypes.ForwardingController) {
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

func (k *Keeper) SetActionControllers(controllers ...orbitertypes.ActionController) {
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

func (k *Keeper) SetAdapterControllers(controllers ...orbitertypes.AdapterController) {
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

// RequireAuthority returns an error is the signer is not the
// keeper authority.
func (k *Keeper) RequireAuthority(signer string) error {
	if k.Authority() != signer {
		return core.ErrUnauthorized
	}

	return nil
}

// setComponents registers all required components in the orbiter keeper.
func (k *Keeper) setComponents(
	cdc codec.Codec,
	logger log.Logger,
	eventService event.Service,
	sb *collections.SchemaBuilder,
	bankKeeper orbitertypes.BankKeeper,
) error {
	executor, err := executorcomp.New(cdc, sb, logger, eventService)
	if err != nil {
		return errorsmod.Wrap(err, "error creating a new action component")
	}

	forwarder, err := forwardercomp.New(cdc, sb, logger, eventService, bankKeeper)
	if err != nil {
		return errorsmod.Wrap(err, "error creating a new forwarding component")
	}

	dispatcher, err := dispatchercomp.New(cdc, sb, logger, forwarder, executor)
	if err != nil {
		return errorsmod.Wrap(err, "error creating a new dispatcher component")
	}

	adapter, err := adaptercomp.New(cdc, sb, logger, eventService, bankKeeper, dispatcher)
	if err != nil {
		return errorsmod.Wrap(err, "error creating a new adapter component")
	}

	k.executor = executor
	k.forwarder = forwarder
	k.dispatcher = dispatcher
	k.adapter = adapter

	return nil
}
