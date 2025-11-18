package noble

import (
	"context"

	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"github.com/noble-assets/orbiter/v2/e2e/refactor/chain"
	"github.com/noble-assets/orbiter/v2/e2e/refactor/suite"
)

// Builder builds a Noble chain with configurable options.
type Builder struct {
	ctx context.Context
	// config is required only during the building process. After that, the config will be available
	// from the Noble type.
	config *chain.Config

	noble *Noble

	preGenesisModifiers chain.PreGenesisModifiers
	genesisModifiers    chain.GenesisModifiers
}

// NewBuilder creates a new default noble build and apply options to it.
func NewBuilder(ctx context.Context, opts ...NobleOption) *Builder {
	nb := &Builder{
		ctx:    ctx,
		config: DefaultNobleConfig(),
		noble:  &Noble{
			// chain is set later, before calling the interchain Build method
		},
	}

	for _, opt := range opts {
		opt(nb)
	}

	return nb
}

// BuildSpec creates the interchaintest chain specification.
func (nb *Builder) BuildSpec() *interchaintest.ChainSpec {
	cfg := nb.config

	return &interchaintest.ChainSpec{
		Name:          cfg.Name,
		Version:       cfg.Version,
		NumValidators: &cfg.NumValidators,
		NumFullNodes:  &cfg.NumFullNodes,
		ChainConfig: ibc.ChainConfig{
			Images: []ibc.DockerImage{
				{Repository: "noble-simd", Version: "local", UIDGID: "1025:1025"},
			},
			Type:                "cosmos",
			Name:                cfg.Name,
			ChainID:             cfg.ChainID,
			Bin:                 cfg.Bin,
			Bech32Prefix:        cfg.Bech32Prefix,
			Denom:               cfg.Denom.Base,
			GasPrices:           string(cfg.GasPrices) + cfg.Denom.Base,
			GasAdjustment:       cfg.GasAdjustment,
			TrustingPeriod:      cfg.TrustingPeriod,
			NoHostMount:         false,
			AdditionalStartArgs: []string{"--log_level", cfg.LogLevel},
			PreGenesis:          nb.preGenesis,
			ModifyGenesis:       nb.modifyGenesis,
		},
	}
}

func (nb *Builder) preGenesis(chain ibc.Chain) error {
	return nb.preGenesisModifiers.Apply(nb.ctx)(chain)
}

func (nb *Builder) modifyGenesis(cfg ibc.ChainConfig, genesis []byte) ([]byte, error) {
	return nb.genesisModifiers.Apply()(cfg, genesis)
}

func (nb *Builder) BuilderOpt() suite.BuilderOpt {
	return func(s *suite.SuiteBuilder) {
		nobleSpec := nb.BuildSpec()

		s.AppendChainSpec(nobleSpec)
	}
}

// func (nb *NobleBuilder) Initialize(chain *cosmos.CosmosChain) *Noble {
// 	// Get the pre-allocated wallets from interchaintest
// 	wallets := make([]ibc.Wallet, nb.config.NumWallets)
//
// 	for i := 0; i < nb.config.NumWallets; i++ {
// 		// These wallets are created by interchaintest during chain setup
// 		wallets[i] := interchaintest.GetAndFundTestUsers(
// 			t,
// 			ctx,
// 			"wallet",
// 			math.NewInt(1_000_000_000),
// 			chain,
// 		)
// 	}
//
// 	return &Noble{
// 		chain:   chain,
// 		wallets: wallets,
// 		Circle:  nb.circle,
// 	}
// }

// SetValidator sets the validator node for genesis setup operations.
// This must be called before Build() if using pre-genesis hooks that need validator access.
// func (nb *NobleBuilder) SetValidator(val *cosmos.ChainNode) {
// 	// Update all CircleGenesisSetup instances with the validator
// 	for _, hook := range nb.preGenesisHooks {
// 		if circleSetup, ok := hook.(*CircleGenesisSetup); ok {
// 			circleSetup.val = val
// 		}
// 	}
// }
