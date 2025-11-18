package e2e

import (
	"context"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

type (
	PreGenesisFn    func(ibc.Chain) error
	ModifyGenesisFn func(cc ibc.ChainConfig, b []byte) ([]byte, error)
)

type AccountsManager interface {
	GetAccountSpecs() map[string]*ibc.Wallet
}

type ValidatorManager interface {
	GetValidator() *cosmos.ChainNode
}

type ChainConfigurator interface {
	Chain() *cosmos.CosmosChain
	ChainSpec() *interchaintest.ChainSpec
	Sender() ibc.Wallet
	WalletNumber(uint32) ibc.Wallet
}

type NodeConfig struct {
	NumValidators int
	NumFullNodes  int
	NumWallet     int
}

func NewDefaultNodeConfig() *NodeConfig {
	return &NodeConfig{
		NumValidators: 1,
		NumFullNodes:  0,
		NumWallet:     3,
	}
}
