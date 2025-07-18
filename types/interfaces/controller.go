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

package interfaces

import (
	"context"

	"orbiter.dev/types"
)

// ControllerOrbit defines the behavior an orbit packet
// controller has to implement.
type ControllerOrbit interface {
	Controller[types.ProtocolID]
	PacketHandler[*types.OrbitPacket]
}

// ControllerAction defines the behavior an action packet
// controller has to implement.
type ControllerAction interface {
	Controller[types.ActionID]
	PacketHandler[*types.ActionPacket]
}

// ControllerAdapter defines the behavior expected from a specific
// protocol adapter.
type ControllerAdapter interface {
	Controller[types.ProtocolID]
	PayloadParser
	// BeforeTransferHook allows to execute logic BEFORE completing
	// the cross-chain transfer.
	BeforeTransferHook(context.Context, *types.Payload) error
	// AfterTransferHook allows to execute logic AFTER completing
	// the cross-chain transfer.
	AfterTransferHook(context.Context, *types.Payload) error
}

// Controller defines the behavior common to
// all controllers.
type Controller[ID IdentifierConstraint] interface {
	Routable[ID]
	Name() string
}
