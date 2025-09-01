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
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/noble-assets/orbiter/controller"
	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var _ types.AdapterController = &IBCAdapter{}

// IBCAdapter is the type component in charge of adapting the
// memo of an IBC ICS20 transfer to the common payload type
// handled by the module.
type IBCAdapter struct {
	*controller.BaseController[core.ProtocolID]

	logger log.Logger
	parser *IBCParser
}

// NewIBCAdapter returns a reference to a new IBCAdapter instance.
func NewIBCAdapter(cdc codec.Codec, logger log.Logger) (*IBCAdapter, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := core.PROTOCOL_IBC
	baseController, err := controller.NewBase(id)
	if err != nil {
		return nil, err
	}

	parser, err := NewIBCParser(cdc)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error during instantiation of IBC adapter")
	}

	return &IBCAdapter{
		logger:         logger.With(core.AdapterControllerName, baseController.Name()),
		BaseController: baseController,
		parser:         parser,
	}, nil
}

// ParsePayload dispatches the payload parsing to the underlying IBC parser.
func (a *IBCAdapter) ParsePayload(
	id core.ProtocolID,
	payloadBz []byte,
) (bool, *core.Payload, error) {
	return a.parser.ParsePayload(id, payloadBz)
}

var _ types.PayloadParser = &IBCParser{}

type IBCParser struct {
	JSONParser
}

// NewIBCParser returns a new instance of an IBC parser.
func NewIBCParser(cdc codec.Codec) (*IBCParser, error) {
	if cdc == nil {
		return nil, core.ErrNilPointer.Wrap("codec cannot be nil")
	}

	jsonParser, err := NewJSONParser(cdc)
	if err != nil {
		return nil, err
	}

	return &IBCParser{
		*jsonParser,
	}, nil
}

// ParsePayload parses the payload from an IBC transfer to retrieve the orbiter
// payload. It returns:
// - bool: whether the payload is intended for the Orbiter module.
// - Payload: the parsed payload.
// - error: an error, if one occurred during parsing.
func (p *IBCParser) ParsePayload(_ core.ProtocolID, payloadBz []byte) (bool, *core.Payload, error) {
	data, err := p.GetICS20PacketData(payloadBz)
	if err != nil {
		// Despite the error is not nil, we don't return it. We
		// want the non fungible token packet data error to be
		// returned from the ICS20 app.
		return false, nil, nil //nolint:nilerr
	}

	if data.GetReceiver() != core.ModuleAddress.String() {
		return false, nil, nil
	}

	payload, err := p.Parse(data.GetMemo())
	if err != nil {
		return true, nil, err
	}

	if err := payload.Validate(); err != nil {
		return true, payload, err
	}

	return true, payload, nil
}

// GetICS20PacketData returns unmarshalled ICS-20 packet data if it is present in the data
// as well as a boolean indicating the successful decoding.
func (p *IBCParser) GetICS20PacketData(data []byte) (transfertypes.FungibleTokenPacketData, error) {
	var ics20Data transfertypes.FungibleTokenPacketData
	err := transfertypes.ModuleCdc.UnmarshalJSON(data, &ics20Data)

	return ics20Data, err
}
