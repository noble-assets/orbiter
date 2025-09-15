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

package hyperlane

import (
	"encoding/binary"
	"errors"
	"math"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"

	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

// OrbiterBody defines the structure of the payload to be passed
// in Hyperlane messages, that can be parsed by the Orbiter module.
//
// Bytes layout:
//   - 0:4 : protocol ID
//   - 4:  : forwarding
type OrbiterBody struct {
	// // TODO: how can these bytes be constructed in the EVM?
	// // TODO: There is no way to do proper serializing of the data according to protobuf types in
	// Solidity,
	// // so we'll need custom bytes building.
	// payload core.Payload

	protocolID core.ProtocolID
	attributes core.ForwardingAttributes
}

func NewHyperlaneOrbiterBody(p core.ProtocolID, a core.ForwardingAttributes) (*OrbiterBody, error) {
	if err := p.Validate(); err != nil {
		return nil, errorsmod.Wrap(err, "invalid protocol id")
	}

	if err := a.Validate(); err != nil {
		return nil, errorsmod.Wrap(err, "invalid forward attributes")
	}

	return &OrbiterBody{
		protocolID: p,
		attributes: a,
	}, nil
}

func (h OrbiterBody) ToOrbiterPayload() (*core.Payload, error) {
	// TODO: currently not supporting passthrough payload
	forwarding, err := core.NewForwarding(h.protocolID, h.attributes, nil)
	if err != nil {
		// sanity check, should not happen because the parsing should only yield valid bodies.
		return nil, err
	}

	return core.NewPayload(forwarding)
}

// ParseHyperlaneOrbiterBody parses the bytes of a Hyperlane message body to retrieve
// the relevant information for handling with the Orbiter implementation.
//
// TODO: for now this is not containing any actions but can only contain forwarding information.
func ParseHyperlaneOrbiterBody(cdc codec.Codec, messageBody []byte) (*OrbiterBody, error) {
	protocolIDu32 := binary.BigEndian.Uint32(messageBody[:4])
	if protocolIDu32 > math.MaxInt32 {
		return nil, errors.New("protocol id out of range")
	}

	protocolID, err := core.NewProtocolID(int32(protocolIDu32))
	if err != nil {
		return nil, errorsmod.Wrap(err, "message body contains invalid protocol id")
	}

	// TODO: This won't work as it is, because this would require marshalling protos inside
	// of Solidity contracts.
	// Instead, we will have to manually pack / unpack the contents from the bytes array.
	forwardingAttributes := messageBody[4:]
	var hypAttrs forwardingtypes.HypAttributes
	if err = cdc.Unmarshal(forwardingAttributes, &hypAttrs); err == nil {
		return NewHyperlaneOrbiterBody(protocolID, &hypAttrs)
	}

	var cctpAttrs forwardingtypes.CCTPAttributes
	if err = cdc.Unmarshal(forwardingAttributes, &cctpAttrs); err == nil {
		return NewHyperlaneOrbiterBody(protocolID, &cctpAttrs)
	}

	// TODO: currently not supporting IBC yet because the attributes are not defined yet

	return nil, errors.New("message body does not contain valid forwarding attributes")
}
