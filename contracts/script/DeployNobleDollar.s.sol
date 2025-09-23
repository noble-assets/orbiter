//SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.24;

import { console } from "forge-std/console.sol";
import { Script } from "forge-std/Script.sol";
import { NobleDollar } from "../src/usdn/USDNHyperlaneOrbiter.sol";
import { TransparentUpgradeableProxy } from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

contract DeployNobleDollar is Script {
    function run() external {
        address mailbox = vm.envAddress("MAILBOX");
        require(mailbox != address(0), "hyperlane mailbox address not set");

        address proxyAdmin = vm.envAddress("PROXYADMIN");
        require(proxyAdmin != address(0), "proxy admin address not set");

        vm.startBroadcast();

        // Deploy the implementation behind a proxy.
        NobleDollar implementation = new NobleDollar(mailbox);
        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(implementation),
            proxyAdmin,
            abi.encodeWithSelector(
                NobleDollar.initialize.selector,
                address(0),
                address(0)
            )
        );

        // Get the proxy implementation using the NobleDollar interface.
        NobleDollar nd = NobleDollar(address(proxy));

        // Sanity check that the minting during initialization was successful.
        uint256 minted = nd.balanceOf(msg.sender);
        require(minted != 0, "expected balance to have been minted for msg sender");

        console.log("deployed NobleDollar as proxy at ", address(nd));

        vm.stopBroadcast();
    }
}