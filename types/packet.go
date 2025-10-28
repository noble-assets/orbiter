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

package types

import (
	"github.com/noble-assets/orbiter/types/core"
)

// OrbiterPacket defines the abstract cross-chain transfer packet used in the Orbiter.
type OrbiterPacket struct {
	TransferAttributes *core.TransferAttributes
	Payload            *core.Payload
}

// ForwardingPacket defines the data structure used to handle a forwarding. The
// forwarding info are extended with the cross chain transfer attributes.
type ForwardingPacket struct {
	TransferAttributes *core.TransferAttributes
	Forwarding         *core.Forwarding
}

// NewForwardingPacket returns a pointer to a validated instance of the
// forwarding packet.
func NewForwardingPacket(
	transferAttr *core.TransferAttributes,
	forwarding *core.Forwarding,
) (*ForwardingPacket, error) {
	forwardingPacket := ForwardingPacket{
		TransferAttributes: transferAttr,
		Forwarding:         forwarding,
	}

	return &forwardingPacket, forwardingPacket.Validate()
}

// Validate returns an error if the instance is not valid.
func (p *ForwardingPacket) Validate() error {
	if p == nil {
		return core.ErrNilPointer.Wrap("forwarding packet is not set")
	}

	err := p.Forwarding.Validate()
	if err != nil {
		return err
	}

	return p.TransferAttributes.Validate()
}

// ActionPacket defines the data structure used to handle an action. The action
// attributes are extended with the cross chain transfer ones.
type ActionPacket struct {
	TransferAttributes *core.TransferAttributes
	Action             *core.Action
}

// NewActionPacket returns a pointer to a validated instance of the
// action packet.
func NewActionPacket(
	transferAttr *core.TransferAttributes,
	action *core.Action,
) (*ActionPacket, error) {
	actionPacket := ActionPacket{
		TransferAttributes: transferAttr,
		Action:             action,
	}

	return &actionPacket, actionPacket.Validate()
}

// Validate returns an error if any of the action packet field is
// not valid.
func (p *ActionPacket) Validate() error {
	if p == nil {
		return core.ErrNilPointer.Wrap("packet is not set")
	}

	err := p.Action.Validate()
	if err != nil {
		return err
	}

	return p.TransferAttributes.Validate()
}
