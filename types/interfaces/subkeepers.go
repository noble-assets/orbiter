package interfaces

import (
	"context"

	"cosmossdk.io/log"

	"orbiter.dev/types"
)

// ActionsSubkeeper defines the behavior the Orbiter module
// expected from a type to act as an actions subkeeper.
type ActionsSubkeeper interface {
	Logger() log.Logger
	PacketHandler[*types.ActionPacket]
	RouterProvider[types.ActionID, ActionController]
	Pause(context.Context, types.ActionID) error
	Unpause(context.Context, types.ActionID) error
}
