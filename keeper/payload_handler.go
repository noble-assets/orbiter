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
	"errors"
	"fmt"
	"time"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/noble-assets/orbiter/types/core"
)

// ExpiredPayloadsLimit defines the maximum number of expired payloads
// removed during the ABCI hooks.
//
// We're limiting the amount of payloads handled here to avoid
// impacts of spam attacks that would slow down the begin block logic
// by iterating over thousands of spam payloads.
const ExpiredPayloadsLimit = 200

// submit adds a new pending payload into the module storage.
// If the payload's hash is already set, an error is returned.
//
// CONTRACT: The payload MUST be validated before using this method.
func (k *Keeper) submit(
	ctx context.Context,
	payload *core.Payload,
) (*core.PayloadHash, error) {
	next, err := k.pendingPayloadsSequence.Next(ctx)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to get next sequence number")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pendingPayload := core.PendingPayload{
		Sequence:  next,
		Payload:   payload,
		Timestamp: sdkCtx.BlockTime().UnixNano(),
	}

	hash, err := pendingPayload.SHA256Hash()
	if err != nil {
		return nil, err
	}

	hashBz := hash.Bytes()

	found, err := k.pendingPayloads.Has(ctx, hashBz)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to check pending payloads")
	}

	if found {
		k.logger.Error("payload hash already registered", "hash", hash.String())

		return nil, errors.New("payload hash already registered")
	}

	k.logger.Debug("payload registered", "hash", hash.String(), "payload", payload.String())

	if err = k.pendingPayloads.Set(ctx, hashBz, pendingPayload); err != nil {
		return nil, errorsmod.Wrap(err, "failed to set pending payload")
	}

	return hash, nil
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
			return fmt.Errorf("action %s is paused", action.ID().String())
		}
	}

	paused, err := k.forwarder.IsProtocolPaused(ctx, payload.Forwarding.ProtocolId)
	if err != nil {
		return errorsmod.Wrap(err, "failed to check if protocol is paused")
	}

	if paused {
		return fmt.Errorf("protocol %s is paused", payload.Forwarding.ProtocolId.String())
	}

	cachedAttrs, err := payload.Forwarding.CachedAttributes()
	if err != nil {
		return err
	}

	ccID := core.CrossChainID{
		ProtocolId:     payload.Forwarding.ProtocolId,
		CounterpartyId: cachedAttrs.CounterpartyID(),
	}

	paused, err = k.forwarder.IsCrossChainPaused(ctx, ccID)
	if err != nil {
		return errorsmod.Wrap(err, "failed to check if cross-chain paused")
	}

	if paused {
		return fmt.Errorf("cross-chain %s is paused", ccID.String())
	}

	return nil
}

// pendingPayload returns the pending payload with the given hash
// if it is found in the module storage.
func (k *Keeper) pendingPayload(
	ctx context.Context,
	hash *core.PayloadHash,
) (*core.PendingPayload, error) {
	if hash == nil {
		return nil, core.ErrNilPointer.Wrap("payload hash")
	}

	payload, err := k.pendingPayloads.Get(ctx, hash.Bytes())
	if err != nil {
		k.Logger().Error(
			"failed to retrieve pending payload",
			"hash", hash.String(),
		)

		return nil, sdkerrors.ErrNotFound.Wrapf("payload with hash %s", hash.String())
	}

	k.Logger().Debug(
		"retrieved pending payload",
		"hash", hash.String(),
		"payload", payload.String(),
	)

	return &payload, nil
}

// RemovePendingPayload removes the pending payload from the module state.
// If a payload is not found, it is a no-op but does not return an error.
func (k *Keeper) RemovePendingPayload(
	ctx context.Context,
	hash *core.PayloadHash,
) error {
	if hash == nil {
		return core.ErrNilPointer.Wrap("payload hash")
	}

	found, err := k.pendingPayloads.Has(ctx, hash.Bytes())
	if err != nil {
		return errorsmod.Wrap(err, "failed to check pending payloads")
	}

	if !found {
		return sdkerrors.ErrNotFound.Wrapf("payload with hash %q", hash.String())
	}

	if err = k.pendingPayloads.Remove(ctx, hash.Bytes()); err != nil {
		return errorsmod.Wrap(err, "failed to remove pending payload")
	}

	k.Logger().Debug("removed pending payload", "hash", hash.String())

	return nil
}

// RemoveExpiredPayloads ranges over the payloads by their submission timestamps
// and removes those that are older than the cutoff date.
func (k *Keeper) RemoveExpiredPayloads(
	ctx context.Context,
	cutoff time.Time,
) error {
	// NOTE: we range over all hashes from zero time UNTIL the cutoff.
	rng := collections.NewPrefixUntilPairRange[int64, []byte](cutoff.UnixNano())

	var count int
	if err := k.pendingPayloads.Indexes.Multi.Walk(
		ctx,
		rng,
		func(_ int64, hash []byte) (stop bool, err error) {
			count++
			if count > ExpiredPayloadsLimit {
				return true, nil
			}

			h := core.PayloadHash(hash)

			err = k.RemovePendingPayload(ctx, &h)
			if err != nil {
				k.Logger().Error(
					"failed to remove pending payload",
					"hash", h.String(),
					"error", err.Error(),
				)

				return true, err
			}

			return false, nil
		},
	); err != nil {
		return errorsmod.Wrap(err, "failed to iterate pending payloads")
	}

	return nil
}
