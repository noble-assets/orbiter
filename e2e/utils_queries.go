package e2e

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"fmt"
	hyperlanepostdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/cosmos/gogoproto/proto"

	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	hyperlanecoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	interchaintestcosmos "github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
)

// getHyperlaneNoOpISM returns the first found No-Op ISM that's registered on the given node.
func getHyperlaneNoOpISM(ctx context.Context, node *interchaintestcosmos.ChainNode) (*ismtypes.NoopISM, error) {
	client := ismtypes.NewQueryClient(node.GrpcConn)

	res, err := client.Isms(ctx, &ismtypes.QueryIsmsRequest{})
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to query isms")
	}

	if len(res.Isms) != 1 {
		return nil, fmt.Errorf("expected exactly 1 ism, got %d", len(res.Isms))
	}

	var ism *ismtypes.NoopISM
	err = proto.Unmarshal(res.Isms[0].Value, ism)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to unmarshal ism")
	}

	return ism, nil
}

// getHyperlaneNoOpHook returns the ID of the first registered hook
func getHyperlaneNoOpHook(ctx context.Context, node *interchaintestcosmos.ChainNode) (*hyperlanepostdispatchtypes.NoopHook, error) {
	res, err := hyperlanepostdispatchtypes.
		NewQueryClient(node.GrpcConn).
		NoopHooks(ctx, &hyperlanepostdispatchtypes.QueryNoopHooksRequest{})
	if err != nil {
		return nil, err
	}

	if len(res.NoopHooks) != 0 {
		return nil, fmt.Errorf("expected exactly 1 noop hook, got %d", len(res.NoopHooks))
	}

	return &res.NoopHooks[0], nil
}

// getHyperlaneMailbox returns the registered Hyperlane mailbox on the given node.
func getHyperlaneMailbox(ctx context.Context, node *interchaintestcosmos.ChainNode) (*hyperlanecoretypes.Mailbox, error) {
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

func getHyperlaneCollateralToken(ctx context.Context, node *interchaintestcosmos.ChainNode) (*warptypes.WrappedHypToken, error) {
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
