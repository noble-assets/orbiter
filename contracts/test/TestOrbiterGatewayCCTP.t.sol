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

import {console} from "forge-std/console.sol";
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
    // bytes public constant PAYLOAD = bytes(
    //     "{'orbiter':{'pre_actions':[{'id':'ACTION_FEE','attributes':{'@type':'/noble.orbiter.controller.action.v1.FeeAttributes','fees_info':[{'recipient':'noble1a3v6t70vsrhx0yraateg8jqavxp2k7ptwlqkqp','basis_points':-100}]}}],'forwarding':{'protocol_id':'PROTOCOL_CCTP','attributes':{'@type':'/noble.orbiter.controller.forwarding.v1.CCTPAttributes','destination_domain':0,'mint_recipient':'AAAAAAAAAAAAAAAAXIYzVGa7yF5U/nCV03DGGd/k+1g=','destination_caller':null},'passthrough_payload':''}}}"
    // );
    bytes constant PAYLOAD =
        hex"1b8e010064ca0e8337064bf0cca84a2a084f4c2f787a482707a2373d5ce4a45f40d23c818c9a071eb209427ec0ba1eb86f6861e878cb1024c4d14c1709c5327c1726427c6c37438b0b50fa93a9b172426c290415880d86e0b86008e241343a99a0c99a26c189491a4c4ab89448e1a93487c069282bd36a4a06afb26c466a7abc7d0f48a0ee401ca4f1f1781249e16a624a7a3292989cc2a29d647c2aa51028cfa6a1929cdc697fc5d01f4c43ba537849a793923a295cd1043a3d351e575945e5e99454012550928074c8feb091e359430b63c713862025991a334623e9b05c63f7aac05f1fdc99a27f942eb7d8ad0e7e59723c1bf9c2f8ecef379df71cbb0abf3b47ececf7f49f79fc6dba28fb3f5ab77461b523e1fbf040ffe9850be76e0c791b879d7f3ea00c06fa9f13cf64ecfb4d6948d5ec87ade6918be3d74053409ecba90658f2bb5e5dfb35d138213e3ed30a9848424fb7a371427c3c";
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
        gateway = new OrbiterGatewayCCTP(address(token), address(tokenMessenger), DESTINATION_CALLER);

        // Setup user balance
        token.mint(USER, TRANSFER_AMOUNT);
    }

    // =============================================================================
    // constructor tests
    // =============================================================================

    function testConstructor() public {
        assertEq(address(gateway.TOKEN()), address(token), "token address should be different");
        assertEq(
            address(gateway.TOKEN_MESSENGER()), address(tokenMessenger), "token messenger address should be different"
        );
        assertEq(
            address(gateway.MESSAGE_TRANSMITTER()),
            address(messageTransmitter),
            "message transmitter address should be different"
        );
        assertEq(gateway.DESTINATION_CALLER(), DESTINATION_CALLER, "destination caller should be different");
        assertEq(gateway.DESTINATION_DOMAIN(), 4, "expected Noble destination domain");
        assertEq(gateway.MINT_RECIPIENT(), ORBITER_ADDRESS, "expected Orbiter address for mint recipient");
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
        emit OrbiterGatewayCCTP.DepositForBurnWithOrbiterPayload(0, 0);

        gateway.depositForBurnWithOrbiterPayload(TRANSFER_AMOUNT, PERMIT_DEADLINE, permitSig, PAYLOAD);

        vm.stopPrank();

        // Verify token transfer
        assertEq(token.balances(USER), 0);
        assertEq(token.balances(address(gateway)), TRANSFER_AMOUNT);
    }

    function testDepositForBurnWithOrbiterPayloadMultipleDeposits() public {
        bytes memory permitSig = abi.encode(uint8(27), bytes32(0), bytes32(0));

        vm.startPrank(USER);

        // First deposit
        gateway.depositForBurnWithOrbiterPayload(TRANSFER_AMOUNT / 2, PERMIT_DEADLINE, permitSig, PAYLOAD);

        // Second deposit
        gateway.depositForBurnWithOrbiterPayload(TRANSFER_AMOUNT / 2, PERMIT_DEADLINE, permitSig, PAYLOAD);

        vm.stopPrank();

        assertEq(token.balances(USER), 0);
        assertEq(token.balances(address(gateway)), TRANSFER_AMOUNT);
    }
}
