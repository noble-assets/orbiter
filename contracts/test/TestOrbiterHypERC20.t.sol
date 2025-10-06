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
import { OrbiterTransientStorage } from "../src/OrbiterTransientStorage.sol";
import { OrbiterGateway } from "../src/OrbiterGateway.sol";

import { Mailbox } from "@hyperlane/Mailbox.sol";
import { IMailbox } from "@hyperlane/interfaces/IMailbox.sol";
import { TypeCasts } from "@hyperlane/libs/TypeCasts.sol";
import { Message } from "@hyperlane/libs/Message.sol";
import { MockMailbox } from "@hyperlane/mock/MockMailbox.sol";
import { TestPostDispatchHook } from "@hyperlane/test/TestPostDispatchHook.sol";
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
        gateway = new OrbiterGateway();

        // Set up Orbiter transient store.
        OrbiterTransientStorage ots = new OrbiterTransientStorage(address(gateway));

        // Deploy Orbiter compatible token with a proxy.
        localToken = deployOrbiterHypERC20(
            address(originMailbox),
            address(noopHook),
            HYP_OWNER,
            address(ots)
        );

        remoteToken = deployOrbiterHypERC20(
            address(remoteMailbox),
            address(noopHook),
            HYP_OWNER,
            address(ots)
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

        bytes32 sentPayloadHash = bytes32(uint256(1234));

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
                abi.encodePacked(sentPayloadHash)
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
            DESTINATION,
            BOB.addressToBytes32(),
            sentAmount,
            sentPayloadHash
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

    /// @notice This test shows that the Orbiter transient storage cannot be
    /// called from external addresses but only from the Gateway contract
    /// that is its owner.
    function testSetPendingPayload() public {
        OrbiterTransientStorage ots = localToken.getOrbiterTransientStore();

        vm.prank(address(gateway));
        ots.setPendingPayloadHash(bytes32(uint256(123)));

        vm.prank(ALICE);
        vm.expectRevert("Ownable: caller is not the owner");
        ots.setPendingPayloadHash(bytes32(uint256(123)));
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
    /// @param _otsAddress Address of the Orbiter transient storage associated with this contract.
    function deployOrbiterHypERC20(
        address _mailboxAddress,
        address _hook,
        address _owner,
        address _otsAddress
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
                OrbiterHypERC20.initialize.selector,
                // default HypERC20 initialization arguments
                TOTAL_SUPPLY,
                NAME,
                SYMBOL,
                _hook,
                address(0), // using no IGP here
                _owner,
                // custom OrbiterHypERC20 initialization arguments
                _otsAddress
            )
        );

        return OrbiterHypERC20(address(proxy));
    }
}
