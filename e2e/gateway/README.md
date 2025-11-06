# Gateway

This folder contains the code used to test the Orbiter CCTP gateway as an external user would do.

## Example

The example folder contains a `main` function that can be run to execute a deposit for burn with
Orbiter as destination. To test the user flow, you must first have a running Ethereum client and:

- A signer with both `ETH` and `USDC`.
- The Gateway contract deployed.
- The Circle CCTP smart contracts deployed.

In the following, we will describe how to test the workflow using a Sepolia testnet fork. Testnet
tokens can be found at:

- [ETH Faucet](https://sepolia-faucet.pk910.de/)
- [USDC Faucet](https://faucet.circle.com/)

The following commands are meant to be executed from the root of the project. The recommended way to
setup the example environment is via [`anvil`](https://getfoundry.sh/anvil/overview):

```sh
anvil --fork-url https://ethereum-sepolia-rpc.publicnode.com -vvvv
```

Once the node is running via anvil, we can deploy the Gateway contract:

```sh
forge script ./contracts/script/OrbiterGatewayCCTP.s.sol:OrbiterGatewayCCTPScript_testnet --root ./contracts/ --fork-url http://localhost:8545  --broadcast --interactives 1 -vvvv
```

It will be asked you o insert the private key for the deployment, and then the gateway will be ready
to be used. At this point, we can execute the example:

```sh
go run ./examples/deposit_for_burn_orbiter.go
```

The template configuration file for the Sepolia testnet and anvil is
`config_sepolia_local.template.toml`. What you have to do is to copy the file, rename it as
`config_sepolia_local.toml`, and add the private key of the signer into it.
