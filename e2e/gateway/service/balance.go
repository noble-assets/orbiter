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

package service

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BalanceAt retrieves the balance of an account at the current block.
func (s *Service) Balance(
	ctx context.Context,
	account common.Address,
) (*big.Int, error) {
	return s.client.BalanceAt(ctx, account, nil)
}

func (s *Service) SignerBalance(ctx context.Context) (*big.Int, error) {
	return s.Balance(ctx, s.signer.Address())
}

func (s *Service) USDCBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return s.usdc.Instance().BalanceOf(nil, address)
}

func (s *Service) SignerUSDCBalance(ctx context.Context) (*big.Int, error) {
	return s.usdc.Instance().BalanceOf(nil, s.signer.Address())
}
