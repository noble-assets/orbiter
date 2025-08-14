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

package dispatcher

import (
	"cosmossdk.io/errors"

	"orbiter.dev/types/core"
)

// DefaultGenesisState returns the default values for the component initial state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		DispatchedAmounts: []DispatchedAmountEntry{},
		DispatchedCounts:  []DispatchCountEntry{},
	}
}

// Validate returns an error if any of the genesis field is not valid.
func (g *GenesisState) Validate() error {
	if g == nil {
		return core.ErrNilPointer.Wrap("dispatcher genesis")
	}

	for _, a := range g.DispatchedAmounts {
		if a.Denom == "" {
			return errors.Wrap(core.ErrValidation, "cannot set empty denom")
		}
		if a.SourceId == nil {
			return errors.Wrap(core.ErrValidation, "missing source cross-chain ID")
		}
		if err := a.SourceId.Validate(); err != nil {
			return errors.Wrap(err, "failed to create source cross-chain ID")
		}
		if a.DestinationId == nil {
			return errors.Wrap(core.ErrValidation, "missing destination cross-chain ID")
		}
		if err := a.DestinationId.Validate(); err != nil {
			return errors.Wrap(err, "failed to create destination cross-chain ID")
		}
		if a.AmountDispatched.Incoming.IsNegative() || a.AmountDispatched.Outgoing.IsNegative() {
			return errors.Wrap(core.ErrValidation, "cannot set negative amounts")
		}
		if !a.AmountDispatched.Incoming.IsPositive() && !a.AmountDispatched.Outgoing.IsPositive() {
			return errors.Wrap(
				core.ErrValidation,
				"cannot set incoming and outgoing amounts equal to zero",
			)
		}
	}

	for _, c := range g.DispatchedCounts {
		if c.Count == 0 {
			return errors.Wrap(core.ErrValidation, "cannot set zero count")
		}
		if c.SourceId == nil {
			return errors.Wrap(core.ErrValidation, "missing source cross-chain ID")
		}
		if err := c.SourceId.Validate(); err != nil {
			return errors.Wrap(err, "failed to create source cross-chain ID")
		}
		if c.DestinationId == nil {
			return errors.Wrap(core.ErrValidation, "missing destination cross-chain ID")
		}
		if err := c.DestinationId.Validate(); err != nil {
			return errors.Wrap(err, "failed to create destination cross-chain ID")
		}
	}

	return nil
}
