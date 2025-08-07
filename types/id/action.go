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

package id

import "fmt"

// NewActionID returns a validated action ID from an int32. If
// the validation fails, the returned value signals an unsupported
// action and an error is returned along with it.
func NewActionID(id int32) (ActionID, error) {
	actionID := ActionID(id)
	if err := actionID.Validate(); err != nil {
		return ACTION_UNSUPPORTED, err
	}

	return actionID, nil
}

// Validate returns an error if the ID is not valid.
func (id ActionID) Validate() error {
	if id == ACTION_UNSUPPORTED {
		return fmt.Errorf("action id is not supported: %s", id.String())
	}
	if _, found := ActionID_name[int32(id)]; !found {
		return fmt.Errorf("action id is unknown: %d", int32(id))
	}

	return nil
}
