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
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/noble-assets/orbiter/controller"
	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
	"github.com/noble-assets/orbiter/types/hyperlane"
)

var _ types.AdapterController = &HyperlaneAdapter{}

// HyperlaneAdapter is the type component to convert
// an incoming Hyperlane message body to the common payload
// type handled by the module.
type HyperlaneAdapter struct {
	*controller.BaseController[core.ProtocolID]

	logger log.Logger
	parser *HyperlaneParser
}

// NewHyperlaneAdapter returns a reference to a new HyperlaneAdapter instance.
func NewHyperlaneAdapter(cdc codec.Codec, logger log.Logger) (*HyperlaneAdapter, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	baseController, err := controller.NewBase(core.PROTOCOL_HYPERLANE)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to create base controller")
	}

	parser, err := NewHyperlaneParser(cdc)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to create hyperlane parser")
	}

	return &HyperlaneAdapter{
		BaseController: baseController,
		logger:         logger.With(core.AdapterControllerName, baseController.Name()),
		parser:         parser,
	}, nil
}

// ParsePayload delegates the parsing of a Hyperlane message body to the underlying
// Parser implementation.
func (h *HyperlaneAdapter) ParsePayload(
	protocolID core.ProtocolID,
	payloadBz []byte,
) (bool, *core.Payload, error) {
	return h.parser.ParsePayload(protocolID, payloadBz)
}

// HyperlaneParser is used to parse Orbiter payloads from an incoming Hyperlane message body.
type HyperlaneParser struct {
	cdc codec.Codec
}

func NewHyperlaneParser(cdc codec.Codec) (*HyperlaneParser, error) {
	if cdc == nil {
		return nil, core.ErrNilPointer.Wrap("codec cannot be nil")
	}

	return &HyperlaneParser{cdc: cdc}, nil
}

// ParsePayload parses the payload from a Hyperlane message body to retrieve
// the Orbiter payload.
//
// NOTE: This parser is only ever called in the Handle method of the Hyperlane application,
// which means that all message bodies handled by this parser were intended for the
// Orbiter. Hence, the first return value is ALWAYS true.
//
// TODO: can protocol ID be removed here? Why is it included?
func (p *HyperlaneParser) ParsePayload(
	_ core.ProtocolID,
	payloadBz []byte,
) (bool, *core.Payload, error) {
	parsedBody, err := hyperlane.ParseHyperlaneOrbiterBody(p.cdc, payloadBz)
	if err != nil {
		return true, nil, errorsmod.Wrap(err, "failed to parse hyperlane body")
	}

	payload, err := parsedBody.ToOrbiterPayload()
	if err != nil {
		return true, nil, errorsmod.Wrap(err, "failed to convert hyperlane body to payload")
	}

	return true, payload, nil
}
