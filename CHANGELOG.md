# CHANGELOG

## v1.0.0 - Oct 20, 2025

### Features

- Add queries for action and protocol IDs. ([#54](https://github.com/noble-assets/orbiter/pull/54))
- Internal forwarding controller. ([#41](https://github.com/noble-assets/orbiter/pull/41))
- Add Hyperlane forwarding controller. ([#24](https://github.com/noble-assets/orbiter/pull/24))
- Add cli commands. ([#15](https://github.com/noble-assets/orbiter/pull/15))
- Add dispatcher query server and genesis. ([#10](https://github.com/noble-assets/orbiter/pull/10))
- Add duplicate actions check. ([#13](https://github.com/noble-assets/orbiter/pull/13))
- Add genesis and servers for forwarder, executor, and adapter components.
  ([#6](https://github.com/noble-assets/orbiter/pull/6))
- Add max bytes validation for passthrough payload.
  ([#7](https://github.com/noble-assets/orbiter/pull/7))
- Add IBC adapter, CCTP forwarder, and fee action.
  ([#2](https://github.com/noble-assets/orbiter/pull/2))
- Implement module architecture. ([#1](https://github.com/noble-assets/orbiter/pull/1))
- Scaffold module.

### Improvements

- General codebase cleanup and maintenance. ([#53](https://github.com/noble-assets/orbiter/pull/53))
- Update FQN for interface registration. ([#44](https://github.com/noble-assets/orbiter/pull/44))
- Add validate method to forwarding attributes interface.
  ([#46](https://github.com/noble-assets/orbiter/pull/46))
- Update project status in README and repo name in LICENSE.
  ([#40](https://github.com/noble-assets/orbiter/pull/40))
- Add CCTP controller docs. ([#37](https://github.com/noble-assets/orbiter/pull/37))
- Add Hyperlane controller docs. ([#30](https://github.com/noble-assets/orbiter/pull/30))
- Improve integration and payload docs. ([#33](https://github.com/noble-assets/orbiter/pull/33))
- Allow empty cctp destination caller. ([#39](https://github.com/noble-assets/orbiter/pull/39))
- Refactor validation of dispatcher genesis state into types.
  ([#36](https://github.com/noble-assets/orbiter/pull/36))
- Rename controllers to use correct grammar.
  ([#35](https://github.com/noble-assets/orbiter/pull/35))
- Implement e2e tests for hyperlane controller.
  ([#34](https://github.com/noble-assets/orbiter/pull/34))
- Add event emitting. ([#27](https://github.com/noble-assets/orbiter/pull/27))
- Handle dust collector as a module account.
  ([#28](https://github.com/noble-assets/orbiter/pull/28))
- Check for nil forwarding packet in CCTP controller.
  ([#26](https://github.com/noble-assets/orbiter/pull/26))
- Add more CI actions. ([#25](https://github.com/noble-assets/orbiter/pull/25))
- Rename orbiter package. ([#23](https://github.com/noble-assets/orbiter/pull/23))
- Support pushing protobuf to registry. ([#22](https://github.com/noble-assets/orbiter/pull/22))
- Improve IBC to CCTP e2e tests. ([#19](https://github.com/noble-assets/orbiter/pull/19))
- Add module and architecture files. ([#3](https://github.com/noble-assets/orbiter/pull/3))
- Update integration docs with correct payload example.
  ([#21](https://github.com/noble-assets/orbiter/pull/21))
- Add custom ibc ack for middleware. ([#17](https://github.com/noble-assets/orbiter/pull/17))
- Unify adapter interfaces and remove unused ones.
  ([#18](https://github.com/noble-assets/orbiter/pull/18))
- Move fee validation to type. ([#16](https://github.com/noble-assets/orbiter/pull/16))
- Use errorsmod throughout repo. ([#14](https://github.com/noble-assets/orbiter/pull/14))
- Add nancy vulnerability checker. ([#11](https://github.com/noble-assets/orbiter/pull/11))
- Add method to get all paused cross-chain IDs.
  ([#12](https://github.com/noble-assets/orbiter/pull/12))
- Add genesis tests and improve id validation.
  ([#9](https://github.com/noble-assets/orbiter/pull/9))
- Restructure types package dependency hierarchy.
  ([#8](https://github.com/noble-assets/orbiter/pull/8))
- Update project structure, component names, and codec usage.
  ([#4](https://github.com/noble-assets/orbiter/pull/4))
- Change method visibility and return concrete types over interfaces.
  ([#5](https://github.com/noble-assets/orbiter/pull/5))

### Fixes

- Revert IBC counterparty ID. ([#51](https://github.com/noble-assets/orbiter/pull/51))
- Apply changes from external audit. ([#42](https://github.com/noble-assets/orbiter/pull/42))
- Get denom balance instead of all balances.
  ([#49](https://github.com/noble-assets/orbiter/pull/49))
- Use correct counterparty channel ID for IBC cross-chain IDs.
  ([#48](https://github.com/noble-assets/orbiter/pull/48))
- Fix tx commands not working with `EnhanceCustomCommand`.
  ([#43](https://github.com/noble-assets/orbiter/pull/43))
- Register rest endpoints. ([#31](https://github.com/noble-assets/orbiter/pull/31))
