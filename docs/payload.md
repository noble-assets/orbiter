# Payload

The Orbiter expected metadata is defined in a [proto3](https://protobuf.dev/) message named
[`Payload`](https://github.com/noble-assets/orbiter/blob/main/proto/noble/orbiter/core/v1/orbiter.proto#L57-L67).
This message consists of two main components:

- A list of actions.
- A forwarding.

The payload is defined as a proto message to guarantee efficiency in the data transmission and
security during its unmarshaling/decoding phase.

Once created, the metadata has to be encoded in the format expected by the initiating bridge
protocol, and sent along with the cross-chain transaction.

An example of a JSON-structured string payload, used in the IBC protocol, is:

```json
{
  "orbiter": {
    "pre_actions": [
      {
        "id": "ACTION_FEE",
        "attributes": {
          "@type": "/noble.orbiter.controller.action.v1.FeeAttributes",
          "fees_info": [
            {
              "recipient": "noble1shrlcs09fl2gghvystkfemewgzkccpyvudch7y",
              "basis_points": 100
            }
          ]
        }
      }
    ],
    "forwarding": {
      "protocol_id": "PROTOCOL_CCTP",
      "attributes": {
        "@type": "/noble.orbiter.controller.forwarding.v1.CCTPAttributes",
        "destination_domain": 0,
        "mint_recipient": "PNWAxASH2RPmgMV+/Tb4e78ON1WL8SoFGnwbWWHxfuA=",
        "destination_caller": "xWtN0TuqjWo90XiknI61JUxYexN2JgZaEaWGxhA/rXE="
      },
      "passthrough_payload": ""
    }
  }
}
```

## Actions

An
[`Action`](https://github.com/noble-assets/orbiter/blob/main/proto/noble/orbiter/core/v1/orbiter.proto#L12-L33)
is a state transition request which can be executed on Noble and is defined by two fields:

1. An action
   [identifier](https://github.com/noble-assets/orbiter/blob/main/proto/noble/orbiter/core/v1/id.proto#L9-L24).
2. An attributes field containing all the information to execute the action.

The attributes field is defined as an `Any` type implementing the
[`ActionAttributes`](https://github.com/noble-assets/orbiter/blob/main/types/core/attributes.go#L27-L31)

Actions are executed in the order in which they are passed into the field `pre_actions`, and each of
them operates on the state resulting from the execution of the previous one. For example, if the pre
actions field is composed of a **swap** and a **fee** payment, the fee payment will be applied on
the amount and denomination of the coin resulting from the swap.

If any of the specified action fails, the entire state transition defined by the execution of the
transaction will be reverted.

### Fee

Is it possible to specify fee payments as a single action by using the
[`FeeAttributes`](https://github.com/noble-assets/orbiter/blob/main/types/controller/action/fee.pb.go#L26-L31).
A single fee is defined by:

- `Recipient`: A Noble address which will receive the fee.
- `BasisPoints (BPS)`: A number between 0 and 10000 which defines the percentage of the transferred
  amount that has to be paid as a fee. The fee amount will be defined as
  $fee = amount \cdot \frac{BPS}{10000}$

## Forwarding

A
[`Forwarding`](https://github.com/noble-assets/orbiter/blob/main/proto/noble/orbiter/core/v1/orbiter.proto#L35-L55)
is an operation defined as a combination of:

1. A bridge protocol
   [identifier](https://github.com/noble-assets/orbiter/blob/main/proto/noble/orbiter/core/v1/id.proto#L26-L44).
2. Attributes required to operate via the bridge protocol.
3. A pass-through payload to forward with the outgoing transfer protocol.

The attributes field is defined as an `Any` type implementing the
[`ForwardingAttributes`](https://github.com/noble-assets/orbiter/blob/main/types/core/attributes.go#L33-L39)

Based on the bridge protocol, the following conditions may result from the execution of an orbit:

- Synchronous protocols: Protocols like CCTP and Hyperlane follow a commit-and-forget style. In this
  case, once the associated server informs the Orbiter that the bridge request has been stored to
  state, the Orbiter execution is complete. If the server returns an error, the entire tx will be
  marked as unsuccessful.

- Asynchronous protocol: Protocols like IBC are asynchronous in nature, and for this reason
  book-keeping of in-flight packets is required in the module.

### CCTP

The CCTP information required to perform a CCTP forwarding are defined in the
[`CCTPAttributes`](https://github.com/noble-assets/orbiter/blob/main/proto/noble/orbiter/controller/forwarding/v1/cctp.proto#L9-L26)

The mint recipient and the destination caller are 32 bytes addresses. Notice that the denom is not
specified in the attributes. This is because the denom used is the same sent to the Orbiter module,
or the result of the actions specified. The denom that the module can forward via CCTP are those
supported by the Fiat Token Factory module.
