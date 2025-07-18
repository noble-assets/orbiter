package interfaces

import (
	"orbiter.dev/types"
)

// PayloadParser defines the behavior expected by a type capable of
// parsing a payload from its bytes representation.
type PayloadParser interface {
	// ParsePayload handle bytes and parse them into the
	// orbiter payload. It returns a boolean to inform if
	// the bytes represent an orbiter payload or not. The
	// parsing is executed only if the boolean is true.
	ParsePayload([]byte) (bool, *types.Payload, error)
}
