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
	errorsmod "cosmossdk.io/errors"

	"github.com/noble-assets/orbiter/types/core"
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
		if err := a.Validate(); err != nil {
			return errorsmod.Wrap(err, "invalid dispatched amount")
		}
	}

	for _, c := range g.DispatchedCounts {
		if err := c.Validate(); err != nil {
			return errorsmod.Wrap(err, "invalid dispatched count")
		}
	}

	return nil
}
