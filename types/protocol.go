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

// NewProtocolID returns a validated protocol ID from an int32. If
// the validation fails, the returned ID is the default ID.
func NewProtocolID(id int32) (ProtocolID, error) {
	protocolID := ProtocolID(id)
	if err := protocolID.Validate(); err != nil {
		return PROTOCOL_UNSUPPORTED, err
	}
	return protocolID, nil
}

// Validate returns an error if the ID is not valid.
func (id ProtocolID) Validate() error {
	if id == PROTOCOL_UNSUPPORTED {
		return ErrIDNotSupported.Wrapf("protocol id %s", id.String())
	}
	// Check if the protocol ID exists in the proto generated enum map
	if _, found := ProtocolID_name[int32(id)]; !found {
		return ErrIDNotSupported.Wrapf("unknown protocol id %d", int32(id))
	}
	return nil
}

func (id ProtocolID) Uint32() uint32 {
	return uint32(id) //nolint:gosec
}
