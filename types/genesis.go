package types

func DefaultGenesisState() *GenesisState {
	return &GenesisState{}
}

func (genesis *GenesisState) Validate() error {
	return nil
}
