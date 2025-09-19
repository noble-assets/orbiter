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

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/noble-assets/orbiter/types/controller/forwarding"
)

var _ forwarding.InternalHandler = InternalHandler{}

type InternalHandler struct {
	*BankKeeper
}

func NewInternalHandler() *InternalHandler {
	return &InternalHandler{
		NewBankKeeper(),
	}
}

// Send implements forwarding.InternalHandler.
func (h InternalHandler) Send(
	ctx context.Context,
	msg *banktypes.MsgSend,
) (*banktypes.MsgSendResponse, error) {
	fromBal := h.Balances[msg.FromAddress]

	if msg.Amount.Len() != 1 {
		return nil, errors.New("only one coin is accepter")
	}

	coin := msg.Amount[0]

	if fromBal.AmountOf(coin.Denom).LT(coin.Amount) {
		return nil, errors.New("not enough balance")
	}

	h.Balances[msg.FromAddress] = fromBal.Sub(coin)
	h.Balances[msg.ToAddress] = h.Balances[msg.ToAddress].Add(coin)

	return &banktypes.MsgSendResponse{}, nil
}
