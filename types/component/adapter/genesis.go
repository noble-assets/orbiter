package adapter

// DefaultGenesisState returns the default values for the adapter
// component initial state.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: Params{
			MaxPassthroughPayloadSize: 0,
		},
	}
}

// Validate retusn an error if any of the genesis field is not valid.
func (g *GenesisState) Validate() error { return nil }
