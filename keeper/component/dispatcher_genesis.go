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

	"orbiter.dev/types/component/dispatcher"
)

// InitGenesis initialize the state of the component with a genesis state.
func (d *Dispatcher) InitGenesis(ctx context.Context, g *dispatcher.GenesisState) error {
	for _, a := range g.DispatchedAmounts {
		if err := d.SetDispatchedAmount(ctx, *a.SourceId, *a.DestinationId, a.Denom, a.AmountDispatched); err != nil {
			return fmt.Errorf("error setting genesis dispatched amount: %w", err)
		}
	}

	for _, c := range g.DispatchedCounts {
		if err := d.SetDispatchedCounts(ctx, *c.SourceId, *c.DestinationId, uint32(c.Count)); err != nil {
			return fmt.Errorf("error setting genesis dispatched amount: %w", err)
		}
	}

	return nil
}

// ExportGenesis returns the current state of the component into a genesis state.
func (d *Dispatcher) ExportGenesis(ctx context.Context) *dispatcher.GenesisState {
	// TODO: add the methods to reteive all the data.
	return &dispatcher.GenesisState{
		DispatchedAmounts: []dispatcher.DispatchedAmountEntry{},
		DispatchedCounts:  []dispatcher.DispatchCountEntry{},
	}
}
