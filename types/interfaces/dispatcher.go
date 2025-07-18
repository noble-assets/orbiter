package interfaces

import (
	"context"

	"orbiter.dev/types"
)

// PayloadDispatcher defines the expected behavior from a type
// to be used as a payload dispatcher.
type PayloadDispatcher interface {
	// Dispatch the payload component to the proper handler.
	DispatchPayload(context.Context, *types.TransferAttributes, *types.Payload) error
}
