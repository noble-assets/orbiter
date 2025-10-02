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

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/noble-assets/orbiter/controller"
	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
	"github.com/noble-assets/orbiter/types/hyperlane"
)

var _ orbitertypes.AdapterController = &HyperlaneAdapter{}

// HyperlaneAdapter is the type component to convert
// an incoming Hyperlane message body to the common payload
// type handled by the module.
type HyperlaneAdapter struct {
	*controller.BaseController[core.ProtocolID]

	logger log.Logger

	stateHandler orbitertypes.PendingPayloadsHandler
}

// NewHyperlaneAdapter returns a reference to a new HyperlaneAdapter instance.
func NewHyperlaneAdapter(
	logger log.Logger,
	orbiterStateHandler orbitertypes.PendingPayloadsHandler,
) (*HyperlaneAdapter, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	baseController, err := controller.NewBase(core.PROTOCOL_HYPERLANE)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to create base controller")
	}

	if orbiterStateHandler == nil {
		return nil, core.ErrNilPointer.Wrap("orbiter state handler cannot be nil")
	}

	return &HyperlaneAdapter{
		BaseController: baseController,
		logger:         logger.With(core.AdapterControllerName, baseController.Name()),
		stateHandler:   orbiterStateHandler,
	}, nil
}

// ParsePayload delegates the parsing of a Hyperlane message body to the underlying
// Parser implementation.
func (h *HyperlaneAdapter) ParsePayload(
	ctx context.Context,
	_ core.ProtocolID,
	payloadBz []byte,
) (bool, *core.Payload, error) {
	payloadHash, err := hyperlane.GetPayloadHashFromWarpMessageBody(payloadBz)
	if err != nil {
		return false, nil, errorsmod.Wrap(err, "failed to parse payload")
	}

	pendingPayload, err := h.stateHandler.PendingPayload(ctx, payloadHash)
	if err != nil {
		return false, nil, errorsmod.Wrap(err, "failed to get pending payload")
	}

	return true, pendingPayload.Payload, nil
}
