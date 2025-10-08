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
	"fmt"

	errorsmod "cosmossdk.io/errors"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

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
		return errorsmod.Wrap(ErrNilPointer, "action is not set")
	}
	if err := a.Id.Validate(); err != nil {
		return err
	}

	if a.Attributes == nil {
		return errorsmod.Wrap(ErrNilPointer, "action attributes are not set")
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
		return nil, errorsmod.Wrap(ErrNilPointer, "action is not set")
	}

	if a.Attributes == nil {
		return nil, errorsmod.Wrap(ErrNilPointer, "action attributes are not set")
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
		return errorsmod.Wrap(ErrNilPointer, "action is not set")
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
		return errorsmod.Wrap(ErrNilPointer, "action is not set")
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
		return errorsmod.Wrap(ErrNilPointer, "forwarding is not set")
	}
	if err := f.ProtocolId.Validate(); err != nil {
		return err
	}
	if f.Attributes == nil {
		return errorsmod.Wrap(ErrNilPointer, "forwarding attributes are not set")
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
		return nil, errorsmod.Wrap(ErrNilPointer, "forwarding is not set")
	}
	if f.Attributes == nil {
		return nil, errorsmod.Wrap(ErrNilPointer, "forwarding attributes are not set")
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
		return errorsmod.Wrap(ErrNilPointer, "forwarding is not set")
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
		return errorsmod.Wrap(ErrNilPointer, "forwarding is not set")
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
		return errorsmod.Wrap(ErrNilPointer, "payload is not set")
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
		return errorsmod.Wrap(ErrNilPointer, "payload is not set")
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
		return errorsmod.Wrap(ErrNilPointer, "payload wrapper is not set")
	}

	return pw.Orbiter.Validate()
}

var _ cdctypes.UnpackInterfacesMessage = &PayloadWrapper{}

func (pw *PayloadWrapper) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	return pw.Orbiter.UnpackInterfaces(unpacker)
}
