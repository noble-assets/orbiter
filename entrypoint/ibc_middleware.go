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
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var _ porttypes.Middleware = &IBCMiddleware{}

// IBCMiddleware implements the ICS26 callbacks (IBCModule) and ICS4Wrapper.
type IBCMiddleware struct {
	porttypes.IBCModule
	porttypes.ICS4Wrapper

	payloadAdapter types.PayloadAdapter
}

func NewIBCMiddleware(
	app porttypes.IBCModule,
	ics4Wrapper porttypes.ICS4Wrapper,
	payloadAdapter types.PayloadAdapter,
) IBCMiddleware {
	if app == nil {
		panic(core.ErrNilPointer.Wrap("IBC module is not set"))
	}

	if ics4Wrapper == nil {
		panic(core.ErrNilPointer.Wrap("ICS4 wrapper module is not set"))
	}

	if payloadAdapter == nil {
		panic(core.ErrNilPointer.Wrap("payload adapter is not set"))
	}

	return IBCMiddleware{
		IBCModule:      app,
		ICS4Wrapper:    ics4Wrapper,
		payloadAdapter: payloadAdapter,
	}
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
		core.PROTOCOL_IBC,
		packet.GetData(),
	)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	if !isOrbiterPayload {
		return i.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// NOTE: we are using destination channel here since that is the channel identifier of the
	// source chain on Noble.
	ccID, err := core.NewCrossChainID(core.PROTOCOL_IBC, packet.DestinationChannel)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	err = i.payloadAdapter.BeforeTransferHook(ctx, ccID, orbiterPayload)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	ack := i.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	transferAttr, err := i.payloadAdapter.AfterTransferHook(ctx, ccID, orbiterPayload)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	err = i.payloadAdapter.ProcessPayload(ctx, transferAttr, orbiterPayload)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	return ack
}

func newErrorAcknowledgement(err error) channeltypes.Acknowledgement {
	return channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: errorsmod.Wrap(err, "orbiter-middleware error").Error(),
		},
	}
}
