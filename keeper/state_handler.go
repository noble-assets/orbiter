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
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var _ orbitertypes.PendingPayloadsHandler = &Keeper{}

// AcceptPayload adds a new pending payload into the module storage.
// If the payload's hash is already set, an error is returned.
func (k *Keeper) AcceptPayload(
	ctx context.Context,
	payload *core.Payload,
) ([]byte, error) {
	if err := payload.Validate(); err != nil {
		return nil, errorsmod.Wrap(err, "invalid payload")
	}

	if err := k.validatePayloadAgainstState(ctx, payload); err != nil {
		return nil, errorsmod.Wrap(err, "payload failed stateful checks")
	}

	next, err := k.PendingPayloadsSequence.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get next sequence number")
	}

	pendingPayload := orbitertypes.PendingPayload{
		Sequence: next,
		Payload:  payload,
	}

	hash, err := pendingPayload.Keccak256Hash()
	if err != nil {
		return nil, err
	}

	found, err := k.pendingPayloads.Has(ctx, hash.Bytes())
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to check pending payloads")
	}

	if found {
		k.Logger().Error("payload hash already registered", "hash", hash.String())

		return nil, errors.New("payload hash already registered")
	}

	k.Logger().Debug("payload registered", "hash", hash.String(), "payload", payload.String())

	return hash.Bytes(), k.pendingPayloads.Set(ctx, hash.Bytes(), pendingPayload)
}

// validatePayloadAgainstState checks if the payload is valid with respect
// to the current state of the chain.
// This asserts that no actions or forwarding configurations contained in the payload
// are paused.
func (k *Keeper) validatePayloadAgainstState(
	ctx context.Context,
	payload *core.Payload,
) error {
	for _, action := range payload.PreActions {
		paused, err := k.executor.IsActionPaused(ctx, action.ID())
		if err != nil {
			return errorsmod.Wrap(err, "failed to check if action is paused")
		}

		if paused {
			return fmt.Errorf("action %s is paused", action.ID())
		}
	}

	paused, err := k.forwarder.IsProtocolPaused(ctx, payload.Forwarding.ProtocolId)
	if err != nil {
		return errorsmod.Wrap(err, "failed to check if forwarder is paused")
	}

	if paused {
		return fmt.Errorf("protocol %s is paused", payload.Forwarding.ProtocolId)
	}

	cachedAttrs, err := payload.Forwarding.CachedAttributes()
	if err != nil {
		return err
	}

	paused, err = k.forwarder.IsCrossChainPaused(ctx, core.CrossChainID{
		ProtocolId:     payload.Forwarding.ProtocolId,
		CounterpartyId: cachedAttrs.CounterpartyID(),
	})
	if err != nil {
		return errorsmod.Wrap(err, "failed to check if cross-chain paused")
	}

	if paused {
		return fmt.Errorf("cross-chain paused")
	}

	return nil
}

// PendingPayload returns the pending payload with the given hash
// if it is found in the module storage.
//
// TODO: move into own abstraction type (Hyperlane state handler or smth.?)
func (k *Keeper) PendingPayload(
	ctx context.Context,
	hash []byte,
) (*orbitertypes.PendingPayload, error) {
	payload, err := k.pendingPayloads.Get(ctx, hash)
	if err != nil {
		return nil, errorsmod.Wrap(err, "pending payload not found")
	}

	k.Logger().Debug(
		"retrieved pending payload",
		"hash", hex.EncodeToString(hash),
		"payload", payload.String(),
	)

	return &payload, nil
}

// RemovePendingPayload removes the pending payload from the module state.
// If a payload is not found, it is a no-op but does not return an error.
func (k *Keeper) RemovePendingPayload(
	ctx context.Context,
	hash []byte,
) error {
	k.Logger().Debug("completing payload", "hash", hex.EncodeToString(hash))

	return k.pendingPayloads.Remove(ctx, hash)
}
