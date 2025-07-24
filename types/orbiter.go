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
)

// NewPayload returns a validated instance reference of
// an orbiter payload. Empty preActions slice is normalized to nil.
func NewPayload(
	orbit *Orbit,
	preActions []*Action,
) (*Payload, error) {
	if len(preActions) == 0 {
		preActions = nil
	}

	payload := Payload{
		Orbit:      orbit,
		PreActions: preActions,
	}

	return &payload, payload.Validate()
}

// Validate returns an error if the payload fields are
// not valid.
func (p *Payload) Validate() error {
	if p == nil {
		return ErrNilPointer.Wrap("payload is a nil pointer")
	}

	for _, action := range p.PreActions {
		if err := action.Validate(); err != nil {
			return err
		}
	}

	return p.Orbit.Validate()
}

var _ cdctypes.UnpackInterfacesMessage = &Payload{}

func (p *Payload) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	if p == nil {
		return ErrNilPointer.Wrap("payload is a nil pointer")
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

	if p.Orbit != nil {
		if err := p.Orbit.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}

	return nil
}

// NewPayloadWrapper returns a validated instance reference
// to a payload wrapper.
func NewPayloadWrapper(
	orbit *Orbit,
	preActions []*Action,
) (*PayloadWrapper, error) {
	payload, err := NewPayload(orbit, preActions)
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
		return ErrNilPointer.Wrap("payload wrapper is a nil pointer")
	}
	return pw.Orbiter.Validate()
}

var _ cdctypes.UnpackInterfacesMessage = &PayloadWrapper{}

func (pw *PayloadWrapper) UnpackInterfaces(unpacker cdctypes.AnyUnpacker) error {
	return pw.Orbiter.UnpackInterfaces(unpacker)
}
