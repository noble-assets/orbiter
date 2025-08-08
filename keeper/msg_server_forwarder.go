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
	"errors"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"orbiter.dev/controller/forwarding"
	"orbiter.dev/types/component/forwarder"
	"orbiter.dev/types/core"
)

var _ forwarder.MsgServer = &msgServerForwarder{}

// msgServerForwarder is the server used to handle messages
// for the forwarder component.
type msgServerForwarder struct {
	// Keeper is the main Orbiter keeper.
	*Keeper
}

func NewMsgServerForwarder(keeper *Keeper) msgServerForwarder {
	return msgServerForwarder{Keeper: keeper}
}

// PauseProtocol implements forwarder.MsgServer.
func (s msgServerForwarder) PauseProtocol(
	ctx context.Context,
	msg *forwarder.MsgPauseProtocol,
) (*forwarder.MsgPauseProtocolResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	f := s.Forwarder()

	if err := f.Pause(ctx, msg.ProtocolId, nil); err != nil {
		return nil, core.ErrUnableToPause.Wrapf(
			"protocol: %s", err.Error(),
		)
	}

	return &forwarder.MsgPauseProtocolResponse{}, nil
}

// UnpauseProtocol implements forwarder.MsgServer.
func (s msgServerForwarder) UnpauseProtocol(
	ctx context.Context,
	msg *forwarder.MsgUnpauseProtocol,
) (*forwarder.MsgUnpauseProtocolResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	f := s.Forwarder()

	if err := f.Unpause(ctx, msg.ProtocolId, nil); err != nil {
		return nil, core.ErrUnableToUnpause.Wrapf(
			"protocol: %s", err.Error(),
		)
	}

	return &forwarder.MsgUnpauseProtocolResponse{}, nil
}

// PauseCounterparties implements forwarder.MsgServer.
func (s msgServerForwarder) PauseCounterparties(
	ctx context.Context,
	msg *forwarder.MsgPauseCounterparties,
) (*forwarder.MsgPauseCounterpartiesResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	f := s.Forwarder()

	if err := f.Pause(ctx, msg.ProtocolId, msg.CounterpartyIds); err != nil {
		return nil, core.ErrUnableToPause.Wrapf(
			"counterparties: %s", err.Error(),
		)
	}

	return &forwarder.MsgPauseCounterpartiesResponse{}, nil
}

// UnpauseCounterparties implements forwarder.MsgServer.
func (s msgServerForwarder) UnpauseCounterparties(
	ctx context.Context,
	msg *forwarder.MsgUnpauseCounterparties,
) (*forwarder.MsgUnpauseCounterpartiesResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	f := s.Forwarder()

	if err := f.Unpause(ctx, msg.ProtocolId, msg.CounterpartyIds); err != nil {
		return nil, core.ErrUnableToUnpause.Wrapf(
			"counterparties: %s", err.Error(),
		)
	}

	return &forwarder.MsgUnpauseCounterpartiesResponse{}, nil
}

// ReplaceDepositForBurn implements forwarder.MsgServer.
func (s msgServerForwarder) ReplaceDepositForBurn(
	ctx context.Context,
	msg *forwarder.MsgReplaceDepositForBurn,
) (*forwarder.MsgReplaceDepositForBurnResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	controller, found := s.Forwarder().Router().Route(core.PROTOCOL_CCTP)
	if !found {
		return nil, errors.New("cctp controller not found")
	}

	cctpController, ok := controller.(*forwarding.CCTPController)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			(*forwarding.CCTPController)(nil),
			controller,
		)
	}

	handler := cctpController.GetHandler()

	msgReplace := cctptypes.MsgReplaceDepositForBurn{
		From:                 core.ModuleAddress.String(),
		OriginalMessage:      msg.OriginalMessage,
		OriginalAttestation:  msg.OriginalAttestation,
		NewDestinationCaller: msg.NewDestinationCaller,
		NewMintRecipient:     msg.NewMintRecipient,
	}
	_, err := handler.ReplaceDepositForBurn(ctx, &msgReplace)
	if err != nil {
		return nil, err
	}

	return &forwarder.MsgReplaceDepositForBurnResponse{}, nil
}
