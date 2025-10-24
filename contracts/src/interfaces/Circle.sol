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
    // https://github.com/circlefin/stablecoin-evm/blob/f2f8b3bb1659e3f1cf23ead72d5cdf58a2f4ebfe/contracts/v2/EIP712Domain.sol#L31-L37
    // forge-lint: disable-next-line(mixed-case-function)
    function DOMAIN_SEPARATOR() external view returns (bytes32);

    // https://github.com/circlefin/stablecoin-evm/blob/f2f8b3bb1659e3f1cf23ead72d5cdf58a2f4ebfe/contracts/v2/EIP2612.sol#L31-L33
    // forge-lint: disable-next-line(mixed-case-function)
    function PERMIT_TYPEHASH() external view returns (bytes32);

    // https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.4.0/contracts/token/ERC20/IERC20.sol#L52-L67
    function approve(address spender, uint256 value) external returns (bool);

    // https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.4.0/contracts/token/ERC20/IERC20.sol#L29-L32
    function balanceOf(address account) external view returns (uint256);

    // https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.4.0/contracts/token/ERC20/IERC20.sol#L69-L78
    function transferFrom(address from, address to, uint256 value) external returns (bool);

    // https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.4.0/contracts/token/ERC20/extensions/IERC20Permit.sol#L76-L83
    function nonces(address owner) external view returns (uint256);

    // https://github.com/OpenZeppelin/openzeppelin-contracts/blob/v5.4.0/contracts/token/ERC20/extensions/IERC20Permit.sol#L43-L74
    function permit(
        address owner,
        address spender,
        uint256 value,
        uint256 deadline,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external;
}

interface IMessageTransmitter {
    // https://github.com/circlefin/evm-cctp-contracts/blob/4061786a5726bc05f99fcdb53b0985599f0dbaf7/src/MessageTransmitter.sol#L81-L82
    function nextAvailableNonce() external view returns (uint64);

    // https://github.com/circlefin/evm-cctp-contracts/blob/4061786a5726bc05f99fcdb53b0985599f0dbaf7/src/interfaces/IRelayer.sol#L37-L55
    function sendMessageWithCaller(
        uint32 destinationDomain,
        bytes32 recipient,
        bytes32 destinationCaller,
        bytes calldata messageBody
    ) external returns (uint64);
}

interface ITokenMessenger {
    // https://github.com/circlefin/evm-cctp-contracts/blob/4061786a5726bc05f99fcdb53b0985599f0dbaf7/src/TokenMessenger.sol#L103-L104
    function localMessageTransmitter() external view returns (IMessageTransmitter);

    // https://github.com/circlefin/evm-cctp-contracts/blob/4061786a5726bc05f99fcdb53b0985599f0dbaf7/src/TokenMessenger.sol#L153-L185
    function depositForBurnWithCaller(
        uint256 amount,
        uint32 destinationDomain,
        bytes32 mintRecipient,
        address burnToken,
        bytes32 destinationCaller
    ) external returns (uint64 nonce);
}
