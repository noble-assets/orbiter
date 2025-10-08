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

package e2e

import (
	"errors"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	ethcommon "github.com/ethereum/go-ethereum/common"

	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/types/hyperlane"
)

type orbHypTransferInputs struct {
	amount            sdkmath.Int
	destinationDomain uint32
	nonce             uint32
	originDomain      uint32
	payloadHash       ethcommon.Hash
	recipient         sdktypes.AccAddress
}

func (o *orbHypTransferInputs) Validate() error {
	if o.recipient.String() == (ethcommon.Address{}).String() {
		return errors.New("empty recipient")
	}

	if !o.amount.IsPositive() {
		return errors.New("invalid amount")
	}

	if o.payloadHash.Hex() == (ethcommon.Hash{}).Hex() {
		return errors.New("empty payload hash")
	}

	return nil
}

// buildOrbHypConfig builds the required moving parts for an
// Orbiter transfer coming in through Hyperlane from the given inputs.
func buildHyperlaneOrbiterMessage(
	in *orbHypTransferInputs,
) (*hyperlaneutil.HyperlaneMessage, error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	warpPayload, err := warptypes.NewWarpPayload(in.recipient.Bytes(), *in.amount.BigInt())
	if err != nil {
		return nil, err
	}

	fullPayload := make([]byte, hyperlane.ORBITER_PAYLOAD_SIZE)
	copy(fullPayload[:len(warpPayload.Bytes())], warpPayload.Bytes())
	copy(fullPayload[hyperlane.ORBITER_PAYLOAD_SIZE-len(in.payloadHash):], in.payloadHash.Bytes())

	return &hyperlaneutil.HyperlaneMessage{
		Version:     3,
		Nonce:       in.nonce,
		Origin:      in.originDomain,
		Sender:      hyperlaneutil.HexAddress{},
		Destination: in.destinationDomain,
		Recipient:   hyperlaneutil.HexAddress(in.recipient.Bytes()),
		Body:        fullPayload,
	}, nil
}
