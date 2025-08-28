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
	"bytes"
	"encoding/binary"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	"github.com/ethereum/go-ethereum/crypto"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/noble-assets/orbiter/controller"
	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var (
	_ types.ControllerAdapter = &CCTPAdapter{}
	_ types.PayloadParser     = &CCTPParser{}
)

// CCTPAdapter ... TODO
type CCTPAdapter struct {
	*controller.BaseController[core.ProtocolID]

	logger log.Logger
	parser *CCTPParser
}

// NewCCTPAdapter returns a reference to a new CCTPAdapter instance.
func NewCCTPAdapter(cdc codec.Codec, logger log.Logger) (*CCTPAdapter, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := core.PROTOCOL_CCTP
	baseController, err := controller.NewBase(id)
	if err != nil {
		return nil, err
	}

	parser, err := NewCCTPParser(cdc)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error during instantiation of CCTP adapter")
	}

	return &CCTPAdapter{
		BaseController: baseController,
		logger:         logger.With(core.AdapterControllerName, baseController.Name()),
		parser:         parser,
	}, nil
}

// ParsePayload dispatches the payload parsing to the underlying CCTP parser.
func (a *CCTPAdapter) ParsePayload(
	id core.ProtocolID,
	bz []byte,
) (bool, *core.Payload, error) {
	return a.parser.ParsePayload(id, bz)
}

// CCTPParser ... TODO
type CCTPParser struct {
	cdc codec.Codec
}

// NewCCTPParser returns a new instance of a CCTP parser.
func NewCCTPParser(cdc codec.Codec) (*CCTPParser, error) {
	if cdc == nil {
		return nil, core.ErrNilPointer.Wrap("codec cannot be nil")
	}

	return &CCTPParser{
		cdc,
	}, nil
}

// ParsePayload parses the payload from the CCTP entrypoint to retrieve the
// Orbiter payload. It returns:
// - bool: whether the payload is intended for the Orbiter module.
// - Payload: the parsed payload.
// - error: an error, if one occurred during parsing.
func (p *CCTPParser) ParsePayload(_ core.ProtocolID, bz []byte) (bool, *core.Payload, error) {
	// The wire format of the provided bytes is as follows:
	//   - transferMessageLength - 2 bytes
	//   - transferMessage - `transferMessageLength` bytes
	//   - payloadMessageLength - 2 bytes
	//   - payloadMessage - `payloadMessageLength` bytes
	//   - payload - remaining bytes
	offset := 0

	transferMessageLength := binary.BigEndian.Uint16(bz[offset : offset+2])
	offset += 2

	rawTransferMessage := bz[offset : offset+int(transferMessageLength)]
	offset += int(transferMessageLength)

	payloadMessageLength := binary.BigEndian.Uint16(bz[offset : offset+2])
	offset += 2

	rawPayloadMessage := bz[offset : offset+int(payloadMessageLength)]
	offset += int(payloadMessageLength)

	rawPayload := bz[offset:]

	// ==========

	transferMessage, err := new(cctptypes.Message).Parse(rawTransferMessage)
	if err != nil {
		return false, nil, err
	}
	payloadMessage, err := new(cctptypes.Message).Parse(rawPayloadMessage)
	if err != nil {
		return false, nil, err
	}

	// Ensure that both messages were sent from the same domain.
	if transferMessage.SourceDomain != payloadMessage.SourceDomain {
		// TODO: Return an error!
	}
	// Ensure that both messages were sent from the same sender.
	if !bytes.Equal(transferMessage.Sender, payloadMessage.Sender) {
		// TODO: Return an error!
	}
	// TODO: Ensure transferMessage.MessageBody.MintRecipient and payloadMessage.Recipient are the
	// Orbiter address!

	// Decode the body of the payload message into the provided transfer nonce
	// and payload hash. Those are stored as 8 and 32 bytes respectively.
	if len(payloadMessage.MessageBody) != 40 {
		// TODO: Return an error!
	}
	expectedTransferNonce := binary.BigEndian.Uint64(payloadMessage.MessageBody[0:8])
	expectedPayloadHash := payloadMessage.MessageBody[8:40]

	// Ensure that the nonce of the provided transfer message is the same as
	// the one present in the payload message body.
	if transferMessage.Nonce != expectedTransferNonce {
		// TODO: Return an error!
	}
	// Ensure that the hash of the provided payload is the same as the one
	// present in the payload message body.
	payloadHash := crypto.Keccak256(rawPayload)
	if !bytes.Equal(payloadHash, expectedPayloadHash) {
		// TODO: Return an error!
	}

	// ==========

	var payload *core.Payload
	if err := p.cdc.Unmarshal(rawPayload, payload); err != nil {
		return true, nil, err
	}

	if err := payload.Validate(); err != nil {
		return true, payload, err
	}

	return true, payload, nil
}
