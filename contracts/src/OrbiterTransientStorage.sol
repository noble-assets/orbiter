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
pragma solidity ^0.8.28;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

/// @title Orbiter Transient Storage
/// @author Noble Core Team
/// @notice Holds pending payload hashes of transfers that should be executed
/// through the Orbiter.
contract OrbiterTransientStorage is Ownable {
    bytes32 transient private pendingPayloadHash;

    /// @notice Initializes the contract by setting the owner to be the Orbiter gateway contract.
    /// @param _gateway The address of the associated Orbiter gateway contract, which is set as the owner.
    constructor(address _gateway) Ownable() {
        transferOwnership(_gateway);
    }

    /// @notice Retrieves the currently pending payload hash.
    /// @return pendingPayloadHash The currently pending payload hash.
    function getPendingPayloadHash() external view returns (bytes32) {
        return pendingPayloadHash;
    }

    /// @notice Sets a new pending payload hash to the transient storage.
    /// @param payloadHash The new payload hash to store.
    function setPendingPayloadHash(bytes32 payloadHash) external onlyOwner {
        pendingPayloadHash = payloadHash;
    }
}
