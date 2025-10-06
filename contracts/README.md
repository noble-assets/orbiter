# Orbiter Entrypoint Contracts

This directory contains the entrypoint contracts for interacting with the Orbiter system
through smart contracts on an EVM chain.

## Requirements

To use this repository, you need the following software installed:

- [`bun`](https://bun.sh)
- [`foundry`](https://getfoundry.sh)

## Dependencies

To download the required dependencies, run:

```sh
make deps
```

## Compile contracts

To compile the contracts, run:

```sh
make compile
```

## Things To Note

The `NobleDollar` contract was taken from this commit on the Dollar repository:
https://github.com/noble-assets/dollar/tree/b84f0bb4a0c8058e073c20b523513920b7870b0b

To have some initial supply of tokens available for testing,
I have changed the `patched` dependencies to still allow
setting a `_totalSupply` that gets minted to the owner.
