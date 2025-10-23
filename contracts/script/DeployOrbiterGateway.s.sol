//SPDX-License-Identifer: BUSL-1.1
pragma solidity ^0.8.24;

import { console } from "forge-std/console.sol";
import { Script } from "forge-std/Script.sol";

import { OrbiterGateway } from "../src/OrbiterGateway.sol";

contract DeployOrbiterGateway is Script {
    function run() external {
        uint32 nobleDomain = 1;

        vm.startBroadcast();

        OrbiterGateway gw = new OrbiterGateway(nobleDomain);
        console.log("deployed hyperlane gateway at: ", address(gw));

        vm.stopBroadcast();
    }
}