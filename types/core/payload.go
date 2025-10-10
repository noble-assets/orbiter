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

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	errorsmod "cosmossdk.io/errors"
)

// SHA256Hash returns the SHA-256 hash of the payload contents.
// To guarantee uniqueness the sequence number is included.
//
// CONTRACT: The pending payload should be validated before calling this function.
func (p *PendingPayload) SHA256Hash() (*PayloadHash, error) {
	bz, err := p.Marshal()
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(bz)
	pHash := PayloadHash(hash)

	return &pHash, nil
}

// PayloadHashLength specifies the expected length of Orbiter payload hashes.
const PayloadHashLength = 32

// PayloadHash is a helper type to define a unified interface for interacting with
// the generated SHA256 hashes for the PendingPayload type.
type PayloadHash [PayloadHashLength]byte

// ParsePayloadHash takes a string and tries to parse a PayloadHash from it.
func ParsePayloadHash(s string) (*PayloadHash, error) {
	bz, err := hex.DecodeString(s)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid payload hash")
	}

	if len(bz) != PayloadHashLength {
		return nil, errors.New("malformed payload hash")
	}

	pHash := PayloadHash(bz)

	return &pHash, nil
}

func (p PayloadHash) Bytes() []byte {
	return p[:]
}

func (p PayloadHash) String() string {
	return hex.EncodeToString(p[:])
}
