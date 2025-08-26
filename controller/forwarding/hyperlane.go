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
	"fmt"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/noble-assets/orbiter/controller"
	"github.com/noble-assets/orbiter/types"
	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

var _ types.ControllerForwarding = &HyperlaneController{}

// HyperlaneController is the forwarding controller for the Hyperlane protocol.
type HyperlaneController struct {
	*controller.BaseController[core.ProtocolID]

	logger  log.Logger
	handler forwardingtypes.HyperlaneHandler
}

// NewHyperlaneController returns a validated instance of the Hyperlane
// controller.
func NewHyperlaneController(
	logger log.Logger,
	handler forwardingtypes.HyperlaneHandler,
) (*HyperlaneController, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	b, err := controller.NewBase(core.PROTOCOL_HYPERLANE)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error creating base controller for hyperlane controller")
	}
	c := &HyperlaneController{
		BaseController: b,
		logger:         logger.With(core.ForwardingControllerName, b.Name()),
		handler:        handler,
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// Validate returns an error if any of the Hyperlane controller's field is not valid.
func (c *HyperlaneController) Validate() error {
	if c.logger == nil {
		return core.ErrNilPointer.Wrap("logger is required for the Hyperlane controller")
	}
	if c.BaseController == nil {
		return core.ErrNilPointer.Wrap("base controller is required for the Hyperlane controller")
	}
	if c.handler == nil {
		return core.ErrNilPointer.Wrap("handler is required for the Hyperlance controller")
	}

	return nil
}

// HandlePacket implements types.ControllerForwarding.
func (c *HyperlaneController) HandlePacket(
	ctx context.Context,
	packet *types.ForwardingPacket,
) error {
	if packet == nil {
		return errorsmod.Wrap(core.ErrNilPointer, "Hyperlane controller received nil packet")
	}

	attr, err := c.ExtractAttributes(packet.Forwarding)
	if err != nil {
		return errorsmod.Wrap(err, "error extracting Hyperlane forwarding attributes")
	}

	err = c.ValidateForwarding(ctx, packet.TransferAttributes, attr)
	if err != nil {
		return errorsmod.Wrap(err, "error validating Hyperlane forwarding")
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

// ExtractAttributes returns the hyperlane attributes from the forwarding
// or an error.
func (c *HyperlaneController) ExtractAttributes(
	forwarding *core.Forwarding,
) (*forwardingtypes.HypAttributes, error) {
	attr, err := forwarding.CachedAttributes()
	if err != nil {
		return nil, errorsmod.Wrap(err, "error extracting cached attributes")
	}

	hypAttr, ok := attr.(*forwardingtypes.HypAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&forwardingtypes.HypAttributes{},
			attr,
		)
	}

	return hypAttr, nil
}

// ValidateForwarding checks whether the attributes received for execute a forwarding are
// valid or not.
func (c *HyperlaneController) ValidateForwarding(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	hypAttr *forwardingtypes.HypAttributes,
) error {
	tokenID := hyperlaneutil.HexAddress(hypAttr.GetTokenId())
	req := warptypes.QueryTokenRequest{
		Id: tokenID.String(),
	}
	resp, err := c.handler.Token(ctx, &req)
	if err != nil {
		return errorsmod.Wrap(err, "invalid Hyperlane forwarding")
	}

	if resp.Token.OriginDenom != transferAttr.DestinationDenom() {
		return fmt.Errorf(
			"invalid forwarding token, wanted %s, got %s",
			resp.Token.OriginDenom, transferAttr.DestinationDenom(),
		)
	}

	return nil
}

// executeForwarding initiates an Hyperlane cross-chain transfer.
func (c *HyperlaneController) executeForwarding(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	hypAttr *forwardingtypes.HypAttributes,
	passthroughPayload []byte,
) error {
	hookID := hyperlaneutil.HexAddress(hypAttr.CustomHookId)
	_, err := c.handler.RemoteTransfer(ctx, &warptypes.MsgRemoteTransfer{
		Sender:             core.ModuleAddress.String(),
		TokenId:            hyperlaneutil.HexAddress(hypAttr.GetTokenId()),
		DestinationDomain:  hypAttr.DestinationDomain,
		Recipient:          hyperlaneutil.HexAddress(hypAttr.GetRecipient()),
		Amount:             transferAttr.DestinationAmount(),
		CustomHookId:       &hookID,
		GasLimit:           hypAttr.GasLimit,
		MaxFee:             hypAttr.GetMaxFee(),
		CustomHookMetadata: string(passthroughPayload),
	})
	if err != nil {
		return errorsmod.Wrap(err, "error executing Hyperlane forwarding")
	}

	return nil
}
