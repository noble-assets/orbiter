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

package executor

import (
	"context"
	"fmt"

	executortypes "orbiter.dev/types/component/executor"
	"orbiter.dev/types/core"
)

// InitGenesis initialize the state of the component with a genesis state.
func (e *Executor) InitGenesis(ctx context.Context, g *executortypes.GenesisState) error {
	if g == nil {
		return core.ErrNilPointer.Wrap("executor genesis")
	}
	for _, id := range g.PausedActionIds {
		if err := id.Validate(); err != nil {
			return fmt.Errorf("invalid paused action ID: %w", err)
		}

		if err := e.SetPausedAction(ctx, id); err != nil {
			return fmt.Errorf("error setting genesis paused action ID: %w", err)
		}
	}

	return nil
}

// ExportGenesis returns the current state of the component into a genesis state.
func (e *Executor) ExportGenesis(ctx context.Context) *executortypes.GenesisState {
	paused, err := e.GetPausedActions(ctx)
	if err != nil {
		e.logger.Error("error exporting paused actions", "err", err.Error())
	}

	return &executortypes.GenesisState{
		PausedActionIds: paused,
	}
}
