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
	protocolId ProtocolID,
	counterpartyId string,
) (*OrbitID, error) {
	attr := OrbitID{
		ProtocolID:     protocolId,
		CounterpartyID: counterpartyId,
	}

	return &attr, attr.Validate()
}

// OrbitID is an internal type used to represent
// uniquely a source or a destination of a cross-chain
// transfer and the protocol used.
type OrbitID struct {
	ProtocolID     ProtocolID
	CounterpartyID string
}

func (a *OrbitID) Validate() error {
	if err := a.ProtocolID.Validate(); err != nil {
		return err
	}
	if a.CounterpartyID == "" {
		return errors.New("counterparty identifier cannot be empty string")
	}
	return nil
}

// ID generates an internal identifier for a tuple (bridge protocol, chain).
// The identifier allows to recover the protocol Id and the chain Id from its value.
func (a *OrbitID) ID() string {
	return fmt.Sprintf("%d%s%s", a.ProtocolID.Uint32(), OrbitIDSeparator, a.CounterpartyID)
}

func (a *OrbitID) String() string {
	return a.ID()
}

func (a *OrbitID) FromString(str string) error {
	parts := strings.Split(str, OrbitIDSeparator)
	if len(parts) != 2 {
		return fmt.Errorf("invalid orbit ID format: %s", str)
	}

	id, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid protocol ID: %w", err)
	}

	protocolID := ProtocolID(int32(id))

	a.ProtocolID = protocolID
	a.CounterpartyID = parts[1]

	return nil
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
	if err := o.ProtocolId.Validate(); err != nil {
		return err
	}
	if o.Attributes == nil {
		return ErrNilPointer.Wrap("orbit attributes are not set")
	}
	return nil
}

// ProtocolID returns the ProtocolID. If the ProtocolID is not set,
// the default value is returned.
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
	// The interface we want to pack as any must
	// implement the proto Message interface.
	m, ok := a.(proto.Message)
	if !ok {
		return sdkerrors.ErrPackAny.Wrapf("can't proto marshal %T", m)
	}
	// Now we set the any type with cache. The cache value
	// is the proto message itself before being converted into
	// an any.
	any, err := cdctypes.NewAnyWithValue(m)
	if err != nil {
		return err
	}
	o.Attributes = any
	return nil
}

// UnpackInterfaces is the method required to correctly unpack
// an Any type into an interface registered in the codec.
func (o *Orbit) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	var attributes OrbitAttributes
	return unpacker.UnpackAny(o.Attributes, &attributes)
}
