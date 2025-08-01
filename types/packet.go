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
	"errors"
	fmt "fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TransferAttributes defines the cross-chain transfer information
// passed down the orbiter to handle actions and routing.
type TransferAttributes struct {
	// Source fields have only getter methods.
	sourceOrbitID OrbitID
	sourceCoin    sdk.Coin
	// Destination field have both setters and getters
	// because they can be mutated by actions.
	destinationCoin sdk.Coin
}

// NewTransferAttributes returns a validated reference to a
// transfer attributes type.
func NewTransferAttributes(
	sourceProtocolID ProtocolID,
	sourceCounterpartyID string,
	denom string,
	amount math.Int,
) (*TransferAttributes, error) {
	sourceOrbitID, err := NewOrbitID(sourceProtocolID, sourceCounterpartyID)
	if err != nil {
		return nil, err
	}
	sourceCoin := sdk.Coin{Denom: denom, Amount: amount}
	// Initially, the destination coin is the same as of the
	// incoming coin.
	destinationCoin := sdk.Coin{Denom: denom, Amount: amount}

	transferAttr := TransferAttributes{
		sourceOrbitID:   sourceOrbitID,
		sourceCoin:      sourceCoin,
		destinationCoin: destinationCoin,
	}

	return &transferAttr, transferAttr.Validate()
}

// Validate returns an error if any of the fields is not valid.
func (a *TransferAttributes) Validate() error {
	if a == nil {
		return ErrNilPointer.Wrap("transfer attributes is a nil pointer")
	}
	if err := a.sourceOrbitID.Validate(); err != nil {
		return err
	}
	if err := a.sourceCoin.Validate(); err != nil {
		return fmt.Errorf("source coin validation error: %w", err)
	}
	if !a.sourceCoin.IsPositive() {
		return errors.New("source amount must be positive")
	}
	if err := a.destinationCoin.Validate(); err != nil {
		return fmt.Errorf("destination coin validation error: %w", err)
	}
	if !a.destinationCoin.IsPositive() {
		return errors.New("destination amount must be positive")
	}

	return nil
}

func (a *TransferAttributes) SourceProtocolID() ProtocolID {
	if a != nil {
		return a.sourceOrbitID.ProtocolID
	}

	return PROTOCOL_UNSUPPORTED
}

func (a *TransferAttributes) SourceCounterpartyID() string {
	if a == nil {
		return ""
	}

	return a.sourceOrbitID.CounterpartyID
}

func (a *TransferAttributes) SourceAmount() math.Int {
	if a == nil || a.sourceCoin.Amount.IsNil() {
		return math.ZeroInt()
	}

	return a.sourceCoin.Amount
}

func (a *TransferAttributes) SourceDenom() string {
	if a == nil {
		return ""
	}

	return a.sourceCoin.Denom
}

func (a *TransferAttributes) DestinationAmount() math.Int {
	if a == nil || a.destinationCoin.Amount.IsNil() {
		return math.ZeroInt()
	}

	return a.destinationCoin.Amount
}

func (a *TransferAttributes) DestinationDenom() string {
	if a == nil {
		return ""
	}

	return a.destinationCoin.Denom
}

// SetDestinationAmount set the input amount for the destination
// amount of the transfer attributes.
//
// CONTRACT: receiver should not be nil but we handle
// nil defensively for robustness.
func (a *TransferAttributes) SetDestinationAmount(amount math.Int) {
	if a == nil {
		fmt.Println("Warning: SetDestinaitonAmount() called on nil TransferAttributes")

		return
	}

	if amount.IsNil() || amount.IsNegative() {
		a.destinationCoin.Amount = math.ZeroInt()

		return
	}
	a.destinationCoin.Amount = amount
}

// SetDestinationDenom set the denom for the destination
// denom of the transfer attributes.
//
// CONTRACT: receiver should not be nil but we handle
// nil defensively for robustness.
func (a *TransferAttributes) SetDestinationDenom(denom string) {
	if a == nil {
		fmt.Println("Warning: SetDestinaitonDenom() called on nil TransferAttributes")

		return
	}
	a.destinationCoin.Denom = denom
}

// OrbitPacket defines the data structure used to handle a routing packet. The routing info
// are extended with the cross chain transfer attributes.
type OrbitPacket struct {
	TransferAttributes *TransferAttributes
	Orbit              *Orbit
}

// NewOrbitPacket returns a pointer to a validated instance of the
// orbit packet.
func NewOrbitPacket(transferAttr *TransferAttributes, orbit *Orbit) (*OrbitPacket, error) {
	orbitPacket := OrbitPacket{
		TransferAttributes: transferAttr,
		Orbit:              orbit,
	}

	return &orbitPacket, orbitPacket.Validate()
}

// Validate returns an error if the instance is not valid.
func (p *OrbitPacket) Validate() error {
	if p == nil {
		return ErrNilPointer.Wrap("orbit packet is not set")
	}

	err := p.Orbit.Validate()
	if err != nil {
		return err
	}

	return p.TransferAttributes.Validate()
}

// ActionPacket defines the data structure used to handle an action. The action
// attributes are extended with the cross chain transfer ones.
type ActionPacket struct {
	TransferAttributes *TransferAttributes
	Action             *Action
}

// NewActionPacket returns a pointer to a validated instance of the
// action packet.
func NewActionPacket(transferAttr *TransferAttributes, action *Action) (*ActionPacket, error) {
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
		return ErrNilPointer.Wrap("packet is not set")
	}

	err := p.Action.Validate()
	if err != nil {
		return err
	}

	return p.TransferAttributes.Validate()
}
