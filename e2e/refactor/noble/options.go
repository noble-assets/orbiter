package noble

import "github.com/noble-assets/orbiter/v2/e2e/refactor/chain"

type NobleOption func(*Builder)

// WithGenesisModifier adds a custom genesis modifier.
func WithGenesisModifier(gm chain.GenesisModifierFn) NobleOption {
	return func(nb *Builder) {
		nb.genesisModifiers = append(nb.genesisModifiers, gm)
	}
}

// WithPreGenesisModifier adds a custom pre-genesis hook.
func WithPreGenesisModifier(pg chain.PreGenesisModifierFn) NobleOption {
	return func(nb *Builder) {
		nb.preGenesisModifiers = append(nb.preGenesisModifiers, pg)
	}
}
