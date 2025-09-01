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

package forwarding

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/types/core"
)

const (
	HypTokenIDLen         = 32
	HypRecipientLen       = 32
	HypCustomHookLen      = 32
	HypHookMetadataPrefix = "0x"
	HypNobleTestnetDomain = 1196573006
	HypNobleMainnetDomain = 1313817164
)

// NewHyperlaneAttributes creates and validates new Hyperlane forwarding attributes.
// Returns an error if validation fails.
func NewHyperlaneAttributes(
	tokenID []byte,
	destinationDomain uint32,
	recipient []byte,
	customHookID []byte,
	customHookMetadata string,
	gasLimit math.Int,
	maxFee sdk.Coin,
) (*HypAttributes, error) {
	attr := &HypAttributes{
		TokenId:            tokenID,
		DestinationDomain:  destinationDomain,
		Recipient:          recipient,
		CustomHookId:       customHookID,
		CustomHookMetadata: customHookMetadata,
		GasLimit:           gasLimit,
		MaxFee:             maxFee,
	}

	if err := attr.Validate(); err != nil {
		return nil, err
	}

	return attr, nil
}

// Validate performs validation on the Hyperlane attributes.
func (a *HypAttributes) Validate() error {
	if a == nil {
		return core.ErrNilPointer.Wrap("Hyperlane attributes are not set")
	}

	if len(a.TokenId) != HypTokenIDLen {
		return fmt.Errorf(
			"token ID must be %d bytes, received %d bytes",
			HypTokenIDLen,
			len(a.TokenId),
		)
	}

	if len(a.Recipient) != HypRecipientLen {
		return fmt.Errorf(
			"recipient must be %d bytes, received %d bytes",
			HypRecipientLen,
			len(a.Recipient),
		)
	}

	if l := len(a.CustomHookId); l != 0 && l != HypCustomHookLen {
		return fmt.Errorf(
			"custom hook ID must be %d bytes when set, received %d bytes",
			HypCustomHookLen,
			l,
		)
	}

	if a.DestinationDomain == HypNobleMainnetDomain ||
		a.DestinationDomain == HypNobleTestnetDomain {
		return fmt.Errorf("destination domain %d is a Noble domain", a.DestinationDomain)
	}

	if a.CustomHookMetadata != "" {
		if !strings.HasPrefix(a.CustomHookMetadata, HypHookMetadataPrefix) {
			return fmt.Errorf("hook metadata must have the %s prefix, got: %s",
				HypHookMetadataPrefix, a.CustomHookMetadata)
		}
		if _, err := hex.DecodeString(strings.TrimPrefix(a.CustomHookMetadata, HypHookMetadataPrefix)); err != nil {
			return fmt.Errorf("hook metadata must be hex-encoded: %w", err)
		}
	}

	return nil
}

var _ core.ForwardingAttributes = &HypAttributes{}

// CounterpartyID returns a string representation of the destination domain.
func (a *HypAttributes) CounterpartyID() string {
	return strconv.FormatUint(uint64(a.GetDestinationDomain()), 10)
}

// NewHyperlaneForwarding creates a new validated Hyperlane forwarding instance.
// It creates Hyperlane attributes from the provided parameters and combines them
// with a passthrough payload to create a complete forwarding configuration.
func NewHyperlaneForwarding(
	tokenID []byte,
	destinationDomain uint32,
	recipient []byte,
	customHookID []byte,
	customHookMetadata string,
	gasLimit math.Int,
	maxFee sdk.Coin,
	passthroughPayload []byte,
) (*core.Forwarding, error) {
	attributes, err := NewHyperlaneAttributes(
		tokenID,
		destinationDomain,
		recipient,
		customHookID,
		customHookMetadata,
		gasLimit,
		maxFee,
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to create Hyperlane attributes")
	}

	return core.NewForwarding(core.PROTOCOL_HYPERLANE, attributes, passthroughPayload)
}
