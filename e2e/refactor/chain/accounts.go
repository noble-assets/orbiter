package chain

import (
	"context"

	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	errorsmod "cosmossdk.io/errors"
)

// AccountsManager defines the interface for managing genesis accounts.
type AccountsManager interface {
	GetAccountSpecs() map[string]*ibc.Wallet
}

func CreateGenesisAccount(
	ctx context.Context,
	accountsManager AccountsManager,
	val *cosmos.ChainNode,
) error {
	chain := val.Chain

	for name, walletPtr := range accountsManager.GetAccountSpecs() {
		wallet, err := chain.BuildRelayerWallet(ctx, name)
		if err != nil {
			return errorsmod.Wrapf(err, "failed to create wallet for account %s", name)
		}

		if err := val.RecoverKey(ctx, wallet.KeyName(), wallet.Mnemonic()); err != nil {
			return errorsmod.Wrapf(err, "failed to restore %s wallet", wallet.KeyName())
		}

		*walletPtr = wallet
	}

	return nil
}
