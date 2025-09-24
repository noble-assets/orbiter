// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.24;

import { OrbiterTransientStore } from "./OrbiterTransientStore.sol";

import { HypERC20 } from "@hyperlane/token/HypERC20.sol";
import { TokenMessage } from "@hyperlane/token/libs/TokenMessage.sol";

contract OrbiterHypERC20 is HypERC20 {
    OrbiterTransientStore private ots;

    constructor(
        uint8 _decimals,
        uint256 _scale,
        address _mailbox
    ) HypERC20(_decimals, _scale, _mailbox) {}

    /**
     * @notice Initializes the contract by calling the initialization logic
     * of the HypERC20 contract and setting the Orbiter transient store.
     */
    function initialize(
        uint256 _totalSupply,
        string memory _name,
        string memory _symbol,
        address _hook,
        address _interchainSecurityModule,
        address _owner,
        address _orbiterTransientStore
    ) public virtual initializer {
        super.initialize(
            _totalSupply,
            _name,
            _symbol,
            _hook,
            _interchainSecurityModule,
            _owner
        );

        ots = OrbiterTransientStore(_orbiterTransientStore);
    }

    /*
     * @notice Returns the address of the Orbiter transient store that's
     * associated with this contract.
     */
    function getOrbiterTransientStoreAddress() external view returns (address) {
        return address(ots);
    }

    /*
     * @notice Overrides the standard implementation of HypERC20 to support
     * passing payloads within the same transaction using the Orbiter
     * transient store.
     */
    function _transferRemote(
        uint32 _destination,
        bytes32 _recipient,
        uint256 _amount,
        uint256 _value,
        bytes memory _hookMetadata,
        address _hook
    ) internal virtual override returns (bytes32) {
        // Run default logic for HypERC20 token.
        HypERC20._transferFromSender(_amount);

        // This is where the custom logic is added
        // to bind the metadata into the Hyperlane message.
        //
        // It is designed with inspiration from the CCTP token bridge contract:
        // https://github.com/hyperlane-xyz/hyperlane-monorepo/blob/%40hyperlane-xyz/core%409.0.9/solidity/contracts/token/TokenBridgeCctp.sol#L196-L231
        require(
            address(ots) != address(0),
            "orbiter transient store must be set"
        );
        bytes32 payloadHash = ots.getPendingPayloadHash();

        // Depending if the payload hash is populated or not,
        // we are building the corresponding token messages to be sent
        // via the Warp route.
        bytes memory _tokenMessage;
        if (payloadHash != bytes32(0)) {
            _tokenMessage = TokenMessage.format(
                _recipient,
                _amount,
                abi.encodePacked(payloadHash)
            );
        } else {
            _tokenMessage = TokenMessage.format(
                _recipient,
                _amount
            );
        }

        bytes32 messageID = _Router_dispatch(
            _destination,
            _value,
            _tokenMessage,
            _hookMetadata,
            _hook
        );

        emit SentTransferRemote(
            _destination,
            _recipient,
            _amount
        );

        return messageID;
    }
}
