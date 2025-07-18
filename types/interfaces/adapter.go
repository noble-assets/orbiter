package interfaces

import (
	"context"

	"orbiter.dev/types"
)

// PayloadAdapter defines the behavior expected by the adapter to handle
// a generic orbiter payload.
type PayloadAdapter interface {
	// ParsePayload allows to parse and validate if the
	// input bytes represent an orbiter payload.
	ParsePayload(types.ProtocolID, []byte) (bool, *types.Payload, error)
	// BeforeTransferHook allows to execute logic BEFORE completing
	// the cross-chain transfer with the specific adapter defined by
	// the protocol ID.
	BeforeTransferHook(context.Context, types.ProtocolID, *types.Payload) error
	// AfterTransferHook allows to execute logic AFTER completing
	// the cross-chain transfer with the specific adapter defined by
	// the protocol ID.
	AfterTransferHook(context.Context, types.ProtocolID, string, *types.Payload) error
}
