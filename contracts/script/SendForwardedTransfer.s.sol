//SPDX-License-Identifer: BUSL-1.1
pragma solidity ^0.8.24;

import { console } from "forge-std/console.sol";
import { Script } from "forge-std/Script.sol";
import {HyperlaneEntrypoint} from "../src/HyperlaneEntrypoint.sol";

contract SendForwardedTransfer is Script {
    function run() external {
        address entrypoint = vm.envAddress("ENTRYPOINT");
        require(entrypoint != address(0), "entrypoint not set");

        address nobleDollar = vm.envAddress("NOBLEDOLLAR");
        require(entrypoint != address(0), "noble dollar address not set");

        uint32 destinationDomain = 1;
        bytes32 recipient = bytes32(0); // TODO: adjust to use identifier
        uint256 amount = 123;
        bytes32 payloadHash = bytes32(uint256(10203040)); // TODO: generate valid payload hash

        vm.startBroadcast();

        HyperlaneEntrypoint he = HyperlaneEntrypoint(entrypoint);
        bytes32 messageID = he.sendUSDNWithForwardThroughHyperlane(
            nobleDollar,
            destinationDomain,
            recipient,
            amount,
            payloadHash
        );

        console.log("sent message ID: ");
        console.logBytes32(messageID);

        vm.stopBroadcast();
    }
}