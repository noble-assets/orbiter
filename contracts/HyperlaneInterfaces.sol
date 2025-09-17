/*
 * @dev Contains the required interface to send tokens through a Hyperlane Warp Route.
 */
interface ITokenRouter {
    // https://github.com/hyperlane-xyz/hyperlane-monorepo/blob/%40hyperlane-xyz/core%409.0.9/solidity/contracts/token/libs/TokenRouter.sol#L45-L61
    function transferRemote(
        uint32 _destination,
        bytes32 _recipient,
        uint256 _amountOrId
    ) external payable virtual returns (
        bytes32 messageId
    );

    // TODO: add option for transfer with hook metadata?
}

/*
 * @dev Contains the required interface to dispatch messages using the Hyperlane protocol.
 */
interface IMailbox {
    // https://github.com/hyperlane-xyz/hyperlane-monorepo/blob/%40hyperlane-xyz/core%409.0.9/solidity/contracts/Mailbox.sol#L102-L123
    function dispatch(
        uint32 _destinationDomain,
        bytes32 _recipientAddress,
        bytes calldata _messageBody
    ) external payable override returns (bytes32);

    // TODO: add option for dispatch with hook metadata?
}
