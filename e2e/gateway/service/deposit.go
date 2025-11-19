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
	"math/big"

	bind "github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/noble-assets/orbiter/e2e/gateway/types"
)

func (s *Service) DepositForBurnWithOrbiter(
	txOpts *bind.TransactOpts,
	amount, blocktimeDeadline *big.Int,
	permitSignature, orbiterPayload []byte,
) (*gethtypes.Transaction, error) {
	return s.gateway.Instance().
		DepositForBurnWithOrbiter(txOpts, amount, blocktimeDeadline, permitSignature, orbiterPayload)
}

func (s *Service) ParseDepositForBurnEvents(
	receipt *gethtypes.Receipt,
) ([]*types.OrbiterGatewayCCTPDepositForBurnWithOrbiter, error) {
	var events []*types.OrbiterGatewayCCTPDepositForBurnWithOrbiter

	for _, log := range receipt.Logs {
		if log.Address != s.GatewayAddress() {
			continue
		}

		event, err := s.gateway.Instance().ParseDepositForBurnWithOrbiter(*log)
		if err != nil {
			continue
		}

		events = append(events, event)
	}

	return events, nil
}
