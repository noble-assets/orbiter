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

package types

import (
	"fmt"

	"orbiter.dev/types/component/adapter"
	"orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/component/executor"
	"orbiter.dev/types/component/forwarder"
)

// DefaultGenesisState returns the default values for the Orbiter module
// initial state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		AdapterGenesis:    adapter.DefaultGenesisState(),
		DispatcherGenesis: dispatcher.DefaultGenesisState(),
		ForwarderGenesis:  forwarder.DefaultGenesisState(),
		ExecutorGenesis:   executor.DefaultGenesisState(),
	}
}

// Validate returns an error if any of the genesis fields is not valid.
func (g *GenesisState) Validate() error {
	if err := g.AdapterGenesis.Validate(); err != nil {
		return fmt.Errorf("error validating adapter component genesis state: %w", err)
	}

	if err := g.DispatcherGenesis.Validate(); err != nil {
		return fmt.Errorf("error validating dispatcher component genesis state: %w", err)
	}

	if err := g.ForwarderGenesis.Validate(); err != nil {
		return fmt.Errorf("error validating forwarder component genesis state: %w", err)
	}

	if err := g.ExecutorGenesis.Validate(); err != nil {
		return fmt.Errorf("error validating executor component genesis state: %w", err)
	}

	return nil
}
