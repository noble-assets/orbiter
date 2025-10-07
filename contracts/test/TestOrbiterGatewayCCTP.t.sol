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

import {Test} from "forge-std/Test.sol";
import {OrbiterGatewayCCTP} from "../src/OrbiterGatewayCCTP.sol";
import {MockFiatToken, MockTokenMessenger, MockMessageTransmitter} from "./mocks/MockCircle.sol";

contract TestOrbiterGatewayCCTP is Test {
    MockFiatToken public token;
    MockTokenMessenger public tokenMessenger;
    MockMessageTransmitter public messageTransmitter;

    OrbiterGatewayCCTP public gateway;
    uint32 public constant S = 1_000_000;

    address public constant USER = address(0x1);
    bytes32 public constant DESTINATION_CALLER = bytes32(uint256(uint160(address(0x2))));

    /// @notice The payload hash is not decoded in the gateway so it could be a generic 32 bytes
    bytes32 public constant PAYLOAD_HASH = keccak256("test payload");
    bytes32 public constant ORBITER_ADDRESS =
        bytes32(0x000000000000000000000000a197eb1a9bfe6143b2d6499897fc1e3c1cfacbb2);

    uint256 public constant PERMIT_DEADLINE = 1_000_000;
    uint256 public constant TRANSFER_AMOUNT = 1_000_000e6;

    function setUp() public {
        // Deploy mock contracts
        messageTransmitter = new MockMessageTransmitter();
        tokenMessenger = new MockTokenMessenger(address(messageTransmitter));
        token = new MockFiatToken();

        // Deploy gateway
        gateway =
            new OrbiterGatewayCCTP(address(token), address(tokenMessenger), DESTINATION_CALLER);

        // Setup user balance
        token.mint(USER, TRANSFER_AMOUNT);
    }

    // =============================================================================
    // constructor tests
    // =============================================================================

    function testConstructor() public {
        assertEq(address(gateway.TOKEN()), address(token), "token address should be different");
        assertEq(
            address(gateway.TOKEN_MESSENGER()),
            address(tokenMessenger),
            "token messenger address should be different"
        );
        assertEq(
            address(gateway.MESSAGE_TRANSMITTER()),
            address(messageTransmitter),
            "message transmitter address should be different"
        );
        assertEq(
            gateway.DESTINATION_CALLER(),
            DESTINATION_CALLER,
            "destination caller should be different"
        );
        assertEq(gateway.DESTINATION_DOMAIN(), 4, "expected Noble destination domain");
        assertEq(
            gateway.MINT_RECIPIENT(), ORBITER_ADDRESS, "expected Orbiter address for mint recipient"
        );
    }

    function testConstructorZeroTokenAddressRevert() public {
        vm.expectRevert(OrbiterGatewayCCTP.ZeroTokenAddress.selector);
        new OrbiterGatewayCCTP(address(0), address(tokenMessenger), DESTINATION_CALLER);
    }

    function testConstructorZeroTokenMessengerAddressRevert() public {
        vm.expectRevert(OrbiterGatewayCCTP.ZeroTokenMessengerAddress.selector);
        new OrbiterGatewayCCTP(address(token), address(0), DESTINATION_CALLER);
    }

    function testConstructorZeroDestinationCallerRevert() public {
        vm.expectRevert(OrbiterGatewayCCTP.ZeroDestinationCaller.selector);
        new OrbiterGatewayCCTP(address(token), address(tokenMessenger), bytes32(0));
    }

    // =============================================================================
    // depositForBurnWithOrbiterPayload tests
    // =============================================================================

    function testDepositForBurnWithOrbiterPayload() public {
        // Create permit signature (mock v, r, s)
        bytes memory permitSig = abi.encode(uint8(27), bytes32(0), bytes32(0));

        vm.startPrank(USER);

        // Expect events
        vm.expectEmit(true, true, true, true);
        emit OrbiterGatewayCCTP.DepositForBurnWithOrbiterPayload(0, 0, PAYLOAD_HASH);

        gateway.depositForBurnWithOrbiterPayload(
            TRANSFER_AMOUNT, PERMIT_DEADLINE, permitSig, PAYLOAD_HASH
        );

        vm.stopPrank();

        // Verify token transfer
        assertEq(token.balances(USER), 0);
        assertEq(token.balances(address(gateway)), TRANSFER_AMOUNT);
    }

    function testDepositForBurnWithOrbiterPayloadMultipleDeposits() public {
        bytes memory permitSig = abi.encode(uint8(27), bytes32(0), bytes32(0));

        vm.startPrank(USER);

        // First deposit
        gateway.depositForBurnWithOrbiterPayload(
            TRANSFER_AMOUNT / 2, PERMIT_DEADLINE, permitSig, PAYLOAD_HASH
        );

        // Second deposit
        gateway.depositForBurnWithOrbiterPayload(
            TRANSFER_AMOUNT / 2, PERMIT_DEADLINE, permitSig, PAYLOAD_HASH
        );

        vm.stopPrank();

        assertEq(token.balances(USER), 0);
        assertEq(token.balances(address(gateway)), TRANSFER_AMOUNT);
    }
}
