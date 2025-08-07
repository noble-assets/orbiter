package keeper

import (
	"context"

	"orbiter.dev/types/component/adapter"
)

var _ adapter.QueryServer = &queryServerAdapter{}

type queryServerAdapter struct {
	*Keeper
}

func NewQueryServerAdapter(keeper *Keeper) queryServerAdapter {
	return queryServerAdapter{Keeper: keeper}
}

// Params implements adapter.QueryClient.
func (s queryServerAdapter) Params(
	ctx context.Context,
	req *adapter.QueryParamsRequest,
) (*adapter.QueryParamsResponse, error) {
	a := s.Adapter()

	params := a.GetParams(ctx)

	return &adapter.QueryParamsResponse{
		Params: params,
	}, nil
}
