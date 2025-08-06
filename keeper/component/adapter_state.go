package components

import (
	"context"

	"orbiter.dev/types"
)

// GetParams returns the adapter params from the state. In case of an error
// returns default values and log the error.
//
// NOTE: Returning the default is safe here since return zero
// bytes allowed, which is the restrictive condition..
func (a *AdapterComponent) GetParams(ctx context.Context) types.AdapterParams {
	params, err := a.params.Get(ctx)
	if err != nil {
		a.logger.Error("error getting params", "err", err.Error())
		return types.AdapterParams{
			MaxPassthroughPayloadSize: 0,
		}
	}
	return params
}

func (a *AdapterComponent) SetParams(ctx context.Context, params types.AdapterParams) error {
	return a.params.Set(ctx, params)
}
