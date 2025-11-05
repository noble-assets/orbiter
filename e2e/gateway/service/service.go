package service

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noble-assets/orbiter/e2e/gateway/config"
	"github.com/noble-assets/orbiter/e2e/gateway/types"
)

// Service represent an Ethereum service used to interact with the Ethereum execution client.
type Service struct {
	cfg *config.EthereumConfig

	client  *ethclient.Client
	chainID *big.Int

	privateKey *ecdsa.PrivateKey
	address    common.Address

	// gateway is the instance of the Orbiter Gateway contract for the CCTP protocol.
	gateway *types.OrbiterGatewayCCTP
	// usdc is the instance of the Fiat Token V2 contract.
	usdc *types.FiatToken
}

func NewService(cfg *config.EthereumConfig) (*Service, error) {
	s := &Service{
		cfg: cfg,
	}

	if err := s.setSigner(cfg.PrivateKey); err != nil {
		return nil, err
	}

	if err := s.setClient(cfg.RPCEndpoint); err != nil {
		return nil, err
	}

	if err := s.setContracts(cfg.Contracts); err != nil {
		return nil, err
	}

	return s, nil
}

func (c *Service) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

func (s *Service) setSigner(privateKeyHex string) error {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("failed to cast public key to ECDSA public key")
	}

	s.privateKey = privateKey
	s.address = crypto.PubkeyToAddress(*publicKeyECDSA)

	return nil
}

func (s *Service) setClient(rpcEndpoint string) error {
	client, err := ethclient.Dial(rpcEndpoint)
	if err != nil {
		return fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return fmt.Errorf("failed to retrieve chain ID: %w", err)
	}

	s.client = client
	s.chainID = chainID

	return nil
}

func (s *Service) setContracts(cfg config.ContractsConfig) error {
	gatewayAddr := common.HexToAddress(cfg.Gateway)
	gateway, err := types.NewOrbiterGatewayCCTP(gatewayAddr, s.client)
	if err != nil {
		return fmt.Errorf("failed to create Orbiter gateway instance: %w", err)
	}

	usdcAddr := common.HexToAddress(cfg.USDC)
	usdc, err := types.NewFiatToken(usdcAddr, s.client)
	if err != nil {
		return fmt.Errorf("failed to create usdc token instance: %w", err)
	}

	s.gateway = gateway
	s.usdc = usdc

	return nil
}
