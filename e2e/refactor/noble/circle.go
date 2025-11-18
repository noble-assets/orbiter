package noble

import (
	"context"
	"fmt"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/v2/e2e/refactor/chain"
)

// WithCircleAccounts adds Circle CCTP and FTF accounts to the Noble chain builder. The function
// set the accounts into the noble chain, and register their pre genesis and genesis modifiers.
func WithCircleAccounts() NobleOption {
	return func(nb *Builder) {
		// Initialize circle accounts.
		nb.noble.Circle = &CircleAccounts{}

		// nb.preGenesisModifiers = append(nb.preGenesisModifiers,
		// nb.noble.Circle.modifyPreGenesis())
		nb.genesisModifiers = append(nb.genesisModifiers, nb.noble.Circle.modifyGenesis())
	}
}

var _ chain.AccountsManager = (*CircleAccounts)(nil)

// CircleAccounts manages Circle-specific accounts (CCTP).
type CircleAccounts struct {
	Pauser         ibc.Wallet
	TokenMessenger ibc.Wallet
	Minter         ibc.Wallet
}

func (c *CircleAccounts) GetAccountSpecs() map[string]*ibc.Wallet {
	return map[string]*ibc.Wallet{
		"pauser":          &c.Pauser,
		"token-messenger": &c.TokenMessenger,
		"minter":          &c.Minter,
	}
}

func (c *CircleAccounts) modifyGenesis() func(gen []cosmos.GenesisKV) []cosmos.GenesisKV {
	return func(gen []cosmos.GenesisKV) []cosmos.GenesisKV {
		gen = modifyGenesisFTF(c, gen)
		gen = modifyGenesisCCTP(c, gen)

		return gen
	}
}

func modifyGenesisFTF(circleAcc *CircleAccounts, gen []cosmos.GenesisKV) []cosmos.GenesisKV {
	return append(gen,
		cosmos.NewGenesisKV("app_state.fiat-tokenfactory", fiattokenfactorytypes.GenesisState{
			Paused: &fiattokenfactorytypes.Paused{Paused: false},
			Pauser: &fiattokenfactorytypes.Pauser{
				Address: circleAcc.Pauser.FormattedAddress(),
			},
			MintersList: []fiattokenfactorytypes.Minters{
				{
					Address: circleAcc.Minter.FormattedAddress(),
					Allowance: sdk.Coin{
						Denom:  chain.UsdcMetadata.Base,
						Amount: math.NewInt(InitialWalletBalance),
					},
				},
			},
			MintingDenom: &fiattokenfactorytypes.MintingDenom{Denom: chain.UsdcMetadata.Base},
		}),
	)
}

func modifyGenesisCCTP(circleAcc *CircleAccounts, gen []cosmos.GenesisKV) []cosmos.GenesisKV {
	tokenMessenger := make([]byte, 32)
	copy(tokenMessenger[12:], circleAcc.TokenMessenger.Address())

	return append(gen,
		cosmos.NewGenesisKV("app_state.cctp", cctptypes.GenesisState{
			TokenMessengerList: []cctptypes.RemoteTokenMessenger{
				{DomainId: 0, Address: tokenMessenger},
			},
			Pauser: circleAcc.Pauser.FormattedAddress(),
			BurningAndMintingPaused: &cctptypes.BurningAndMintingPaused{
				Paused: false,
			},
			SendingAndReceivingMessagesPaused: &cctptypes.SendingAndReceivingMessagesPaused{
				Paused: false,
			},
			PerMessageBurnLimitList: []cctptypes.PerMessageBurnLimit{
				{Denom: chain.UsdcMetadata.Base, Amount: math.NewInt(InitialWalletBalance)},
			},
		}),
	)
}

// func (c *CircleAccounts) modifyPreGenesis() func(ctx context.Context, ibcChain ibc.Chain) error {
// 	return func(ctx context.Context, ibcChain ibc.Chain) error {
// 		for name, walletPtr := range c.GetAccountSpecs() {
// 			wallet, err := ibcChain.BuildRelayerWallet(ctx, name)
// 			if err != nil {
// 				return fmt.Errorf("failed to create wallet for account %s: %w", name, err)
// 			}
//
// 			*walletPtr = wallet
// 		}
//
// 		return nil
// 	}
// }
