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

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdktypes "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/testutil"
	"github.com/noble-assets/orbiter/types/controller/action"
	"github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

type orbHypTransferInputs struct {
	amount            sdkmath.Int
	destinationDomain uint32
	nonce             uint32
	originDomain      uint32
	payload           *core.Payload
	recipient         sdktypes.AccAddress
}

func (o *orbHypTransferInputs) Validate() error {
	if o.recipient.String() == (ethcommon.Address{}).String() {
		return errors.New("empty recipient")
	}

	if !o.amount.IsPositive() {
		return errors.New("invalid amount")
	}

	if err := o.payload.Validate(); err != nil {
		return errorsmod.Wrap(err, "invalid payload")
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

	warpPayloadBz := warpPayload.Bytes()

	testPayload, err := createTestPayload()
	if err != nil {
		return nil, err
	}

	testPayloadBz, err := testPayload.Marshal()
	if err != nil {
		return nil, err
	}

	fullPayloadBz := make([]byte, len(warpPayloadBz)+len(testPayloadBz))
	copy(fullPayloadBz[:len(warpPayloadBz)], warpPayloadBz)
	copy(fullPayloadBz[len(warpPayloadBz):], testPayloadBz)

	orbModAddr32Bz, err := LeftPadBytes(core.ModuleAddress.Bytes())
	if err != nil {
		return nil, err
	}

	return &hyperlaneutil.HyperlaneMessage{
		Version:     3,
		Nonce:       in.nonce,
		Origin:      in.originDomain,
		Sender:      hyperlaneutil.HexAddress{},
		Destination: in.destinationDomain,
		// TODO: the recipient here has to be the handler for the custom Hyperlane packages
		Recipient: hyperlaneutil.HexAddress(orbModAddr32Bz),
		Body:      fullPayloadBz,
	}, nil
}

func createTestPayload() (*core.Payload, error) {
	feeAction, err := action.NewFeeAction(&action.FeeInfo{
		Recipient:   "noble1vpq5ecul8v3ecq5hl4dhneekwu08sjwkugm979",
		BasisPoints: 123,
	})
	if err != nil {
		return nil, err
	}

	fw, err := forwarding.NewCCTPForwarding(
		0,
		testutil.RandomBytes(32),
		testutil.RandomBytes(32),
		nil,
	)
	if err != nil {
		return nil, err
	}

	payload, err := core.NewPayload(fw, feeAction)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
