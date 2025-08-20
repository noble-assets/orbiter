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

package action

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/noble-assets/orbiter/controller"
	"github.com/noble-assets/orbiter/types"
	actiontypes "github.com/noble-assets/orbiter/types/controller/action"
	"github.com/noble-assets/orbiter/types/core"
)

var _ types.ControllerAction = &FeeController{}

// FeeController is the controller to execute
// fee payment action.
type FeeController struct {
	*controller.BaseController[core.ActionID]

	logger     log.Logger
	BankKeeper actiontypes.BankKeeperFee
}

// NewFeeController returns a new validated instance of
// the fee controller.
func NewFeeController(
	logger log.Logger,
	bankKeeper actiontypes.BankKeeperFee,
) (*FeeController, error) {
	if logger == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := core.ACTION_FEE
	baseController, err := controller.NewBase(id)
	if err != nil {
		return nil, err
	}

	feeController := FeeController{
		logger:         logger.With("action", baseController.Name()),
		BaseController: baseController,
		BankKeeper:     bankKeeper,
	}

	return &feeController, feeController.Validate()
}

// Validate performs basic validation for the fee controller.
func (c *FeeController) Validate() error {
	if c.BaseController == nil {
		return core.ErrNilPointer.Wrap("base controller cannot be nil")
	}
	if c.BankKeeper == nil {
		return core.ErrNilPointer.Wrap("bank keeper cannot be nil")
	}

	return nil
}

// HandlePacket process a fee collection action packet.
func (c *FeeController) HandlePacket(
	ctx context.Context,
	packet *types.ActionPacket,
) error {
	attr, err := c.GetAttributes(packet.Action)
	if err != nil {
		return err
	}

	transferAttr := packet.TransferAttributes

	feesToDistribute, err := c.ComputeFeesToDistribute(
		transferAttr.DestinationAmount(),
		transferAttr.DestinationDenom(),
		attr.FeesInfo,
	)
	if err != nil {
		return err
	}
	if feesToDistribute.Total.GTE(transferAttr.DestinationAmount()) {
		return core.ErrInvalidAttributes.Wrap("total fees equal or exceed transfer amount")
	}

	err = c.executeAction(ctx, feesToDistribute.Values)
	if err != nil {
		return errorsmod.Wrap(err, "fee controller execution error")
	}

	transferAttr.SetDestinationAmount(
		transferAttr.DestinationAmount().Sub(feesToDistribute.Total),
	)

	return nil
}

// GetAttributes returns the fee attributes concrete type from
// a fee action.
func (c *FeeController) GetAttributes(action *core.Action) (*actiontypes.FeeAttributes, error) {
	attr, err := c.extractAttributes(action)
	if err != nil {
		return nil, core.ErrInvalidAttributes.Wrap(err.Error())
	}
	err = c.ValidateAttributes(attr)
	if err != nil {
		return nil, core.ErrValidation.Wrap(err.Error())
	}

	return attr, nil
}

// ValidateAttributes returns an error if the provided fee attributes are not valid.
func (c *FeeController) ValidateAttributes(attr *actiontypes.FeeAttributes) error {
	return attr.Validate()
}

// ComputeFeesToDistribute computes the fee to distribute based on the
// action fees information and the amount and denom to transfer.
//
// CONTRACT: the inputs have already been validated.
func (c *FeeController) ComputeFeesToDistribute(
	transferAmount math.Int,
	transferDenom string,
	feesInfo []*actiontypes.FeeInfo,
) (*actiontypes.FeesToDistribute, error) {
	fees := actiontypes.NewFeesToDistribute()

	for _, feeInfo := range feesInfo {
		addr, _ := sdk.AccAddressFromBech32(feeInfo.Recipient)

		feeAmount, err := ComputeFeeAmount(transferAmount, uint64(feeInfo.BasisPoints))
		if err != nil {
			return nil, err
		}
		if feeAmount.IsPositive() {
			fee := sdk.NewCoin(transferDenom, feeAmount)
			fees.Values = append(
				fees.Values,
				actiontypes.RecipientAmount{Recipient: addr, Amount: sdk.NewCoins(fee)},
			)
			fees.Total = fees.Total.Add(feeAmount)
		}
	}

	return &fees, nil
}

// executeAction is the core controller logic which perform
// the state transition transferring fees collected to fee
// recipients.
func (c *FeeController) executeAction(
	ctx context.Context,
	fees []actiontypes.RecipientAmount,
) error {
	for _, fee := range fees {
		err := c.BankKeeper.SendCoins(ctx, core.ModuleAddress, fee.Recipient, fee.Amount)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractAttributes extract the fee attributes. Return an error in case
// of invalid attributes.
func (c *FeeController) extractAttributes(
	action *core.Action,
) (*actiontypes.FeeAttributes, error) {
	if action == nil {
		return nil, core.ErrNilPointer.Wrap("received nil fee attributes")
	}
	attr, err := action.CachedAttributes()
	if err != nil {
		return nil, err
	}

	feeAttr, ok := attr.(*actiontypes.FeeAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&actiontypes.FeeAttributes{},
			attr,
		)
	}

	return feeAttr, nil
}

// ComputeFeeAmount returns the fee associated with the provided basis
// points and amount. The function returns an error in case of overflow.
func ComputeFeeAmount(amount math.Int, basisPoints uint64) (math.Int, error) {
	basisPointsInt := math.NewIntFromUint64(basisPoints)
	fee, err := amount.SafeMul(basisPointsInt)
	if err != nil || fee.IsZero() {
		return math.ZeroInt(), err
	}

	return fee.QuoRaw(core.BPSNormalizer), nil
}
