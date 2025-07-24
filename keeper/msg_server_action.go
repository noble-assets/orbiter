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

	"orbiter.dev/types"
)

// PauseAction implements types.MsgServer.
func (m msgServer) PauseAction(
	ctx context.Context,
	msg *types.MsgPauseAction,
) (*types.MsgPauseActionResponse, error) {
	if err := m.CheckIsAuthority(msg.Signer); err != nil {
		return nil, err
	}

	actionComp := m.ActionComponent()

	if err := actionComp.Pause(ctx, msg.ActionId); err != nil {
		return nil, types.ErrUnableToPause.Wrapf(
			"action: %s", err.Error(),
		)
	}

	return &types.MsgPauseActionResponse{}, nil
}

// UnpauseAction implements types.MsgServer.
func (m msgServer) UnpauseAction(
	ctx context.Context,
	msg *types.MsgUnpauseAction,
) (*types.MsgUnpauseActionResponse, error) {
	if err := m.CheckIsAuthority(msg.Signer); err != nil {
		return nil, err
	}

	actionComp := m.ActionComponent()

	if err := actionComp.Unpause(ctx, msg.ActionId); err != nil {
		return nil, types.ErrUnableToUnpause.Wrapf(
			"action: %s", err.Error(),
		)
	}

	return &types.MsgUnpauseActionResponse{}, nil
}
