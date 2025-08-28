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

package adapter

import (
	"context"
	"encoding/binary"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"

	"github.com/noble-assets/orbiter/types"
	adaptertypes "github.com/noble-assets/orbiter/types/component/adapter"
	"github.com/noble-assets/orbiter/types/core"
)

var _ adaptertypes.MsgServer = &msgServer{}

// msgServer is the server used to handle messages for the component.
type msgServer struct {
	*Adapter
	types.Authorizer
}

func NewMsgServer(a *Adapter, auth types.Authorizer) msgServer {
	return msgServer{Adapter: a, Authorizer: auth}
}

// UpdateParams implements adapter.MsgServer.
func (s msgServer) UpdateParams(
	ctx context.Context,
	msg *adaptertypes.MsgUpdateParams,
) (*adaptertypes.MsgUpdateParamsResponse, error) {
	if err := s.RequireAuthority(msg.Signer); err != nil {
		return nil, err
	}

	if err := s.SetParams(ctx, msg.Params); err != nil {
		return nil, err
	}

	return &adaptertypes.MsgUpdateParamsResponse{}, nil
}

// CCTPEntrypoint implements adapter.MsgServer.
func (s msgServer) CCTPEntrypoint(
	ctx context.Context,
	msg *adaptertypes.MsgCCTPEntrypoint,
) (*adaptertypes.MsgCCTPEntrypointResponse, error) {
	rawPayload, err := s.cdc.Marshal(msg.Payload)
	if err != nil {
		return nil, err
	}
	bz := make([]byte, 2+len(msg.TransferMessage)+2+len(msg.PayloadMessage)+len(rawPayload))
	offset := 0
	binary.BigEndian.PutUint16(bz[offset:offset+2], uint16(len(msg.TransferMessage)))
	offset += 2
	copy(bz[offset:offset+len(msg.TransferMessage)], msg.TransferMessage)
	offset += len(msg.TransferMessage)
	binary.BigEndian.PutUint16(bz[offset:offset+2], uint16(len(msg.PayloadMessage)))
	offset += 2
	copy(bz[offset:offset+len(msg.PayloadMessage)], msg.PayloadMessage)
	offset += len(msg.PayloadMessage)
	copy(bz[offset:], rawPayload)

	isOrbiterPayload, _, err := s.ParsePayload(core.PROTOCOL_CCTP, bz)
	if err != nil {
		return nil, err
	}

	if !isOrbiterPayload {
	}

	ccID, err := core.NewCrossChainID(core.PROTOCOL_CCTP, "TODO")
	if err != nil {
		return nil, err
	}

	err = s.BeforeTransferHook(ctx, ccID, msg.Payload)
	if err != nil {
		return nil, err
	}

	transferRes, err := s.cctpServer.ReceiveMessage(ctx, &cctptypes.MsgReceiveMessage{
		From:        msg.Signer,
		Message:     msg.TransferMessage,
		Attestation: msg.TransferAttestation,
	})
	if err != nil || !transferRes.Success {
	}
	payloadRes, err := s.cctpServer.ReceiveMessage(ctx, &cctptypes.MsgReceiveMessage{
		From:        msg.Signer,
		Message:     msg.PayloadMessage,
		Attestation: msg.TransferAttestation,
	})
	if err != nil || !payloadRes.Success {
	}

	transferAttr, err := s.AfterTransferHook(ctx, ccID, msg.Payload)
	if err != nil {
		return nil, err
	}

	err = s.ProcessPayload(ctx, transferAttr, msg.Payload)
	if err != nil {
		return nil, err
	}

	return &adaptertypes.MsgCCTPEntrypointResponse{}, nil
}
