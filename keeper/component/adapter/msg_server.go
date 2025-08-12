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

	"orbiter.dev/types"
	"orbiter.dev/types/component/adapter"
)

var _ adapter.MsgServer = &msgServer{}

// msgServer is the server used to handle messages
// for the executor component.
type msgServer struct {
	*Adapter
	types.Authorizator
}

func NewMsgServer(a *Adapter, auth types.Authorizator) msgServer {
	return msgServer{Adapter: a, Authorizator: auth}
}

// UpdateParams implements adapter.MsgServer.
func (s msgServer) UpdateParams(
	ctx context.Context,
	msg *adapter.MsgUpdateParams,
) (*adapter.MsgUpdateParamsResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	if err := s.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &adapter.MsgUpdateParamsResponse{}, nil
}
