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

	errorsmod "cosmossdk.io/errors"

	orbitertypes "github.com/noble-assets/orbiter/types"
	hyperlaneorbitertypes "github.com/noble-assets/orbiter/types/hyperlane"
)

var _ orbitertypes.HyperlaneStateHandler = &Keeper{}

// AcceptPendingPayload adds a new pending payload into the module storage.
// If the payload's hash is already set, an error is returned.
func (k *Keeper) AcceptPendingPayload(
	ctx context.Context,
	payload *hyperlaneorbitertypes.PendingPayload,
) error {
	hash, err := payload.Hash()
	if err != nil {
		return err
	}

	found, err := k.pendingPayloads.Has(ctx, hash.Bytes())
	if err != nil {
		return errorsmod.Wrap(err, "failed to check pending payloads")
	}

	if found {
		return errors.New("payload hash already exists")
	}

	return k.pendingPayloads.Set(ctx, hash.Bytes(), *payload)
}

// GetPendingPayloadWithHash returns the pending payload with the given hash
// if it is found in the module storage.
//
// TODO: move into own abstraction type (Hyperlane state handler or smth.?)
func (k *Keeper) GetPendingPayloadWithHash(
	ctx context.Context,
	hash []byte,
) (*hyperlaneorbitertypes.PendingPayload, error) {
	payload, err := k.pendingPayloads.Get(ctx, hash)
	if err != nil {
		return nil, errorsmod.Wrap(err, "pending payload not found")
	}

	return &payload, nil
}

// CompletePayloadWithHash removes the pending payload from the module state.
// If a payload is not found, it is a no-op but does not return an error.
//
// TODO: probably this needs to have more fields other than just the payload like the sender and
// origin account
func (k *Keeper) CompletePayloadWithHash(
	ctx context.Context,
	payload *hyperlaneorbitertypes.PendingPayload,
) error {
	if payload == nil {
		return errors.New("payload must not be nil")
	}

	hash, err := payload.Hash()
	if err != nil {
		return err
	}

	return k.pendingPayloads.Remove(ctx, hash.Bytes())
}
