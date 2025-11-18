package e2e

import (
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
)

func NewNoble(ctx context.Context, suite *Suite, nodeCfg *NodeConfig) *Noble {
	gasPrices := "1" + usdcDenomMetadata.Base

	return &Noble{
		chainSpec: &interchaintest.ChainSpec{
			Name:          "noble-e2e",
			Version:       "local",
			NumValidators: &nodeCfg.NumValidators,
			NumFullNodes:  &nodeCfg.NumFullNodes,
			ChainConfig: ibc.ChainConfig{
				Images: []ibc.DockerImage{
					{
						Repository: "noble-simd",
						Version:    "local",
						UIDGID:     "1025:1025",
					},
				},
				Type:                "cosmos",
				Name:                "noble-e2e",
				ChainID:             "noble-e2e-1",
				AdditionalStartArgs: []string{"--log_level", "*:info,orbiter:trace"},
				Bin:                 "simd",
				Bech32Prefix:        "noble",
				Denom:               usdcDenomMetadata.Base,
				GasPrices:           gasPrices,
				GasAdjustment:       1.5,
				TrustingPeriod:      "504h",
				NoHostMount:         false,
				PreGenesis:          preGenesis(ctx, suite),
				ModifyGenesis:       modifyGenesis(suite),
			},
		},
	}
}

var _ AccountsManager = (*CircleAccounts)(nil)

type CircleAccounts struct {
	Pauser         ibc.Wallet
	TokenMessenger ibc.Wallet
}

func (c *CircleAccounts) GetAccountSpecs() map[string]*ibc.Wallet {
	return map[string]*ibc.Wallet{
		"pauser":          &c.Pauser,
		"token-messenger": &c.TokenMessenger,
	}
}

var _ ChainConfigurator = (*Noble)(nil)

type Noble struct {
	chain     *cosmos.CosmosChain
	chainSpec *interchaintest.ChainSpec
	sender    ibc.Wallet
	wallets   []ibc.Wallet
}

func (n *Noble) Chain() *cosmos.CosmosChain {
	return n.chain
}

func (n *Noble) ChainSpec() *interchaintest.ChainSpec {
	return n.chainSpec
}

func (n *Noble) Sender() ibc.Wallet {
	return n.sender
}

func (n *Noble) WalletNumber(num uint32) ibc.Wallet {
	// TODO: handle not enough wallets
	return n.wallets[num]
}
