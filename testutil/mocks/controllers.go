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

	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var _ types.ForwardingController = &ForwardingController{}

type ForwardingController struct {
	id core.ProtocolID
}

func (o *ForwardingController) ID() core.ProtocolID {
	return o.id
}

func (o *ForwardingController) Name() string {
	return o.id.String()
}

func (o *ForwardingController) HandlePacket(ctx context.Context, _ *types.ForwardingPacket) error {
	if CheckIfFailing(ctx) {
		return errors.New("error dispatching the forwarding packet")
	}

	return nil
}

var _ types.ActionController = &NoOpActionController{}

type NoOpActionController struct {
	id core.ActionID
}

func (a *NoOpActionController) ID() core.ActionID {
	return a.id
}

func (a *NoOpActionController) Name() string {
	return a.id.String()
}

func (a *NoOpActionController) HandlePacket(ctx context.Context, _ *types.ActionPacket) error {
	if CheckIfFailing(ctx) {
		return errors.New("error dispatching the action packet")
	}

	return nil
}

var _ types.AdapterController = &NoOpAdapterController{}

type NoOpAdapterController struct {
	id core.ProtocolID
}

func (a *NoOpAdapterController) ID() core.ProtocolID {
	return a.id
}

func (a *NoOpAdapterController) Name() string {
	return a.id.String()
}

// ParsePacket implements types.AdapterController.
func (a *NoOpAdapterController) ParsePacket(bz []byte) (*types.ParsedData, error) {
	if string(bz) == "failing" {
		return nil, errors.New("error parsing payload")
	}

	return &types.ParsedData{}, nil
}
