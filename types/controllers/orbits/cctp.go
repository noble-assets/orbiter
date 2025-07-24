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

package orbits

import (
	"errors"
	"fmt"

	"orbiter.dev/types"
)

var _ types.OrbitAttributes = &CCTPAttributes{}

// CounterpartyID implements types.OrbitAttributes.
func (a *CCTPAttributes) CounterpartyID() string {
	return fmt.Sprintf("%d", a.GetDestinationDomain())
}

// NewCCTPAttributes returns a validated instance of the CCTP attributes.
func NewCCTPAttributes(
	destinationDomain uint32,
	mintRecipient []byte,
	destinationCaller []byte,
) (*CCTPAttributes, error) {
	attr := CCTPAttributes{
		DestinationDomain: destinationDomain,
		MintRecipient:     mintRecipient,
		DestinationCaller: destinationCaller,
	}
	return &attr, attr.Validate()
}

// Validate returns an error if the CCTP attributes are not valid.
func (a *CCTPAttributes) Validate() error {
	if a.DestinationDomain == types.CCTPNobleDomain {
		return errors.New("destination domain cannot be Noble")
	}
	if len(a.MintRecipient) == 0 {
		return errors.New("mint recipient cannot be empty")
	}
	if len(a.DestinationCaller) == 0 {
		return errors.New("destination caller cannot be empty")
	}
	return nil
}

// NewCCTPOrbit returns a reference to a validated CCTP orbit.
func NewCCTPOrbit(
	destinationDomain uint32,
	mintRecipient []byte,
	destinationCaller []byte,
	passthroughPayload []byte,
) (*types.Orbit, error) {
	attributes, err := NewCCTPAttributes(destinationDomain, mintRecipient, destinationCaller)
	if err != nil {
		return nil, err
	}

	return types.NewOrbit(types.PROTOCOL_CCTP, attributes, passthroughPayload)
}
