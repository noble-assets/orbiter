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

package orbits

import (
	"context"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"

	"cosmossdk.io/log"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"orbiter.dev/controllers"
	"orbiter.dev/types"
	"orbiter.dev/types/controllers/orbits"
	"orbiter.dev/types/interfaces"
)

var _ interfaces.ControllerOrbit = &CCTPController{}

// NewCCTPController returns a validated instance of the
// Cross-Chain Transfer Protocol controller.
func NewCCTPController(
	logger log.Logger,
	msgServer orbits.CCTPMsgServer,
) (interfaces.ControllerOrbit, error) {
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := types.PROTOCOL_CCTP
	baseController, err := controllers.NewBaseController(id)
	if err != nil {
		return nil, err
	}

	handler, err := newCCTPHandler(msgServer)
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

// CCTPController is the orbit controller to perform
// a CCTP transfer.
type CCTPController struct {
	logger log.Logger
	*controllers.BaseController[types.ProtocolID]
	handler *cctpHandler
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
func (c *CCTPController) HandlePacket(ctx context.Context, packet *types.OrbitPacket) error {
	attr, err := c.extractAttributes(packet.Orbit)
	if err != nil {
		return types.ErrInvalidAttributes.Wrap(err.Error())
	}

	err = c.validateAttributes(attr)
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

// extractAttributes extract the CCTP orbit attributes. Return an error in case
// of invalid attributes.
func (c *CCTPController) extractAttributes(
	orbit *types.Orbit,
) (*orbits.CCTPAttributes, error) {
	attr, err := orbit.CachedAttributes()
	if err != nil {
		return nil, err
	}

	cctpAttr, ok := attr.(*orbits.CCTPAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&orbits.CCTPAttributes{},
			attr,
		)
	}

	return cctpAttr, nil
}

// validateAttributes returns an error if the provided CCTP attributes are
// not valid.
func (c *CCTPController) validateAttributes(attr *orbits.CCTPAttributes) error {
	if attr == nil {
		return types.ErrNilPointer.Wrap("CCTP attributes")
	}
	return attr.Validate()
}

// executeOrbit is the core controller logic which performs
// the state transition calling into the CCTP server to
// initiate a cross-chain transfer.
func (c *CCTPController) executeOrbit(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	cctpAttr *orbits.CCTPAttributes,
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

// newCCTPHandler returns a validated instance of the CCTP handler.
func newCCTPHandler(
	msgServer orbits.CCTPMsgServer,
) (*cctpHandler, error) {
	handler := cctpHandler{
		CCTPMsgServer: msgServer,
	}
	return &handler, handler.validate()
}

// cctpHandler is the type responsible to initiate a CCTP
// transfer.
type cctpHandler struct {
	orbits.CCTPMsgServer
}

// validate returns an error if the CCTP handler instance is not valid.
func (h *cctpHandler) validate() error {
	if h.CCTPMsgServer == nil {
		return types.ErrNilPointer.Wrap("CCTP message server cannot be nil")
	}
	return nil
}
