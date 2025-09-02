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
	"errors"

	errorsmod "cosmossdk.io/errors"

	"github.com/noble-assets/orbiter/types/core"
)

func (a *AmountDispatched) IsPositive() bool {
	return a.Incoming.IsPositive() || a.Outgoing.IsPositive()
}

func (a DispatchedAmountEntry) Validate() error {
	if a.Denom == "" {
		return errors.New("cannot set empty denom")
	}

	if a.SourceId == nil {
		return errorsmod.Wrap(core.ErrNilPointer, "missing source cross-chain ID")
	}

	if err := a.SourceId.Validate(); err != nil {
		return errorsmod.Wrap(err, "failed to create source cross-chain ID")
	}

	if a.DestinationId == nil {
		return errorsmod.Wrap(core.ErrNilPointer, "missing destination cross-chain ID")
	}

	if err := a.DestinationId.Validate(); err != nil {
		return errorsmod.Wrap(err, "failed to create destination cross-chain ID")
	}

	if a.AmountDispatched.Incoming.IsNegative() || a.AmountDispatched.Outgoing.IsNegative() {
		return errors.New("cannot set negative amounts")
	}

	if !a.AmountDispatched.Incoming.IsPositive() && !a.AmountDispatched.Outgoing.IsPositive() {
		return errors.New(
			"cannot set incoming and outgoing amounts equal to zero",
		)
	}

	return nil
}

func (c DispatchCountEntry) Validate() error {
	if c.Count == 0 {
		return errors.New("cannot set zero count")
	}

	if c.SourceId == nil {
		return errorsmod.Wrap(core.ErrNilPointer, "missing source cross-chain ID")
	}

	if err := c.SourceId.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid source cross-chain ID")
	}

	if c.DestinationId == nil {
		return errorsmod.Wrap(core.ErrNilPointer, "missing destination cross-chain ID")
	}

	if err := c.DestinationId.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid destination cross-chain ID")
	}

	return nil
}
