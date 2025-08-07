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
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

// ForwardingAttributes is the interface every protocol forwarding
// attribute type has to implement.
type ForwardingAttributes interface {
	proto.Message
	// Returns the destination chain identifier.
	CounterpartyID() string
}

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
		return ErrNilPointer.Wrap("forwarding is a nil pointer")
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
		return nil, ErrNilPointer.Wrap("forwarding is a nil pointer")
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
		return ErrNilPointer.Wrap("forwarding is a nil pointer")
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
		return ErrNilPointer.Wrap("forwarding is a nil pointer")
	}

	var attributes ForwardingAttributes

	return unpacker.UnpackAny(f.Attributes, &attributes)
}
