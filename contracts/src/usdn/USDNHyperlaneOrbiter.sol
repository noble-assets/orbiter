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
pragma solidity ^0.8.24;

import {HypERC20} from "@hyperlane/token/HypERC20.sol";
import {TokenMessage} from "@hyperlane/token/libs/TokenMessage.sol";

import {OrbiterTransientStore} from "./OrbiterTransientStore.sol";

import {IndexingMath} from "./utils/IndexingMath.sol";

import {UIntMath} from "./utils/UIntMath.sol";

/*

███╗   ██╗ ██████╗ ██████╗ ██╗     ███████╗
████╗  ██║██╔═══██╗██╔══██╗██║     ██╔════╝
██╔██╗ ██║██║   ██║██████╔╝██║     █████╗
██║╚██╗██║██║   ██║██╔══██╗██║     ██╔══╝
██║ ╚████║╚██████╔╝██████╔╝███████╗███████╗
╚═╝  ╚═══╝ ╚═════╝ ╚═════╝ ╚══════╝╚══════╝

██████╗  ██████╗ ██╗     ██╗      █████╗ ██████╗
██╔══██╗██╔═══██╗██║     ██║     ██╔══██╗██╔══██╗
██║  ██║██║   ██║██║     ██║     ███████║██████╔╝
██║  ██║██║   ██║██║     ██║     ██╔══██║██╔══██╗
██████╔╝╚██████╔╝███████╗███████╗██║  ██║██║  ██║
╚═════╝  ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝

*/

/**
 * @title  NobleDollar
 * @author John Letey <john@noble.xyz>
 * @notice ERC20 Noble Dollar.
 */
contract NobleDollar is HypERC20 {
    /// @notice Thrown when a user attempts to claim yield but has no claimable yield available.
    error NoClaimableYield();

    /// @notice Thrown when an invalid transfer to the contract is attempted.
    error InvalidTransfer();

    /// @notice The transient store contract to retrieve a pending Orbiter payload hash within the same
    /// smart contract call.
    OrbiterTransientStore orbiterTransientStore;

    /**
     * @notice Emitted when the index is updated due to yield accrual.
     * @param oldIndex The previous index value.
     * @param newIndex The new index value.
     * @param totalPrincipal The total principal amount at the time of update.
     * @param yieldAccrued The amount of yield that was accrued.
     */
    event IndexUpdated(uint128 oldIndex, uint128 newIndex, uint112 totalPrincipal, uint256 yieldAccrued);

    /**
     * @notice Emitted when yield is claimed by an account.
     * @param account The account that claimed the yield.
     * @param amount The amount of yield claimed.
     */
    event YieldClaimed(address indexed account, uint256 amount);

    /// @custom:storage-location erc7201:noble.storage.USDN
    struct USDNStorage {
        uint128 index;
        uint112 totalPrincipal;
        mapping(address account => uint112) principal;
    }

    // keccak256(abi.encode(uint256(keccak256("noble.storage.USDN")) - 1)) & ~bytes32(uint256(0xff))
    bytes32 private constant USDNStorageLocation = 0xccec1a0a356b34ea3899fbc248aeaeba5687659563a3acddccc6f1e8a5d84200;

    function _getUSDNStorage() private pure returns (USDNStorage storage $) {
        assembly {
            $.slot := USDNStorageLocation
        }
    }

    constructor(address mailbox_) HypERC20(6, 1, mailbox_) {
//        _disableInitializers();
    }

    function initialize(address hook_, address ism_) public virtual initializer {
//        super.initialize("Noble Dollar", "USDN", hook_, ism_, msg.sender);
        // TODO: added this for testing purposes to have tokens available
        super.initialize(uint256(2e18), "Noble Dollar", "USDN", hook_, ism_, msg.sender);

        _getUSDNStorage().index = IndexingMath.EXP_SCALED_ONE;
    }

    /// @dev Returns the current index used for yield calculations.
    function index() public view returns (uint128) {
        return _getUSDNStorage().index;
    }

    /// @dev Returns the amount of principal in existence.
    function totalPrincipal() public view returns (uint112) {
        return _getUSDNStorage().totalPrincipal;
    }

    /// @dev Returns the amount of principal owned for a given account.
    function principalOf(address account) public view returns (uint112) {
        return _getUSDNStorage().principal[account];
    }

    /**
     * @notice Returns the amount of yield claimable for a given account.
     * @dev Calculates claimable yield by comparing the expected balance (principal * current index)
     *      with the actual token balance. The yield represents the difference between what the
     *      account should have based on yield accrual and what they currently hold.
     *
     *      Formula: max(0, (principal * index / 1e12) - currentBalance)
     *
     *      Returns 0 if the current balance is greater than or equal to the expected balance,
     *      which can happen if the account has already claimed their yield or if no yield
     *      has accrued since their last interaction.
     *
     * @param account The address to check yield for.
     * @return The amount of yield claimable by the account.
     */
    function yield(address account) public view returns (uint256) {
        USDNStorage storage $ = _getUSDNStorage();

        uint256 expectedBalance = IndexingMath.getPresentAmountRoundedDown($.principal[account], $.index);

        uint256 currentBalance = balanceOf(account);

        return expectedBalance > currentBalance ? expectedBalance - currentBalance : 0;
    }

    /**
     * @notice Claims all available yield for the caller.
     * @dev Calculates the claimable yield based on the difference between the expected balance
     *      (principal * current index) and the actual token balance. Transfers the yield amount
     *      from the contract to the caller and emits a YieldClaimed event.
     * @custom:throws NoClaimableYield if the caller has no yield available to claim.
     * @custom:emits YieldClaimed when yield is successfully claimed.
     */
    function claim() public {
        uint256 amount = yield(msg.sender);

        if (amount == 0) revert NoClaimableYield();

        _update(address(this), msg.sender, amount);

        emit YieldClaimed(msg.sender, amount);
    }

    /*
     * @notice Sets the Orbiter transient store.
     */
    function setOTS(address _otsAddress) external {
        orbiterTransientStore = OrbiterTransientStore(_otsAddress);
    }

    /**
     * @notice Internal function that handles token transfers while managing principal accounting.
     * @dev Overrides the base ERC20 _update function to implement yield-bearing token mechanics.
     *      This function manages principal balances and index updates for different transfer scenarios:
     *
     *      1. Yield payout (from contract): No principal updates needed
     *      2. Yield accrual (to contract from zero address): Updates index based on new yield
     *      3. Regular transfers: Updates principal balances for both sender and recipient
     *      4. Minting (from zero address): Increases recipient's principal and total principal
     *      5. Burning (to zero address): Decreases sender's principal and total principal
     *
     * @param from The address tokens are transferred from (zero address for minting)
     * @param to The address tokens are transferred to (zero address for burning)
     * @param value The amount of tokens being transferred
     *
     * @custom:throws InvalidTransfer if attempting to transfer to the contract from a non-zero address
     * @custom:emits IndexUpdated when yield is accrued and the index is updated
     * @custom:security Principal is calculated using ceiling / floor division in favor of protocol.
     */
    function _update(address from, address to, uint256 value) internal virtual override {
        super._update(from, to, value);

        // Special case, no-op operation.
        if (from == address(0) && to == address(0)) return;

        // We don't want to perform any principal updates in the case of yield payout.
        if (from == address(this)) return;

        // We don't want to allow any other transfers to the yield account.
        if (from != address(0) && to == address(this)) revert InvalidTransfer();

        USDNStorage storage $ = _getUSDNStorage();

        // Distribute yield, derive new index from the adjusted total supply.
        // NOTE: We don't want to perform any principal updates in the case of yield accrual.
        if (to == address(this)) {

            if ($.totalPrincipal == 0) return;

            uint128 oldIndex = $.index;

            $.index = IndexingMath.getIndexRoundedDown(totalSupply(), $.totalPrincipal);

            emit IndexUpdated(oldIndex, $.index, $.totalPrincipal, value);

            return;
        }

        // Minting
        if (from == address(0)) {
            uint112 principalDown = IndexingMath.getPrincipalAmountRoundedDown(value, $.index);
            $.principal[to] += principalDown;
            $.totalPrincipal += principalDown;

            return;
        }

        // Safely round up principal in case of transfers or burning.
        uint112 fromPrincipal = $.principal[from];
        uint112 principalUp = IndexingMath.getSafePrincipalAmountRoundedUp(value, $.index, fromPrincipal);
        $.principal[from] = fromPrincipal - principalUp;

        if (to == address(0)) {
            // Burning
            $.totalPrincipal -= principalUp;
        } else {
            // Transfer
            $.principal[to] += principalUp;
        }
    }

    function _transferRemote(
        uint32 _destination,
        bytes32 _recipient,
        uint256 _amount,
        uint256 _value,
        bytes memory _hookMetadata,
        address _hook
    ) internal virtual override returns (bytes32 messageId) {
        // Run default logic for HypERC20 token.
        HypERC20._transferFromSender(_amount);

        // This is where the custom logic is added
        // to bind the metadata comes into play.
        //
        // It is designed with inspiration from the CCTP token bridge contract:
        // https://github.com/hyperlane-xyz/hyperlane-monorepo/blob/%40hyperlane-xyz/core%409.0.9/solidity/contracts/token/TokenBridgeCctp.sol#L196-L231
        //
        // TODO: this is currently requiring the OTS to be set, check if we always want this?
        require(
            address(orbiterTransientStore) != address(0),
            "orbiterTransientStore must be set"
        );
        bytes32 payloadHash = orbiterTransientStore.getPendingPayloadHash();

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

        messageId = _Router_dispatch(
            _destination,
            _value,
            _tokenMessage,
            _hookMetadata,
            _hook
        );

        emit SentTransferRemote(_destination, _recipient, _amount);
    }
}