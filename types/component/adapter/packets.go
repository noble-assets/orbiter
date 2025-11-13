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
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/v2/types/core"
)

var (
	_ CrossChainPacket = (*IBCCrossChainPacket)(nil)
	_ CrossChainPacket = (*CCTPCrossChainPacket)(nil)
)

// CrossChainPacket defines the behavior of an Orbiter abstraction message for a generic
// bridge protocol.
type CrossChainPacket interface {
	// Returns the underlying protocol packet.
	Packet() []byte
}

// CCTPCrossChainPacket is the Orbiter abstraction message for the payload received via CCTP.
type CCTPCrossChainPacket struct {
	// transferNonce is the nonce of the associated transfer message.
	transferNonce uint64
	// coin to be trasferred.
	coin sdk.Coin
	// data contains the bytes of the Orbiter paylaod.
	data []byte
}

// NewCCTPCrossChainPacket returns a reference to an abstracted CCTP packet containing
// the Orbiter payload.
func NewCCTPCrossChainPacket(
	transferNonce uint64,
	coin sdk.Coin,
	data []byte,
) (*CCTPCrossChainPacket, error) {
	if len(data) == 0 {
		return nil, errors.New("received empty data for Orbiter payload")
	}

	return &CCTPCrossChainPacket{
		transferNonce: transferNonce,
		coin:          coin,
		data:          data,
	}, nil
}

// Packet returns the raw packet data bytes.
func (p *CCTPCrossChainPacket) Packet() []byte {
	return p.data
}

// TransferNonce returns the associated transfer message nonce.
func (p *CCTPCrossChainPacket) TransferNonce() uint64 {
	return p.transferNonce
}

// Coin returns the coin received via  CCTP.
func (p *CCTPCrossChainPacket) Coin() sdk.Coin {
	return p.coin
}

// IBCCrossChainPacket represents a cross-chain packet received via IBC with routing metadata.
// It encapsulates the packet data along with source port and channel information required
// for IBC packet processing and acknowledgment handling.
type IBCCrossChainPacket struct {
	sourcePort    string
	sourceChannel string
	data          []byte
}

// NewIBCCrossChainPacket creates a new IBCCrossChainPacket with the provided routing information.
// The data slice is defensively copied to prevent external mutation after construction.
func NewIBCCrossChainPacket(
	sourcePort, sourceChannel string,
	data []byte,
) (*IBCCrossChainPacket, error) {
	if sourcePort == "" || sourceChannel == "" {
		return nil, fmt.Errorf("source port and channel must not be empty")
	}

	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	return &IBCCrossChainPacket{
		sourcePort:    sourcePort,
		sourceChannel: sourceChannel,
		data:          dataCopy,
	}, nil
}

// Packet returns the raw packet data bytes.
func (i *IBCCrossChainPacket) Packet() []byte {
	return i.data
}

// SourcePort returns the IBC source port identifier from which the packet originated.
func (i *IBCCrossChainPacket) SourcePort() string {
	return i.sourcePort
}

// SourceChannel returns the IBC source channel identifier from which the packet originated.
func (i *IBCCrossChainPacket) SourceChannel() string {
	return i.sourceChannel
}
