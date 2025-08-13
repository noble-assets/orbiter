package dispatcher

import (
	"cosmossdk.io/errors"

	"orbiter.dev/types/core"
)

// DefaultGenesisState returns the default values for the component initial state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		DispatchedAmounts: []DispatchedAmountEntry{},
		DispatchedCounts:  []DispatchCountEntry{},
	}
}

// Validate returns an error if any of the genesis field is not valid.
func (g *GenesisState) Validate() error {
	for _, a := range g.DispatchedAmounts {
		if a.Denom == "" {
			return errors.Wrap(core.ErrValidation, "cannot set empty denom")
		}
		if err := a.SourceId.Validate(); err != nil {
			return errors.Wrap(err, "failed to create source cross-chain ID")
		}
		if err := a.DestinationId.Validate(); err != nil {
			return errors.Wrap(err, "failed to create destination cross-chain ID")
		}
		if !a.AmountDispatched.Incoming.IsPositive() && !a.AmountDispatched.Outgoing.IsPositive() {
			return errors.Wrap(
				core.ErrValidation,
				"cannot set incoming and outgoing amounts equal to zero",
			)
		}
	}

	for _, c := range g.DispatchedCounts {
		if c.Count == 0 {
			return errors.Wrap(core.ErrValidation, "cannot set zero count")
		}
		if err := c.SourceId.Validate(); err != nil {
			return errors.Wrap(err, "failed to create source cross-chain ID")
		}
		if err := c.DestinationId.Validate(); err != nil {
			return errors.Wrap(err, "failed to create destination cross-chain ID")
		}
	}

	return nil
}
