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
	"context"
	"fmt"

	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	hyperlanepostdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	hyperlanecoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	interchaintestcosmos "github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"
)

// getHyperlaneNoOpISM returns the first found No-Op ISM that's registered on the given node.
func getHyperlaneNoOpISM(
	ctx context.Context,
	node *interchaintestcosmos.ChainNode,
) (*ismtypes.NoopISM, error) {
	res, err := ismtypes.NewQueryClient(node.GrpcConn).
		Isms(ctx, &ismtypes.QueryIsmsRequest{})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to query isms")
	}

	if len(res.Isms) != 1 {
		return nil, fmt.Errorf("expected exactly 1 ism, got %d", len(res.Isms))
	}

	var ism ismtypes.NoopISM
	err = proto.Unmarshal(res.Isms[0].Value, &ism)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to unmarshal ism")
	}

	return &ism, nil
}

// getHyperlaneNoOpHook returns the ID of the first registered hook
func getHyperlaneNoOpHook(
	ctx context.Context,
	node *interchaintestcosmos.ChainNode,
) (*hyperlanepostdispatchtypes.NoopHook, error) {
	res, err := hyperlanepostdispatchtypes.
		NewQueryClient(node.GrpcConn).
		NoopHooks(ctx, &hyperlanepostdispatchtypes.QueryNoopHooksRequest{})
	if err != nil {
		return nil, err
	}

	if len(res.NoopHooks) != 1 {
		return nil, fmt.Errorf("expected exactly 1 noop hook, got %d", len(res.NoopHooks))
	}

	return &res.NoopHooks[0], nil
}

// getHyperlaneMailbox returns the registered Hyperlane mailbox on the given node.
func getHyperlaneMailbox(
	ctx context.Context,
	node *interchaintestcosmos.ChainNode,
) (*hyperlanecoretypes.Mailbox, error) {
	res, err := hyperlanecoretypes.NewQueryClient(node.GrpcConn).
		Mailboxes(ctx, &hyperlanecoretypes.QueryMailboxesRequest{})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to query mailboxes")
	}

	if len(res.Mailboxes) != 1 {
		return nil, fmt.Errorf("expected exactly 1 mailbox; found %d", len(res.Mailboxes))
	}

	return &res.Mailboxes[0], nil
}

func getHyperlaneCollateralToken(
	ctx context.Context,
	node *interchaintestcosmos.ChainNode,
) (*warptypes.WrappedHypToken, error) {
	res, err := warptypes.NewQueryClient(node.GrpcConn).
		Tokens(ctx, &warptypes.QueryTokensRequest{})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to query tokens")
	}

	if len(res.Tokens) != 1 {
		return nil, fmt.Errorf("expected exactly 1 token; found %d", len(res.Tokens))
	}

	return &res.Tokens[0], nil
}
