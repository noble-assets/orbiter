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

package types

import (
	"cosmossdk.io/math"
)

func NewAmountDispatched(
	incoming math.Int,
	outgoing math.Int,
) *AmountDispatched {
	return &AmountDispatched{
		Incoming: incoming,
		Outgoing: outgoing,
	}
}

type ChainAmountDispatched struct {
	orbitID          OrbitID
	amountDispatched AmountDispatched
}

func NewChainAmountDispatched(
	orbitID OrbitID,
	amountDispatched AmountDispatched,
) *ChainAmountDispatched {
	return &ChainAmountDispatched{
		orbitID:          orbitID,
		amountDispatched: amountDispatched,
	}
}

func (cad *ChainAmountDispatched) OrbitID() OrbitID {
	return cad.orbitID
}

func (cad *ChainAmountDispatched) AmountDispatched() AmountDispatched {
	return cad.amountDispatched
}

type TotalDispatched struct {
	chainsAmount map[string]ChainAmountDispatched
}

func NewTotalDispatched() *TotalDispatched {
	return &TotalDispatched{
		chainsAmount: make(map[string]ChainAmountDispatched),
	}
}

func (td *TotalDispatched) ChainAmount(counterpartyID string) ChainAmountDispatched {
	return td.chainsAmount[counterpartyID]
}

func (td *TotalDispatched) ChainsAmount() map[string]ChainAmountDispatched {
	return td.chainsAmount
}

func (td *TotalDispatched) SetAmountDispatched(counterpartyID string, cad ChainAmountDispatched) {
	td.chainsAmount[counterpartyID] = cad
}
