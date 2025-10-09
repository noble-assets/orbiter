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

	ethcommon "github.com/ethereum/go-ethereum/common"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var _ orbitertypes.QueryServer = &queryServer{}

type queryServer struct {
	*Keeper
}

func NewQueryServer(k *Keeper) orbitertypes.QueryServer {
	return &queryServer{k}
}

func (s *queryServer) PendingPayloads(
	ctx context.Context,
	req *orbitertypes.QueryPendingPayloadsRequest,
) (*orbitertypes.QueryPendingPayloadsResponse, error) {
	hashes, pageRes, err := query.CollectionPaginate(
		ctx,
		s.pendingPayloads,
		req.Pagination,
		func(hash []byte, _ core.PendingPayload) (string, error) {
			return ethcommon.BytesToHash(hash).Hex(), nil
		},
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to paginate pending payloads")
	}

	return &orbitertypes.QueryPendingPayloadsResponse{
		Hashes:     hashes,
		Pagination: pageRes,
	}, nil
}
