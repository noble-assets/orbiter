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

package entrypoint

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"orbiter.dev/types"
	"orbiter.dev/types/interfaces"
)

var _ porttypes.Middleware = &IBCMiddleware{}

func NewIBCMiddleware(
	app porttypes.IBCModule,
	ics4Wrapper porttypes.ICS4Wrapper,
	payloadAdapter interfaces.PayloadAdapter,
) IBCMiddleware {
	if app == nil {
		panic(errors.New("IBC module cannot be nil"))
	}

	if ics4Wrapper == nil {
		panic(errors.New("ICS4 wrapper cannot be nil"))
	}

	if payloadAdapter == nil {
		panic(errors.New("payload adapter cannot be nil"))
	}

	return IBCMiddleware{
		IBCModule:      app,
		ICS4Wrapper:    ics4Wrapper,
		payloadAdapter: payloadAdapter,
	}
}

// IBCMiddleware implements the ICS26 callbacks (IBCModule) and ICS4Wrapper.
type IBCMiddleware struct {
	porttypes.IBCModule
	porttypes.ICS4Wrapper

	payloadAdapter interfaces.PayloadAdapter
}

// ====================================================================================================
// IBCModule interface
// ====================================================================================================

// OnRecvPacket implements types.Middleware.
func (i IBCMiddleware) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	isOrbiterPayload, orbiterPayload, err := i.payloadAdapter.ParsePayload(
		types.PROTOCOL_IBC,
		packet.GetData(),
	)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	} else if !isOrbiterPayload {
		return i.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	orbitID, err := types.NewOrbitID(types.PROTOCOL_IBC, packet.SourceChannel)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	err = i.payloadAdapter.BeforeTransferHook(ctx, orbitID, orbiterPayload)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	ack := i.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	transferAttr, err := i.payloadAdapter.AfterTransferHook(ctx, orbitID, orbiterPayload)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	err = i.payloadAdapter.ProcessPayload(ctx, transferAttr, orbiterPayload)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	return ack
}
