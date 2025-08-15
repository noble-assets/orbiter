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

package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"orbiter.dev/testutil/testdata"
	"orbiter.dev/types"
	"orbiter.dev/types/core"
)

// CreateValidIBCPacketData creates a valid IBC FungibleTokenPacketData with given parameters.
func CreateValidIBCPacketData(sender, receiver, memo string) []byte {
	packetData := transfertypes.NewFungibleTokenPacketData(
		"uusdc",
		"1000000",
		sender,
		receiver,
		memo,
	)

	return packetData.GetBytes()
}

func CreateValidOrbiterPayloadWithActions() string {
	return `{"orbiter": {"forwarding": {"protocol_id": 2, "attributes": { "@type" : "/testpb.TestForwardingAttr", "planet": "earth" }}, "pre_actions": [{"id": 1, "attributes": { "@type" : "/testpb.TestActionAttr", "whatever": "it takes" }}]}}`
}

func CreateValidOrbiterPayload() string {
	return `{"orbiter": {"forwarding": {"protocol_id": 2, "attributes": { "@type" : "/testpb.TestForwardingAttr", "planet": "earth" }}}}`
}

func CreatePayloadWrapperJSON(t *testing.T) (*core.Payload, string) {
	t.Helper()

	encCfg := MakeTestEncodingConfig("noble")
	encCfg.InterfaceRegistry.RegisterImplementations(
		(*core.ForwardingAttributes)(nil),
		&testdata.TestForwardingAttr{},
	)

	forwardingAttributes := testdata.TestForwardingAttr{Planet: "venus"}
	forwarding, err := core.NewForwarding(
		core.PROTOCOL_IBC,
		&forwardingAttributes,
		[]byte("payload"),
	)
	require.NoError(t, err)
	payloadWrapper, err := core.NewPayloadWrapper(forwarding, []*core.Action{})
	require.NoError(t, err)
	bz, err := types.MarshalJSON(encCfg.Codec, payloadWrapper)
	require.NoError(t, err)

	return payloadWrapper.Orbiter, string(bz)
}

func CreatePayloadWithAction(t *testing.T) (*core.Payload, string) {
	t.Helper()

	encCfg := MakeTestEncodingConfig("noble")
	encCfg.InterfaceRegistry.RegisterImplementations(
		(*core.ForwardingAttributes)(nil),
		&testdata.TestForwardingAttr{},
	)
	encCfg.InterfaceRegistry.RegisterImplementations(
		(*core.ActionAttributes)(nil),
		&testdata.TestActionAttr{},
	)

	forwardingAttributes := testdata.TestForwardingAttr{Planet: "venus"}
	forwarding, err := core.NewForwarding(
		core.PROTOCOL_IBC,
		&forwardingAttributes,
		[]byte("payload"),
	)
	require.NoError(t, err)

	actionAttributes := testdata.TestActionAttr{Whatever: "it takes"}
	action, err := core.NewAction(core.ACTION_FEE, &actionAttributes)
	require.NoError(t, err)

	payloadWrapper, err := core.NewPayloadWrapper(forwarding, []*core.Action{action})
	require.NoError(t, err)
	bz, err := types.MarshalJSON(encCfg.Codec, payloadWrapper)
	require.NoError(t, err)

	return payloadWrapper.Orbiter, string(bz)
}
