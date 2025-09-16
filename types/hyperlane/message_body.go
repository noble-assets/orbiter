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
	errorsmod "cosmossdk.io/errors"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"math"

	"github.com/noble-assets/orbiter/types/core"
)

// OrbiterBody defines the structure of the payload to be passed
// in Hyperlane messages, that can be parsed by the Orbiter module.
//
// Bytes layout:
//   - 0:4 : protocol ID
//   - 4:  : forwarding attributes
type OrbiterBody struct {
	// // TODO: how can these bytes be constructed in the EVM?
	// // TODO: There is no way to do proper serializing of the data according to protobuf types in
	// Solidity,
	// // so we'll need custom bytes building.
	// payload core.Payload

	protocolID core.ProtocolID
	attributes core.ForwardingAttributes
}

// TODO: remove if unused?
func NewHyperlaneOrbiterBody(p core.ProtocolID, a core.ForwardingAttributes) (*OrbiterBody, error) {
	o := &OrbiterBody{
		protocolID: p,
		attributes: a,
	}

	return o, o.Validate()
}

func (o OrbiterBody) Validate() error {
	if err := o.protocolID.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid protocol id")
	}

	if err := o.attributes.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid forwarding attributes")
	}

	return nil
}

func (o OrbiterBody) ToOrbiterPayload() (*core.Payload, error) {
	// TODO: currently not supporting passthrough payload
	forwarding, err := core.NewForwarding(o.protocolID, o.attributes, nil)
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
// TODO: cdc can be removed I think :eyes:
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
	var attr core.ForwardingAttributes
	switch protocolID {
	case core.PROTOCOL_CCTP:
		panic("TODO: implement")
		attr, err = unpackCCTPAttributes(messageBody[4:])
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to unpack cctp attributes")
		}
	case core.PROTOCOL_HYPERLANE:
		panic("TODO: implement")
		attr, err = unpackHypAttributes(messageBody[4:])
		if err != nil {
			return nil, errorsmod.Wrap(err, "failed to unpack hyperlane attributes")
		}
	default:
		panic(fmt.Sprintf("protocol %s not implemented", protocolID.String()))
	}

	return &OrbiterBody{
		protocolID: protocolID,
		attributes: attr,
	}, errors.New("message body does not contain valid forwarding attributes")
}
