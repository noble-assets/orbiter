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

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";

/*
 * @notice Holds pending payload hashes of transfers that should be executed
 * through the Orbiter.
 */
contract OrbiterTransientStore {
    bytes32 transient private pendingPayloadHash;

//    constructor() Ownable(msg.sender) {}
    constructor() {}

    function getPendingPayloadHash() external view returns (bytes32) {
        return pendingPayloadHash;
    }

    // TODO: this should be possible to be set by everyone? since it can only be set in the same transaction
//    function setPendingPayloadHash(bytes32 payloadHash) external onlyOwner {
    function setPendingPayloadHash(bytes32 payloadHash) external {
        pendingPayloadHash = payloadHash;
    }
}
