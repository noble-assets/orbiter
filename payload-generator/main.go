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

package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"orbiter.dev/testutil"
)

func main() {
	// NOTE: this is required to be called to correctly set the bech32 prefix
	testutil.SetSDKConfig()

	// Setup the TUI model and run it
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	runModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Print the full payload to stdout when exiting
	//
	// NOTE: This is not handled within the charm stuff to enable copying the full thing.
	// Within the charm TUI, the output would be truncated to the size of the window.
	if runModel != nil {
		m, ok := runModel.(model)
		if !ok {
			log.Fatal(fmt.Errorf("unexpected model; got %T", runModel))
		}

		fmt.Println(m.payload)
	}
}
