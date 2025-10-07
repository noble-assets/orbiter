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

/*
	███╗   ██╗ ██████╗ ██████╗ ██╗     ███████╗     ██████╗ ██████╗ ██████╗ ██╗████████╗███████╗██████╗     
	████╗  ██║██╔═══██╗██╔══██╗██║     ██╔════╝    ██╔═══██╗██╔══██╗██╔══██╗██║╚══██╔══╝██╔════╝██╔══██╗    
	██╔██╗ ██║██║   ██║██████╔╝██║     █████╗      ██║   ██║██████╔╝██████╔╝██║   ██║   █████╗  ██████╔╝    
	██║╚██╗██║██║   ██║██╔══██╗██║     ██╔══╝      ██║   ██║██╔══██╗██╔══██╗██║   ██║   ██╔══╝  ██╔══██╗    
	██║ ╚████║╚██████╔╝██████╔╝███████╗███████╗    ╚██████╔╝██║  ██║██████╔╝██║   ██║   ███████╗██║  ██║    
	╚═╝  ╚═══╝ ╚═════╝ ╚═════╝ ╚══════╝╚══════╝     ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝    
																											
	 ██████╗  █████╗ ████████╗███████╗██╗    ██╗ █████╗ ██╗   ██╗     ██████╗ ██████╗████████╗██████╗       
	██╔════╝ ██╔══██╗╚══██╔══╝██╔════╝██║    ██║██╔══██╗╚██╗ ██╔╝    ██╔════╝██╔════╝╚══██╔══╝██╔══██╗      
	██║  ███╗███████║   ██║   █████╗  ██║ █╗ ██║███████║ ╚████╔╝     ██║     ██║        ██║   ██████╔╝      
	██║   ██║██╔══██║   ██║   ██╔══╝  ██║███╗██║██╔══██║  ╚██╔╝      ██║     ██║        ██║   ██╔═══╝       
	╚██████╔╝██║  ██║   ██║   ███████╗╚███╔███╔╝██║  ██║   ██║       ╚██████╗╚██████╗   ██║   ██║           
	 ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝ ╚══╝╚══╝ ╚═╝  ╚═╝   ╚═╝        ╚═════╝ ╚═════╝   ╚═╝   ╚═╝           
*/

import {IFiatToken, ITokenMessenger, IMessageTransmitter} from "./interfaces/Circle.sol";

/**
 * @title OrbiterGatewayCCTP
 * @author Noble Team
 * @notice Allows to initiate a metadata extended CCTP token transfer to the Noble Orbiter module.
 */
contract OrbiterGatewayCCTP {
    /// @notice Thrown when the address of the token to transfer is the zero address.
    error ZeroTokenAddress();
    /// @notice Thrown when the address of the token messenger is the zero address.
    error ZeroTokenMessengerAddress();
    /// @notice Thrown when the address of the destination caller is the zero address.
    error ZeroDestinationCaller();

    /**
     * @notice Emitted when the deposit for burn and the general message
     * passing are executed successfully.
     * @param transferNonce Nonce of the CCTP deposit for burn message.
     * @param payloadNonce Nonce of the GMP message containing the payload hash.
     * @param payloadHash Hash of the Orbiter payload.
     */
    event DepositForBurnWithOrbiterPayload(
        uint256 indexed transferNonce, uint256 indexed payloadNonce, bytes32 indexed payloadHash
    );

    /// @notice Noble chain identifier
    uint32 public constant DESTINATION_DOMAIN = 4;
    /// @notice Padded address of the CCTP module on Noble:
    /// https://github.com/circlefin/noble-cctp/blob/master/x/cctp/types/keys.go#L52-L58
    bytes32 public constant MINT_RECIPIENT =
        0x000000000000000000000000a197eb1a9bfe6143b2d6499897fc1e3c1cfacbb2;

    /// @notice Token transferred to the Orbiter
    IFiatToken public immutable TOKEN;
    /// @notice Token messenger used by the CCTP protocol to exchange token transfer messages
    /// with remote chains
    ITokenMessenger public immutable TOKEN_MESSENGER;
    /// @notice Token transmitter used by the CCTP protocol to exchange generic messages with
    /// remote chains
    IMessageTransmitter public immutable MESSAGE_TRANSMITTER;
    /// @notice The only address allowed to complete the transfer
    /// on the receiving chain
    bytes32 public immutable DESTINATION_CALLER;

    /**
     * @notice Initialize the OrbiterGatewayCCTP contract.
     * @param token_ Address of the token to transfer.
     * @param tokenMessenger_ Address of the CCTP TokenMessenger contract.
     * @param destinationCaller_ Address of the relayer that will complete the transfer to the
     * Noble chain.
     */
    constructor(address token_, address tokenMessenger_, bytes32 destinationCaller_) {
        if (token_ == address(0)) revert ZeroTokenAddress();
        if (tokenMessenger_ == address(0)) revert ZeroTokenMessengerAddress();
        if (destinationCaller_ == bytes32(0)) revert ZeroDestinationCaller();

        TOKEN = IFiatToken(token_);
        TOKEN_MESSENGER = ITokenMessenger(tokenMessenger_);
        MESSAGE_TRANSMITTER = IMessageTransmitter(TOKEN_MESSENGER.localMessageTransmitter());

        DESTINATION_CALLER = destinationCaller_;
    }

    /**
     * @notice Initiates a CCTP token transfer to the Orbiter module on the Noble chain, and a
     * a generic message transfer containing the hash of the payload that will be executed by
     * the Orbiter. The function requires a pre signed permit to transfer funds from the user
     * to the module as per EIP-2612 (https://eips.ethereum.org/EIPS/eip-2612).
     * The Orbiter module requires the two intents created with this method to be relayed to the
     * Noble chain in the same tx.
     * @param amount Amount of tokens to transfer.
     * @param blocktimeDeadline Blocktime after which the permit expires.
     * @param permitSignature ABI encoded secp256k1 signature of the user.
     * @param payloadHash Hash of the Orbiter payload that will be executed.
     */
    function depositForBurnWithOrbiterPayload(
        uint256 amount,
        uint256 blocktimeDeadline,
        bytes calldata permitSignature,
        bytes32 payloadHash
    ) external {
        // NOTE: maybe we can save gas passing directly v, s, and r.
        (uint8 v, bytes32 r, bytes32 s) = abi.decode(permitSignature, (uint8, bytes32, bytes32));

        TOKEN.permit(msg.sender, address(this), amount, blocktimeDeadline, v, r, s);
        TOKEN.transferFrom(msg.sender, address(this), amount);
        TOKEN.approve(address(TOKEN_MESSENGER), amount);

        uint256 transferNonce = TOKEN_MESSENGER.depositForBurnWithCaller(
            amount, DESTINATION_DOMAIN, MINT_RECIPIENT, address(TOKEN), DESTINATION_CALLER
        );

        uint256 payloadNonce = MESSAGE_TRANSMITTER.sendMessageWithCaller(
            DESTINATION_DOMAIN,
            MINT_RECIPIENT,
            DESTINATION_CALLER,
            abi.encodePacked(transferNonce, payloadHash)
        );

        emit DepositForBurnWithOrbiterPayload(transferNonce, payloadNonce, payloadHash);
    }
}
