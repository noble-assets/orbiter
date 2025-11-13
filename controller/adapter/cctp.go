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
	"encoding/binary"
	"errors"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/noble-assets/orbiter/v2/controller"
	"github.com/noble-assets/orbiter/v2/types"
	adaptertypes "github.com/noble-assets/orbiter/v2/types/component/adapter"
	"github.com/noble-assets/orbiter/v2/types/core"
)

var _ types.AdapterController = &CCTPAdapter{}

type CCTPKeeper interface {
	GetTokenPair(
		context.Context,
		uint32,
		[]byte,
	) (cctptypes.TokenPair, bool)
}

type CCTPAdapter struct {
	*controller.BaseController[core.ProtocolID]

	logger log.Logger
	parser *BytesParser

	cctpKeeper CCTPKeeper
}

func NewCCTPAdapter(
	cdc codec.Codec,
	logger log.Logger,
	cctpKeeper CCTPKeeper,
) (*CCTPAdapter, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	if cctpKeeper == nil {
		return nil, core.ErrNilPointer.Wrap("CCTP keeper cannot be nil")
	}

	id := core.PROTOCOL_CCTP
	baseController, err := controller.NewBase(id)
	if err != nil {
		return nil, err
	}

	parser, err := NewBytesParser(cdc)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to create bytes parser")
	}

	return &CCTPAdapter{
		logger:         logger.With(core.AdapterControllerName, baseController.Name()),
		BaseController: baseController,
		parser:         parser,
		cctpKeeper:     cctpKeeper,
	}, nil
}

func (c *CCTPAdapter) ParsePayloadMessage(bz []byte) (uint64, *core.Payload, error) {
	nonceLen := cctptypes.SenderIndex - cctptypes.NonceIndex
	if len(bz) < nonceLen {
		return 0, nil, errors.New("payload received is rong")
	}

	nonce := binary.BigEndian.Uint64(bz[:nonceLen])
	payload, err := c.parser.ParsePayload(bz[nonceLen:])
	if err != nil {
		return 0, nil, err
	}

	return nonce, payload, nil
}

// ParsePacket implements types.AdapterController.
func (c *CCTPAdapter) ParsePacket(
	ctx context.Context,
	ccPacket adaptertypes.CrossChainPacket,
) (*types.ParsedData, error) {
	cctpPacket, ok := ccPacket.(*adaptertypes.CCTPCrossChainPacket)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&adaptertypes.CCTPCrossChainPacket{},
			ccPacket,
		)
	}

	transferNonce, payload, err := c.ParsePayloadMessage(ccPacket.Packet())
	if err != nil {
		return nil, errors.New("errrrrrror")
	}

	if transferNonce != cctpPacket.TransferNonce() {
		return nil, errors.New("nonce does not match")
	}

	return &types.ParsedData{
		Payload: *payload,
		Coin:    cctpPacket.Coin(),
	}, nil
}
