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

var _ CrossChainPacket = (*IBCCrossChainPacket)(nil)

type CrossChainPacket interface {
	// Returns the underlying protocol packet.
	Packet() []byte
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
) *IBCCrossChainPacket {
	if sourcePort == "" || sourceChannel == "" {
		panic("source port and channel must not be empty")
	}

	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)

	return &IBCCrossChainPacket{
		sourcePort:    sourcePort,
		sourceChannel: sourceChannel,
		data:          dataCopy,
	}
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
