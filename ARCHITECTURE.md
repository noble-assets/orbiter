# Orbiter Architecture

This document provides a detailed overview of the `x/orbiter` module architecture and logic.

## Data Flow

1. **Entrypoint**: An external protocol sends a cross-chain transfer with attached metadata to the
   Orbiter exposed entrypoints.
2. **Payload Parsing**: The entrypoint, which wires the adapter, extracts the payload. Adapter
   validates and parses protocol-specific payload format.
3. **Hooks Execution**: The adapter clears previous balances and validates incoming funds to create
   the expected initial condition for the state transition. Based on the incoming protocol used,
   specific hook can be used.
4. **Payload Processing**: Adapter forwards validated payload to Dispatcher.
5. **Payload Dispatching**: The dispatcher coordinate the dispatch of the orbiter payload content.
6. **Action Handling**: The dispatcher dispatches pre-actions sequentially. For every action, a
   specific action controller is required to execute the business logic.
7. **Orbit Handling**: The dispatcher dispatches the cross-chain forwarding operation. Similarly to
   the action processing, every protocol makes use of a specific controller.
8. **Statistics Update**: The dispatcher records metrics for monitoring.

```mermaid
flowchart TD
    A["CCTP, IBC, Hyperlane"]

    subgraph adaptation
      A1["Entrypoint"]
      B["Adapter Component"]
    end

    subgraph "state transition"
      C["Dispatcher Component"]
      D["Action Component"]
      E["Orbit Component"]
      F["Statistics Storage"]
    end

    Z["Outgoing Transfer<br/>(CCTP, IBC, Hyperlane)"]

    A -->|incoming transfer</br>with payload| A1
    A1 -->|parse payload| B
    A1 -->|transfer hooks| B
    A1 -->|process payload| B
    B -.->|always return| A1

    B -->|dispatch payload| C

    C -->|handle action| D
    C -->|handle orbit| E
    C -->|update stats| F

    E -->|outgoing transfer</br>with payload| Z
```

## Class Diagram

The following diagram provides a high-level overview of the interfaces and concrete structures
defined in the module. From the diagram, it is possible to define 4 main groups:

- The main keeper.
- Keeper's components.
- Routers for the controllers.
- Controllers.

```mermaid
classDiagram
    namespace Components {
        class RouterProvider {
            <<interface>>
            Router()
            SetRouter()
        }

        class Loggable {
            <<interface>>
            Logger()
        }

        class PacketHandler {
            <<interface>>
            HandlePacket()
        }

        class PayloadDispatcher {
            <<interface>>
            DispatchPayload()
        }

        class PayloadAdapter {
            <<interface>>
            ParsePayload()
            BeforeTransferHook()
            AfterTransferHook()
        }

        class OrbitComponentI["OrbitComponent"] {
            <<interface>>
            Pause()
            Unpause()
        }

        class ActionComponentI["ActionComponent"] {
            <<interface>>
            Pause()
            Unpause()
        }

        class DispatcherComponentI["DispatcherComponent"] {
            <<interface>>
        }

        class AdapterComponentI["AdapterComponent"] {
            <<interface>>
        }

        class OrbitComponent["OrbitComponent"] {
            Router
        }

        class ActionComponent["ActionComponent"] {
            Router
        }

        class DispatcherComponent["DispatcherComponent"] {
            Router
        }

        class AdapterComponent["AdapterComponent"] {
            Router
        }
    }

    namespace Orbiter_Keeper {
        class Keeper {
        }
    }

    %% Interface relationships
    Loggable <-- OrbitComponentI : embeds
    RouterProvider <-- OrbitComponentI : embeds
    PacketHandler <-- OrbitComponentI : embeds

    Loggable <-- ActionComponentI : embeds
    RouterProvider <-- ActionComponentI : embeds
    PacketHandler <-- ActionComponentI : embeds

    Loggable <-- DispatcherComponentI : embeds
    PayloadDispatcher <-- DispatcherComponentI : embeds

    Loggable <-- AdapterComponentI : embeds
    PayloadAdapter <-- AdapterComponentI : embeds

    %% Implementation relationships
    ActionComponentI <|.. ActionComponent : implements
    OrbitComponentI <|.. OrbitComponent : implements
    DispatcherComponentI <|.. DispatcherComponent : implements
    AdapterComponentI <|.. AdapterComponent : implements

    %% Keeper relationships
    Keeper *-- ActionComponent
    Keeper *-- OrbitComponent
    Keeper *-- DispatcherComponent
    Keeper *-- AdapterComponent

    namespace Routers {
        class RouterInterface["Router"] {
            <<interface>>
            AddRoute()
            HasRoute()
            Route()
        }

        class RouterType["Router"] {
        }

        class OrbitRouter {
        }

        class ActionRouter {
        }

        class AdapterRouter {
        }
    }

    %% Router relationships
    RouterType <.. ActionComponent
    RouterType <.. OrbitComponent
    RouterType <.. AdapterComponent

    ControllerOrbit <-- OrbitRouter : orchestrate
    ControllerAction <-- ActionRouter : orchestrate
    ControllerAdapter <-- AdapterRouter : orchestrate

    RouterInterface <|.. RouterType : implements
    RouterType <|-- ActionRouter : instance of
    RouterType <|-- OrbitRouter : instance of
    RouterType <|-- AdapterRouter : instance of

    namespace Controllers {
        class Controller {
            <<interface>>
            Id()
            Name()
        }

        class PayloadParser {
            <<interface>>
            ParsePayload()
        }

        class ControllerAdapter {
            <<interface>>
            BeforeTransferHook()
            AfterTransferHook()
        }

        class PacketHandlerC["PacketHandler"] {
            <<interface>>
            HandlePacket()
        }

        class ControllerAction {
            <<interface>>
        }

        class ControllerOrbit {
            <<interface>>
        }

        class IBCAdapter {
        }

        class CCTPAdapter {
        }

        class HyperlaneAdapter {
        }

        class FeeController {
        }

        class SwapController {
        }

        class CCTPController {
        }

        class IBCController {
        }

        class HyperlaneController {
        }
    }

    %% Controller interface relationships
    Controller <-- ControllerAdapter : embeds
    PayloadParser <-- ControllerAdapter : embeds

    Controller <-- ControllerAction : embeds
    PacketHandlerC <-- ControllerAction : embeds

    Controller <-- ControllerOrbit : embeds
    PacketHandlerC <-- ControllerOrbit : embeds

    %% Controller implementations
    ControllerAdapter <|.. IBCAdapter : implements
    ControllerAdapter <|.. CCTPAdapter : implements
    ControllerAdapter <|.. HyperlaneAdapter : implements

    ControllerAction <|.. FeeController : implements
    ControllerAction <|.. SwapController : implements

    ControllerOrbit <|.. CCTPController : implements
    ControllerOrbit <|.. IBCController : implements
    ControllerOrbit <|.. HyperlaneController : implements
```

## Keeper

The Orbiter design follows a components-based approach. As with any standard Cosmos SDK module,
there is a central keeper that controls access to the underlying module state, both for read and
write operations. The keeper manages the state and business logic by splitting responsibilities
across components. Each component is responsible for a single functionality, and all together they
allow forwarding cross-chain funds with pre-transfer custom state transitions.

## Components

Components are used to allow the Orbiter keeper to perform the three fundamental operations:

1. Adapt the bridge protocol by creating a unique internal request type.
2. Execute actions on the Noble core with the received funds.
3. Forward the funds resulting from the internal actions to the destination.

```mermaid
classDiagram

    namespace Components {
    class RouterProvider {
        <<interface>>
        Router()
        SetRouter()
    }

    class Loggable {
        <<interface>>
        Logger()
    }

    class PacketHandler {
        <<interface>>
        HandlePacket()
    }

    class PayloadDispatcher {
        <<interface>>
        DispatchPayload()
    }

    class PayloadAdapter {
        <<interface>>
        ParsePayload()
        BeforeTransferHook()
        AfterTransferHook()
        ProcessPayload()
    }

    class ComponentOrbit["ComponentOrbit"] {
        <<interface>>
        Pause()
        Unpause()
    }

    class ComponentAction["ComponentAction"] {
        <<interface>>
        Pause()
        Unpause()
    }

    class ComponentDispatcher["ComponentDispatcher"] {
        <<interface>>
    }

    class ComponentAdapter["ComponentAdapter"] {
        <<interface>>
    }

    class OrbitComponent["OrbitComponent"] {
    }

    class ActionComponent["ActionComponent"] {
    }

    class DispatcherComponent["DispatcherComponent"] {
    }

    class AdapterComponent["AdapterComponent"] {
    }
    }
    namespace Orbiter Keeper{
    class Keeper {
    }
    }


    Loggable <-- ComponentOrbit : embeds
    RouterProvider <-- ComponentOrbit : embeds
    PacketHandler <-- ComponentOrbit : embeds

    Loggable <-- ComponentAction : embeds
    RouterProvider <-- ComponentAction : embeds
    PacketHandler <-- ComponentAction : embeds

    Loggable <-- ComponentDispatcher : embeds
    PayloadDispatcher <-- ComponentDispatcher : embeds

    Loggable <-- ComponentAdapter : embeds
    PayloadAdapter <-- ComponentAdapter : embeds

    ComponentAction <|.. ActionComponent : implements
    ComponentOrbit <|.. OrbitComponent : implements
    ComponentDispatcher <|.. DispatcherComponent : implements
    ComponentAdapter <|.. AdapterComponent : implements

    Keeper *-- ActionComponent
    Keeper *-- OrbitComponent
    Keeper *-- DispatcherComponent
    Keeper *-- AdapterComponent
```

### Adapter Component

The `AdapterComponent` (`keeper/components/adapter.go`) serves as the interface between external
cross-chain communication protocols and the internal handling of the orbiter packets. The role of
this component is to create the expected orbiter payload out of the cross-chain metadata received.

This component does not directly adapt the incoming metadata, but keeps track internally of the
available adapter controllers and routes the incoming metadata to the correct one.

**Key Responsibilities**:

- **Payload Parsing**: Validates and parses incoming cross-chain payloads. This phase is required to
  convert cross-chain metadata formatted into different standards based on the bridge, into an
  internal payload type.
- **Adapter Controllers Routing**: Routes to the correct adapter the incoming data.
- **Transfer Hooks**: Executes pre/post transfer logic. In this phase, the adapter creates and
  verifies the initial conditions to execute an Orbiter state transition.
- **Protocol Routing**: Routes operations defined in the payload to the proper forwarding (orbit) or
  action handler. // TODO

### Dispatcher Component

The `DispatcherComponent` (`keeper/components/dispatcher.go`) orchestrates payload execution by
coordinating actions and orbit operations. This component is created by injecting the action and
orbit component.

**Key Responsibilities**:

- **Payload Validation**: Ensures payload structure and content validity
- **Action Dispatching**: Dispatches pre-actions sequentially (fees, swaps, etc.) to the proper
  handler.
- **Orbit Execution**: Dispatches cross-chain forwarding operations to the proper handler.
- **Statistics Tracking**: Maintains dispatch counts and amount metrics.

### Action Component

The `ActionComponent` (`keeper/components/action.go`) handles action operations by performing state
transitions on the Noble chain.

This component does not execute any actions, but keeps track internally of the available action
controllers and routes the incoming request to the correct one.

**Key Responsibilities**:

- **Packet Handling**: Handles an incoming action packet.
- **Action Packet Validation**: Validates if an action packet is valid and can be executed.
- **Action Controllers Routing**: Stores and routes the incoming action request to the proper
  controller.

### Orbit Component

The `OrbitComponent` (`keeper/components/orbit.go`) handles the outgoing cross-chain transfer by
forwarding the orbiter balance to the destination. This module operates on the resulting denom and
amount of all the actions executions.

This component does not execute any cross-chain transfers, but keeps track internally of the
available orbit controllers and routes the incoming request to the correct one.

**Key Responsibilities**:

- **Packet Handling**: Handles an incoming orbit packet.
- **Orbit Packet Validation**: Validates if an orbit packet is valid and can be executed.
- **Orbit Controllers Routing**: Stores and routes the incoming orbit request to the proper
  controller.

## Controllers

To provide loose coupling between the actions and the supported bridges within the orbiter keeper, a
controller pattern has been implemented. Using controllers, the specific logic associated with an
action or a bridge protocol is not implemented directly into the associated component. The specific
logic is implemented through controllers that are injected into the components during app
initialization. This way, components are responsible for executing only high-level logic that is
independent of specific requests, and for routing the low-level execution to the associated
controller.

```mermaid
classDiagram
    class Controller {
        <<interface>>
        Id()
        Name()
    }

    class PayloadParser {
        <<interface>>
        ParsePayload()
    }

    class ControllerAdapter {
        <<interface>>
        BeforeTransferHook()
        AfterTransferHook()
    }

    Controller <-- ControllerAdapter : embeds
    PayloadParser <-- ControllerAdapter : embeds

    ControllerAdapter <|.. IBCAdapter : implements
    ControllerAdapter <|.. CCTPAdapter : implements
    ControllerAdapter <|.. HyperlaneAdapter : implements

    class PacketHandler {
        <<interface>>
        HandlePacket()
    }

    class ControllerAction {
        <<interface>>
    }

    class ControllerOrbit {
        <<interface>>
    }

    Controller <-- ControllerAction : embeds
    PacketHandler <-- ControllerAction : embeds

    Controller <-- ControllerOrbit : embeds
    PacketHandler <-- ControllerOrbit : embeds

    ControllerAction <|.. FeeController : implements
    ControllerAction <|.. SwapController : implements


    ControllerOrbit <|.. CCTPController : implements
    ControllerOrbit <|.. IBCController : implements
    ControllerOrbit <|.. HyperlaneController : implements
```

## Router

The router is a custom type that facilitates the in-memory storage of all orbiter controllers and
their invocation. Components that require the coordination of controllers embed the generic router
type and expose methods to set controllers. When a specific controller is needed, it is requested
from the router and then its public methods can be called.

```mermaid
classDiagram
    class ControllerAdapter {
        <<interface>>
    }
    class ControllerAction {
        <<interface>>
    }
    class ControllerOrbit {
        <<interface>>
    }

    class RouterInterface["Router"] {
        <<interface>>
        AddRoute()
        HasRoute()
        Route()
    }

    class RouterType["Router"] {
    }

    class OrbitRouter {
    }
    class ActionRouter {
    }
    class AdapterRouter {
    }

    %% Relationships
    ControllerOrbit <-- OrbitRouter : orchestrate
    ControllerAction <-- ActionRouter : orchestrate
    ControllerAdapter <-- AdapterRouter : orchestrate

    RouterInterface <|.. RouterType : implements
    RouterType <|-- ActionRouter : instance of
    RouterType <|-- OrbitRouter : instance of
    RouterType <|-- AdapterRouter : instance of
```
