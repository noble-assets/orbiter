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
	"strconv"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"

	errorsmod "cosmossdk.io/errors"

	"github.com/noble-assets/orbiter/types/core"
)

var _ hyperlaneutil.HyperlaneApp = &Keeper{}

const OrbiterHyperlaneApp uint8 = iota

func (k *Keeper) RegisterHyperlaneAppRoute() {
	k.hyperlaneCoreKeeper.AppRouter().RegisterModule(uint8(0), k)
}

// TODO: do we need this? The Warp implementation is using this to check if a token is registered..
// but we don't need that?
func (k *Keeper) Exists(
	ctx context.Context,
	handledPayload hyperlaneutil.HexAddress,
) (bool, error) {
	// return k.HandledHyperlaneTransfers.Has(ctx, handledPayload.GetInternalId())
	return true, nil
}

// TODO: do we need this? The Warp implementation is using this to check for a certain ISM
// per-token,
// but we won't keep track of any specific assets that would need to be registered.
func (k *Keeper) ReceiverIsmId(
	ctx context.Context,
	recipient hyperlaneutil.HexAddress,
) (*hyperlaneutil.HexAddress, error) {
	return nil, errors.New("not implemented")
}

func (k *Keeper) Handle(
	ctx context.Context,
	mailboxId hyperlaneutil.HexAddress,
	message hyperlaneutil.HyperlaneMessage,
) error {
	_, payload, err := k.adapter.ParsePayload(core.PROTOCOL_HYPERLANE, message.Body)
	if err != nil {
		return errorsmod.Wrap(err, "failed to parse payload")
	}

	ccID, err := core.NewCrossChainID(core.PROTOCOL_HYPERLANE, strconv.Itoa(int(message.Origin)))
	if err != nil {
		return errorsmod.Wrap(err, "failed to parse cross chain ID")
	}

	// TODO: I guess here should be where the BeforeTransferHook is called? however the transfer
	// should already have been processed at this point
	if err = k.adapter.BeforeTransferHook(
		ctx, ccID, payload,
	); err != nil {
		return errorsmod.Wrap(err, "failed during before transfer hook")
	}

	// TODO: what to do with mailbox ID? do we need to check this?

	// TODO: theoretically we'd need a different after-transfer hook here?
	// The transfer attributes should be populated from the information from the Warp transfer
	// and not based on the balance of the module account since the transfer is in a separate
	// message.
	transferAttr, err := k.adapter.AfterTransferHook(ctx, ccID, payload)
	if err != nil {
		return errorsmod.Wrap(err, "failed to run AfterTransferHook")
	}

	if err = k.adapter.ProcessPayload(ctx, transferAttr, payload); err != nil {
		return errorsmod.Wrap(err, "failed to process payload")
	}

	return nil
}
