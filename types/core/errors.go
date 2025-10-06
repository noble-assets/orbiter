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

package core

import errorsmod "cosmossdk.io/errors"

var (
	ErrUnauthorized      = errorsmod.Register(ModuleName, 1, "signer must be the authority")
	ErrIDNotSupported    = errorsmod.Register(ModuleName, 2, "ID is not supported")
	ErrNilPointer        = errorsmod.Register(ModuleName, 3, "invalid nil pointer")
	ErrEmptyString       = errorsmod.Register(ModuleName, 4, "string cannot be empty")
	ErrInvalidAttributes = errorsmod.Register(ModuleName, 5, "invalid attributes")
	ErrValidation        = errorsmod.Register(ModuleName, 6, "validation failed")
	ErrParsingPayload    = errorsmod.Register(ModuleName, 7, "parsing payload failed")
	ErrUnableToPause     = errorsmod.Register(ModuleName, 8, "unable to pause")
	ErrUnableToUnpause   = errorsmod.Register(ModuleName, 9, "unable to unpause")
	ErrAlreadySet        = errorsmod.Register(ModuleName, 10, "value already set")
	ErrSubmitPayload     = errorsmod.Register(ModuleName, 11, "payload submission failed")
	ErrRemovePayload     = errorsmod.Register(ModuleName, 12, "payload removal failed")
)
