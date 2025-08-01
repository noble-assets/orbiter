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

// ActionAttributes is the interface defining the expected behavior
// for a type to be used to perform actions on the orbiter module.
type ActionAttributes interface {
	proto.Message
}

// ====================================================================================================
// ID
// ====================================================================================================

// NewActionID returns a validated action ID from an int32. If
// the validation fails, the returned value signals an unsupported
// action and an error is returned along with it.
func NewActionID(id int32) (ActionID, error) {
	actionID := ActionID(id)
	if err := actionID.Validate(); err != nil {
		return ACTION_UNSUPPORTED, err
	}

	return actionID, nil
}

// Validate returns an error if the ID is not valid.
func (id ActionID) Validate() error {
	if id == ACTION_UNSUPPORTED {
		return ErrIDNotSupported.Wrapf("action id %s", id.String())
	}
	if _, found := ActionID_name[int32(id)]; !found {
		return ErrIDNotSupported.Wrapf("unknown action id %d", int32(id))
	}

	return nil
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
		return ErrNilPointer.Wrap("action is a nil pointer")
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
		return nil, ErrNilPointer.Wrap("action is a nil pointer")
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
		return ErrNilPointer.Wrap("action is a nil pointer")
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
		return ErrNilPointer.Wrap("action is a nil pointer")
	}
	var attributes ActionAttributes

	return unpacker.UnpackAny(a.Attributes, &attributes)
}
