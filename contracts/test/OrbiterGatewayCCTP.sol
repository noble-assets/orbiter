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

import { Test } from "forge-std/Test.sol";
import { console } from "forge-std/console.sol";
import { IFiatToken, IMessageTransmitter, ITokenMessenger } from "../src/interfaces/Circle.sol";
import { OrbiterGatewayCCTP } from "../src/OrbiterGatewayCCTP.sol";

contract TestOrbiterGatewayCCTP is Test {
    // https://developers.circle.com/stablecoins/usdc-contract-addresses
    address public immutable TOKEN_ADDRESS = 0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48;
    // https://developers.circle.com/cctp/v1/evm-smart-contracts
    address public immutable TOKEN_MESSENGER_ADDRESS = 0xBd3fa81B58Ba92a82136038B25aDec7066af3155;

    OrbiterGatewayCCTP public gateway;
    IFiatToken public token;
    ITokenMessenger public tokenMessenger;
    IMessageTransmitter public messageTransmitter;

    address public user;
    uint256 private userKey;

    /// @notice The payload hash is not decoded in the gateway so it could be a generic 32 bytes
    // bytes public constant PAYLOAD = bytes(
    //     "{'orbiter':{'pre_actions':[{'id':'ACTION_FEE','attributes':{'@type':'/noble.orbiter.controller.action.v1.FeeAttributes','fees_info':[{'recipient':'noble1a3v6t70vsrhx0yraateg8jqavxp2k7ptwlqkqp','basis_points':-100}]}}],'forwarding':{'protocol_id':'PROTOCOL_CCTP','attributes':{'@type':'/noble.orbiter.controller.forwarding.v1.CCTPAttributes','destination_domain':0,'mint_recipient':'AAAAAAAAAAAAAAAAXIYzVGa7yF5U/nCV03DGGd/k+1g=','destination_caller':null},'passthrough_payload':''}}}"
    // );
    bytes public constant PAYLOAD =
        hex"1b8e010064ca0e8337064bf0cca84a2a084f4c2f787a482707a2373d5ce4a45f40d23c818c9a071eb209427ec0ba1eb86f6861e878cb1024c4d14c1709c5327c1726427c6c37438b0b50fa93a9b172426c290415880d86e0b86008e241343a99a0c99a26c189491a4c4ab89448e1a93487c069282bd36a4a06afb26c466a7abc7d0f48a0ee401ca4f1f1781249e16a624a7a3292989cc2a29d647c2aa51028cfa6a1929cdc697fc5d01f4c43ba537849a793923a295cd1043a3d351e575945e5e99454012550928074c8feb091e359430b63c713862025991a334623e9b05c63f7aac05f1fdc99a27f942eb7d8ad0e7e59723c1bf9c2f8ecef379df71cbb0abf3b47ececf7f49f79fc6dba28fb3f5ab77461b523e1fbf040ffe9850be76e0c791b879d7f3ea00c06fa9f13cf64ecfb4d6948d5ec87ade6918be3d74053409ecba90658f2bb5e5dfb35d138213e3ed30a9848424fb7a371427c3c";

    bytes32 public immutable DESTINATION_CALLER_ADDRESS = bytes32(uint256(uint160(address(0x2))));
    uint256 public immutable TRANSFER_AMOUNT = 1_000_000e6;

    function setUp() public {
        vm.createSelectFork("mainnet");

        token = IFiatToken(TOKEN_ADDRESS);
        vm.label(TOKEN_ADDRESS, "USDC");
        tokenMessenger = ITokenMessenger(TOKEN_MESSENGER_ADDRESS);
        vm.label(TOKEN_MESSENGER_ADDRESS, "TokenMessenger");
        messageTransmitter = tokenMessenger.localMessageTransmitter();
        vm.label(address(messageTransmitter), "MessageTransmitter");

        gateway = new OrbiterGatewayCCTP(
            TOKEN_ADDRESS, TOKEN_MESSENGER_ADDRESS, DESTINATION_CALLER_ADDRESS
        );

        (user, userKey) = makeAddrAndKey("user");
        deal(TOKEN_ADDRESS, user, TRANSFER_AMOUNT);
    }

    function generatePermit(uint256 amount) internal view returns (bytes memory, uint256) {
        uint256 permitDeadline = block.timestamp + 60;
        bytes32 structHash = keccak256(
            abi.encode(
                token.PERMIT_TYPEHASH(), user, gateway, amount, token.nonces(user), permitDeadline
            )
        );
        // prefix is: hex"1901"
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(
            userKey, keccak256(abi.encodePacked("\x19\x01", token.DOMAIN_SEPARATOR(), structHash))
        );
        bytes memory permitSig = abi.encode(v, r, s);
        // console.logBytes(permitSig);
        return (permitSig, permitDeadline);
    }

    // =============================================================================
    // constructor tests
    // =============================================================================

    function testConstructor() public view {
        assertEq(address(gateway.TOKEN()), TOKEN_ADDRESS, "token address should be different");
        assertEq(
            address(gateway.TOKEN_MESSENGER()),
            TOKEN_MESSENGER_ADDRESS,
            "token messenger address should be different"
        );
        assertEq(
            address(gateway.MESSAGE_TRANSMITTER()),
            address(messageTransmitter),
            "message transmitter address should be different"
        );
        assertEq(
            gateway.DESTINATION_CALLER(),
            DESTINATION_CALLER_ADDRESS,
            "destination caller should be different"
        );
    }

    function testConstructorZeroTokenAddressRevert() public {
        vm.expectRevert(OrbiterGatewayCCTP.ZeroTokenAddress.selector);
        new OrbiterGatewayCCTP(address(0), TOKEN_MESSENGER_ADDRESS, DESTINATION_CALLER_ADDRESS);
    }

    function testConstructorZeroTokenMessengerAddressRevert() public {
        vm.expectRevert(OrbiterGatewayCCTP.ZeroTokenMessengerAddress.selector);
        new OrbiterGatewayCCTP(TOKEN_ADDRESS, address(0), DESTINATION_CALLER_ADDRESS);
    }

    function testConstructorZeroDestinationCallerRevert() public {
        vm.expectRevert(OrbiterGatewayCCTP.ZeroDestinationCaller.selector);
        new OrbiterGatewayCCTP(TOKEN_ADDRESS, TOKEN_MESSENGER_ADDRESS, bytes32(0));
    }

    // =============================================================================
    // depositForBurnWithOrbiter tests
    // =============================================================================

    function testDepositForBurnWithOrbiter() public {
        vm.startPrank(user);

        uint64 nonce = messageTransmitter.nextAvailableNonce();
        vm.expectEmit(true, true, true, true);
        emit OrbiterGatewayCCTP.DepositForBurnWithOrbiter(nonce, nonce + 1);

        (bytes memory permitSig, uint256 permitDeadline) = generatePermit(TRANSFER_AMOUNT);
        gateway.depositForBurnWithOrbiter(
            TRANSFER_AMOUNT, permitDeadline, permitSig, PAYLOAD
        );

        vm.stopPrank();
    }

    function testDepositForBurnWithOrbiterMultipleDeposits() public {
        vm.startPrank(user);

        // First deposit
        uint64 nonce = messageTransmitter.nextAvailableNonce();
        vm.expectEmit(true, true, true, true);
        emit OrbiterGatewayCCTP.DepositForBurnWithOrbiter(nonce, nonce + 1);

        (bytes memory permitSig1, uint256 permitDeadline1) = generatePermit(TRANSFER_AMOUNT / 2);
        gateway.depositForBurnWithOrbiter(
            TRANSFER_AMOUNT / 2, permitDeadline1, permitSig1, PAYLOAD
        );

        // Second deposit
        nonce = messageTransmitter.nextAvailableNonce();
        vm.expectEmit(true, true, true, true);
        emit OrbiterGatewayCCTP.DepositForBurnWithOrbiter(nonce, nonce + 1);

        (bytes memory permitSig2, uint256 permitDeadline2) = generatePermit(TRANSFER_AMOUNT / 2);
        gateway.depositForBurnWithOrbiter(
            TRANSFER_AMOUNT / 2, permitDeadline2, permitSig2, PAYLOAD
        );

        vm.stopPrank();
    }
}
