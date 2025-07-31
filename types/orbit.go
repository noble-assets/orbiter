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
	"strconv"
	"strings"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"
)

// OrbitAttributes is the interface every protocol attribute type
// has to implement to be a valid orbit attribute.
type OrbitAttributes interface {
	proto.Message
	// Returns the destination chain identifier.
	CounterpartyID() string
}

// ====================================================================================================
// ID
// ====================================================================================================

// NewOrbitID returns a validated orbit identifier instance.
func NewOrbitID(
	protocolID ProtocolID,
	counterpartyID string,
) (OrbitID, error) {
	attr := OrbitID{
		ProtocolID:     protocolID,
		CounterpartyID: counterpartyID,
	}

	return attr, attr.Validate()
}

// OrbitID is an internal type used to uniquely
// represent a source or a destination of a cross-chain
// transfer and the bridge protocol used.
type OrbitID struct {
	ProtocolID ProtocolID
	// Protocol specific identifier of a counterparty.
	CounterpartyID string
}

// Validate returns an error if any of the orbit id field
// is not valid.
func (i OrbitID) Validate() error {
	if err := i.ProtocolID.Validate(); err != nil {
		return err
	}
	if i.CounterpartyID == "" {
		return errors.New("counterparty id cannot be empty string")
	}
	return nil
}

// ID generates an internal identifier for a tuple (bridge protocol, chain).
// The identifier allows to recover the protocol Id and the chain Id from its value.
func (i OrbitID) ID() string {
	return fmt.Sprintf("%d%s%s", i.ProtocolID.Uint32(), orbitIDSeparator, i.CounterpartyID)
}

// String returns the string representation of the id.
func (i OrbitID) String() string {
	return i.ID()
}

// ParseOrbitID returns a new orbit id instance from the string.
// Returns an error if the string is not a valid orbit
// id string.
func ParseOrbitID(str string) (OrbitID, error) {
	sepIndex := strings.Index(str, orbitIDSeparator)
	if sepIndex == -1 {
		return OrbitID{}, fmt.Errorf("invalid orbit ID format: missing separator in %s", str)
	}

	protocolIDStr := str[:sepIndex]
	counterpartyID := str[sepIndex+1:]

	id, err := strconv.ParseInt(protocolIDStr, 10, 32)
	if err != nil {
		return OrbitID{}, fmt.Errorf("invalid protocol ID: %w", err)
	}

	protocolID := ProtocolID(int32(id))
	orbitID, err := NewOrbitID(protocolID, counterpartyID)
	if err != nil {
		return OrbitID{}, fmt.Errorf("invalid orbit ID string %s: %w", str, err)
	}

	return orbitID, nil
}

// ====================================================================================================
// Orbits
// ====================================================================================================

var _ cdctypes.UnpackInterfacesMessage = &Orbit{}

// NewOrbit returns a reference to a validated orbit. This utility
// function automatically sets the attributes in the Any type of the
// return orbit.
//
// NOTE: passthroughPayload is currently ignored since CCTP does not allow to pass
// the payload along with the transfer.
func NewOrbit(id ProtocolID, a OrbitAttributes, passthroughPayload []byte) (*Orbit, error) {
	o := Orbit{
		ProtocolId:         id,
		PassthroughPayload: passthroughPayload,
	}
	err := o.SetAttributes(a)
	if err != nil {
		return nil, err
	}

	return &o, o.Validate()
}

// Validate returns an error if the orbit is not valid.
func (o *Orbit) Validate() error {
	if o == nil {
		return ErrNilPointer.Wrap("orbit is a nil pointer")
	}
	if err := o.ProtocolId.Validate(); err != nil {
		return err
	}
	if o.Attributes == nil {
		return ErrNilPointer.Wrap("orbit attributes are not set")
	}
	return nil
}

// ProtocolID returns the protocol ID associated with the orbit. If
// the id is not set, the default value is returned.
func (o *Orbit) ProtocolID() ProtocolID {
	if o != nil {
		return o.ProtocolId
	}
	return PROTOCOL_UNSUPPORTED
}

// CachedAttributes returns the attributes interface from the
// codec Any type. Returns nil if the orbit does not have
// attributes set.
func (o *Orbit) CachedAttributes() (OrbitAttributes, error) {
	if o == nil {
		return nil, ErrNilPointer.Wrap("orbit is a nil pointer")
	}
	if o.Attributes == nil {
		return nil, ErrNilPointer.Wrap("orbit attributes are not set")
	}
	av := o.Attributes.GetCachedValue()
	a, ok := av.(OrbitAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			(OrbitAttributes)(nil),
			av,
		)
	}
	return a, nil
}

// SetAttributes sets the orbit attributes into the orbit as codec Any type.
func (o *Orbit) SetAttributes(a OrbitAttributes) error {
	if o == nil {
		return ErrNilPointer.Wrap("orbit is a nil pointer")
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
	o.Attributes = anyValue
	return nil
}

// UnpackInterfaces is the method required to correctly unpack
// an Any type into an interface registered in the codec.
func (o *Orbit) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if o == nil {
		return ErrNilPointer.Wrap("orbit is a nil pointer")
	}
	var attributes OrbitAttributes
	return unpacker.UnpackAny(o.Attributes, &attributes)
}
