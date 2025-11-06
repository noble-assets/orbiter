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

package dispatcher

import (
	"context"
	"errors"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	dispatchertypes "github.com/noble-assets/orbiter/types/component/dispatcher"
	"github.com/noble-assets/orbiter/types/core"
)

// UpdateStats updates all the statistics the module keep track of.
func (d *Dispatcher) UpdateStats(
	ctx context.Context,
	attr *core.TransferAttributes,
	forwarding *core.Forwarding,
) error {
	if attr == nil {
		return core.ErrNilPointer.Wrap("received nil transfer attributes")
	}
	if forwarding == nil {
		return core.ErrNilPointer.Wrap("received nil forwarding")
	}

	forwardingAttr, err := forwarding.CachedAttributes()
	if err != nil {
		return err
	}

	var sourceID core.CrossChainID
	if sourceID, err = core.NewCrossChainID(attr.SourceProtocolID(), attr.SourceCounterpartyID()); err != nil {
		return errorsmod.Wrap(err, "failed to create source cross-chain ID")
	}

	var destID core.CrossChainID
	if destID, err = core.NewCrossChainID(forwarding.ProtocolID(), forwardingAttr.CounterpartyID()); err != nil {
		return errorsmod.Wrap(err, "failed to create destination cross-chain ID")
	}

	// Since the denom is part of the stored key, if it changed during the execution
	// of some actions, we will have to store multiple dispatched amount entries.
	amounts, err := d.BuildDenomDispatchedAmounts(attr)
	if err != nil {
		return errorsmod.Wrap(err, "error building denom dispatched amounts")
	}

	for _, a := range amounts {
		if err := d.updateDispatchedAmount(ctx, &sourceID, &destID, a.Denom, a.AmountDispatched); err != nil {
			return errorsmod.Wrap(err, "update dispatched amounts stats failure")
		}
	}

	if err := d.updateDispatchedCounts(ctx, &sourceID, &destID); err != nil {
		return errorsmod.Wrap(err, "update dispatch counts stats failure")
	}

	return nil
}

// updateDispatchedAmount updates the amount dispatched
// values on the store. A boolean flag is used to indicate
// if the amount to be added is an incoming or outgoing amount.
// It is important to keep track of incoming and outgoing
// information because fees, swaps, or other actions can change
// the coins delivered to the destination chain.
func (d *Dispatcher) updateDispatchedAmount(
	ctx context.Context,
	sourceID *core.CrossChainID,
	destID *core.CrossChainID,
	denom string,
	newAmount dispatchertypes.AmountDispatched,
) error {
	da := d.GetDispatchedAmount(ctx, sourceID, destID, denom)
	amount := da.AmountDispatched

	if newAmount.Incoming.IsPositive() {
		amount.Incoming = amount.Incoming.Add(newAmount.Incoming)
	}
	if newAmount.Outgoing.IsPositive() {
		amount.Outgoing = amount.Outgoing.Add(newAmount.Outgoing)
	}

	return d.SetDispatchedAmount(ctx, sourceID, destID, denom, amount)
}

// updateDispatchedCounts updates the counter of the
// number of dispatches executed between the two cross-chain IDs.
// ID without consider.
func (d *Dispatcher) updateDispatchedCounts(
	ctx context.Context,
	sourceID *core.CrossChainID,
	destID *core.CrossChainID,
) error {
	dc := d.GetDispatchedCounts(ctx, sourceID, destID)
	if dc.Count == math.MaxUint64 {
		return errors.New("dispatch count overflow")
	}
	count := dc.Count + 1

	return d.SetDispatchedCounts(ctx, sourceID, destID, count)
}

// ====================================================================================================
// Helpers
// ====================================================================================================

// denomDispatchedAmount is an helper type used only
// to group types returned by the helper method
// buildDispatchedAmounts.
type denomDispatchedAmount struct {
	Denom            string
	AmountDispatched dispatchertypes.AmountDispatched
}

// BuildDenomDispatchedAmounts is an helper method used to
// extract the amounts dispatched that have to be stored in state.
func (d *Dispatcher) BuildDenomDispatchedAmounts(
	attr *core.TransferAttributes,
) ([]denomDispatchedAmount, error) {
	if attr == nil {
		return nil, core.ErrNilPointer.Wrap("received nil transfer attributes")
	}
	sourceDenom, sourceAmount := attr.SourceDenom(), attr.SourceAmount()
	destDenom, destAmount := attr.DestinationDenom(), attr.DestinationAmount()

	// We can have at maximum two entries, and at least one.
	ddas := make([]denomDispatchedAmount, 1, 2)
	ddas[0] = denomDispatchedAmount{
		Denom: sourceDenom,
		AmountDispatched: dispatchertypes.AmountDispatched{
			Incoming: sourceAmount,
			Outgoing: sdkmath.ZeroInt(),
		},
	}

	// We can have two situations here:
	// - No action changed the destination denom: In this case we have to
	//   set the destination amount, which can be different than the source
	//   amount (e.g. fees)
	// - An action changed the destination denom (e.g. swap): In this case we have to
	//   append a new entry in the slice.
	if sourceDenom == destDenom {
		ddas[0].AmountDispatched.Outgoing = destAmount
	} else {
		ddas = append(ddas, denomDispatchedAmount{
			Denom: destDenom,
			AmountDispatched: dispatchertypes.AmountDispatched{
				Incoming: sdkmath.ZeroInt(),
				Outgoing: destAmount,
			},
		})
	}

	return ddas, nil
}
