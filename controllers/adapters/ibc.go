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

package adapters

import (
	"context"
	"fmt"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"orbiter.dev/controllers"
	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

var _ interfaces.ControllerAdapter = &IBCAdapter{}

// NewIBCAdapter returns a reference to a new IBCAdapter instance.
func NewIBCAdapter(cdc codec.Codec, logger log.Logger) (*IBCAdapter, error) {
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := types.PROTOCOL_IBC
	baseController, err := controllers.NewBaseController(id)
	if err != nil {
		return nil, err
	}

	parser, err := NewIBCParser(cdc)
	if err != nil {
		return nil, fmt.Errorf("error during instantiation of IBC adapter: %w", err)
	}

	return &IBCAdapter{
		logger:         logger.With(types.AdapterControllerName, baseController.Name()),
		BaseController: baseController,
		parser:         parser,
	}, nil
}

// IBCAdapter is the type component in charge of adapting the
// memo of an IBC ICS20 transfer to the common payload type
// handled by the module.
type IBCAdapter struct {
	logger log.Logger
	*controllers.BaseController[types.ProtocolID]
	parser *IBCParser
}

// ParsePayload dispatch the payload parsing to the underlying IBC parser.
func (a *IBCAdapter) ParsePayload(payloadBz []byte) (bool, *types.Payload, error) {
	return a.parser.ParsePayload(payloadBz)
}

// BeforeTransferHook run logic before executing the IBC transfer to the
// orbiter module.
func (a *IBCAdapter) BeforeTransferHook(context.Context, *types.Payload) error {
	return nil
}

// AfterTransferHook run logic after executing the IBC transfer to the orbiter
// module.
func (a *IBCAdapter) AfterTransferHook(context.Context, *types.Payload) error {
	return nil
}

var _ interfaces.PayloadParser = &IBCParser{}

// NewIBCParser returns a new instance of an IBC parser.
func NewIBCParser(cdc codec.Codec) (*IBCParser, error) {
	jsonParser, err := NewJSONParser(cdc)
	if err != nil {
		return nil, err
	}

	return &IBCParser{
		*jsonParser,
	}, nil
}

type IBCParser struct {
	JSONParser
}

// ParsePayload parses the payload from an IBC memo to retrieve the orbiter
// payload. It returns:
// - bool: whether the payload is intended for the Orbiter module.
// - Payload: the parsed payload.
// - error: an error, if one occurred during parsing.
func (p *IBCParser) ParsePayload(payloadBz []byte) (bool, *types.Payload, error) {
	isIcs20Packet, data := p.IsIcs20Packet(payloadBz)
	if !isIcs20Packet {
		return false, nil, nil
	}

	if data.GetReceiver() != types.ModuleAddress.String() {
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

// IsIcs20Packet returns a boolean indicating wheater the data is an ICS20
// packet, and if true, the fungible token packet data.
func (p *IBCParser) IsIcs20Packet(data []byte) (bool, transfertypes.FungibleTokenPacketData) {
	var ics20Data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(data, &ics20Data); err != nil {
		return false, transfertypes.FungibleTokenPacketData{}
	}
	return true, ics20Data
}
