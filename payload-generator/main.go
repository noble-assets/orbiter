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

package main

import (
	"fmt"
	"log"

	"orbiter.dev"
	"orbiter.dev/testutil"
	"orbiter.dev/types"
	"orbiter.dev/types/controller/action"
	"orbiter.dev/types/controller/forwarding"
	"orbiter.dev/types/core"
)

func main() {
	destinationDomain := uint32(0)
	mintRecipient := testutil.RandomBytes(32)
	destinationCaller := testutil.RandomBytes(32)
	passthroughPayload := []byte("")

	cctpForwarding, err := forwarding.NewCCTPForwarding(
		destinationDomain,
		mintRecipient,
		destinationCaller,
		passthroughPayload,
	)
	if err != nil {
		log.Fatalf("Failed to create CCTP forwarding: %v", err)
	}

	feeRecipientAddr := testutil.NewNobleAddress()
	feeAttr := action.FeeAttributes{
		FeesInfo: []*action.FeeInfo{
			{
				Recipient:   feeRecipientAddr,
				BasisPoints: 100,
			},
		},
	}

	feeAction := core.Action{
		Id: core.ACTION_FEE,
	}
	err = feeAction.SetAttributes(&feeAttr)
	if err != nil {
		log.Fatalf("Failed to set action attributes: %v", err)
	}

	payload, err := core.NewPayloadWrapper(cctpForwarding, []*core.Action{&feeAction})
	if err != nil {
		log.Fatalf("Failed to create payload wrapper: %v", err)
	}

	encCfg := testutil.MakeTestEncodingConfig("noble")
	orbiter.RegisterInterfaces(encCfg.InterfaceRegistry)
	payloadStr, err := types.MarshalJSON(encCfg.Codec, payload)
	if err != nil {
		log.Fatalf("Failed to marshal payload: %v", err)
	}

	fmt.Printf("Generated Orbiter Payload:\n%s\n\n", string(payloadStr))
}
