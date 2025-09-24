// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {OrbiterTransientStore} from "./OrbiterTransientStore.sol";
import {OrbiterHypERC20} from "./OrbiterHypERC20.sol";

import { TokenRouter } from "@hyperlane/token/libs/TokenRouter.sol";

/*
 * @dev The canonical entrypoint contract to use Noble's Orbiter implementation through Hyperlane.
 *
 * The Orbiter (https://github.com/noble-assets/orbiter) allows to send cross-chain transfers
 * using various bridge mechanisms, execute actions on the Noble blockchain (e.g. fee payments),
 * and eventually forward the resulting assets to another destination through one of the available
 * bridging mechanisms (e.g. IBC, CCTP).
 *
 * TODO: make upgradeable
 */
contract HyperlaneEntrypoint {
    /*
     * @notice Send an asset transfer to the Orbiter, that should be forwarded to another Hyperlane domain.
     */
    function sendForwardedTransfer(
        address tokenAddress,
        uint32 destinationDomain,
        bytes32 recipient,
        uint256 amount,
        bytes32 payloadHash
    ) external returns (bytes32 messageID) {
        OrbiterHypERC20 token = OrbiterHypERC20(tokenAddress);

        // TODO: this returns the address and then we instantiate the contract again -- can this simply return the OTS implementation?
        address otsAddress = token.getOrbiterTransientStoreAddress();
        require(otsAddress != address(0), "orbiter transient store not set on token");

        // set the pending payload into the transient storage
        OrbiterTransientStore ots = OrbiterTransientStore(otsAddress);
        ots.setPendingPayloadHash(payloadHash);

        /*
         * Transfer tokens from the user to this contract first.
         * This ensures that when transferRemote burns tokens, this contract has them.
         */
        require(
            token.transferFrom(msg.sender, address(this), amount),
            "failed to transfer tokens to entrypoint"
        );

        /*
         * Call transferRemote directly on the token contract.
         * The token contract will handle the transfer and return the message ID.
         */
        return token.transferRemote(destinationDomain, recipient, amount);
    }
}