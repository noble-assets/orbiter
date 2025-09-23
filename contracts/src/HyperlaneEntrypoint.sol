// SPDX-License-Identifier: Apache-2.0
pragma solidity ^0.8.24;

import {NobleDollar} from "./usdn/USDNHyperlaneOrbiter.sol";
import {OrbiterTransientStore} from "./usdn/OrbiterTransientStore.sol";

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
    address private otsAddress;

    constructor(address _otsAddress) {
        otsAddress = _otsAddress;
    }

    /*
     * @dev Send an asset transfer to the Orbiter, that should be forwarded to another Hyperlane domain.
     */
    function sendUSDNWithForwardThroughHyperlane(
        address tokenAddress,
        uint32 destinationDomain,
        bytes32 recipient,
        uint256 amount,
        bytes32 payloadHash
    ) external returns (bytes32 messageID) {
        OrbiterTransientStore ots = OrbiterTransientStore(otsAddress);

        // 1. set the pending payload into the transient storage
        ots.setPendingPayloadHash(payloadHash);

        // 2. send the warp transfer with the NobleDollar contract,
        //    which will get the payload hash to build the metadata
        //    in the overriden _transferRemote internal method.
        NobleDollar token = NobleDollar(tokenAddress);
        return token.transferRemote(
            destinationDomain,
            recipient,
            amount
        );
    }
}