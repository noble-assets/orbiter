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

package types

import "cosmossdk.io/errors"

var (
	ErrUnauthorized        = errors.Register(ModuleName, 1, "signer must be the authority")
	ErrIDNotSupported      = errors.Register(ModuleName, 2, "id is not supported")
	ErrNilPointer          = errors.Register(ModuleName, 3, "invalid nil pointer")
	ErrControllerExecution = errors.Register(ModuleName, 4, "controller execution failed")
	ErrInvalidAttributes   = errors.Register(ModuleName, 5, "invalid attributes")
	ErrValidation          = errors.Register(ModuleName, 6, "validation failed")
	ErrParsingPayload      = errors.Register(ModuleName, 7, "parsing payload failed")
	ErrUnableToPause       = errors.Register(ModuleName, 8, "unable to pause")
	ErrUnableToUnpause     = errors.Register(ModuleName, 9, "unable to unpause")
)
