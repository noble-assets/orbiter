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
	"context"

	"cosmossdk.io/log"

	"orbiter.dev/types/core"
	"orbiter.dev/types/router"
)

type Loggable interface {
	Logger() log.Logger
}

// Forwarder defines the behavior a components must
// have to process forwardings.
type Forwarder interface {
	Loggable
	PacketHandler[*ForwardingPacket]
	router.RouterProvider[core.ProtocolID, ControllerForwarding]
	Pause(context.Context, core.ProtocolID, []string) error
	Unpause(context.Context, core.ProtocolID, []string) error
}

// Executor defines the behavior a components must
// have to process pre-actios.
type Executor interface {
	Loggable
	PacketHandler[*ActionPacket]
	router.RouterProvider[core.ActionID, ControllerAction]
	Pause(context.Context, core.ActionID) error
	Unpause(context.Context, core.ActionID) error
}

// Dispatcher defines the behavior a components must
// have to dispatch packets.
type Dispatcher interface {
	Loggable
	PayloadDispatcher
}

// Adapter defines the behavior a component must
// have to adapt cross-chain packets to the Orbiter.
type Adapter interface {
	Loggable
	PayloadAdapter
	router.RouterProvider[core.ProtocolID, ControllerAdapter]
}
