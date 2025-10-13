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

package core

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

const (
	MaxCounterpartyIDLength = 32
)

type IdentifierConstraint interface {
	ProtocolID | ActionID
	Validate() error
	String() string
}

func NewActionIDFromString(id string) (ActionID, error) {
	val, exists := ActionID_value[id]
	if !exists {
		return ACTION_UNSUPPORTED, fmt.Errorf("action ID %s does not exist", id)
	}
	actionID, err := NewActionID(val)
	if err != nil {
		return ACTION_UNSUPPORTED, fmt.Errorf("action ID %s is not supported", err.Error())
	}

	return actionID, nil
}

// NewActionID returns a validated action ID from an int32. If
// the validation fails, the returned value signals an unsupported
// action and an error is returned along with it.
func NewActionID(id int32) (ActionID, error) {
	actionID := ActionID(id)
	if err := actionID.Validate(); err != nil {
		return ACTION_UNSUPPORTED, err
	}

	return actionID, nil
}

// Validate returns an error if the ID is not valid.
func (id ActionID) Validate() error {
	if id == ACTION_UNSUPPORTED {
		return errorsmod.Wrapf(ErrIDNotSupported, "action ID: %s", id.String())
	}
	if _, found := ActionID_name[int32(id)]; !found {
		return fmt.Errorf("action ID is unknown: %d", int32(id))
	}

	return nil
}

func NewProtocolIDFromString(id string) (ProtocolID, error) {
	val, exists := ProtocolID_value[id]
	if !exists {
		return PROTOCOL_UNSUPPORTED, fmt.Errorf("protocol ID %s does not exist", id)
	}
	protocolID, err := NewProtocolID(val)
	if err != nil {
		return PROTOCOL_UNSUPPORTED, fmt.Errorf("protocol ID %s is not supported", err.Error())
	}

	return protocolID, nil
}

// NewProtocolID returns a validated protocol ID from an int32. If
// the validation fails, the returned ID is the default ID.
func NewProtocolID(id int32) (ProtocolID, error) {
	protocolID := ProtocolID(id)
	if err := protocolID.Validate(); err != nil {
		return PROTOCOL_UNSUPPORTED, err
	}

	return protocolID, nil
}

// Validate returns an error if the ID is not valid.
func (id ProtocolID) Validate() error {
	if id == PROTOCOL_UNSUPPORTED {
		return errorsmod.Wrapf(ErrIDNotSupported, "protocol ID: %s", id.String())
	}
	// Check if the protocol ID exists in the proto generated enum map
	if _, found := ProtocolID_name[int32(id)]; !found {
		return fmt.Errorf("protocol ID is unknown: %d", int32(id))
	}

	return nil
}

func (id ProtocolID) Uint32() uint32 {
	return uint32(id) //nolint:gosec
}

// NewCrossChainID returns a validated cross-chain identifier instance.
func NewCrossChainID(
	protocolID ProtocolID,
	counterpartyID string,
) (CrossChainID, error) {
	id := CrossChainID{
		ProtocolId:     protocolID,
		CounterpartyId: counterpartyID,
	}

	err := id.Validate()
	if err != nil {
		return CrossChainID{}, errorsmod.Wrap(err, "invalid cross-chain ID")
	}

	return id, nil
}

// Validate returns an error if any of the cross-chain ID field
// is not valid.
func (i CrossChainID) Validate() error {
	if err := i.ProtocolId.Validate(); err != nil {
		return err
	}
	if err := ValidateCounterpartyID(i.CounterpartyId, i.ProtocolId); err != nil {
		return err
	}

	return nil
}

func ValidateCounterpartyID(id string, protocol ProtocolID) error {
	if id == "" {
		return errors.New("counterparty ID cannot be empty string")
	}

	if len(id) > MaxCounterpartyIDLength {
		return fmt.Errorf(
			"counterparty ID cannot contain more than %d characters",
			MaxCounterpartyIDLength,
		)
	}

	var valid bool
	switch protocol {
	case PROTOCOL_IBC:
		valid = channeltypes.IsValidChannelID(id)
	case PROTOCOL_CCTP, PROTOCOL_HYPERLANE:
		valid = isInteger(id)
	case PROTOCOL_INTERNAL:
		valid = true
	case PROTOCOL_UNSUPPORTED:
		valid = false
	default:
		valid = false
	}

	if !valid {
		return fmt.Errorf("invalid counterparty ID for protocol %s", protocol.String())
	}

	return nil
}

// isInteger returns true if the string can be converted to
// an integer, false otherwise.
func isInteger(s string) bool {
	_, err := strconv.Atoi(s)

	return err == nil
}

// ID generates an internal identifier for a tuple (bridge protocol, chain).
// The identifier allows to recover the protocol Id and the chain Id from its value.
func (i CrossChainID) ID() string {
	return fmt.Sprintf("%d%s%s", i.ProtocolId.Uint32(), crosschainIDSeparator, i.CounterpartyId)
}

// String returns the string representation of the id.
func (i CrossChainID) String() string {
	return i.ID()
}

// ParseCrossChainID returns a new cross-chain ID instance from the string.
// Returns an error if the string is not a valid orbit
// id string.
func ParseCrossChainID(str string) (CrossChainID, error) {
	sepIndex := strings.Index(str, crosschainIDSeparator)
	if sepIndex == -1 {
		return CrossChainID{}, fmt.Errorf(
			"invalid cross-chain ID format: missing separator in %s",
			str,
		)
	}

	protocolIDStr := str[:sepIndex]
	counterpartyID := str[sepIndex+1:]

	id, err := strconv.ParseInt(protocolIDStr, 10, 32)
	if err != nil {
		return CrossChainID{}, errorsmod.Wrap(err, "invalid protocol ID")
	}

	protocolID := ProtocolID(int32(id))
	ccID, err := NewCrossChainID(protocolID, counterpartyID)
	if err != nil {
		return CrossChainID{}, errorsmod.Wrapf(err, "invalid cross-chain ID string %s", str)
	}

	return ccID, nil
}
