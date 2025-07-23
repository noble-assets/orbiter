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

package components

import (
	"context"
	"fmt"

	"cosmossdk.io/math"

	"orbiter.dev/types"
)

// updateStats updates all the statistics the module keep track of.
// CONTRACT: transferAttr and orbit are not nil pointers.
func (d *DispatcherComponent) updateStats(
	ctx context.Context,
	transferAttr *types.TransferAttributes,
	orbit *types.Orbit,
) error {
	attr, err := orbit.CachedAttributes()
	if err != nil {
		return err
	}

	var sourceAttr types.OrbitID
	if sourceAttr, err = types.NewOrbitID(
		transferAttr.SourceProtocolID(),
		transferAttr.SourceCounterpartyID(),
	); err != nil {
		return err
	}

	var destinationAttr types.OrbitID
	if destinationAttr, err = types.NewOrbitID(
		orbit.ProtocolID(),
		attr.CounterpartyID(),
	); err != nil {
		return err
	}

	// Since incoming denom can be different than the outgoing one,
	// we have to check here how many amount dispatched types to store.
	// If the denom is not changed, we can set a single type with incoming
	// and outgoing amount, which could be different too, but are not part
	// of the key. If the denom changed, we have to set two values with
	// a different denom.
	denomDispatchedAmounts := d.buildDenomDispatchedAmounts(transferAttr)

	for _, dda := range denomDispatchedAmounts {
		if err := d.updateDispatchedAmountStats(ctx, &sourceAttr, &destinationAttr, dda.Denom, dda.AmountDispatched); err != nil {
			return fmt.Errorf("update incoming stats failure: %w", err)
		}
	}

	if err := d.updateDispatchedCountsStats(ctx, &sourceAttr, &destinationAttr); err != nil {
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
func (d *DispatcherComponent) updateDispatchedAmountStats(
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
func (d *DispatcherComponent) updateDispatchedCountsStats(
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

// buildDenomDispatchedAmounts is an helper method used to
// extract the amounts dispatched that have to be dumped to state.
func (d *DispatcherComponent) buildDenomDispatchedAmounts(
	transferAttributes *types.TransferAttributes,
) []denomDispatchedAmount {
	sourceDenom := transferAttributes.SourceDenom()
	sourceAmount := transferAttributes.SourceAmount()
	destDenom := transferAttributes.DestinationDenom()
	destAmount := transferAttributes.DestinationAmount()

	dda := make([]denomDispatchedAmount, 1, 2)
	dda[0] = denomDispatchedAmount{
		Denom: sourceDenom,
		AmountDispatched: types.AmountDispatched{
			Incoming: sourceAmount,
			Outgoing: math.ZeroInt(),
		},
	}

	if sourceDenom == destDenom {
		dda[0].AmountDispatched.Outgoing = destAmount
	} else {
		dda = append(dda, denomDispatchedAmount{
			Denom: destDenom,
			AmountDispatched: types.AmountDispatched{
				Incoming: math.ZeroInt(),
				Outgoing: destAmount,
			},
		})
	}

	return dda
}
