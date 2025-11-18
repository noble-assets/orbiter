package chain

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

type Manager interface {
	Chain() *cosmos.CosmosChain
	Wallet(int) (ibc.Wallet, error)
}

// Config represents the configuration for a chain before it's built.
type Config struct {
	Name           string
	ChainID        string
	Version        string
	Denom          *banktypes.Metadata
	NumValidators  int
	NumFullNodes   int
	NumWallets     int
	GasPrices      int
	LogLevel       string
	Bech32Prefix   string
	Bin            string
	TrustingPeriod string
	GasAdjustment  float64
}

// GenesisModifier modifies the genesis file of a chain.
type GenesisModifierFn func([]cosmos.GenesisKV) []cosmos.GenesisKV

// GenesisModifiers is a composable list of genesis modifiers.
type GenesisModifiers []GenesisModifierFn

func (gm GenesisModifiers) Apply() func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
	return func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		genesis := []cosmos.GenesisKV{}

		for _, modifier := range gm {
			genesis = modifier(genesis)
		}

		return cosmos.ModifyGenesis(genesis)(cc, b)
	}
}

// PreGenesisModifier executes before genesis is created.
type PreGenesisModifierFn func(ctx context.Context, chain ibc.Chain) error

type PreGenesisModifiers []PreGenesisModifierFn

func (pg PreGenesisModifiers) Apply(ctx context.Context) func(ibc.Chain) error {
	return func(c ibc.Chain) error {
		for _, modifier := range pg {
			if err := modifier(ctx, c); err != nil {
				return err
			}
		}

		return nil
	}
}
