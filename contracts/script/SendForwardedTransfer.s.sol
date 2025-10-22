//SPDX-License-Identifer: BUSL-1.1
pragma solidity ^0.8.24;

import { console } from "forge-std/console.sol";
import { Script } from "forge-std/Script.sol";
import { OrbiterGateway } from "../src/OrbiterGateway.sol";

contract SendForwardedTransfer is Script {
    function run() external {
        address gateway = vm.envAddress("GATEWAY");
        require(gateway != address(0), "orbiter gateway not set");

        address tokenAddress = vm.envAddress("NOBLEDOLLAR");
        require(tokenAddress != address(0), "noble dollar address not set");

        uint32 destinationDomain = 1;
        bytes32 recipient = bytes32(0);
        uint256 amount = 123;
        bytes memory payload = abi.encodePacked(uint256(10203040));

        vm.startBroadcast();

        OrbiterGateway gw = OrbiterGateway(gateway);
        bytes32 messageID = gw.sendForwardedTransfer(
            tokenAddress,
            recipient,
            amount,
            payload
        );

        console.log("sent message ID: ");
        console.logBytes32(messageID);

        vm.stopBroadcast();
    }
}