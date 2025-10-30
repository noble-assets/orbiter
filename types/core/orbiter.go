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

package core

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

// TransferAttributes defines the cross-chain transfer information
// passed down the orbiter to handle actions and routing.
type TransferAttributes struct {
	// Source fields have only getter methods.
	sourceID   CrossChainID
	sourceCoin sdk.Coin
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
	sourceID, err := NewCrossChainID(sourceProtocolID, sourceCounterpartyID)
	if err != nil {
		return nil, err
	}
	sourceCoin := sdk.Coin{Denom: denom, Amount: amount}
	// Initially, the destination coin is the same as of the
	// incoming coin.
	destinationCoin := sdk.Coin{Denom: denom, Amount: amount}

	transferAttr := TransferAttributes{
		sourceID:        sourceID,
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
	if err := a.sourceID.Validate(); err != nil {
		return err
	}
	if err := a.sourceCoin.Validate(); err != nil {
		return errorsmod.Wrap(err, "source coin validation error")
	}
	if !a.sourceCoin.IsPositive() {
		return errors.New("source amount must be positive")
	}
	if err := a.destinationCoin.Validate(); err != nil {
		return errorsmod.Wrap(err, "destination coin validation error")
	}
	if !a.destinationCoin.IsPositive() {
		return errors.New("destination amount must be positive")
	}

	return nil
}

func (a *TransferAttributes) SourceProtocolID() ProtocolID {
	return a.sourceID.GetProtocolId()
}

func (a *TransferAttributes) SourceCounterpartyID() string {
	return a.sourceID.GetCounterpartyId()
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
		fmt.Println("Warning: SetDestinationAmount() called on nil TransferAttributes")

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
		fmt.Println("Warning: SetDestinationDenom() called on nil TransferAttributes")

		return
	}
	a.destinationCoin.Denom = denom
}

// ====================================================================================================
// Action
// ====================================================================================================

var _ cdctypes.UnpackInterfacesMessage = &Action{}

// NewAction returns a reference to a validated action. This utility
// function automatically set the attributes in the Any type of the
// return action.
func NewAction(id ActionID, attr ActionAttributes) (*Action, error) {
	a := Action{
		Id: id,
	}
	err := a.SetAttributes(attr)
	if err != nil {
		return nil, err
	}

	return &a, a.Validate()
}

// Validate returns an error if the action is not valid.
func (a *Action) Validate() error {
	if a == nil {
		return ErrNilPointer.Wrap("action is not set")
	}
	if err := a.Id.Validate(); err != nil {
		return err
	}

	if a.Attributes == nil {
		return ErrNilPointer.Wrap("action attributes are not set")
	}

	return nil
}

// ID returns the ID. If the ID is not set,
// the default value is returned.
func (a *Action) ID() ActionID {
	if a != nil {
		return a.Id
	}

	return ACTION_UNSUPPORTED
}

// CachedAttributes returns the attributes interface from the
// codec Any type. Returns nil if the action does not have
// attributes set.
func (a *Action) CachedAttributes() (ActionAttributes, error) {
	if a == nil {
		return nil, ErrNilPointer.Wrap("action is not set")
	}

	if a.Attributes == nil {
		return nil, ErrNilPointer.Wrap("action attributes are not set")
	}
	av := a.Attributes.GetCachedValue()
	attr, ok := av.(ActionAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			(ActionAttributes)(nil),
			av,
		)
	}

	return attr, nil
}

// SetAttributes sets the action attributes into the action as codec Any type.
func (a *Action) SetAttributes(attr ActionAttributes) error {
	if a == nil {
		return ErrNilPointer.Wrap("action is not set")
	}

	m, ok := attr.(proto.Message)
	if !ok {
		return sdkerrors.ErrPackAny.Wrapf("can't proto marshal %T", m)
	}

	anyValue, err := cdctypes.NewAnyWithValue(m)
	if err != nil {
		return err
	}
	a.Attributes = anyValue

	return nil
}

// UnpackInterfaces is the method required to correctly unpack
// an Any type into an interface registered in the codec.
func (a *Action) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if a == nil {
		return ErrNilPointer.Wrap("action is not set")
	}
	var attributes ActionAttributes

	return unpacker.UnpackAny(a.Attributes, &attributes)
}

// ====================================================================================================
// Forwarding
// ====================================================================================================

var _ cdctypes.UnpackInterfacesMessage = &Forwarding{}

// NewForwarding returns a reference to a validated forwarding. The
// function automatically sets the attributes in the Any type of the
// return instance.
//
// NOTE: passthroughPayload is currently ignored.
func NewForwarding(
	id ProtocolID,
	a ForwardingAttributes,
	passthroughPayload []byte,
) (*Forwarding, error) {
	o := Forwarding{
		ProtocolId:         id,
		PassthroughPayload: passthroughPayload,
	}
	err := o.SetAttributes(a)
	if err != nil {
		return nil, err
	}

	return &o, o.Validate()
}

// Validate returns an error if the forwarding is not valid.
func (f *Forwarding) Validate() error {
	if f == nil {
		return ErrNilPointer.Wrap("forwarding is not set")
	}
	if err := f.ProtocolId.Validate(); err != nil {
		return err
	}
	if f.Attributes == nil {
		return ErrNilPointer.Wrap("forwarding attributes are not set")
	}

	return nil
}

// ProtocolID returns the protocol ID associated with the forwarding. If
// the id is not set, the default value is returned.
func (f *Forwarding) ProtocolID() ProtocolID {
	if f != nil {
		return f.ProtocolId
	}

	return PROTOCOL_UNSUPPORTED
}

// CachedAttributes returns the attributes interface from the
// codec Any type. Returns nil if the forwarding does not have
// attributes set.
func (f *Forwarding) CachedAttributes() (ForwardingAttributes, error) {
	if f == nil {
		return nil, ErrNilPointer.Wrap("forwarding is not set")
	}
	if f.Attributes == nil {
		return nil, ErrNilPointer.Wrap("forwarding attributes are not set")
	}
	av := f.Attributes.GetCachedValue()
	a, ok := av.(ForwardingAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			(ForwardingAttributes)(nil),
			av,
		)
	}

	return a, nil
}

// SetAttributes sets the attributes as codec Any type.
func (f *Forwarding) SetAttributes(a ForwardingAttributes) error {
	if f == nil {
		return ErrNilPointer.Wrap("forwarding is not set")
	}
	// The interface we want to pack as any must
	// implement the proto Message interface.
	m, ok := a.(proto.Message)
	if !ok {
		return sdkerrors.ErrPackAny.Wrapf("can't proto marshal %T", m)
	}
	// Now we set the anyValue type with cache. The cache value
	// is the proto message itself before being converted into
	// an anyValue.
	anyValue, err := cdctypes.NewAnyWithValue(m)
	if err != nil {
		return err
	}
	f.Attributes = anyValue

	return nil
}

// UnpackInterfaces is the method required to correctly unpack
// an Any type into an interface registered in the codec.
func (f *Forwarding) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if f == nil {
		return ErrNilPointer.Wrap("forwarding is not set")
	}

	var attributes ForwardingAttributes

	return unpacker.UnpackAny(f.Attributes, &attributes)
}

// ====================================================================================================
// Payload
// ====================================================================================================

// NewPayload returns a validated instance reference of
// an orbiter payload. Empty preActions slice is normalized to nil.
func NewPayload(
	forwarding *Forwarding,
	preActions ...*Action,
) (*Payload, error) {
	if len(preActions) == 0 {
		preActions = nil
	}

	payload := Payload{
		Forwarding: forwarding,
		PreActions: preActions,
	}

	return &payload, payload.Validate()
}

// Validate returns an error if the payload fields are
// not valid.
func (p *Payload) Validate() error {
	if p == nil {
		return ErrNilPointer.Wrap("payload is not set")
	}

	visitedIDs := make(map[int32]any)
	for _, action := range p.PreActions {
		if _, found := visitedIDs[int32(action.Id)]; found {
			return fmt.Errorf("received repeated action ID: %v", action.ID())
		}
		visitedIDs[int32(action.Id)] = nil
	}

	for _, action := range p.PreActions {
		if err := action.Validate(); err != nil {
			return err
		}
	}

	return p.Forwarding.Validate()
}

var _ cdctypes.UnpackInterfacesMessage = &Payload{}

func (p *Payload) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if p == nil {
		return ErrNilPointer.Wrap("payload is not set")
	}

	if p.PreActions != nil {
		for _, a := range p.PreActions {
			if a != nil {
				if err := a.UnpackInterfaces(unpacker); err != nil {
					return err
				}
			}
		}
	}

	if p.Forwarding != nil {
		if err := p.Forwarding.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}

	return nil
}

// NewPayloadWrapper returns a validated instance reference
// to a payload wrapper.
func NewPayloadWrapper(
	forwarding *Forwarding,
	preActions ...*Action,
) (*PayloadWrapper, error) {
	payload, err := NewPayload(forwarding, preActions...)
	if err != nil {
		return nil, err
	}
	payloadWrapper := PayloadWrapper{
		Orbiter: payload,
	}

	return &payloadWrapper, nil
}

// Validate returns an error if the orbiter payload wrapper
// contains non valid fields.
func (pw *PayloadWrapper) Validate() error {
	if pw == nil {
		return ErrNilPointer.Wrap("payload wrapper is not set")
	}

	return pw.Orbiter.Validate()
}

var _ cdctypes.UnpackInterfacesMessage = &PayloadWrapper{}

func (pw *PayloadWrapper) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	return pw.Orbiter.UnpackInterfaces(unpacker)
}
