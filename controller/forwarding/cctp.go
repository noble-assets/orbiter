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

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/noble-assets/orbiter/v2/controller"
	"github.com/noble-assets/orbiter/v2/types"
	forwardingtypes "github.com/noble-assets/orbiter/v2/types/controller/forwarding"
	"github.com/noble-assets/orbiter/v2/types/core"
)

var _ types.ForwardingController = &CCTPController{}

// CCTPController is the forwarding controller to perform
// a CCTP transfer.
type CCTPController struct {
	*controller.BaseController[core.ProtocolID]

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
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := core.PROTOCOL_CCTP
	baseController, err := controller.NewBase(id)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error creating base controller for CCTP controller")
	}

	handler, err := NewCCTPHandler(msgServer)
	if err != nil {
		return nil, err
	}

	c := &CCTPController{
		logger:         logger.With(core.ForwardingControllerName, baseController.Name()),
		BaseController: baseController,
		handler:        handler,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// Validate returns an error if the instance of the controller
// is not valid.
func (c *CCTPController) Validate() error {
	if c.logger == nil {
		return core.ErrNilPointer.Wrap("logger")
	}
	if c.BaseController == nil {
		return core.ErrNilPointer.Wrap("base controller")
	}
	if c.handler == nil {
		return core.ErrNilPointer.Wrap("CCTP handler")
	}

	return nil
}

// HandlePacket validates and process a CCTP cross-chain transfer.
func (c *CCTPController) HandlePacket(ctx context.Context, packet *types.ForwardingPacket) error {
	c.logger.Debug("handling CCTP packet")
	if packet == nil {
		return core.ErrNilPointer.Wrap("CCTP controller received nil packet")
	}

	attr, err := c.ExtractAttributes(packet.Forwarding)
	if err != nil {
		return core.ErrValidation.Wrapf("invalid CCTP forwarding: %s", err.Error())
	}

	err = c.ValidateAttributes(attr)
	if err != nil {
		return core.ErrValidation.Wrap(err.Error())
	}

	err = c.executeForwarding(ctx, packet.TransferAttributes, attr)
	if err != nil {
		return errorsmod.Wrap(err, "CCTP controller execution error")
	}

	return nil
}

func (c *CCTPController) GetHandler() *cctpHandler {
	return c.handler
}

// ExtractAttributes extract the CCTP forwarding attributes. Return an error in case
// of invalid attributes.
func (c *CCTPController) ExtractAttributes(
	forwarding *core.Forwarding,
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

// executeForwarding is the core controller logic which performs
// the state transition calling into the CCTP server to
// initiate a cross-chain transfer.
func (c *CCTPController) executeForwarding(
	ctx context.Context,
	transferAttr *core.TransferAttributes,
	cctpAttr *forwardingtypes.CCTPAttributes,
) error {
	if len(cctpAttr.DestinationCaller) == 0 {
		msg := cctptypes.MsgDepositForBurn{
			From:              core.ModuleAddress.String(),
			Amount:            transferAttr.DestinationAmount(),
			DestinationDomain: cctpAttr.DestinationDomain,
			MintRecipient:     cctpAttr.MintRecipient,
			BurnToken:         transferAttr.DestinationDenom(),
		}

		_, err := c.handler.DepositForBurn(ctx, &msg)
		if err != nil {
			return err
		}
	} else {
		msg := cctptypes.MsgDepositForBurnWithCaller{
			From:              core.ModuleAddress.String(),
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
		return core.ErrNilPointer.Wrap("CCTP message server cannot be nil")
	}

	return nil
}
