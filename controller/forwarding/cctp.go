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

package forwarding

import (
	"context"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"

	"cosmossdk.io/log"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"orbiter.dev/controller"
	"orbiter.dev/types"
	forwardingtypes "orbiter.dev/types/controller/forwarding"
	"orbiter.dev/types/interfaces"
)

var _ interfaces.ControllerOrbit = &CCTPController{}

// CCTPController is the forwarding controller to perform
// a CCTP transfer.
type CCTPController struct {
	*controller.BaseController[types.ProtocolID]

	logger  log.Logger
	handler *cctpHandler
}

// NewCCTPController returns a validated instance of the
// Cross-Chain Transfer Protocol controller.
func NewCCTPController(
	logger log.Logger,
	msgServer forwardingtypes.CCTPMsgServer,
) (*CCTPController, error) {
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := types.PROTOCOL_CCTP
	baseController, err := controller.NewBaseController(id)
	if err != nil {
		return nil, err
	}

	handler, err := NewCCTPHandler(msgServer)
	if err != nil {
		return nil, err
	}

	cctpController := CCTPController{
		logger:         logger.With(types.OrbitControllerName, baseController.Name()),
		BaseController: baseController,
		handler:        handler,
	}

	return &cctpController, cctpController.Validate()
}

// Validate returns an error if the instance of the controller
// is not valid.
func (c *CCTPController) Validate() error {
	if c.logger == nil {
		return types.ErrNilPointer.Wrap("logger")
	}
	if c.BaseController == nil {
		return types.ErrNilPointer.Wrap("base controller")
	}
	if c.handler == nil {
		return types.ErrNilPointer.Wrap("CCTP handler")
	}

	return nil
}

// HandlePacket validates and process a CCTP cross-chain transfer.
func (c *CCTPController) HandlePacket(ctx context.Context, packet *types.ForwardingPacket) error {
	attr, err := c.ExtractAttributes(packet.Forwarding)
	if err != nil {
		return types.ErrInvalidAttributes.Wrap(err.Error())
	}

	err = c.ValidateAttributes(attr)
	if err != nil {
		return types.ErrValidation.Wrap(err.Error())
	}

	err = c.executeOrbit(ctx, packet.TransferAttributes, attr)
	if err != nil {
		return types.ErrControllerExecution.Wrapf(
			"an error occurred executing the orbit: %s",
			err.Error(),
		)
	}

	return nil
}

func (c *CCTPController) GetHandler() *cctpHandler {
	return c.handler
}

// ExtractAttributes extract the CCTP forwarding attributes. Return an error in case
// of invalid attributes.
func (c *CCTPController) ExtractAttributes(
	forwarding *types.Forwarding,
) (*forwardingtypes.CCTPAttributes, error) {
	attr, err := forwarding.CachedAttributes()
	if err != nil {
		return nil, err
	}

	cctpAttr, ok := attr.(*forwardingtypes.CCTPAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&forwardingtypes.CCTPAttributes{},
			attr,
		)
	}

	return cctpAttr, nil
}

// ValidateAttributes returns an error if the provided CCTP attributes are
// not valid.
func (c *CCTPController) ValidateAttributes(attr *forwardingtypes.CCTPAttributes) error {
	return attr.Validate()
}

// executeOrbit is the core controller logic which performs
// the state transition calling into the CCTP server to
// initiate a cross-chain transfer.
func (c *CCTPController) executeOrbit(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	cctpAttr *forwardingtypes.CCTPAttributes,
) error {
	msg := cctptypes.MsgDepositForBurnWithCaller{
		From:              types.ModuleAddress.String(),
		Amount:            transferAttr.DestinationAmount(),
		DestinationDomain: cctpAttr.DestinationDomain,
		MintRecipient:     cctpAttr.MintRecipient,
		BurnToken:         transferAttr.DestinationDenom(),
		DestinationCaller: cctpAttr.DestinationCaller,
	}

	_, err := c.handler.DepositForBurnWithCaller(ctx, &msg)
	if err != nil {
		return err
	}

	return nil
}

// cctpHandler is the type responsible to initiate a CCTP
// transfer.
type cctpHandler struct {
	forwardingtypes.CCTPMsgServer
}

// NewCCTPHandler returns a validated instance of the CCTP handler.
func NewCCTPHandler(
	msgServer forwardingtypes.CCTPMsgServer,
) (*cctpHandler, error) {
	handler := cctpHandler{
		CCTPMsgServer: msgServer,
	}

	return &handler, handler.validate()
}

// validate returns an error if the CCTP handler instance is not valid.
func (h *cctpHandler) validate() error {
	if h.CCTPMsgServer == nil {
		return types.ErrNilPointer.Wrap("CCTP message server cannot be nil")
	}

	return nil
}
