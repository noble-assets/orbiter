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

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/noble-assets/orbiter/types/core"

	"cosmossdk.io/collections/codec"
)

// Keccak256Hash returns the keccak 256 hash of the payload contents.
// To guarantee uniqueness the sequence number is included.
func (p *PendingPayload) Keccak256Hash() (common.Hash, error) {
	bz, err := p.Marshal()
	if err != nil {
		return common.Hash{}, err
	}

	return crypto.Keccak256Hash(bz), nil
}

// Validate checks that the pending payload contents are valid.
func (p *PendingPayload) Validate() error {
	if p == nil {
		return core.ErrNilPointer.Wrap("pending payload")
	}

	return p.Payload.Validate()
}

// TODO: check if this is required?
var _ codec.ValueCodec[PendingPayload] = &PendingPayloadCollValue{}

// PendingPayloadCollValue implements the ValueCodec interface so that PendingPayload
// can be used with collections maps.
type PendingPayloadCollValue struct{}

// TODO: this could use e.g. `abi` encoding to be aligned with Ethereum?
func (v *PendingPayloadCollValue) Encode(p PendingPayload) ([]byte, error) {
	return p.Marshal()
}

func (v *PendingPayloadCollValue) Decode(data []byte) (PendingPayload, error) {
	var p PendingPayload
	if err := (&p).Unmarshal(data); err != nil {
		return PendingPayload{}, err
	}

	return p, nil
}

func (v *PendingPayloadCollValue) EncodeJSON(payload PendingPayload) ([]byte, error) {
	panic("implement me (EncodeJSON)")
}

func (v *PendingPayloadCollValue) DecodeJSON(data []byte) (PendingPayload, error) {
	panic("implement me (DecodeJSON)")
}

func (v *PendingPayloadCollValue) Stringify(_ PendingPayload) string {
	panic("implement me (Stringify)")
}

func (v *PendingPayloadCollValue) ValueType() string {
	return "noble/orbiter/v1/PendingPayload"
}
