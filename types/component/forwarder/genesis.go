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

package forwarder

import (
	core "orbiter.dev/types/core"
)

// DefaultGenesisState returns the default values for the component initial state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		PausedProtocolIds:   []core.ProtocolID{},
		PausedCrossChainIds: []*core.CrossChainID{},
	}
}

// Validate returns an error if any of the genesis field is not valid.
func (g *GenesisState) Validate() error {
	for _, id := range g.PausedProtocolIds {
		if err := id.Validate(); err != nil {
			return err
		}
	}

	for _, id := range g.PausedCrossChainIds {
		if err := id.Validate(); err != nil {
			return err
		}
	}

	return nil
}
