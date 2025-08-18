# Commands

This document provides CLI commands to execute transactions and queries against the Orbiter module
in the simapp.

## Env

Define the aliases:

```sh
SIMD="./simapp/build/simd"
HOME_DIR=".orbiter"
CHAIN_ID="orbiter-1"
KEYRING_BACKEND="test"
```

## Adapter

```sh
$SIMD tx orbiter adapter update-params '{"max_passthrough_payload_size": 1000}' --from authority --home $HOME_DIR --keyring-backend $KEYRING_BACKEND --chain-id "$CHAIN_ID"
```

```sh
$SIMD q orbiter adapter params
```
