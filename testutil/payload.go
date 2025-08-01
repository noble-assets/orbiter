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
	return `{"orbiter": {"orbit": {"protocol_id": 2, "attributes": { "@type" : "/testpb.TestOrbitAttr", "planet": "earth" }}, "pre_actions": [{"id": 1, "attributes": { "@type" : "/testpb.TestActionAttr", "whatever": "it takes" }}]}}`
}

func CreateValidOrbiterPayload() string {
	return `{"orbiter": {"orbit": {"protocol_id": 2, "attributes": { "@type" : "/testpb.TestOrbitAttr", "planet": "earth" }}}}`
}

func CreatePayloadWrapperJSON(t *testing.T) (*types.Payload, string) {
	t.Helper()

	encCfg := MakeTestEncodingConfig("noble")
	encCfg.InterfaceRegistry.RegisterImplementations(
		(*types.OrbitAttributes)(nil),
		&testdata.TestOrbitAttr{},
	)

	orbitAttributes := testdata.TestOrbitAttr{Planet: "venus"}
	orbit, err := types.NewOrbit(
		types.PROTOCOL_IBC,
		&orbitAttributes,
		[]byte("payload"),
	)
	require.NoError(t, err)
	payloadWrapper, err := types.NewPayloadWrapper(orbit, []*types.Action{})
	require.NoError(t, err)
	bz, err := types.MarshalJSON(encCfg.Codec, payloadWrapper)
	require.NoError(t, err)

	return payloadWrapper.Orbiter, string(bz)
}

func CreatePayloadWrapperWithActionJSON(t *testing.T) (*types.Payload, string) {
	t.Helper()

	encCfg := MakeTestEncodingConfig("noble")
	encCfg.InterfaceRegistry.RegisterImplementations(
		(*types.OrbitAttributes)(nil),
		&testdata.TestOrbitAttr{},
	)
	encCfg.InterfaceRegistry.RegisterImplementations(
		(*types.ActionAttributes)(nil),
		&testdata.TestActionAttr{},
	)

	orbitAttributes := testdata.TestOrbitAttr{Planet: "venus"}
	orbit, err := types.NewOrbit(
		types.PROTOCOL_IBC,
		&orbitAttributes,
		[]byte("payload"),
	)
	require.NoError(t, err)

	actionAttributes1 := testdata.TestActionAttr{Whatever: "it takes"}
	action1, err := types.NewAction(types.ACTION_FEE, &actionAttributes1)
	require.NoError(t, err)

	actionAttributes2 := testdata.TestActionAttr{Whatever: "whatever"}
	action2, err := types.NewAction(types.ACTION_FEE, &actionAttributes2)
	require.NoError(t, err)

	payloadWrapper, err := types.NewPayloadWrapper(orbit, []*types.Action{
		action1,
		action2,
	})
	require.NoError(t, err)
	bz, err := types.MarshalJSON(encCfg.Codec, payloadWrapper)
	require.NoError(t, err)

	return payloadWrapper.Orbiter, string(bz)
}
