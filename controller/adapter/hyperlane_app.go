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

package adapter

import (
	"context"
	"errors"
	"strconv"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"

	errorsmod "cosmossdk.io/errors"

	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
	hyperlaneorbitertypes "github.com/noble-assets/orbiter/types/hyperlane"
)

// Hyperlane interface compliance.
var _ hyperlaneutil.HyperlaneApp = &HyperlaneAdapter{}

// OrbiterHyperlaneAppID defines the module identifier of the Orbiter Hyperlane application.
const OrbiterHyperlaneAppID uint8 = 255

// RegisterHyperlaneAppRoute registers the Orbiter Hyperlane application in the main application
// router.
func (ha *HyperlaneAdapter) RegisterHyperlaneAppRoute() {
	ha.hyperlaneCore.AppRouter().RegisterModule(OrbiterHyperlaneAppID, ha)
}

func (ha *HyperlaneAdapter) Handle(
	ctx context.Context,
	mailboxId hyperlaneutil.HexAddress,
	message hyperlaneutil.HyperlaneMessage,
) error {
	_, payload, err := ha.ParsePayload(ctx, core.PROTOCOL_HYPERLANE, message.Body)
	if err != nil {
		return errorsmod.Wrap(err, "failed to parse payload")
	}

	ccID, err := core.NewCrossChainID(core.PROTOCOL_HYPERLANE, strconv.Itoa(int(message.Origin)))
	if err != nil {
		return errorsmod.Wrap(err, "failed to parse cross chain ID")
	}

	if err = types.Adapter(ha).BeforeTransferHook(
		ctx, ccID, payload,
	); err != nil {
		return errorsmod.Wrap(err, "failed during before transfer hook")
	}

	// TODO: what to do with mailbox ID? do we need to check this?

	reducedWarpMessage, err := hyperlaneorbitertypes.GetReducedWarpMessageFromOrbiterMessage(
		message,
	)
	if err != nil {
		return errorsmod.Wrap(err, "failed to create reduced warp message")
	}

	if err = ha.hyperlaneWarp.Handle(
		ctx,
		mailboxId,
		reducedWarpMessage,
	); err != nil {
		return errorsmod.Wrap(err, "internal warp handling failed")
	}

	transferAttr, err := ha.AfterTransferHook(ctx, ccID, payload)
	if err != nil {
		return errorsmod.Wrap(err, "failed to run AfterTransferHook")
	}

	if err = ha.ProcessPayload(ctx, transferAttr, payload); err != nil {
		return errorsmod.Wrap(err, "failed to process payload")
	}

	return nil
}

// TODO: do we need this? The Warp implementation is using this to check if a token is registered..
// but we don't need that?
//
// TODO: should we delegate to the Warp method here?
func (ha *HyperlaneAdapter) Exists(
	_ context.Context,
	_ hyperlaneutil.HexAddress,
) (bool, error) {
	// return ha.HandledHyperlaneTransfers.Has(ctx, handledPayload.GetInternalId())
	return true, nil
}

// TODO: do we need this? The Warp implementation is using this to check for a certain ISM
// per-token,
// but we won't keep track of any specific assets that would need to be registered.
//
// TODO: should we delegate to the Warp method here?
func (ha *HyperlaneAdapter) ReceiverIsmId(
	_ context.Context,
	_ hyperlaneutil.HexAddress,
) (*hyperlaneutil.HexAddress, error) {
	return nil, errors.New("not implemented")
}
