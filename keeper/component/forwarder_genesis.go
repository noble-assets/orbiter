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

package component

import (
	"context"
	"fmt"

	"orbiter.dev/types/component/forwarder"
	"orbiter.dev/types/core"
)

// InitGenesis initialize the state of the component with a genesis state.
func (f *Forwarder) InitGenesis(ctx context.Context, g *forwarder.GenesisState) error {
	for _, id := range g.PausedProtocolId {
		if err := f.SetPausedProtocol(ctx, id); err != nil {
			return fmt.Errorf("error setting genesis paused protocol id: %w", err)
		}
	}

	for _, id := range g.PausedOrbitId {
		if err := f.SetPausedProtocolCounterparty(ctx, *id); err != nil {
			return fmt.Errorf("error setting genesis paused protocol-counterpaty: %w", err)
		}
	}

	return nil
}

// ExportGenesis returns the current state of the adapter component into a genesis state.
func (f *Forwarder) ExportGenesis(ctx context.Context) *forwarder.GenesisState {
	pausedProtocols, err := f.GetPausedProtocols(ctx)
	if err != nil {
		f.logger.Error("error exporting paused protocols", "err", err.Error())
	}
	// TODO: add method to get all protocols and counterparties.
	return &forwarder.GenesisState{
		PausedProtocolId: pausedProtocols,
		PausedOrbitId:    []*core.OrbitID{},
	}
}
