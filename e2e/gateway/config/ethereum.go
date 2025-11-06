package config

import (
	"fmt"
)

// EthereumConfig holds Ethereum-specific configuration.
type EthereumConfig struct {
	RPCEndpoint string `toml:"rpc_endpoint"`
	// PrivateKey used to send transactions on the EVM chain.
	PrivateKey string          `toml:"private_key"`
	Contracts  ContractsConfig `toml:"contracts"`
}

func (c *EthereumConfig) Validate() error {
	if c.RPCEndpoint == "" {
		return fmt.Errorf("ethereum.rpc_endpoint is required")
	}
	if c.PrivateKey == "" {
		return fmt.Errorf("ethereum.private_key is required")
	}

	if err := c.Contracts.Validate(); err != nil {
		return fmt.Errorf("ethereum.contracts configuration error: %w", err)
	}

	return nil
}

// ContractsConfig defines the addresses of the smart contracts deployed on the chain.
type ContractsConfig struct {
	USDC           string `toml:"usdc"`
	TokenMessenger string `toml:"token_messenger"`
	Gateway        string `toml:"gateway"`
}

func (c *ContractsConfig) Validate() error {
	if c.USDC == "" {
		return fmt.Errorf("contracts.usdc is required")
	}
	if c.Gateway == "" {
		return fmt.Errorf("contracts.gateway is required")
	}

	return nil
}
