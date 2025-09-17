// SPDX-License-Identifier: Apache-2.0
pragma Solidity ^v0.8.13;

import {ITokenRouter} from "./HyperlaneInterfaces.sol";

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
    // The canonical mailbox to dispatch orbiter messages.
    //
    // TODO: check if should be public and immutable?
    IMailbox public immutable mailbox;

    // TODO: add constructor logic
    constructor(address _mailbox) {
        mailbox = IMailbox(_mailbox);
    }

    // TODO: add initialize function so we can use an upgradeable proxy?
    function initialize(
        address mailbox
    ) external initializer {
        __Ownable_init();
    }

    /*
     * @dev Send an asset transfer to the Orbiter, that should be forwarded to another Hyperlane domain.
     */
    function sendWithForwardThroughHyperlane(
        uint32 destinationDomain,
        bytes32 recipient,
        uint256 amount
    ) external returns (bytes32 messageID) {
        ITokenRouter token = new ITokenRouter(tokenAddress);
        bytes32 warpMessageID = token.remoteTransfer(

        );

        bytes32 orbiterMessageID = sendHyperlaneForwardInformation(
            warpMessageID,
            orbiterPayload
        );

        mailbox.dispatch(

        )
    }

    /* @dev Packs the given attributes into a bytes array.
     *
     * This metadata is used to be passed as payload bytes to the Orbiter implementation
     * via the Hyperlane protocol.
     *
     * TODO: check if used types are correct.
     */
    function packHyperlaneForwardingAttributes(
        bytes32 tokenID, // checked
        uint32 destDomain, // checked
        bytes32 recipient, // checked
        bytes32 customHookID, // TODO: maybe instead use the IPostDispatchHook thing here?
        string memory customHookMetadata,
        uint256 gasLimit, // checked
        uint256 maxFeeAmount, // checked
        string memory maxFeeDenom // TODO: even pass this? Shouldn't just the denom on Noble be used?
    ) internal returns (bytes memory) {
        return abi.encode(
            tokenID,
            destDomain,
            recipient,
            customHookID,
            gasLimit,
            maxFeeAmount,
            maxFeeDenom,
            customHookMetadata
        );
    }
}