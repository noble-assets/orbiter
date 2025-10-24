//SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.24;

import { console } from "forge-std/console.sol";
import { Script } from "forge-std/Script.sol";

import { OrbiterHypERC20 } from "../src/OrbiterHypERC20.sol";

import { HypERC20 } from "@hyperlane/token/HypERC20.sol";
import { TransparentUpgradeableProxy } from "@openzeppelin/contracts/proxy/transparent/TransparentUpgradeableProxy.sol";

contract DeployOrbiterHypERC20 is Script {
    function run() external {
        address mailbox = vm.envAddress("MAILBOX");
        require(mailbox != address(0), "hyperlane mailbox address not set");

        address proxyAdmin = vm.envAddress("PROXYADMIN");
        require(proxyAdmin != address(0), "proxy admin address not set");

        address gateway = vm.envAddress("GATEWAY");
        require(gateway != address(0), "gateway address not set");

        uint8 decimals = 6;
        uint256 scale = 1;
        uint256 initialSupply = 2e6;
        string memory name = "Example";
        string memory symbol = "XMPL";
        address hook = address(0);
        address ism = address(0);

        vm.startBroadcast();

        // Deploy the implementation behind a proxy.
        OrbiterHypERC20 implementation = new OrbiterHypERC20(
            decimals,
            scale,
            mailbox
        );

        TransparentUpgradeableProxy proxy = new TransparentUpgradeableProxy(
            address(implementation),
            proxyAdmin,
            abi.encodeWithSelector(
                HypERC20.initialize.selector,
                initialSupply,
                name,
                symbol,
                hook,
                ism,
                msg.sender
            )
        );

        // Get the proxy implementation using the NobleDollar interface.
        OrbiterHypERC20 token = OrbiterHypERC20(address(proxy));

        // Sanity check that the minting during initialization was successful.
        uint256 minted = token.balanceOf(msg.sender);
        require(minted != 0, "expected balance to have been minted for msg sender");

        console.log("deployed proxy for OrbiterHypERC20 at ", address(token));

        vm.stopBroadcast();
    }
}