package dispatcher

import (
	"context"

	"cosmossdk.io/errors"

	dispatchertypes "orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

// InitGenesis initialize the state of the component with a genesis state.
func (d *Dispatcher) InitGenesis(ctx context.Context, g *dispatchertypes.GenesisState) error {
	if g == nil {
		return core.ErrNilPointer.Wrap("dispatcher genesis")
	}

	for _, a := range g.DispatchedAmounts {
		if err := d.SetDispatchedAmount(ctx, a.SourceId, a.DestinationId, a.Denom, a.AmountDispatched); err != nil {
			return errors.Wrap(err, "failed to set dispatcher amount during genesis initialization")
		}
	}

	for _, c := range g.DispatchedCounts {
		if err := d.SetDispatchedCounts(ctx, c.SourceId, c.DestinationId, c.Count); err != nil {
			return errors.Wrap(err, "failed to set dispatches count during genesis initialization")
		}
	}

	return nil
}

// ExportGenesis returns the current state of the component into a genesis state.
func (d *Dispatcher) ExportGenesis(ctx context.Context) *dispatchertypes.GenesisState {
	return &dispatchertypes.GenesisState{
		DispatchedAmounts: d.GetAllDispatchedAmounts(ctx),
		DispatchedCounts:  d.GetAllDispatchedCounts(ctx),
	}
}
