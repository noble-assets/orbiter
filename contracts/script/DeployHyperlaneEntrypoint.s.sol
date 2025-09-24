//SPDX-License-Identifer: BUSL-1.1
pragma solidity ^0.8.24;

import { console } from "forge-std/console.sol";
import { Script } from "forge-std/Script.sol";

import { HyperlaneEntrypoint } from "../src/HyperlaneEntrypoint.sol";

contract DeployHyperlaneEntrypoint is Script {
    function run() external {
        vm.startBroadcast();

        HyperlaneEntrypoint he = new HyperlaneEntrypoint();
        console.log("deployed hyperlane entrypoint at: ", address(he));

        vm.stopBroadcast();
    }
}