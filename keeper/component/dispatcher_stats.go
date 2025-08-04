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

package component

import (
	"context"
	"fmt"

	"cosmossdk.io/math"

	"orbiter.dev/types"
)

// UpdateStats updates all the statistics the module keep track of.
func (d *Dispatcher) UpdateStats(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	forwarding *types.Forwarding,
) error {
	if transferAttr == nil {
		return types.ErrNilPointer.Wrap("received nil transfer attributes")
	}
	if forwarding == nil {
		return types.ErrNilPointer.Wrap("received nil forwarding")
	}

	attr, err := forwarding.CachedAttributes()
	if err != nil {
		return err
	}

	var sourceOrbitID types.OrbitID
	if sourceOrbitID, err = types.NewOrbitID(transferAttr.SourceProtocolID(), transferAttr.SourceCounterpartyID()); err != nil {
		return err
	}

	var destOrbitID types.OrbitID
	if destOrbitID, err = types.NewOrbitID(forwarding.ProtocolID(), attr.CounterpartyID()); err != nil {
		return err
	}

	// Since incoming denom can be different than the outgoing one,
	// we have to check here how many amount dispatched types to store.
	// If the denom is not changed, we can set a single type with incoming
	// and outgoing amount, which could be different too, but are not part
	// of the key. If the denom changed, we have to set two values with
	// a different key.
	denomDispatchedAmounts, err := d.BuildDenomDispatchedAmounts(transferAttr)
	if err != nil {
		return err
	}

	for _, dda := range denomDispatchedAmounts {
		if err := d.updateDispatchedAmountStats(ctx, &sourceOrbitID, &destOrbitID, dda.Denom, dda.AmountDispatched); err != nil {
			return fmt.Errorf("update dispatched amount stats failure: %w", err)
		}
	}

	if err := d.updateDispatchedCountsStats(ctx, &sourceOrbitID, &destOrbitID); err != nil {
		return fmt.Errorf("update dispatch counts stats failure: %w", err)
	}

	return nil
}

// updateDispatchedAmountStats updates the amount dispatched
// values on the store. A boolean flag is used to indicate
// if the amount to be added is an incoming or outgoing amount.
// It is important to keep track of incoming and outgoing
// information because fees, swaps, or other actions can change
// the coins delivered to the destination chain.
func (d *Dispatcher) updateDispatchedAmountStats(
	ctx context.Context,
	sourceOrbitID *types.OrbitID,
	destinationOrbitID *types.OrbitID,
	denom string,
	newAmountDispatched types.AmountDispatched,
) error {
	amountDispatched := d.GetDispatchedAmount(
		ctx,
		*sourceOrbitID,
		*destinationOrbitID,
		denom,
	)

	if newAmountDispatched.Incoming.IsPositive() {
		amountDispatched.Incoming = amountDispatched.Incoming.Add(newAmountDispatched.Incoming)
	}
	if newAmountDispatched.Outgoing.IsPositive() {
		amountDispatched.Outgoing = amountDispatched.Outgoing.Add(newAmountDispatched.Outgoing)
	}

	return d.SetDispatchedAmount(
		ctx,
		*sourceOrbitID,
		*destinationOrbitID,
		denom,
		amountDispatched,
	)
}

// updateDispatchedCountsStats updates the counter of the
// number of dispatches executed.
func (d *Dispatcher) updateDispatchedCountsStats(
	ctx context.Context,
	sourceInfo *types.OrbitID,
	destinationInfo *types.OrbitID,
) error {
	countDispatches := d.GetDispatchedCounts(
		ctx,
		*sourceInfo,
		*destinationInfo,
	)
	countDispatches++

	return d.SetDispatchedCounts(
		ctx,
		*sourceInfo,
		*destinationInfo,
		countDispatches,
	)
}

// ====================================================================================================
// Helpers
// ====================================================================================================

// denomDispatchedAmount is an helper type used only
// to group types returned by the helper method
// buildDispatchedAmounts.
type denomDispatchedAmount struct {
	Denom            string
	AmountDispatched types.AmountDispatched
}

// BuildDenomDispatchedAmounts is an helper method used to
// extract the amounts dispatched that have to be dumped to state.
func (d *Dispatcher) BuildDenomDispatchedAmounts(
	transferAttributes *types.TransferAttributes,
) ([]denomDispatchedAmount, error) {
	if transferAttributes == nil {
		return nil, types.ErrNilPointer.Wrap("received nil transfer attributes")
	}
	sourceDenom := transferAttributes.SourceDenom()
	sourceAmount := transferAttributes.SourceAmount()
	destDenom := transferAttributes.DestinationDenom()
	destAmount := transferAttributes.DestinationAmount()

	ddas := make([]denomDispatchedAmount, 1, 2)

	ddas[0] = denomDispatchedAmount{
		Denom: sourceDenom,
		AmountDispatched: types.AmountDispatched{
			Incoming: sourceAmount,
			Outgoing: math.ZeroInt(),
		},
	}

	// We can have two situations here:
	// - An action changed the destination denom (e.g. swap): In this case we have to
	//   append a new entry in the slice.
	// - No action changed the destination denom: In this case we have to
	//   set the destination amount, which can be different than the source
	//   amount (e.g. fees)
	if sourceDenom == destDenom {
		ddas[0].AmountDispatched.Outgoing = destAmount
	} else {
		ddas = append(ddas, denomDispatchedAmount{
			Denom: destDenom,
			AmountDispatched: types.AmountDispatched{
				Incoming: math.ZeroInt(),
				Outgoing: destAmount,
			},
		})
	}

	return ddas, nil
}
