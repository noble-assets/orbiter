package adapter

import (
	"context"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
)

// HyperlaneCoreKeeper specifies the expected interface of Hyperlane
// core functionality that is required for the Orbiter execution.
type HyperlaneCoreKeeper interface {
	AppRouter() *util.Router[util.HyperlaneApp]
}

// HyperlaneWarpKeeper specifies the expected interface of Hyperlane
// warp functionality that is required for the Orbiter execution.
type HyperlaneWarpKeeper interface {
	Handle(
		ctx context.Context,
		mailboxId util.HexAddress,
		message util.HyperlaneMessage,
	) error
}
