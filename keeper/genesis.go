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

package keeper

import (
	"context"
	"errors"
	"fmt"

	"orbiter.dev/types"
)

// InitGenesis initialize the state of the Orbiter module with
// a genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, g types.GenesisState) {
	a := k.Adapter()
	if g.AdapterGenesis == nil {
		panic(errors.New("nil pointer: missing adapter genesis state"))
	}
	if err := a.InitGenesis(ctx, g.AdapterGenesis); err != nil {
		panic(fmt.Errorf("unable to initialize adapter genesis state %w", err))
	}

	// TODO: add dispatcher

	f := k.Forwarder()
	if g.ForwarderGenesis == nil {
		panic(errors.New("nil pointer: missing forwarder genesis state"))
	}
	if err := f.InitGenesis(ctx, g.ForwarderGenesis); err != nil {
		panic(fmt.Errorf("unable to initialize forwarder genesis state %w", err))
	}

	e := k.Executor()
	if g.ExecutorGenesis == nil {
		panic(errors.New("nil pointer: missing forwarder genesis state"))
	}
	if err := e.InitGenesis(ctx, g.ExecutorGenesis); err != nil {
		panic(fmt.Errorf("unable to initialize executor genesis state %w", err))
	}
}

// ExportGenesis returns the current state of the Orbiter module
// into a genesis state.
func (k *Keeper) ExportGenesis(ctx context.Context) *types.GenesisState {
	return &types.GenesisState{
		AdapterGenesis:   k.adapter.ExportGenesis(ctx),
		ForwarderGenesis: k.forwarder.ExportGenesis(ctx),
		ExecutorGenesis:  k.executor.ExportGenesis(ctx),
	}
}
