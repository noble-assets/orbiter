package identifier

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const orbitIDSeparator = ":"

// OrbitID is an internal type used to uniquely
// represent a source or a destination of a cross-chain
// transfer and the bridge protocol used.
type OrbitID struct {
	ProtocolID ProtocolID
	// Protocol specific identifier of a counterparty.
	CounterpartyID string
}

// NewOrbitID returns a validated orbit identifier instance.
func NewOrbitID(
	protocolID ProtocolID,
	counterpartyID string,
) (OrbitID, error) {
	attr := OrbitID{
		ProtocolID:     protocolID,
		CounterpartyID: counterpartyID,
	}

	return attr, attr.Validate()
}

// Validate returns an error if any of the orbit id field
// is not valid.
func (i OrbitID) Validate() error {
	if err := i.ProtocolID.Validate(); err != nil {
		return err
	}
	if i.CounterpartyID == "" {
		return errors.New("counterparty id cannot be empty string")
	}

	return nil
}

// ID generates an internal identifier for a tuple (bridge protocol, chain).
// The identifier allows to recover the protocol Id and the chain Id from its value.
func (i OrbitID) ID() string {
	return fmt.Sprintf("%d%s%s", i.ProtocolID.Uint32(), orbitIDSeparator, i.CounterpartyID)
}

// String returns the string representation of the id.
func (i OrbitID) String() string {
	return i.ID()
}

// ParseOrbitID returns a new orbit id instance from the string.
// Returns an error if the string is not a valid orbit
// id string.
func ParseOrbitID(str string) (OrbitID, error) {
	sepIndex := strings.Index(str, orbitIDSeparator)
	if sepIndex == -1 {
		return OrbitID{}, fmt.Errorf("invalid orbit ID format: missing separator in %s", str)
	}

	protocolIDStr := str[:sepIndex]
	counterpartyID := str[sepIndex+1:]

	id, err := strconv.ParseInt(protocolIDStr, 10, 32)
	if err != nil {
		return OrbitID{}, fmt.Errorf("invalid protocol ID: %w", err)
	}

	protocolID := ProtocolID(int32(id))
	orbitID, err := NewOrbitID(protocolID, counterpartyID)
	if err != nil {
		return OrbitID{}, fmt.Errorf("invalid orbit ID string %s: %w", str, err)
	}

	return orbitID, nil
}
