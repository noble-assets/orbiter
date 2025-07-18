package interfaces

import (
	"context"

	"orbiter.dev/types"
)

// PacketConstraint defines the packet types supported
// by the module.
type PacketConstraint interface {
	*types.OrbitPacket | *types.ActionPacket
}

// PacketHandler defines the behavior expected by a type
// capable of handling the information contained in a
// supported packet type.
type PacketHandler[T PacketConstraint] interface {
	HandlePacket(context.Context, T) error
}
