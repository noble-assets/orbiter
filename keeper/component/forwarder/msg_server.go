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

package forwarder

import (
	"context"
	"errors"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/noble-assets/orbiter/v2/controller/forwarding"
	"github.com/noble-assets/orbiter/v2/types"
	forwardertypes "github.com/noble-assets/orbiter/v2/types/component/forwarder"
	"github.com/noble-assets/orbiter/v2/types/core"
)

var _ forwardertypes.MsgServer = &msgServer{}

// msgServer is the server used to handle messages
// for the forwarder component.
type msgServer struct {
	*Forwarder
	types.Authorizer
}

func NewMsgServer(f *Forwarder, a types.Authorizer) msgServer {
	return msgServer{Forwarder: f, Authorizer: a}
}

func (s msgServer) PauseProtocol(
	ctx context.Context,
	msg *forwardertypes.MsgPauseProtocol,
) (*forwardertypes.MsgPauseProtocolResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	protocolID, err := core.NewProtocolIDFromString(msg.ProtocolId)
	if err != nil {
		return nil, core.ErrUnableToPause.Wrap(err.Error())
	}

	if err := s.Pause(ctx, protocolID, nil); err != nil {
		return nil, core.ErrUnableToPause.Wrapf(
			"error setting paused state: %s", err.Error(),
		)
	}

	if err := s.eventService.EventManager(ctx).Emit(
		ctx,
		&forwardertypes.EventProtocolPaused{ProtocolId: protocolID},
	); err != nil {
		return nil, core.ErrUnableToPause.Wrapf("failed to emit event: %s", err.Error())
	}

	return &forwardertypes.MsgPauseProtocolResponse{}, nil
}

func (s msgServer) UnpauseProtocol(
	ctx context.Context,
	msg *forwardertypes.MsgUnpauseProtocol,
) (*forwardertypes.MsgUnpauseProtocolResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	protocolID, err := core.NewProtocolIDFromString(msg.ProtocolId)
	if err != nil {
		return nil, core.ErrUnableToUnpause.Wrap(err.Error())
	}

	if err := s.Unpause(ctx, protocolID, nil); err != nil {
		return nil, core.ErrUnableToUnpause.Wrapf(
			"error setting unpaused state: %s", err.Error(),
		)
	}

	if err := s.eventService.EventManager(ctx).Emit(
		ctx,
		&forwardertypes.EventProtocolUnpaused{ProtocolId: protocolID},
	); err != nil {
		return nil, core.ErrUnableToUnpause.Wrapf("failed to emit event: %s", err.Error())
	}

	return &forwardertypes.MsgUnpauseProtocolResponse{}, nil
}

func (s msgServer) PauseCrossChains(
	ctx context.Context,
	msg *forwardertypes.MsgPauseCrossChains,
) (*forwardertypes.MsgPauseCrossChainsResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	protocolID, err := core.NewProtocolIDFromString(msg.ProtocolId)
	if err != nil {
		return nil, core.ErrUnableToPause.Wrapf("invalid protocol id: %s", err.Error())
	}

	if len(msg.CounterpartyIds) > core.MaxTargetCounterparties {
		return nil, core.ErrUnableToPause.Wrapf(
			"cannot pause more than %d counterparties in a transaction",
			core.MaxTargetCounterparties,
		)
	}

	if err := s.Pause(ctx, protocolID, msg.CounterpartyIds); err != nil {
		return nil, core.ErrUnableToPause.Wrapf(
			"error setting paused state: %s", err.Error(),
		)
	}

	if err := s.eventService.EventManager(ctx).Emit(
		ctx,
		&forwardertypes.EventCrossChainsPaused{
			ProtocolId:      protocolID,
			CounterpartyIds: msg.CounterpartyIds,
		},
	); err != nil {
		return nil, core.ErrUnableToPause.Wrapf("failed to emit event: %s", err.Error())
	}

	return &forwardertypes.MsgPauseCrossChainsResponse{}, nil
}

func (s msgServer) UnpauseCrossChains(
	ctx context.Context,
	msg *forwardertypes.MsgUnpauseCrossChains,
) (*forwardertypes.MsgUnpauseCrossChainsResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	protocolID, err := core.NewProtocolIDFromString(msg.ProtocolId)
	if err != nil {
		return nil, core.ErrUnableToUnpause.Wrapf("invalid protocol id: %s", err.Error())
	}

	if len(msg.CounterpartyIds) > core.MaxTargetCounterparties {
		return nil, core.ErrUnableToUnpause.Wrapf(
			"cannot unpause more than %d counterparties in a transaction",
			core.MaxTargetCounterparties,
		)
	}

	if err := s.Unpause(ctx, protocolID, msg.CounterpartyIds); err != nil {
		return nil, core.ErrUnableToUnpause.Wrapf(
			"error setting unpaused state: %s", err.Error(),
		)
	}

	if err := s.eventService.EventManager(ctx).Emit(
		ctx,
		&forwardertypes.EventCrossChainsUnpaused{
			ProtocolId:      protocolID,
			CounterpartyIds: msg.CounterpartyIds,
		},
	); err != nil {
		return nil, core.ErrUnableToUnpause.Wrapf("failed to emit event: %s", err.Error())
	}

	return &forwardertypes.MsgUnpauseCrossChainsResponse{}, nil
}

func (s msgServer) ReplaceDepositForBurn(
	ctx context.Context,
	msg *forwardertypes.MsgReplaceDepositForBurn,
) (*forwardertypes.MsgReplaceDepositForBurnResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	controller, found := s.router.Route(core.PROTOCOL_CCTP)
	if !found {
		return nil, errors.New("CCTP controller not found")
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

	return &forwardertypes.MsgReplaceDepositForBurnResponse{}, nil
}
