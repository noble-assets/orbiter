/*
 * Copyright 2025 NASD Inc. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
pragma solidity 0.8.30;

import { Script, console } from "forge-std/Script.sol";
import { OrbiterGatewayCCTP } from "../src/OrbiterGatewayCCTP.sol";

contract OrbiterGatewayCCTPScript_mainnet is Script {
    // https://developers.circle.com/stablecoins/usdc-contract-addresses
    address public immutable TOKEN_ADDRESS = 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48;
    // https://developers.circle.com/cctp/v1/evm-smart-contracts
    address public immutable TOKEN_MESSENGER_ADDRESS = 0xBd3fa81B58Ba92a82136038B25aDec7066af3155;

    // bytes32 paddedAddress =
    //     bytes32(abi.encodePacked(bytes12(0),
    // address(0xc125931f8Fc15B0EE6e873ab769602588E7Dee6e)));
    bytes32 public immutable DESTINATION_CALLER_ADDRESS =
        bytes32(uint256(uint160(address(0xc125931f8Fc15B0EE6e873ab769602588E7Dee6e))));

    OrbiterGatewayCCTP public gateway;

    function setUp() public { }

    function run() public {
        vm.startBroadcast();

        gateway = new OrbiterGatewayCCTP(
            TOKEN_ADDRESS, TOKEN_MESSENGER_ADDRESS, DESTINATION_CALLER_ADDRESS
        );

        vm.stopBroadcast();
    }
}

contract OrbiterGatewayCCTPScript_testnet is Script {
    // https://developers.circle.com/stablecoins/usdc-contract-addresses
    address public immutable TOKEN_ADDRESS = 0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238;
    // https://developers.circle.com/cctp/v1/evm-smart-contracts
    address public immutable TOKEN_MESSENGER_ADDRESS = 0x9f3B8679c73C2Fef8b59B4f3444d4e156fb70AA5;

    // bytes32 paddedAddress =
    //     bytes32(abi.encodePacked(bytes12(0),
    // address(0xc125931f8Fc15B0EE6e873ab769602588E7Dee6e)));
    bytes32 public immutable DESTINATION_CALLER_ADDRESS =
        bytes32(uint256(uint160(address(0xc125931f8Fc15B0EE6e873ab769602588E7Dee6e))));

    OrbiterGatewayCCTP public gateway;

    function setUp() public { }

    function run() public {
        vm.startBroadcast();

        gateway = new OrbiterGatewayCCTP(
            TOKEN_ADDRESS, TOKEN_MESSENGER_ADDRESS, DESTINATION_CALLER_ADDRESS
        );

        vm.stopBroadcast();
    }
}
