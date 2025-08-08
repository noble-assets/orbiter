# Types Package Dependencies Graph

This document visualizes the internal dependencies between packages in the `types/` directory of the
Orbiter module.

## Dependency Graph

```mermaid
graph TD
    %% Core Foundation - No internal dependencies
    Core[types/core/]

    %% Level 1 - Direct dependencies on core
    DispatcherComponent[types/component/dispatcher/]
    ActionController[types/controller/action/]

    %% Level 2 - Dependencies on core + root types
    ForwardingController[types/controller/forwarding/]

    %% Level 3 - Interface definitions and router
    Interfaces[types/interfaces/]
    Router[types/router/]

    %% Level 4 - Package aggregators
    Component[types/component/]
    Controller[types/controller/]

    %% Level 5 - Root types package
    Types[types/]

    %% Sub-packages (leaf nodes)
    AdapterComponent[types/component/adapter/]
    ExecutorComponent[types/component/executor/]
    ForwarderComponent[types/component/forwarder/]

    %% Dependencies
    DispatcherComponent --> Core
    ActionController --> Core
    ForwardingController --> Core
    ForwardingController --> Types

    Interfaces --> Core
    Interfaces --> Types
    Router --> Interfaces

    Component --> AdapterComponent
    Component --> ExecutorComponent
    Component --> ForwarderComponent
    Controller --> ActionController
    Controller --> ForwardingController

    Types --> Core
    Types --> Component
    Types --> Controller

    %% Styling
    classDef coreStyle fill:#e1f5fe,stroke:#01579b,stroke-width:3px
    classDef componentStyle fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef controllerStyle fill:#e8f5e8,stroke:#1b5e20,stroke-width:2px
    classDef interfaceStyle fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef rootStyle fill:#ffebee,stroke:#b71c1c,stroke-width:3px

    class Core coreStyle
    class Component,DispatcherComponent,AdapterComponent,ExecutorComponent,ForwarderComponent componentStyle
    class Controller,ActionController,ForwardingController controllerStyle
    class Interfaces,Router interfaceStyle
    class Types rootStyle
```

## Package Descriptions

### Core Package (`types/core/`)

- **Purpose**: Foundational types and structures
- **Files**: `id.go`, `orbiter.go`, `attributes.go`, `keys.go`, `errors.go`
- **Dependencies**: None (foundation layer)
- **Key Types**: Core ID types, Orbiter structures, error definitions

### Component Packages

- **`types/component/`**: Aggregates all component sub-packages
- **`types/component/dispatcher/`**: Handles payload processing and execution
- **`types/component/adapter/`**: Interfaces with external protocols
- **`types/component/executor/`**: Transaction execution logic
- **`types/component/forwarder/`**: Message forwarding functionality

### Controller Packages

- **`types/controller/`**: Aggregates all controller sub-packages
- **`types/controller/action/`**: Pre-execution actions (fee payments)
- **`types/controller/forwarding/`**: Cross-chain forwarding (CCTP implementation)

### Interface and Router Packages

- **`types/interfaces/`**: Interface definitions for all components
- **`types/router/`**: Router implementations for extensibility

### Root Package (`types/`)

- **Purpose**: Main entry point with codec registration
- **Files**: `codec.go`, `genesis.go`, `packet.go`
- **Dependencies**: Imports from component, controller, and core packages

## Dependency Levels

1. **Level 0**: `core/` - Foundation with no internal dependencies
2. **Level 1**: Components and controllers that only depend on core
3. **Level 2**: Packages depending on core + root types
4. **Level 3**: Interface definitions and routing
5. **Level 4**: Package aggregators (component/, controller/)
6. **Level 5**: Root types package (main entry point)

## Key Design Principles

- **No Circular Dependencies**: Clean hierarchical structure
- **Core Foundation**: All packages ultimately depend on `types/core/`
- **Modular Design**: Clear separation between components, controllers, and interfaces
- **Aggregation Pattern**: Parent packages aggregate their sub-packages through codec files

