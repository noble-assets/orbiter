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

package service

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noble-assets/orbiter/e2e/gateway/config"
	"github.com/noble-assets/orbiter/e2e/gateway/types"
)

// Service represent an Ethereum service used to interact with the Ethereum execution client.
type Service struct {
	cfg *config.EthereumConfig

	client  EVMClient
	chainID *big.Int

	signer *types.Signer

	// gateway is the instance of the Orbiter Gateway contract for the CCTP protocol.
	gateway *types.Contract[types.OrbiterGatewayCCTP]
	// usdc is the instance of the Fiat Token V2 contract.
	usdc *types.Contract[types.FiatToken]
}

func NewService(ctx context.Context, cfg *config.EthereumConfig) (*Service, error) {
	s := &Service{
		cfg: cfg,
	}

	if err := s.setSigner(cfg.PrivateKey); err != nil {
		return nil, err
	}

	if err := s.setClient(ctx, cfg.RPCEndpoint); err != nil {
		return nil, err
	}

	if err := s.setContracts(cfg.Contracts); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) Close() {
	if s.client != nil {
		s.client.Close()
	}
}

func (s *Service) Signer() *types.Signer {
	return s.signer
}

func (s *Service) GatewayAddress() common.Address {
	return s.gateway.Address()
}

func (s *Service) USDCAddress() common.Address {
	return s.usdc.Address()
}

func (s *Service) BlockTime(ctx context.Context) (uint64, error) {
	block, err := s.client.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}

	return block.Time(), nil
}

func (s Service) TxOpts(ctx context.Context) (*bind.TransactOpts, error) {
	signer := s.Signer()

	txOpts, err := signer.CreateTransactor(s.chainID)
	if err != nil {
		return nil, err
	}

	nonce, err := s.client.PendingNonceAt(ctx, signer.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasPrice, err := s.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	txOpts.Nonce = big.NewInt(int64(nonce))
	txOpts.GasPrice = gasPrice
	txOpts.GasLimit = uint64(300000)

	return txOpts, nil
}

func (s *Service) WaitForTransaction(
	ctx context.Context,
	txHash common.Hash,
) (*gethtypes.Receipt, error) {
	receipt, err := bind.WaitMined(ctx, s.client, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction: %w", err)
	}

	if receipt.Status == 0 {
		return receipt, fmt.Errorf("transaction failed with status 0")
	}

	return receipt, nil
}

func (s *Service) setSigner(privateKeyHex string) error {
	signer, err := types.NewSigner(privateKeyHex)
	if err != nil {
		return err
	}
	s.signer = signer

	return nil
}

func (s *Service) setClient(ctx context.Context, rpcEndpoint string) error {
	client, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve chain ID: %w", err)
	}

	s.client = client
	s.chainID = chainID

	return nil
}

func (s *Service) setContracts(cfg config.ContractsConfig) error {
	gateway, err := types.NewContract(s.client, cfg.Gateway, types.NewOrbiterGatewayCCTP)
	if err != nil {
		return fmt.Errorf("failed to create Orbiter gateway instance: %w", err)
	}

	usdc, err := types.NewContract(s.client, cfg.USDC, types.NewFiatToken)
	if err != nil {
		return fmt.Errorf("failed to create USDC instance: %w", err)
	}

	s.gateway = gateway
	s.usdc = usdc

	return nil
}
