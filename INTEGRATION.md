# Integration

## Introduction

This document describes how to integrate with the Orbiter module to leverage Noble cross-chain
functionalities.

## Definitions

- **Forwarding via CCTP**: A packet flow incoming from a bridge protocol and leaving Noble via CCTP
  through the Orbiter, is called an **AutoCCTP** flow.

- **Forwarding via IBC**: A packet flow incoming from a bridge protocol and leaving Noble via IBC
  through the Orbiter, is called an **AutoIBC** flow. (TODO)

- **Forwarding via Hyperlane**: A packet flow incoming from a bridge protocol and leaving Noble via
  Hyperlane through the Orbiter, is called an **AutoLane** flow. (TODO)

## Payload

The Orbiter module is a payload-based module, which is capable of parsing cross-chain metadata from
different sources, executing state transitions, and then forwarding funds to a destination
counterparty.

The Orbiter payload is defined as a [proto3](https://protobuf.dev/) message in the `./proto` folder.
This type is composed of two parts:

- A list of actions.
- An orbit.

An example of a JSON structured payload is:

```json
{
  "orbiter": {
    "pre_actions": [
      {
        "id": "ACTION_FEE",
        "attributes": {
          "@type": "/noble.orbiter.controllers.actions.v1.FeeAttributes",
          "fees_info": [
            {
              "recipient": "noble1shrlcs09fl2gghvystkfemewgzkccpyvudch7y",
              "basis_points": 100
            }
          ]
        }
      }
    ],
    "orbit": {
      "protocol_id": "PROTOCOL_CCTP",
      "attributes": {
        "@type": "/noble.orbiter.controllers.orbits.v1.CCTPAttributes",
        // Note: mint_recipient and destination_caller are 32-byte values encoded as base64
        "destination_domain": 0,
        "mint_recipient": "PNWAxASH2RPmgMV+/Tb4e78ON1WL8SoFGnwbWWHxfuA=",
        "destination_caller": "xWtN0TuqjWo90XiknI61JUxYexN2JgZaEaWGxhA/rXE="
      },
      "passthrough_payload": ""
    }
  }
}
```

### Actions

Actions are defined as state transition requests which can be executed on Noble. An action to be
valid must satisfy the `ActionAttributes` defined in the Orbiter `types` package. An action is
defined by two fields:

1. An identifier.
2. An attributes field containing all the relevant information to satisfy the request.

Actions are executed in the order in which they are passed into the field `pre_actions`, and each of
them operates on the state resulting from the execution of the previous one. For example, if the pre
actions field is composed of a **swap** and a **fee** payment actions, the fee payment will be
applied based on the amount and denomination of the coin resulting from the swap.

The field is called `pre_actions` because all the requests specified are executed before performing
the forwarding via the orbit specified. If any of them fails, the entire state transition defined by
the execution of the Cosmos SDK transaction will be reverted.

Actions supported by the Orbiter module are:

<div align="center">

| Action      | Status |
| ----------- | ------ |
| Fee payment | done   |
| Swap        | todo   |

</div>

They are identified by a unique ID:

```go
type ActionID int32

const (
 ACTION_UNSUPPORTED ActionID = 0
 ACTION_FEE ActionID = 1
 ACTION_SWAP ActionID = 2
)
```

### Orbit

An orbit is a forwarding operation defined as a combination of:

1. A bridge protocol.
2. A counterparty chain. (defined in the attributes but required to define an orbit ID)
3. Attributes required to operate via the bridge protocol.
4. A pass-through payload to forward with the outgoing transfer protocol. (TODO)

Based on the bridge protocol, the following conditions may result from the execution of an orbit:

- Synchronous protocols: Protocols like CCTP and Hyperlane follow a commit-and-forget style. In this
  case, once the associated server informs the Orbiter that the bridge request has been stored to
  state, the Orbiter execution is complete. Conversely, if the server returns an error, the entire
  tx will be marked as unsuccessful.

- Asynchronous protocol: Protocols like IBC are asynchronous in nature, and for this reason require
  book-keeping of in-flight packets in the module. (TODO how to handle them)

Orbits supported by the Orbiter module are:

<div align="center">

| Action    | Status |
| --------- | ------ |
| CCTP      | done   |
| IBC       | todo   |
| Hyperlane | todo   |

</div>

They are identified by a unique protocol ID:

```go
type ProtocolID int32

const (
 PROTOCOL_IBC ProtocolID = 1
 PROTOCOL_CCTP ProtocolID = 2
 PROTOCOL_HYPERLANE ProtocolID = 3
)
```

## Payload Creation

### IBC

This section describes how to create a valid IBC payload in Golang for the Orbiter module:

1. Import the required packages from the Orbiter repo:

```go
 "orbiter.dev/types"
 "orbiter.dev/types/controllers/actions"
 "orbiter.dev/types/controllers/orbits"
 "orbiter.dev/testutil"
```

2. Define the orbit attributes:

```go
 destinationDomain := uint32(0)
 mintRecipient := testutil.RandomBytes(32)
 destinationCaller := testutil.RandomBytes(32)
 passthroughPayload := []byte("")
```

3. Create an orbit via a factory function:

```go
 orbit, err := orbits.NewCCTPOrbit(
  destinationDomain,
  mintRecipient,
  destinationCaller,
  passthroughPayload,
 )
```

4. If an action is required, start defining the attributes:

```go
 feeAttr := actions.FeeAttributes{
  FeesInfo: []*actions.FeeInfo{
   {
    Recipient:   feeRecipientAddr,
    BasisPoints: 100,
   },
  },
 }
```

5. Define the action and set the attributes:

```go
 action := types.Action{
  Id: types.ACTION_FEE,
 }
 err = action.SetAttributes(&feeAttr)
```

6. Create a wrapped payload:

```go
 payload, err := types.NewPayloadWrapper(orbit, []*types.Action{&action})
```

7. Marshal the payload structure into JSON using the codec with registered interfaces:

```go
 encCfg := testutil.MakeTestEncodingConfig("noble")
 orbiter.RegisterInterfaces(encCfg.InterfaceRegistry)
 payloadBz, err := types.MarshalJSON(encCfg.Codec, payload)
 payloadStr := string(payloadBz)
```

8. The payload is now ready to be added in the ICS20 memo field.

A working example for the payload creation can be found in the file `e2e/ibc_to_cctp_test.go`.
