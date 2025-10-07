package keeper

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	orbitertypes "github.com/noble-assets/orbiter/types"
	"github.com/noble-assets/orbiter/types/core"
)

var _ orbitertypes.QueryServer = &Keeper{}

func (k *Keeper) PendingPayloads(
	ctx context.Context,
	req *orbitertypes.QueryPendingPayloadsRequest,
) (*orbitertypes.QueryPendingPayloadsResponse, error) {
	hashes, pageRes, err := query.CollectionPaginate(
		ctx,
		k.pendingPayloads,
		req.Pagination,
		func(hash []byte, _ core.PendingPayload) (string, error) {
			return ethcommon.BytesToHash(hash).Hex(), nil
		},
	)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to paginate pending payloads")
	}

	return &orbitertypes.QueryPendingPayloadsResponse{
		Hashes:     hashes,
		Pagination: pageRes,
	}, nil
}
