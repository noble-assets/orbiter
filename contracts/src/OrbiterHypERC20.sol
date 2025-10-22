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

import { HypERC20 } from "@hyperlane/token/HypERC20.sol";
import { TokenMessage } from "@hyperlane/token/libs/TokenMessage.sol";

/// @title Orbiter HypERC-20 Extension
/// @author Noble Core Team
/// @notice Extends the HypERC20 contracts to include custom payload handling
/// for cross-chain transfers using Noble's Orbiter.
contract OrbiterHypERC20 is HypERC20 {
    /// @notice Calls the constructor method of the underlying HypERC20 contract.
    /// @param _decimals The decimals of the ERC-20 token.
    /// @param _scale The scaling factor to apply for cross-chain transfers.
    /// @param _mailbox The address of the associated mailbox.
    constructor(
        uint8 _decimals,
        uint256 _scale,
        address _mailbox
    ) HypERC20(_decimals, _scale, _mailbox) {}

    /// @notice Transfer assets cross-chain using the Hyperlane message passing
    /// framework and send a payload along with it.
    /// @param _destination The destination domain for the cross-chain transfer.
    /// @param _recipient The bytes32 representation of the recipient address.
    /// @param _amount The sent token amount.
    /// @return messageID The message ID of the dispatched Hyperlane message.
    function transferRemoteWithPayload(
        uint32 _destination,
        bytes32 _recipient,
        uint256 _amount,
        bytes memory _payload
    ) external payable virtual returns (bytes32 messageID) {
        return _transferRemoteWithPayload(
            _destination,
            _recipient,
            _amount,
            msg.value,
            _GasRouter_hookMetadata(_destination),
            address(hook),
            _payload
        );
    }

    function _transferRemoteWithPayload(
        uint32 _destination,
        bytes32 _recipient,
        uint256 _amount,
        uint256 _value,
        bytes memory _hookMetadata,
        address _hook,
        bytes memory _payload
    ) internal virtual returns (bytes32) {
        // Run default logic for HypERC20 token.
        HypERC20._transferFromSender(_amount);

        // We only want to handle transfers with a non-zero payload.
        require(_payload.length != 0, "not supporting empty payloads");
        bytes memory _tokenMessage = TokenMessage.format(
            _recipient,
            _amount,
            _payload
        );

        bytes32 messageID = _Router_dispatch(
            _destination,
            _value,
            _tokenMessage,
            _hookMetadata,
            _hook
        );

        emit SentTransferRemote(
            _destination,
            _recipient,
            _amount
        );

        return messageID;
    }
}
