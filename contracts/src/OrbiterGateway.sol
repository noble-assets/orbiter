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

import { OrbiterHypERC20 } from "./OrbiterHypERC20.sol";


/// @title Orbiter Gateway Contract
/// @author Noble Core Team
/// @notice The canonical portal contract to use Noble's Orbiter implementation through Hyperlane.
/// @dev The Orbiter (https://github.com/noble-assets/orbiter) allows to send cross-chain transfers
/// using various bridge mechanisms, execute actions on the Noble blockchain (e.g. fee payments),
/// and eventually forward the resulting assets to another destination through one of the available
/// bridging mechanisms (e.g. IBC, CCTP).
///
/// TODO: make upgradeable?
contract OrbiterGateway {
    uint32 private destinationDomain;

    constructor(uint32 _domain) {
        destinationDomain = _domain;
    }

    /// @notice Send an asset transfer to the Orbiter, that should be forwarded to another Hyperlane domain.
    /// @param _tokenAddress Address of the token to forward using Orbiter.
    /// @param _recipient A bytes32 representation of the token recipient on the receiving chain.
    /// @param _amount The amount of tokens to transfer.
    /// @param _payload The payload passed along with the asset transfer.
    /// @return messageID The ID of the dispatched Hyperlane message.
    function sendForwardedTransfer(
        address _tokenAddress,
        bytes32 _recipient,
        uint256 _amount,
        bytes calldata _payload
    ) external returns (bytes32 messageID) {
        OrbiterHypERC20 token = OrbiterHypERC20(_tokenAddress);

        /*
         * Transfer tokens from the user to this contract first.
         * This is required since transferRemote is burning from msg.sender,
         * which is this contract.
         */
        token.transferFrom(msg.sender, address(this), _amount);

        /*
         * Call transferRemote directly on the token contract.
         * The token contract will handle the transfer and return the message ID.
         */
        return token.transferRemoteWithPayload(
            destinationDomain,
            _recipient,
            _amount,
            _payload
        );
    }
}