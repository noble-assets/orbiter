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

package actions

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"orbiter.dev/controllers"
	"orbiter.dev/types"
	"orbiter.dev/types/controllers/actions"
	"orbiter.dev/types/interfaces"
)

var _ interfaces.ControllerAction = &FeeController{}

// NewFeeController returns a new validated instance of
// the fee controller.
func NewFeeController(
	logger log.Logger,
	bankKeeper actions.BankKeeperFee,
) (*FeeController, error) {
	if logger == nil {
		return nil, types.ErrNilPointer.Wrap("logger cannot be nil")
	}

	id := types.ACTION_FEE
	baseController, err := controllers.NewBaseController(id)
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

// FeeController is the controller to execute
// fee payment actions.
type FeeController struct {
	logger log.Logger
	*controllers.BaseController[types.ActionID]
	BankKeeper actions.BankKeeperFee
}

// Validate performs basic validation for the fee controller.
func (c *FeeController) Validate() error {
	if c.BaseController == nil {
		return types.ErrNilPointer.Wrap("base controller cannot be nil")
	}
	if c.BankKeeper == nil {
		return types.ErrNilPointer.Wrap("bank keeper cannot be nil")
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

	feesToDistribute := c.ComputeFeesToDistribute(
		transferAttr.DestinationAmount(),
		transferAttr.DestinationDenom(),
		attr.FeesInfo,
	)
	if feesToDistribute.Total.GTE(transferAttr.DestinationAmount()) {
		return types.ErrInvalidAttributes.Wrap("total fees equal or exceed transfer amount")
	}

	err = c.executeAction(ctx, feesToDistribute.Values)
	if err != nil {
		return types.ErrControllerExecution.Wrapf(
			"an error occurred executing the action %s",
			err.Error(),
		)
	}

	transferAttr.SetDestinationAmount(
		transferAttr.DestinationAmount().Sub(feesToDistribute.Total),
	)

	return nil
}

// GetAttributes returns the fee attributes concrete type from
// a fee action.
func (c *FeeController) GetAttributes(action *types.Action) (*actions.FeeAttributes, error) {
	attr, err := c.extractAttributes(action)
	if err != nil {
		return nil, types.ErrInvalidAttributes.Wrap(err.Error())
	}
	err = c.ValidateAttributes(attr)
	if err != nil {
		return nil, types.ErrValidation.Wrap(err.Error())
	}

	return attr, nil
}

// extractAttributes extract the fee attributes. Return an error in case
// of invalid attributes.
func (c *FeeController) extractAttributes(
	action *types.Action,
) (*actions.FeeAttributes, error) {
	if action == nil {
		return nil, types.ErrNilPointer.Wrap("received nil fee attributes")
	}
	attr, err := action.CachedAttributes()
	if err != nil {
		return nil, err
	}

	feeAttr, ok := attr.(*actions.FeeAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&actions.FeeAttributes{},
			attr,
		)
	}
	return feeAttr, nil
}

// ValidateAttributes returns an error if the provided fee attributes are
// not valid.
func (c *FeeController) ValidateAttributes(attr *actions.FeeAttributes) error {
	if attr == nil {
		return types.ErrNilPointer.Wrap("fee attributes")
	}
	for _, feeInfo := range attr.FeesInfo {
		if err := c.ValidateFee(feeInfo); err != nil {
			return err
		}
	}
	return nil
}

// ValidateFee returns an error if the provided fee information are
// not valid.
func (c *FeeController) ValidateFee(feeInfo *actions.FeeInfo) error {
	if feeInfo == nil {
		return types.ErrNilPointer.Wrap("fee info")
	}
	if feeInfo.BasisPoints == 0 {
		return errors.New("fee basis point must be greater than zero")
	}
	if feeInfo.BasisPoints > types.BPSNormalizer {
		return fmt.Errorf("fee basis point cannot be higher than %d", types.BPSNormalizer)
	}

	_, err := sdk.AccAddressFromBech32(feeInfo.Recipient)
	return err
}

// ComputeFeesToDistribute computes the fee to distribute based on the
// action fees information and the amount and denom to transfer.
//
// CONTRACT: the input has already been validated.
func (c *FeeController) ComputeFeesToDistribute(
	transferAmount math.Int,
	transferDenom string,
	feesInfo []*actions.FeeInfo,
) *actions.FeesToDistribute {
	fees := actions.FeesToDistribute{
		Total:  math.ZeroInt(),
		Values: make([]actions.RecipientAmount, 0),
	}

	for _, feeInfo := range feesInfo {
		addr, _ := sdk.AccAddressFromBech32(feeInfo.Recipient)

		feeAmount := ComputeFeeAmount(transferAmount, uint64(feeInfo.BasisPoints))
		if feeAmount.IsPositive() {
			fee := sdk.NewCoin(transferDenom, feeAmount)
			fees.Values = append(
				fees.Values,
				actions.RecipientAmount{Recipient: addr, Amount: sdk.NewCoins(fee)},
			)
			fees.Total = fees.Total.Add(feeAmount)
		}
	}

	return &fees
}

// executeAction is the core controller logic which perform
// the state transition transferring fees collected to fee
// recipients.
func (c *FeeController) executeAction(
	ctx context.Context,
	fees []actions.RecipientAmount,
) error {
	for _, fee := range fees {
		err := c.BankKeeper.SendCoins(ctx, types.ModuleAddress, fee.Recipient, fee.Amount)
		if err != nil {
			return err
		}
	}

	return nil
}

// ComputeFeeAmount returns the fee associated with the provided basis
// points and amount. The function returns zero if overflow.
func ComputeFeeAmount(amount math.Int, basisPoints uint64) math.Int {
	basisPointsInt := math.NewIntFromUint64(basisPoints)
	fee, err := amount.SafeMul(basisPointsInt)
	if err != nil || fee.IsZero() {
		return math.ZeroInt()
	}
	return fee.QuoRaw(types.BPSNormalizer)
}
