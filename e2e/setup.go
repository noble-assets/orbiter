// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2025, NASD Inc. All rights reserved.
// Use of this software is governed by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN "AS IS" BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package e2e

import (
	"context"
	"fmt"
	"testing"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	"github.com/ethereum/go-ethereum/common"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"orbiter.dev/testutil"
)

var (
	numValidators = 1
	numFullNodes  = 0
)

type IBC struct {
	CounterpartyChain *cosmos.CosmosChain
	RelayerReporter   *testreporter.RelayerExecReporter
	Relayer           ibc.Relayer
	Account           ibc.Wallet
	PathName          string
}

type Suite struct {
	Chain *cosmos.CosmosChain

	IBC *IBC

	// Addresses
	CircleRoles       CircleRoles
	sender            ibc.Wallet
	fallbackRecipient ibc.Wallet
	mintRecipient     string
	destinationCaller string

	destinationDomain uint32
}

func NewSuite(t *testing.T, isZeroFees bool, isIBC bool) (context.Context, Suite) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	reporter := testreporter.NewNopReporter()
	execReporter := reporter.RelayerExecReporter(t)

	client, network := interchaintest.DockerSetup(t)
	skipPathCreation := true

	suite := Suite{}

	// Setup chains
	chainSpecs := []*interchaintest.ChainSpec{
		createOrbiterChainSpec(ctx, &suite, isZeroFees),
	}
	if isIBC {
		chainSpecs = append(
			chainSpecs,
			&interchaintest.ChainSpec{
				Name:    "ibc-go-simd",
				Version: "v8.5.1",
				ChainConfig: ibc.ChainConfig{
					ChainID: "ibc-1",
					Denom:   "uibc",
				},
				NumValidators: &numValidators,
				NumFullNodes:  &numFullNodes,
			},
		)
	}

	factory := interchaintest.NewBuiltinChainFactory(
		logger,
		chainSpecs,
	)

	chains, err := factory.Chains(t.Name())
	require.NoError(t, err)

	suite.Chain = chains[0].(*cosmos.CosmosChain)

	if isIBC {
		rf := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, logger)
		suite.IBC = &IBC{
			CounterpartyChain: chains[1].(*cosmos.CosmosChain),
			RelayerReporter:   execReporter,
			PathName:          "transfer-path",
		}

		suite.IBC.Relayer = rf.Build(t, client, network)
	}

	// Build testing environment
	interchain := interchaintest.NewInterchain().AddChain(suite.Chain)
	if isIBC {
		interchain.AddChain(suite.IBC.CounterpartyChain).
			AddRelayer(suite.IBC.Relayer, "test-relayer").
			AddLink(interchaintest.InterchainLink{
				Chain1:  suite.Chain,
				Chain2:  suite.IBC.CounterpartyChain,
				Relayer: suite.IBC.Relayer,
				Path:    suite.IBC.PathName,
			})
		skipPathCreation = false
	}

	require.NoError(t, interchain.Build(ctx, execReporter, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: skipPathCreation,
	}))

	t.Cleanup(func() {
		_ = interchain.Close()
	})

	wallets := interchaintest.GetAndFundTestUsers(
		t,
		ctx,
		"wallet",
		math.NewInt(1_000_000_000),
		suite.Chain,
		suite.Chain,
	)

	suite.sender = wallets[0]
	suite.fallbackRecipient = wallets[1]

	suite.destinationDomain = 0

	addr := testutil.AddressBytes()
	suite.mintRecipient = common.BytesToAddress(addr).String()
	addr = testutil.AddressBytes()
	suite.destinationCaller = common.BytesToAddress(addr).String()

	if isIBC {
		wallets := interchaintest.GetAndFundTestUsers(
			t,
			ctx,
			"wallet",
			math.NewInt(1_000_000_000),
			suite.IBC.CounterpartyChain,
		)
		suite.IBC.Account = wallets[0]
	}

	return ctx, suite
}

var DenomMetadataUsdc = banktypes.Metadata{
	Description: "USD Coin",
	DenomUnits: []*banktypes.DenomUnit{
		{
			Denom:    "uusdc",
			Exponent: 0,
			Aliases: []string{
				"microusdc",
			},
		},
		{
			Denom:    "usdc",
			Exponent: 6,
			Aliases:  []string{},
		},
	},
	Base:    "uusdc",
	Display: "usdc",
	Name:    "usdc",
	Symbol:  "USDC",
}

type CircleRoles struct {
	Pauser         ibc.Wallet
	TokenMessenger ibc.Wallet
}

func createOrbiterChainSpec(
	ctx context.Context,
	suite *Suite,
	isZeroFees bool,
) *interchaintest.ChainSpec {
	gasPrices := "1uusdc"
	if isZeroFees {
		gasPrices = "0uusdc"
	}
	return &interchaintest.ChainSpec{
		Name:          "orbiter",
		Version:       "local",
		NumValidators: &numValidators,
		NumFullNodes:  &numFullNodes,
		ChainConfig: ibc.ChainConfig{
			Images: []ibc.DockerImage{
				{
					Repository: "orbiter-simd",
					Version:    "local",
					UIDGID:     "1025:1025",
				},
			},
			Type:           "cosmos",
			Name:           "orbiter",
			ChainID:        "orbiter-1",
			Bin:            "simd",
			Bech32Prefix:   "noble",
			Denom:          "uusdc",
			GasPrices:      gasPrices,
			GasAdjustment:  1.5,
			TrustingPeriod: "504h",
			NoHostMount:    false,
			PreGenesis:     preGenesis(ctx, suite),
			ModifyGenesis:  modifyGenesis(suite),
		},
	}
}

func preGenesis(ctx context.Context, suite *Suite) func(ibc.Chain) error {
	return func(cc ibc.Chain) error {
		val := suite.Chain.Validators[0]

		nobleVal := val.Chain

		fiatTfRoles := CircleRoles{}

		var err error
		fiatTfRoles.Pauser, err = nobleVal.BuildRelayerWallet(ctx, "pauser-ftf")
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}
		if err := val.RecoverKey(ctx, fiatTfRoles.Pauser.KeyName(), fiatTfRoles.Pauser.Mnemonic()); err != nil {
			return fmt.Errorf("failed to restore %s wallet: %w", fiatTfRoles.Pauser.KeyName(), err)
		}

		genesisWallet := ibc.WalletAmount{
			Address: fiatTfRoles.Pauser.FormattedAddress(),
			Denom:   "uusdc",
			Amount:  math.NewIntFromUint64(1_000_000_000),
		}
		err = val.AddGenesisAccount(
			ctx,
			genesisWallet.Address,
			[]sdk.Coin{sdk.NewCoin(genesisWallet.Denom, genesisWallet.Amount)},
		)
		if err != nil {
			return err
		}

		fiatTfRoles.TokenMessenger, err = nobleVal.BuildRelayerWallet(ctx, "token-messenger-ftf")
		if err != nil {
			return fmt.Errorf("failed to create wallet: %w", err)
		}

		suite.CircleRoles = fiatTfRoles

		return nil
	}
}

func modifyGenesis(suite *Suite) func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
	return func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		tokenMessenger := make([]byte, 32)
		copy(tokenMessenger[12:], suite.CircleRoles.Pauser.Address())

		updatedGenesis := []cosmos.GenesisKV{
			cosmos.NewGenesisKV(
				"app_state.bank.denom_metadata",
				[]banktypes.Metadata{DenomMetadataUsdc},
			),

			cosmos.NewGenesisKV("app_state.fiat-tokenfactory", fiattokenfactorytypes.GenesisState{
				Paused: &fiattokenfactorytypes.Paused{Paused: false},
				Pauser: &fiattokenfactorytypes.Pauser{
					Address: suite.CircleRoles.Pauser.FormattedAddress(),
				},
				MintersList: []fiattokenfactorytypes.Minters{
					{
						Address: "noble12l2w4ugfz4m6dd73yysz477jszqnfughxvkss5",
						Allowance: sdk.Coin{
							Denom:  DenomMetadataUsdc.Base,
							Amount: math.NewInt(1_000_000_000),
						},
					},
				},
				MintingDenom: &fiattokenfactorytypes.MintingDenom{Denom: DenomMetadataUsdc.Base},
			}),

			cosmos.NewGenesisKV("app_state.cctp", cctptypes.GenesisState{
				TokenMessengerList: []cctptypes.RemoteTokenMessenger{
					{DomainId: 0, Address: tokenMessenger},
				},
				Pauser: suite.CircleRoles.Pauser.FormattedAddress(),
				BurningAndMintingPaused: &cctptypes.BurningAndMintingPaused{
					Paused: false,
				},
				SendingAndReceivingMessagesPaused: &cctptypes.SendingAndReceivingMessagesPaused{
					Paused: false,
				},
				PerMessageBurnLimitList: []cctptypes.PerMessageBurnLimit{
					{Denom: DenomMetadataUsdc.Base, Amount: math.NewInt(1_000_000_000)},
				},
			}),
		}

		return cosmos.ModifyGenesis(updatedGenesis)(cc, b)
	}
}
