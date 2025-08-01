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

package mocks

import (
	"context"
	"errors"

	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

var _ interfaces.ControllerOrbit = &OrbitController{}

type OrbitController struct {
	Id types.ProtocolID
}

// ID implements types.OrbitController.
func (o *OrbitController) ID() types.ProtocolID {
	return o.Id
}

// Name implements types.OrbitController.
func (o *OrbitController) Name() string {
	return o.Id.String()
}

// HandlePacket implements types.OrbitController.
func (o *OrbitController) HandlePacket(ctx context.Context, _ *types.OrbitPacket) error {
	if CheckIfFailing(ctx) {
		return errors.New("error dispatching the orbit packet")
	}

	return nil
}

var _ interfaces.ControllerAction = &NoOpActionController{}

type NoOpActionController struct {
	Id types.ActionID
}

// ID implements types.ActionController.
func (a *NoOpActionController) ID() types.ActionID {
	return a.Id
}

// Name implements types.ActionController.
func (a *NoOpActionController) Name() string {
	return a.Id.String()
}

// HandlePacket implements types.ActionController.
func (a *NoOpActionController) HandlePacket(ctx context.Context, _ *types.ActionPacket) error {
	if CheckIfFailing(ctx) {
		return errors.New("error dispatching the action packet")
	}

	return nil
}

var _ interfaces.ControllerAdapter = &NoOpAdapterController{}

type NoOpAdapterController struct {
	Id types.ProtocolID
}

func (a *NoOpAdapterController) ID() types.ProtocolID {
	return a.Id
}

func (a *NoOpAdapterController) Name() string {
	return a.Id.String()
}

// AfterTransferHook implements types.AdapterProtocol.
func (a *NoOpAdapterController) AfterTransferHook(ctx context.Context, _ *types.Payload) error {
	if CheckIfFailing(ctx) {
		return errors.New("error in after transfer hook")
	}

	return nil
}

// BeforeTransferHook implements types.AdapterProtocol.
func (a *NoOpAdapterController) BeforeTransferHook(ctx context.Context, _ *types.Payload) error {
	if CheckIfFailing(ctx) {
		return errors.New("error in before transfer hook")
	}

	return nil
}

// ParsePayload implements types.AdapterProtocol.
func (a *NoOpAdapterController) ParsePayload(bz []byte) (bool, *types.Payload, error) {
	if string(bz) == "failing" {
		return false, nil, errors.New("error parsing payload")
	}

	return true, &types.Payload{}, nil
}
