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

package controllers

import (
	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

var (
	_ interfaces.Controller[types.ActionID]   = &BaseController[types.ActionID]{}
	_ interfaces.Controller[types.ProtocolID] = &BaseController[types.ProtocolID]{}
)

// NewBaseController returns a new instance of a validated BaseController.
func NewBaseController[ID interfaces.IdentifierConstraint](id ID) (*BaseController[ID], error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}

	return &BaseController[ID]{
		id: id,
	}, nil
}

// BaseController is a generic types that implements the types.BaseController interface.
type BaseController[ID interfaces.IdentifierConstraint] struct {
	id ID
}

// ID returns the controller identifier.
func (b *BaseController[ID]) ID() ID {
	return b.id
}

// Name returns the controller name.
func (b *BaseController[ID]) Name() string {
	return b.id.String()
}
