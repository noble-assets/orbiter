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

import {IFiatToken, ITokenMessenger, IMessageTransmitter} from "../../src/interfaces/Circle.sol";

contract MockFiatToken is IFiatToken {
    mapping(address => mapping(address => uint256)) public allowances;
    mapping(address => uint256) public balances;

    function approve(address spender, uint256 value) external returns (bool) {
        allowances[msg.sender][spender] = value;
        return true;
    }

    function transferFrom(address from, address to, uint256 value) external returns (bool) {
        require(balances[from] >= value, "Insufficient balance");
        require(allowances[from][msg.sender] >= value, "Insufficient allowance");

        balances[from] -= value;
        balances[to] += value;
        allowances[from][msg.sender] -= value;

        return true;
    }

    function permit(address owner, address spender, uint256 value, uint256, uint8, bytes32, bytes32)
        external
    {
        allowances[owner][spender] = value;
    }

    function mint(address to, uint256 amount) external {
        balances[to] += amount;
    }
}

contract MockMessageTransmitter is IMessageTransmitter {
    uint64 public nonce;

    function sendMessageWithCaller(uint32, bytes32, bytes32, bytes calldata)
        external
        returns (uint64)
    {
        return nonce++;
    }
}

contract MockTokenMessenger is ITokenMessenger {
    IMessageTransmitter public immutable MESSAGE_TRANSMITTER;
    uint64 public nonce;

    constructor(address _messageTransmitter) {
        MESSAGE_TRANSMITTER = IMessageTransmitter(_messageTransmitter);
    }

    function localMessageTransmitter() external view returns (IMessageTransmitter) {
        return MESSAGE_TRANSMITTER;
    }

    function depositForBurnWithCaller(uint256, uint32, bytes32, address, bytes32)
        external
        returns (uint64)
    {
        return nonce++;
    }
}
