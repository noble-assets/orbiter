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

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	orbitertypes "github.com/noble-assets/orbiter/v2/types"
	"github.com/noble-assets/orbiter/v2/types/core"
)

var _ orbitertypes.QueryServer = &queryServer{}

type queryServer struct {
	*Keeper
}

func NewQueryServer(k *Keeper) orbitertypes.QueryServer {
	return &queryServer{Keeper: k}
}

func (q *queryServer) ActionIDs(
	_ context.Context,
	req *orbitertypes.QueryActionIDsRequest,
) (*orbitertypes.QueryActionIDsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	ids := map[int32]string{}
	for id, action := range core.ActionID_name {
		if action == core.ACTION_UNSUPPORTED.String() {
			continue
		}
		ids[id] = action
	}

	return &orbitertypes.QueryActionIDsResponse{
		ActionIds: ids,
	}, nil
}

func (q *queryServer) ProtocolIDs(
	_ context.Context,
	req *orbitertypes.QueryProtocolIDsRequest,
) (*orbitertypes.QueryProtocolIDsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	ids := map[int32]string{}
	for id, protocol := range core.ProtocolID_name {
		if protocol == core.PROTOCOL_UNSUPPORTED.String() {
			continue
		}
		ids[id] = protocol
	}

	return &orbitertypes.QueryProtocolIDsResponse{
		ProtocolIds: ids,
	}, nil
}
