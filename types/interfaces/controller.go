package interfaces

import (
	"context"

	"orbiter.dev/types"
)

// OrbitController defines the behavior an orbit packet
// controller has to implement.
type OrbitController interface {
	BaseController[types.ProtocolID]
	PacketHandler[*types.OrbitPacket]
}

// ActionController defines the behavior an action packet
// controller has to implement.
type ActionController interface {
	BaseController[types.ActionID]
	PacketHandler[*types.ActionPacket]
}

// AdapterController defines the behavior expected from a specific
// protocol adapter.
type AdapterController interface {
	BaseController[types.ProtocolID]
	PayloadParser
	// BeforeTransferHook allows to execute logic BEFORE completing
	// the cross-chain transfer.
	BeforeTransferHook(context.Context, *types.Payload) error
	// AfterTransferHook allows to execute logic AFTER completing
	// the cross-chain transfer.
	AfterTransferHook(context.Context, *types.Payload) error
}

// BaseController defines the behavior common to
// all controllers.
type BaseController[ID IdentifierConstraint] interface {
	Routable[ID]
	Name() string
}
