package interfaces

import (
	"context"

	"cosmossdk.io/log"

	"orbiter.dev/types"
)

// OrbitSubkeeper defines the behavior the Orbiter module
// expected from a type to act as an orbits subkeeper.
type OrbitSubkeeper interface {
	Logger() log.Logger
	PacketHandler[*types.OrbitPacket]
	RouterProvider[types.ProtocolID, OrbitController]
	Pause(context.Context, types.ProtocolID, []string) error
	Unpause(context.Context, types.ProtocolID, []string) error
}

// ActionSubkeeper defines the behavior the Orbiter module
// expected from a type to act as an actions subkeeper.
type ActionSubkeeper interface {
	Logger() log.Logger
	PacketHandler[*types.ActionPacket]
	RouterProvider[types.ActionID, ActionController]
	Pause(context.Context, types.ActionID) error
	Unpause(context.Context, types.ActionID) error
}

// AdapterSubkeeper defines the behavior the Orbiter module
// expected from a type to act as a cross-chain adapter subkeeper.
type AdapterSubkeeper interface {
	Logger() log.Logger
	PayloadAdapter
	RouterProvider[types.ProtocolID, AdapterController]
}
