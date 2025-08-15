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

package adapter

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	adaptertypes "orbiter.dev/types/component/adapter"
)

// InitGenesis initialize the state of the adapter component with a genesis state.
func (a *Adapter) InitGenesis(ctx context.Context, g *adaptertypes.GenesisState) error {
	if err := g.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid adapter genesis state")
	}
	if err := a.SetParams(ctx, g.Params); err != nil {
		return errorsmod.Wrap(err, "error setting genesis params")
	}

	return nil
}

// ExportGenesis returns the current state of the adapter component into a genesis state.
func (a *Adapter) ExportGenesis(ctx context.Context) *adaptertypes.GenesisState {
	return &adaptertypes.GenesisState{
		Params: a.GetParams(ctx),
	}
}
