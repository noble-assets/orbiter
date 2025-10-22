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
pragma solidity ^0.8.24;

import "forge-std/Test.sol";

import { OrbiterHypERC20 } from "../src/OrbiterHypERC20.sol";
import { OrbiterGateway } from "../src/OrbiterGateway.sol";

import { Mailbox } from "@hyperlane/Mailbox.sol";
import { IMailbox } from "@hyperlane/interfaces/IMailbox.sol";
import { TypeCasts } from "@hyperlane/libs/TypeCasts.sol";
import { Message } from "@hyperlane/libs/Message.sol";
import { MockMailbox } from "@hyperlane/mock/MockMailbox.sol";
import { TestPostDispatchHook } from "@hyperlane/test/TestPostDispatchHook.sol";
import { HypERC20 } from "@hyperlane/token/HypERC20.sol";
import { TokenMessage } from "@hyperlane/token/libs/TokenMessage.sol";
import { TokenRouter } from "@hyperlane/token/libs/TokenRouter.sol";

import { TransparentUpgradeableProxy, ITransparentUpgradeableProxy } from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

/// @notice Tests for the Orbiter extension of the Hyperlane ERC-20 contracts.
/// @author Noble Core Team
contract TestOrbiterHypERC20 is Test {
    // NOTE: this is adding the utilities for converting address to Hyperlane expected bytes32.
    using TypeCasts for address;

    /*
     * CONSTANTS
     */
    uint32 internal constant ORIGIN = 1;
    uint32 internal constant DESTINATION = 6;
    uint8 internal DECIMALS = 6;
    uint256 internal SCALE = 1;
    uint256 internal TOTAL_SUPPLY = 2e7; // 20 $ORB

    string internal constant NAME = "Orbiter";
    string internal constant SYMBOL = "ORB";

    /*
     * HYPERLANE CONTRACTS
     */
    MockMailbox internal originMailbox;
    MockMailbox internal remoteMailbox;
    TestPostDispatchHook internal noopHook;

    OrbiterHypERC20 internal localToken;
    OrbiterHypERC20 internal remoteToken;
    OrbiterGateway internal gateway;

    /*
     * TESTING ACCOUNTS
     */
    address internal constant ALICE = address(0x1);
    address internal constant BOB = address(0x2);
    address internal constant ADMIN = address(0x3);
    address internal constant HYP_OWNER = address(0x4);

    /// @notice Shared setup for all test scenarios.
    /// @dev Deploys mocked mailboxes, Orbiter tokens and sets up the required routing.
    function setUp() public virtual {
        // Run setup from ADMIN to make it the owner of contracts.
        //
        // NOTE: This MUST be a different account than the caller contract,
        // because the TransparentUpgradeableProxy does not forward calls to the contracts
        // if sending a transaction from its admin.
        // This is an additional security mechanism to only have external accounts interact with the proxy's methods
        // and the proxy admin to only be able to call configuration / settings methods.
        vm.startPrank(ADMIN);

        // Set up testing instances of Hyperlane dependencies.
        originMailbox = new MockMailbox(ORIGIN);
        remoteMailbox = new MockMailbox(DESTINATION);
        originMailbox.addRemoteMailbox(DESTINATION, remoteMailbox);
        remoteMailbox.addRemoteMailbox(ORIGIN, originMailbox);

        noopHook = new TestPostDispatchHook();

        address owner = remoteMailbox.owner();
        require(owner == ADMIN, "expected admin to be owner of mailbox");

        remoteMailbox.setDefaultHook(address(noopHook));
        remoteMailbox.setRequiredHook(address(noopHook));

        // Deploy the Orbiter gateway contract.
        uint32 nobleDomain = 6;
        gateway = new OrbiterGateway(nobleDomain);

        // Deploy Orbiter compatible token with a proxy.
        localToken = deployOrbiterHypERC20(
            address(originMailbox),
            address(noopHook),
            HYP_OWNER
        );

        remoteToken = deployOrbiterHypERC20(
            address(remoteMailbox),
            address(noopHook),
            HYP_OWNER
        );

        // After setting up the state we need to fund the test accounts
        // with the ERC-20s.
        //
        // NOTE: the msg.sender of the `initialize` call has the supply of tokens
        // minted to the corresponding address.
        require(localToken.balanceOf(ADMIN) == TOTAL_SUPPLY, "expected tokens to be minted");
        require(localToken.balanceOf(ALICE) == 0, "expected alice to have no tokens before transfer");

        uint256 sentAmount = 1e7;
        require(localToken.transfer(ALICE, sentAmount), "failed to send tokens to alice");
        require(localToken.balanceOf(ALICE) == sentAmount, "expected tokens to have been sent to alice");

        vm.stopPrank();

        /*
         * Enrolling routers has to be done by the HYP token owner.
         */
        vm.startPrank(HYP_OWNER);

        localToken.enrollRemoteRouter(
            DESTINATION,
            address(remoteToken).addressToBytes32()
        );

        remoteToken.enrollRemoteRouter(
            ORIGIN,
            address(localToken).addressToBytes32()
        );

        vm.stopPrank();
    }

    /// @notice This test checks that the setup was successful by asserting
    /// expected token balances.
    function testSetupWorked() public view {
        assertNotEq(
            localToken.balanceOf(ALICE),
            0,
            "expected alice to have non-zero token balance after setup"
        );
    }

    /// @notice This tests that sending a Hyperlane forwarded transfer
    /// using the gateway contract is going to work and emits
    /// the expected dispatch event, which contains the payload hash.
    function testForwardedTransfer() public {
        vm.startPrank(ALICE);

        uint256 sentAmount = 1230;
        assertGe(localToken.balanceOf(ALICE), sentAmount, "sent amount exceeds available token balance");

        // Approve the gateway contract to spend ALICE's tokens
        require(localToken.approve(address(gateway), sentAmount), "failed to approve gateway");

        // bytes memory sentPayload = hex"0a9f010801129a010a312f6e6f626c652e6f7262697465722e636f6e74726f6c6c65722e616374696f6e2e76312e4665654174747269627574657312650a300a2c6e6f626c6531717277647a386d796366393636617a75706e383472633438737032786a6e373337363979763910640a310a2c6e6f626c6531673430667536666d3579646e6a356e7639646b6d327071706c6330363070386377617665616410c80112d801080312b1010a352f6e6f626c652e6f7262697465722e636f6e74726f6c6c65722e666f7277617264696e672e76312e4879704174747269627574657312780a20378c68bc8ea4319e981adde0623559c2f6c5f1b6fe57299b8c25693fd0eb458e1a202ded8657ed21a1dfd7779b467b22a93d3ecf12f3613c977bbc78df5aa1c62832222068224b2ae4f9a07b354685339605745f92e333f4b63fef0e91e9d000efc5bb5232033130303a0b0a047573646312033130301a209bdf0925ab244b2839a1c5863e93c7059eb22ee70c57d0c396100c31abac83c1";
        // bytes memory sentPayload = hex"1f8b08000000000000ffe29acfc8c128348b91cb503f2f3f2927552fbf2829b324b5482f393fafa4283f2727b5482f31b924333f4fafcc50cf2d35d5b1a4a42833a9b424b5582895cb804b07accbd0d42cafb82cb7dc22232d2dab2cb72a39abacc028b1bca8a02cc5a83c37adc42c2325aba0d2442085cb10a6a3cc34b1dc20a5c0343bb130b7aa2823c3d8b2a4d0d4a8243bb7b0c2a8b228d522abd42cbdc05ce004a3d00d460e66a18d8c5ca6b85d98965f549e58949299970e72a5476501922b2bb8147e7d35eb58749b275b27af6fe5ec2b7e21cf39f644b10706c7bf9db2256693aae6722905b7920fff6be7cee70b787afeed9ed9a6975a1e1c4acfe59150f2f698d3d8a87b535e4961ceeb27129ce7d6cd9af444beaaf7d49bb01ec75596c593791ae5b9dd769d3afed1de88d9d0c0c08a9b8ba5b438255908c49152d8ad7afa9fe8fe5d258af2269719ae5bed71d1b45b75f867e7cd5bc2052e9ed3cde600020000ffffb252e9217d010000";
        bytes memory sentPayload = hex"1b8e010064ca0e8337064bf0cca84a2a084f4c2f787a482707a2373d5ce4a45f40d23c818c9a071eb209427ec0ba1eb86f6861e878cb1024c4d14c1709c5327c1726427c6c37438b0b50fa93a9b172426c290415880d86e0b86008e241343a99a0c99a26c189491a4c4ab89448e1a93487c069282bd36a4a06afb26c466a7abc7d0f48a0ee401ca4f1f1781249e16a624a7a3292989cc2a29d647c2aa51028cfa6a1929cdc697fc5d01f4c43ba537849a793923a295cd1043a3d351e575945e5e99454012550928074c8feb091e359430b63c713862025991a334623e9b05c63f7aac05f1fdc99a27f942eb7d8ad0e7e59723c1bf9c2f8ecef379df71cbb0abf3b47ececf7f49f79fc6dba28fb3f5ab77461b523e1fbf040ffe9850be76e0c791b879d7f3ea00c06fa9f13cf64ecfb4d6948d5ec87ade6918be3d74053409ecba90658f2bb5e5dfb35d138213e3ed30a9848424fb7a371427c3c";

        // NOTE: the expected message is the wrapped token message with the contained payload.
        bytes memory expectedMessage = _formatMessageWithMemoryBody(
            0,
            ORIGIN,
            address(localToken).addressToBytes32(),
            DESTINATION,
            address(remoteToken).addressToBytes32(),
            TokenMessage.format(
                BOB.addressToBytes32(),
                sentAmount,
                sentPayload
            )
        );

        vm.expectEmit();
        emit IMailbox.Dispatch(
            address(localToken),
            DESTINATION,
            address(remoteToken).addressToBytes32(),
            expectedMessage
        );

        bytes32 messageID = gateway.sendForwardedTransfer(
            address(localToken),
            BOB.addressToBytes32(),
            sentAmount,
            sentPayload
        );
        assertNotEq32(messageID, 0, "expected non-zero message ID");

        vm.stopPrank();
    }

    /// @notice This test shows that the token contract can still be used
    /// for direct `remoteTransfer` calls that do not go through the Hyperlane gateway,
    /// and therefore don't insert the payload hash into the emitted token message.
    function testStandardRemoteTransfer() public {
        uint256 sentAmount = 123;
        uint256 initialBalance = localToken.balanceOf(ALICE);
        require(initialBalance >= sentAmount);

        // NOTE: this tracks the next emitted event and checks if it was emitted in the following
        // function call.
        vm.expectEmit();
        emit TokenRouter.SentTransferRemote(
            DESTINATION,
            address(BOB).addressToBytes32(),
            sentAmount
        );

        vm.prank(ALICE);
        localToken.transferRemote(
            DESTINATION,
            address(BOB).addressToBytes32(),
            sentAmount
        );

        require(localToken.balanceOf(ALICE) == initialBalance - sentAmount, "expected tokens to be sent");
    }

    /// @notice Helper function to format a message with bytes memory body
    /// @dev This is needed because Message.formatMessage expects bytes calldata.
    ///
    /// It's based on the implementation on Message.sol:
    /// https://github.com/hyperlane-xyz/hyperlane-monorepo/blob/%40hyperlane-xyz/core%409.0.9/solidity/contracts/libs/Message.sol#L21-L52
    function _formatMessageWithMemoryBody(
        uint32 _nonce,
        uint32 _originDomain,
        bytes32 _sender,
        uint32 _destinationDomain,
        bytes32 _recipient,
        bytes memory _messageBody
    ) internal pure returns (bytes memory) {
        uint8 _version = 3; // used version of the Hyperlane message implementation
        return abi.encodePacked(
            _version,
            _nonce,
            _originDomain,
            _sender,
            _destinationDomain,
            _recipient,
            _messageBody
        );
    }

    /// @notice Deploy an instance of the OrbiterHypERC20 contract for testing purposes.
    ///
    /// @param _mailboxAddress Address of the used mailbox for this Hyperlane token.
    /// @param _hook Address of the used post-dispatch hook.
    /// @param _owner Address of the contract owner.
    function deployOrbiterHypERC20(
        address _mailboxAddress,
        address _hook,
        address _owner
    ) internal returns (OrbiterHypERC20) {
        OrbiterHypERC20 implementation = new OrbiterHypERC20(
            DECIMALS,
            SCALE,
            _mailboxAddress
        );

        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(implementation),
            msg.sender,
            abi.encodeWithSelector(
                HypERC20.initialize.selector,
                // default HypERC20 initialization arguments
                TOTAL_SUPPLY,
                NAME,
                SYMBOL,
                _hook,
                address(0), // using no IGP here
                _owner
            )
        );

        return OrbiterHypERC20(address(proxy));
    }
}
