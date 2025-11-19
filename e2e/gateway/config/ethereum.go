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
