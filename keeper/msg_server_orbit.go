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

// PauseProtocol implements types.MsgServer.
func (m msgServer) PauseProtocol(
	ctx context.Context,
	msg *types.MsgPauseProtocol,
) (*types.MsgPauseProtocolResponse, error) {
	if err := m.CheckIsAuthority(msg.Signer); err != nil {
		return nil, err
	}

	orbitComp := m.OrbitComponent()

	if err := orbitComp.Pause(ctx, msg.ProtocolId, nil); err != nil {
		return nil, types.ErrUnableToPause.Wrapf(
			"protocol: %s", err.Error(),
		)
	}

	return &types.MsgPauseProtocolResponse{}, nil
}

// PauseCounterparties implements types.MsgServer.
func (m msgServer) PauseCounterparties(
	ctx context.Context,
	msg *types.MsgPauseCounterparties,
) (*types.MsgPauseCounterpartiesResponse, error) {
	if err := m.CheckIsAuthority(msg.Signer); err != nil {
		return nil, err
	}

	orbitComp := m.OrbitComponent()

	if err := orbitComp.Pause(ctx, msg.ProtocolId, msg.CounterpartyIds); err != nil {
		return nil, types.ErrUnableToPause.Wrapf(
			"counterparties: %s", err.Error(),
		)
	}
	return &types.MsgPauseCounterpartiesResponse{}, nil
}

// UnpauseProtocol implements types.MsgServer.
func (m msgServer) UnpauseProtocol(
	ctx context.Context,
	msg *types.MsgUnpauseProtocol,
) (*types.MsgUnpauseProtocolResponse, error) {
	if err := m.CheckIsAuthority(msg.Signer); err != nil {
		return nil, err
	}

	orbitComp := m.OrbitComponent()

	if err := orbitComp.Unpause(ctx, msg.ProtocolId, nil); err != nil {
		return nil, types.ErrUnableToUnpause.Wrapf(
			"protocol: %s", err.Error(),
		)
	}

	return &types.MsgUnpauseProtocolResponse{}, nil
}

// UnpauseCounterparties implements types.MsgServer.
func (m msgServer) UnpauseCounterparties(
	ctx context.Context,
	msg *types.MsgUnpauseCounterparties,
) (*types.MsgUnpauseCounterpartiesResponse, error) {
	if err := m.CheckIsAuthority(msg.Signer); err != nil {
		return nil, err
	}

	orbitComp := m.OrbitComponent()

	if err := orbitComp.Unpause(ctx, msg.ProtocolId, msg.CounterpartyIds); err != nil {
		return nil, types.ErrUnableToUnpause.Wrapf(
			"counterparties: %s", err.Error(),
		)
	}

	return &types.MsgUnpauseCounterpartiesResponse{}, nil
}
