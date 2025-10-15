package keeper

import (
	"context"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const PendingPayloadLifespan = 24 * time.Hour

func (k *Keeper) BeginBlock(ctx context.Context) error {
	blockTime := sdk.UnwrapSDKContext(ctx).BlockTime().UnixNano()

	cutoff := time.Unix(
		0,
		blockTime-PendingPayloadLifespan.Nanoseconds(),
	)

	return k.RemoveExpiredPayloads(ctx, cutoff)
}
