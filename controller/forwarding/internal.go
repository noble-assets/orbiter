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

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/noble-assets/orbiter/controller"
	"github.com/noble-assets/orbiter/types"
	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

var _ types.ForwardingController = &InternalController{}

// InternalController is the forwarding controller to perform
// an internal transfer.
type InternalController struct {
	*controller.BaseController[core.ProtocolID]

	logger  log.Logger
	handler forwardingtypes.InternalHandler
}

func NewInternalController(
	logger log.Logger,
	handler forwardingtypes.InternalHandler,
) (*InternalController, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := core.PROTOCOL_INTERNAL
	b, err := controller.NewBase(id)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error creating base controller for internal controller")
	}

	c := &InternalController{
		BaseController: b,
		logger:         logger,
		handler:        handler,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// Validate returns an error if any of the internal controller's field is not valid.
func (c *InternalController) Validate() error {
	if c.logger == nil {
		return core.ErrNilPointer.Wrap("logger is required for the internal controller")
	}
	if c.BaseController == nil {
		return core.ErrNilPointer.Wrap("base controller is required for the internal controller")
	}
	if c.handler == nil {
		return core.ErrNilPointer.Wrap("handler is required for the internal controller")
	}

	return nil
}

// HandlePacket implements types.ForwardingController.
func (c *InternalController) HandlePacket(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	c.logger.Debug("Handling internal packet")

	if packet == nil {
		return core.ErrNilPointer.Wrap("internal controller received nil packet")
	}

	attr, err := c.ExtractAttributes(packet.Forwarding)
	if err != nil {
		return errorsmod.Wrap(err, "error extracting internal forwarding attributes")
	}

	err = c.ValidateForwarding(ctx, packet.TransferAttributes, attr)
	if err != nil {
		return errorsmod.Wrap(err, "error validating internal forwarding")
	}

	err = c.executeForwarding(
		ctx,
		packet.TransferAttributes,
		attr,
		packet.Forwarding.PassthroughPayload,
	)
	if err != nil {
		return errorsmod.Wrap(err, "Hyperlane controller execution error")
	}

	return nil
}

// ExtractAttributes returns the internal attributes from the forwarding or an error.
func (c *InternalController) ExtractAttributes(
	forwarding *core.Forwarding,
) (*forwardingtypes.InternalAttributes, error) {
	attr, err := forwarding.CachedAttributes()
	if err != nil {
		return nil, errorsmod.Wrap(err, "error extracting cached attributes")
	}

	intAttr, ok := attr.(*forwardingtypes.InternalAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&forwardingtypes.InternalAttributes{},
			attr,
		)
	}

	return intAttr, nil
}

// ValidateForwarding checks whether the forwarding attributes are valid or not.
func (c *InternalController) ValidateForwarding(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	intAttr *forwardingtypes.InternalAttributes,
) error {
	if err := transferAttr.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid transfer attributes")
	}

	if err := intAttr.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid internal forwarding attributes")
	}

	return nil
}

// executeForwarding initiates an internal transfer.
func (c *InternalController) executeForwarding(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	intAttr *forwardingtypes.InternalAttributes,
	_ []byte,
) error {
	_, err := c.handler.Send(ctx, &banktypes.MsgSend{
		FromAddress: core.ModuleAddress.String(),
		ToAddress:   intAttr.Recipient,
		Amount: sdk.Coins{
			sdk.NewCoin(transferAttr.DestinationDenom(), transferAttr.DestinationAmount()),
		},
	})
	if err != nil {
		return errorsmod.Wrap(err, "error executing internal forwarding")
	}

	return nil
}
