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
    /// @notice The error raised when receiving empty payload bytes.
    error EmptyPayload();

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
    /// @param _payload The orbiter payload sent along with the token transfer.
    /// @return messageID The message ID of the dispatched Hyperlane message.
    function transferRemoteWithPayload(
        uint32 _destination,
        bytes32 _recipient,
        uint256 _amount,
        bytes calldata _payload
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

    /// @notice Transfer assets cross-chain using the Hyperlane message passing
    /// framework and send a payload along with it.
    /// @param _destination The destination domain for the cross-chain transfer.
    /// @param _recipient The bytes32 representation of the recipient address.
    /// @param _amount The sent token amount.
    /// @param _value The sent amount of native denomination sent along with the contract call.
    /// @param _hookMetadata Any metadata required for the registered hook for this token.
    /// @param _hook The address of the hook contract.
    /// @param _payload The orbiter payload sent along with the token transfer.
    /// @return messageID The message ID of the dispatched Hyperlane message.
    function _transferRemoteWithPayload(
        uint32 _destination,
        bytes32 _recipient,
        uint256 _amount,
        uint256 _value,
        bytes memory _hookMetadata,
        address _hook,
        bytes calldata _payload
    ) internal virtual returns (bytes32 messageID) {
        // Run default logic for HypERC20 token.
        HypERC20._transferFromSender(_amount);

        // We only want to handle transfers with a non-zero payload.
        if (_payload.length == 0) {
            revert EmptyPayload();
        }

        bytes memory _tokenMessage = TokenMessage.format(
            _recipient,
            _amount,
            _payload
        );

        messageID = _Router_dispatch(
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
