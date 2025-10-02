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

import {IFiatToken, ITokenMessenger, IMessageTransmitter} from "./interfaces/Circle.sol";

/*

    ███╗   ██╗ ██████╗ ██████╗ ██╗     ███████╗     ██████╗ ██████╗ ██████╗ ██╗████████╗███████╗██████╗
    ████╗  ██║██╔═══██╗██╔══██╗██║     ██╔════╝    ██╔═══██╗██╔══██╗██╔══██╗██║╚══██╔══╝██╔════╝██╔══██╗
    ██╔██╗ ██║██║   ██║██████╔╝██║     █████╗      ██║   ██║██████╔╝██████╔╝██║   ██║   █████╗  ██████╔╝
    ██║╚██╗██║██║   ██║██╔══██╗██║     ██╔══╝      ██║   ██║██╔══██╗██╔══██╗██║   ██║   ██╔══╝  ██╔══██╗
    ██║ ╚████║╚██████╔╝██████╔╝███████╗███████╗    ╚██████╔╝██║  ██║██████╔╝██║   ██║   ███████╗██║  ██║
    ╚═╝  ╚═══╝ ╚═════╝ ╚═════╝ ╚══════╝╚══════╝     ╚═════╝ ╚═╝  ╚═╝╚═════╝ ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝

     ██████╗ ██████╗████████╗██████╗     ██╗    ██╗██████╗  █████╗ ██████╗ ██████╗ ███████╗██████╗
    ██╔════╝██╔════╝╚══██╔══╝██╔══██╗    ██║    ██║██╔══██╗██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗
    ██║     ██║        ██║   ██████╔╝    ██║ █╗ ██║██████╔╝███████║██████╔╝██████╔╝█████╗  ██████╔╝
    ██║     ██║        ██║   ██╔═══╝     ██║███╗██║██╔══██╗██╔══██║██╔═══╝ ██╔═══╝ ██╔══╝  ██╔══██╗
    ╚██████╗╚██████╗   ██║   ██║         ╚███╔███╔╝██║  ██║██║  ██║██║     ██║     ███████╗██║  ██║
     ╚═════╝ ╚═════╝   ╚═╝   ╚═╝          ╚══╝╚══╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝     ╚═╝     ╚══════╝╚═╝  ╚═╝

 */

contract CCTPWrapper {
    IFiatToken public immutable token;
    ITokenMessenger public immutable tokenMessenger;
    // TODO: Consider retrieving the messageTransmitter directly the tokenMessenger!
    IMessageTransmitter public immutable messageTransmitter;

    uint32  public constant DESTINATION_DOMAIN = 4;
    bytes32 public constant MINT_RECIPIENT     = 0x000000000000000000000000a197eb1a9bfe6143b2d6499897fc1e3c1cfacbb2;
    bytes32 public immutable destinationCaller;

    constructor(address token_, address tokenMessenger_, address messageTransmitter_, bytes32 destinationCaller_)  {
        token              = IFiatToken(token_);
        tokenMessenger     = ITokenMessenger(tokenMessenger_);
        messageTransmitter = IMessageTransmitter(messageTransmitter_);
        destinationCaller  = destinationCaller_;
    }

    function entrypoint(
        uint amount,
        uint permitDeadline,
        bytes calldata permitSignature,
        bytes32 payloadHash
    )
        external
    {
        (uint8 v, bytes32 r, bytes32 s) = abi.decode(permitSignature, (uint8, bytes32, bytes32));
        token.permit(msg.sender, address(this), amount, permitDeadline, v, r, s);
        token.transferFrom(msg.sender, address(this), amount);
        token.approve(address(tokenMessenger), amount);

        uint transferNonce = tokenMessenger.depositForBurnWithCaller(
            amount, DESTINATION_DOMAIN, MINT_RECIPIENT, address(token), destinationCaller
        );

        uint payloadNonce = messageTransmitter.sendMessageWithCaller(
            DESTINATION_DOMAIN, MINT_RECIPIENT, destinationCaller, abi.encodePacked(transferNonce, payloadHash)
        );

        // TODO: Emit an event!
    }
}
