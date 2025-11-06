# Contracts

Smart contracts used to extend cross-chain communication protocols to support the Noble's Orbiter
module.

For more information regarding the gateways, please refer to the [`docs`](../../docs/gateway.md).

## Requirements

To use this repository, you need the following software installed:

- [`bun`](https://bun.sh)
- [`foundry`](https://getfoundry.sh)

## Build

```sh
bun install
forge build
```

## Tests

```shell
forge test
```

## Format

```shell
forge fmt
```

## Deploy

It is possible to deploy the CCTP Gateway contract using the Solidity scripts contained in the
`./scripts/` folder. To deploy the gateway on a testnet fork running locally with anvil, you should
first start the Ethereum node:

```sh
anvil --fork-url https://ethereum-sepolia-rpc.publicnode.com -vvvv
```

Then, deploy the contract with the testnet dependencies:

```sh
forge script ./contracts/script/OrbiterGatewayCCTP.s.sol:OrbiterGatewayCCTPScript_testnet \
--fork-url http://localhost:8545 \
--broadcast --interactives 1 -vvvv
```

You will be asked to insert the private key to sign the transaction, and then the gateway will be
deployed.
