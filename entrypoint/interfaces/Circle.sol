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

interface IFiatToken {
    // https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.4.0/contracts/token/ERC20/IERC20.sol#L52-L67
    function approve(address spender, uint256 value) external returns (bool);

    // https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.4.0/contracts/token/ERC20/IERC20.sol#L69-L78
    function transferFrom(address from, address to, uint256 value) external returns (bool);

    // https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.4.0/contracts/token/ERC20/extensions/IERC20Permit.sol#L43-L74
    function permit(address owner, address spender, uint256 value, uint256 deadline, uint8 v, bytes32 r, bytes32 s) external;
}

interface ITokenMessenger {
    // https://github.com/circlefin/evm-cctp-contracts/blob/1ddc5057e2a686194d481d04239387cf095ec760/src/TokenMessenger.sol#L187-L227
    function depositForBurnWithCaller(uint256 amount, uint32 destinationDomain, bytes32 mintRecipient, address burnToken, bytes32 destinationCaller) external returns (uint64 nonce);
}

interface IMessageTransmitter {
    // https://github.com/circlefin/evm-cctp-contracts/blob/1ddc5057e2a686194d481d04239387cf095ec760/src/interfaces/IRelayer.sol#L37-L55
    function sendMessageWithCaller(uint32 destinationDomain, bytes32 recipient, bytes32 destinationCaller, bytes calldata messageBody) external returns (uint64);
}
