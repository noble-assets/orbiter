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

import {OrbiterTransientStorage} from "./OrbiterTransientStorage.sol";

import { HypERC20 } from "@hyperlane/token/HypERC20.sol";
import { TokenMessage } from "@hyperlane/token/libs/TokenMessage.sol";

/// @title Orbiter HypERC-20 Extension
/// @author Noble Core Team
/// @notice Extends the HypERC20 contracts to include custom payload handling
/// for cross-chain transfers using Noble's Orbiter.
contract OrbiterHypERC20 is HypERC20 {
    OrbiterTransientStorage private ots;

    /// @notice Calls the constructor method of the underlying HypERC20 contract.
    /// @param _decimals The decimals of the ERC-20 token.
    /// @param _scale The scaling factor to apply for cross-chain transfers.
    /// @param _mailbox The address of the associated mailbox.
    constructor(
        uint8 _decimals,
        uint256 _scale,
        address _mailbox
    ) HypERC20(_decimals, _scale, _mailbox) {}

    /// @notice Initializes the contract by calling the initialization logic
    /// of the HypERC20 contract and setting the Orbiter transient store.
    ///
    /// @param _totalSupply The initially minted total supply of the Orbiter HypERC20 token.
    /// @param _name The token name.
    /// @param _symbol The token symbol.
    /// @param _hook Address of the used post-dispatch hook.
    /// @param _interchainSecurityModule Address of the used ISM.
    /// @param _owner Address of the contract owner.
    /// @param _orbiterTransientStorage Address of the Orbiter transient storage contract.
    function initialize(
        uint256 _totalSupply,
        string memory _name,
        string memory _symbol,
        address _hook,
        address _interchainSecurityModule,
        address _owner,
        address _orbiterTransientStorage
    ) public virtual initializer {
        super.initialize(
            _totalSupply,
            _name,
            _symbol,
            _hook,
            _interchainSecurityModule,
            _owner
        );

        ots = OrbiterTransientStorage(_orbiterTransientStorage);
    }

    /// @notice Returns the address of the Orbiter transient store that's
    /// associated with this contract.
    /// @return The instance of the Orbiter transient storage associated with this token.
    function getOrbiterTransientStore() external view returns (OrbiterTransientStorage) {
        return ots;
    }

    /// @notice Overrides the standard implementation of HypERC20 to support
    /// passing payloads within the same transaction using the Orbiter
    /// transient store.
    /// @param _destination The destination domain for the cross-chain transfer.
    /// @param _recipient The bytes32 representation of the recipient address.
    /// @param _amount The sent token amount.
    /// @param _value The native denomination sent along with the transaction.
    /// @param _hookMetadata The metadata to pass along to the post-dispatch hook.
    /// @param _hook Address of the post-dispatch hook.
    /// @return The message ID of the dispatched Hyperlane message.
    function _transferRemote(
        uint32 _destination,
        bytes32 _recipient,
        uint256 _amount,
        uint256 _value,
        bytes memory _hookMetadata,
        address _hook
    ) internal virtual override returns (bytes32) {
        // Run default logic for HypERC20 token.
        HypERC20._transferFromSender(_amount);

        // This is where the custom logic is added
        // to bind the metadata into the Hyperlane message.
        //
        // It is designed with inspiration from the CCTP token bridge contract:
        // https://github.com/hyperlane-xyz/hyperlane-monorepo/blob/%40hyperlane-xyz/core%409.0.9/solidity/contracts/token/TokenBridgeCctp.sol#L196-L231
        bytes32 payloadHash = ots.getPendingPayloadHash();

        // Depending if the payload hash is populated or not,
        // we are building the corresponding token messages to be sent
        // via the Warp route.
        bytes memory _tokenMessage;
        if (payloadHash != bytes32(0)) {
            _tokenMessage = TokenMessage.format(
                _recipient,
                _amount,
                abi.encodePacked(payloadHash)
            );
        } else {
            _tokenMessage = TokenMessage.format(
                _recipient,
                _amount
            );
        }

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
