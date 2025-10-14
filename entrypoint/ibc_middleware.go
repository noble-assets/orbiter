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
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
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
	// NOTE: we are using destination channel here since that is the channel identifier of the
	// source chain on Noble.
	ccID, err := core.NewCrossChainID(core.PROTOCOL_IBC, packet.DestinationChannel)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	orbiterPacket, err := i.payloadAdapter.AdaptPacket(ctx, ccID, packet.GetData())
	// If the error is the sentinel error, we call the next middleware/app in the ICS20 stack.
	if err != nil && !errors.Is(err, core.ErrNoOrbiterPacket) {
		return newErrorAcknowledgement(err)
	}
	if orbiterPacket == nil {
		return i.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	// In IBC the denom specified in the packet is the sending chain representation. We have to
	// convert the denom into the Noble representation.
	denom, err := recoverNativeDenom(
		orbiterPacket.TransferAttributes.SourceDenom(),
		packet.GetSourcePort(),
		packet.GetSourceChannel(),
	)
	if err != nil {
		return newErrorAcknowledgement(errorsmod.Wrap(err, "coin is not native"))
	}

	orbiterPacket.TransferAttributes.SetDestinationDenom(denom)

	err = i.payloadAdapter.BeforeTransferHook(ctx, orbiterPacket)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	ack := i.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	err = i.payloadAdapter.AfterTransferHook(ctx, orbiterPacket)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	err = i.payloadAdapter.ProcessPayload(ctx, orbiterPacket)
	if err != nil {
		return newErrorAcknowledgement(err)
	}

	return ack
}

func recoverNativeDenom(denom, sourcePort, sourceChannel string) (string, error) {
	if transfertypes.SenderChainIsSource(sourcePort, sourceChannel, denom) {
		return "", errors.New("coin is native of source chain")
	}

	voucherPrefix := transfertypes.GetDenomPrefix(sourcePort, sourceChannel)

	// Remove from the denom the prefix created on the source chain when it received
	// the coin from Noble.
	if !strings.HasPrefix(denom, voucherPrefix) {
		return "", fmt.Errorf(
			"denom %q missing expected IBC prefix %q",
			denom,
			voucherPrefix,
		)
	}
	unprefixedDenom := strings.TrimPrefix(denom, voucherPrefix)

	// The denomination used to send the coins is either the native denom or the hash of the path
	// if the denomination is not native.
	denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
	if !denomTrace.IsNativeDenom() {
		return "", errors.New("orbiter supports only native tokens")
	}

	return unprefixedDenom, nil
}

func newErrorAcknowledgement(err error) channeltypes.Acknowledgement {
	return channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: errorsmod.Wrap(err, "orbiter-middleware error").Error(),
		},
	}
}
