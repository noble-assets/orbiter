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

	errorsmod "cosmossdk.io/errors"

	"orbiter.dev/types/core"
)

func (e *Executor) IsActionPaused(ctx context.Context, id core.ActionID) (bool, error) {
	return e.PausedActions.Has(ctx, int32(id))
}

func (e *Executor) SetPausedAction(ctx context.Context, id core.ActionID) error {
	if err := id.Validate(); err != nil {
		return err
	}

	paused, err := e.IsActionPaused(ctx, id)
	if err != nil {
		return err
	}
	// Already paused, no-op
	if paused {
		return nil
	}

	return e.PausedActions.Set(ctx, int32(id))
}

func (e *Executor) SetUnpausedAction(ctx context.Context, id core.ActionID) error {
	if err := id.Validate(); err != nil {
		return err
	}

	paused, err := e.IsActionPaused(ctx, id)
	if err != nil {
		return err
	}
	// Already unpaused, no-op
	if !paused {
		return nil
	}

	return e.PausedActions.Remove(ctx, int32(id))
}

func (e *Executor) GetPausedActions(
	ctx context.Context,
) ([]core.ActionID, error) {
	iter, err := e.PausedActions.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = iter.Close()
	}()

	paused := make([]core.ActionID, 0, len(core.ActionID_name))
	for ; iter.Valid(); iter.Next() {
		k, err := iter.Key()
		if err != nil {
			return nil, err
		}

		id, err := core.NewActionID(k)
		if err != nil {
			return nil, errorsmod.Wrap(err, "cannot create action ID from iterator key")
		}
		paused = append(paused, id)
	}

	return paused, nil
}
