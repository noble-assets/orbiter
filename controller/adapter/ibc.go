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
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/noble-assets/orbiter/v2/controller"
	"github.com/noble-assets/orbiter/v2/types"
	adaptertypes "github.com/noble-assets/orbiter/v2/types/component/adapter"
	"github.com/noble-assets/orbiter/v2/types/core"
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

func (a *IBCAdapter) ParsePacket(
	ccPacket adaptertypes.CrossChainPacket,
) (*types.ParsedData, error) {
	ibcPacket, ok := ccPacket.(*adaptertypes.IBCCrossChainPacket)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&adaptertypes.IBCCrossChainPacket{},
			ccPacket,
		)
	}

	packet, err := GetICS20PacketData(ibcPacket.Packet())
	if err != nil {
		return nil, core.ErrNoOrbiterPacket.Wrap("data is not ICS20 packet")
	}

	if packet.GetReceiver() != core.ModuleAddress.String() {
		return nil, core.ErrNoOrbiterPacket.Wrap("receiver is not Orbiter module")
	}

	payload, err := a.parser.ParsePayload([]byte(packet.GetMemo()))
	if err != nil {
		return nil, err
	}

	amount, ok := sdkmath.NewIntFromString(packet.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount: %s", packet.Amount)
	}

	// In IBC the denom specified in the packet is the sending chain representation. We have to
	// convert the denom into the Noble representation.
	denom, err := RecoverNativeDenom(
		packet.Denom,
		ibcPacket.SourcePort(),
		ibcPacket.SourceChannel(),
	)
	if err != nil {
		return nil, err
	}

	return &types.ParsedData{
		Coin:    sdk.NewCoin(denom, amount),
		Payload: *payload,
	}, nil
}

var _ types.PayloadParser = &IBCParser{}

// NOTE: maybe we get rid of the IBC parser and directly use the JSON one.
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
func (p *IBCParser) ParsePayload(memoBz []byte) (*core.Payload, error) {
	payload, err := p.Parse(string(memoBz))
	if err != nil {
		return nil, err
	}

	// NOTE: validation should probably be not here.
	if err := payload.Validate(); err != nil {
		return payload, err
	}

	return payload, nil
}

// GetICS20PacketData returns the unmarshalled ICS-20 packet data.
// It returns an error if the data cannot be unmarshalled.
func GetICS20PacketData(data []byte) (transfertypes.FungibleTokenPacketData, error) {
	var ics20Data transfertypes.FungibleTokenPacketData
	err := transfertypes.ModuleCdc.UnmarshalJSON(data, &ics20Data)

	return ics20Data, err
}
