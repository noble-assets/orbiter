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
	"strconv"
	"testing"

	hyperlanepostdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	fiattokenfactorytypes "github.com/circlefin/noble-fiattokenfactory/x/fiattokenfactory/types"
	interchaintest "github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/strangelove-ventures/interchaintest/v8/testreporter"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/noble-assets/orbiter/testutil"
	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
)

const (
	usdcDenom  = "usdc"
	uusdcDenom = "uusdc"
)

var (
	numValidators = 1
	numFullNodes  = 0
	burnLimit     = math.NewInt(1_000_000_000)
)

type IBC struct {
	CounterpartyChain  *cosmos.CosmosChain
	RelayerReporter    *testreporter.RelayerExecReporter
	Relayer            ibc.Relayer
	CounterpartySender ibc.Wallet
	PathName           string
}

type Suite struct {
	Chain *cosmos.CosmosChain

	IBC *IBC

	// -----------------------
	// CCTP fields
	CircleRoles       CircleRoles
	sender            ibc.Wallet
	fallbackRecipient ibc.Wallet
	mintRecipient     []byte
	destinationCaller []byte

	destinationDomain uint32

	// -----------------------
	// Hyperlane fields
	hyperlaneToken             *warptypes.WrappedHypToken
	hyperlaneHook              *hyperlanepostdispatchtypes.NoopHook
	hyperlaneDestinationDomain uint32
}

func NewSuite(t *testing.T, isZeroFees bool, isIBC, isHyperlane bool) (context.Context, Suite) {
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
		err := interchain.Close()
		if err != nil {
			t.Logf("failed to close interchain: %v", err)
		}
	})

	wallets := interchaintest.GetAndFundTestUsers(
		t,
		ctx,
		"wallet",
		math.NewInt(1_000_000_000),
		suite.Chain,
		suite.Chain,
		suite.Chain,
	)

	suite.sender = wallets[0]
	suite.fallbackRecipient = wallets[1]
	hyperlaneWallet := wallets[2]

	suite.destinationDomain = 0

	addr := testutil.AddressBytes()
	bz, err := LeftPadBytes(addr)
	require.NoError(t, err, "expected no error padding mint recipient address")
	suite.mintRecipient = bz

	addr = testutil.AddressBytes()
	bz, err = LeftPadBytes(addr)
	require.NoError(t, err, "expected no error padding destination caller address")
	suite.destinationCaller = bz

	if isIBC {
		wallets := interchaintest.GetAndFundTestUsers(
			t,
			ctx,
			"wallet",
			math.NewInt(1_000_000_000),
			suite.IBC.CounterpartyChain,
		)
		suite.IBC.CounterpartySender = wallets[0]
	}

	if isHyperlane {
		// Create the ISM -- for testing purposes it's enough to use the No-Op ISM
		node := suite.Chain.GetNode()
		hyperlaneKey := hyperlaneWallet.KeyName()

		_, err = node.ExecTx(ctx, hyperlaneKey, "hyperlane", "ism", "create-noop")
		require.NoError(t, err, "failed to create noop ism")
		ism, err := getHyperlaneNoOpISM(ctx, node)
		require.NoError(t, err, "unexpected result getting hyperlane ISM")
		require.NotNil(t, ism, "expected hyperlane ISM to be in state")

		_, err = node.ExecTx(ctx, hyperlaneKey, "hyperlane", "hooks", "noop", "create")
		require.NoError(t, err, "failed to create hyperlane hook")
		hook, err := getHyperlaneNoOpHook(ctx, node)
		require.NoError(t, err, "failed to get hyperlane hook")

		suite.hyperlaneHook = hook

		_, err = node.ExecTx(
			ctx,
			hyperlaneKey,
			"hyperlane",
			"mailbox",
			"create",
			ism.Id.String(),
			strconv.FormatInt(forwardingtypes.HypNobleMainnetDomain, 10),
		)
		require.NoError(t, err, "failed to create mailbox")
		mailbox, err := getHyperlaneMailbox(ctx, node)
		require.NoError(t, err, "failed to get hyperlane mailbox")

		_, err = node.ExecTx(
			ctx,
			hyperlaneKey,
			"hyperlane",
			"mailbox",
			"set",
			mailbox.Id.String(),
			"--default-ism",
			ism.Id.String(),
			"--required-hook",
			hook.Id.String(),
			"--default-hook",
			hook.Id.String(),
		)
		require.NoError(t, err, "failed to create set mailbox")

		_, err = node.ExecTx(
			ctx,
			hyperlaneKey,
			"hyperlane-transfer",
			"create-collateral-token",
			mailbox.Id.String(),
			uusdcDenom,
		)
		require.NoError(t, err, "failed to create collateral token")
		collateralToken, err := getHyperlaneCollateralToken(ctx, node)
		require.NoError(t, err, "failed to get hyperlane collateral token")

		suite.hyperlaneToken = collateralToken

		suite.hyperlaneDestinationDomain = 1
		receiverDomain := strconv.Itoa(int(suite.hyperlaneDestinationDomain))
		receiverContract := "0x0000000000000000000000000000000000000000000000000000000000000000"
		gasAmount := "0"
		_, err = node.ExecTx(
			ctx,
			hyperlaneKey,
			"hyperlane-transfer",
			"enroll-remote-router",
			collateralToken.Id,
			receiverDomain,
			receiverContract,
			gasAmount,
		)
		require.NoError(t, err, "failed to create enroll remote router for token")
	}

	return ctx, suite
}

var DenomMetadataUsdc = banktypes.Metadata{
	Description: "USD Coin",
	DenomUnits: []*banktypes.DenomUnit{
		{
			Denom:    uusdcDenom,
			Exponent: 0,
			Aliases: []string{
				"microusdc",
			},
		},
		{
			Denom:    usdcDenom,
			Exponent: 6,
			Aliases:  []string{},
		},
	},
	Base:    uusdcDenom,
	Display: usdcDenom,
	Name:    usdcDenom,
	Symbol:  usdcDenom,
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
			Type:                "cosmos",
			Name:                "orbiter",
			ChainID:             "orbiter-1",
			AdditionalStartArgs: []string{"--log_level", "*:info,orbiter:trace"},
			Bin:                 "simd",
			Bech32Prefix:        "noble",
			Denom:               uusdcDenom,
			GasPrices:           gasPrices,
			GasAdjustment:       1.5,
			TrustingPeriod:      "504h",
			NoHostMount:         false,
			PreGenesis:          preGenesis(ctx, suite),
			ModifyGenesis:       modifyGenesis(suite),
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
			return errorsmod.Wrap(err, "failed to create wallet")
		}
		if err := val.RecoverKey(ctx, fiatTfRoles.Pauser.KeyName(), fiatTfRoles.Pauser.Mnemonic()); err != nil {
			return errorsmod.Wrapf(err, "failed to restore %s wallet", fiatTfRoles.Pauser.KeyName())
		}

		genesisWallet := ibc.WalletAmount{
			Address: fiatTfRoles.Pauser.FormattedAddress(),
			Denom:   uusdcDenom,
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
			return errorsmod.Wrap(err, "failed to create wallet")
		}

		suite.CircleRoles = fiatTfRoles

		return nil
	}
}

func modifyGenesis(suite *Suite) func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
	return func(cc ibc.ChainConfig, b []byte) ([]byte, error) {
		tokenMessenger := make([]byte, 32)
		copy(tokenMessenger[12:], suite.CircleRoles.TokenMessenger.Address())

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
					{Denom: DenomMetadataUsdc.Base, Amount: burnLimit},
				},
			}),
		}

		return cosmos.ModifyGenesis(updatedGenesis)(cc, b)
	}
}
