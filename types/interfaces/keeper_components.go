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

	"cosmossdk.io/log"

	"orbiter.dev/types"
)

// OrbitComponent defines the behavior the Orbiter module
// expected from a type to act as an orbits component.
type OrbitComponent interface {
	Logger() log.Logger
	PacketHandler[*types.OrbitPacket]
	RouterProvider[types.ProtocolID, OrbitController]
	Pause(context.Context, types.ProtocolID, []string) error
	Unpause(context.Context, types.ProtocolID, []string) error
}

// ActionComponent defines the behavior the Orbiter module
// expected from a type to act as an actions component.
type ActionComponent interface {
	Logger() log.Logger
	PacketHandler[*types.ActionPacket]
	RouterProvider[types.ActionID, ActionController]
	Pause(context.Context, types.ActionID) error
	Unpause(context.Context, types.ActionID) error
}

// DispatcherComponent defines the behavior the Orbiter module
// expected from a type to act as a dispatcher.
type DispatcherComponent interface {
	Logger() log.Logger
	PayloadDispatcher
}

// AdapterComponent defines the behavior the Orbiter module
// expected from a type to act as a cross-chain adapter.
type AdapterComponent interface {
	Logger() log.Logger
	PayloadAdapter
	RouterProvider[types.ProtocolID, AdapterController]
}
