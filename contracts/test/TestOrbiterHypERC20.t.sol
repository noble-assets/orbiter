// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.24;

import "forge-std/Test.sol";

import { OrbiterHypERC20 } from "../src/OrbiterHypERC20.sol";
import { OrbiterTransientStore } from "../src/OrbiterTransientStore.sol";
import { HyperlaneEntrypoint } from "../src/HyperlaneEntrypoint.sol";

import { Mailbox } from "@hyperlane/Mailbox.sol";
import { MockMailbox } from "@hyperlane/mock/MockMailbox.sol";
import { TestPostDispatchHook } from "@hyperlane/test/TestPostDispatchHook.sol";
import { TypeCasts } from "@hyperlane/libs/TypeCasts.sol";

import { TransparentUpgradeableProxy, ITransparentUpgradeableProxy } from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

contract TestOrbiterHypERC20 is Test {
    // NOTE: this is adding the utilities for converting address to Hyperlane expected bytes32.
    using TypeCasts for address;

    /*
     * CONSTANTS
     */
    uint32 internal constant ORIGIN = 1;
    // TODO: check noble destination domain for Hyperlane
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
    HyperlaneEntrypoint internal entrypoint;

    /*
     * TESTING ACCOUNTS
     */
    address internal constant ALICE = address(0x1);
    address internal constant BOB = address(0x2);
    address internal constant ADMIN = address(0x3);
    address internal constant HYP_OWNER = address(0x4);

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

        // Set up Orbiter transient store.
        OrbiterTransientStore ots = new OrbiterTransientStore();

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

        entrypoint = new HyperlaneEntrypoint();

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

        vm.startPrank(HYP_OWNER);

        // Enroll the routers
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

    /*
     * @notice This test checks that the setup was successful by asserting
     * expected token balances and correct wiring of the interdependent contracts.
     */
    function testSetupWorked() public {
        assertNotEq(
            localToken.balanceOf(ALICE),
            0,
            "expected alice to have non-zero token balance after setup"
        );
    }

    /*
     * @notice This tests that sending a Hyperlane forwarded transfer
     * using the entrypoint contract is going to work.
     */
    function testForwardedTransfer() public {
        vm.startPrank(ALICE);

        uint256 sentAmount = 1230;
        assertGe(localToken.balanceOf(ALICE), sentAmount, "sent amount exceeds available token balance");

        // Approve the entrypoint contract to spend ALICE's tokens
        require(localToken.approve(address(entrypoint), sentAmount), "failed to approve entrypoint");

        bytes32 sentPayloadHash = bytes32(uint256(1234));

        bytes32 messageID = entrypoint.sendForwardedTransfer(
            address(localToken),
            DESTINATION,
            bytes32(uint256(uint160(BOB))), // This converts the 20-byte address to a bytes32 value.
            sentAmount,
            sentPayloadHash
        );
        assertNotEq32(messageID, 0, "expected non-zero message ID");

        vm.stopPrank();
    }

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
                address(0), // TODO: check if InterchainGasPaymaster has to be created?
                _owner,
                // custom OrbiterHypERC20 initialization arguments
                _otsAddress
            )
        );

        remoteToken = OrbiterHypERC20(address(proxy));

        return remoteToken;
    }
}
