//SPDX-License-Identifer: BUSL-1.1
pragma solidity ^0.8.24;

import { console } from "forge-std/console.sol";
import { Script } from "forge-std/Script.sol";
import { HyperlaneEntrypoint } from "../src/HyperlaneEntrypoint.sol";
import { OrbiterTransientStore } from "../src/usdn/OrbiterTransientStore.sol";
import { NobleDollar } from "../src/usdn/USDNHyperlaneOrbiter.sol";

contract DeployHyperlaneEntrypoint is Script {
    function run() external {
        address nobleDollar = vm.envAddress("NOBLEDOLLAR");
        require(nobleDollar != address(0), "noble dollar address not set");

        vm.startBroadcast();

        OrbiterTransientStore ots = new OrbiterTransientStore();
        NobleDollar nd = NobleDollar(nobleDollar);
        nd.setOTS(address(ots));

        HyperlaneEntrypoint he = new HyperlaneEntrypoint(address(ots));

        ots.transferOwnership(address(he));
        require(ots.owner() == address(he), "expected ownership to be transferred");

        console.log("deployed hyperlane entrypoint at: ", address(he));

        vm.stopBroadcast();
    }
}