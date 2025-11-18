package noble

import (
	"errors"
	"fmt"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/noble-assets/orbiter/v2/e2e/refactor/chain"
)

const InitialWalletBalance = 1_000_000_000

var _ chain.Manager = (*Noble)(nil)

// Noble represents the Noble chain runtime state after building.
type Noble struct {
	chain *cosmos.CosmosChain

	// Generic chain accounts
	Accounts []ibc.Wallet

	// Specific chain accounts
	Circle *CircleAccounts
}

func (n *Noble) Chain() *cosmos.CosmosChain {
	return n.chain
}

func (n *Noble) Sender() (ibc.Wallet, error) {
	if len(n.Accounts) == 0 {
		return nil, errors.New("no wallets available")
	}
	return n.Accounts[0], nil
}

func (n *Noble) Wallet(idx int) (ibc.Wallet, error) {
	if idx >= len(n.Accounts) {
		return nil, fmt.Errorf(
			"wallet index %d out of range (have %d wallets)",
			idx,
			len(n.Accounts),
		)
	}
	return n.Accounts[idx], nil
}

// DefaultNobleConfig returns the default configuration for Noble chain.
func DefaultNobleConfig() *chain.Config {
	nodesConfig := chain.DefaultNodesConfig()

	denom := chain.UsdcMetadata
	name := "noble-e2e"

	return &chain.Config{
		Name:           name,
		ChainID:        name + "-1",
		Version:        "local",
		Denom:          denom,
		NumValidators:  nodesConfig.NumValidator,
		NumFullNodes:   nodesConfig.NumFullNodes,
		NumWallets:     nodesConfig.NumWallets,
		GasPrices:      1,
		GasAdjustment:  1.5,
		Bin:            "simd",
		LogLevel:       "*:info",
		Bech32Prefix:   "noble",
		TrustingPeriod: "504h",
	}
}

func WithDenom(denom *banktypes.Metadata) NobleOption {
	return func(nb *Builder) {
		nb.config.Denom = denom
	}
}

func WithNumWallets(n int) NobleOption {
	return func(nb *Builder) {
		nb.config.NumWallets = n
	}
}

func WithOrbiterLogLevel(level string) NobleOption {
	return func(nb *Builder) {
		orbiterLevel := fmt.Sprintf("orbiter:%s", level)
		nb.config.LogLevel = fmt.Sprintf("%s,%s", nb.config.LogLevel, orbiterLevel)
	}
}
