package forwarding

import (
	"context"
	"fmt"

	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/noble-assets/orbiter/controller"
	"github.com/noble-assets/orbiter/types"
	forwardingtypes "github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
)

var _ types.ControllerForwarding = &HyperlaneController{}

// HyperlaneController is the forwarding controller for the Hyperlane protocol.
type HyperlaneController struct {
	*controller.BaseController[core.ProtocolID]
	logger  log.Logger
	handler forwardingtypes.HyperlaneHandler
}

// NewHyperlaneController returns a validated instance of the Hyperlane
// controller.
func NewHyperlaneController(
	l log.Logger,
	handler forwardingtypes.HyperlaneHandler,
) (*HyperlaneController, error) {
	if l == nil {
		return nil, core.ErrNilPointer.Wrap("logger cannot be nil")
	}

	b, err := controller.NewBase(core.PROTOCOL_HYPERLANE)
	if err != nil {
		return nil, errorsmod.Wrap(err, "error creating base controller for hyperlane controller")
	}
	c := HyperlaneController{
		BaseController: b,
		logger:         l,
		handler:        handler,
	}

	return &c, c.Validate()
}

// Validate returns an error if any of the Hyperlane controller's field is not valid.
func (c *HyperlaneController) Validate() error {
	if c.logger == nil {
		return core.ErrNilPointer.Wrap("logger is required for the Hyperlane controller")
	}
	if c.BaseController == nil {
		return core.ErrNilPointer.Wrap("base controller is required for the Hyperlane controller")
	}
	if c.handler == nil {
		return core.ErrNilPointer.Wrap("server is required for the Hyperlance controller")
	}
	return nil
}

// HandlePacket implements types.ControllerForwarding.
func (c *HyperlaneController) HandlePacket(ctx context.Context, p *types.ForwardingPacket) error {
	attr, err := c.ExtractAttributes(p.Forwarding)
	if err != nil {
		return core.ErrInvalidAttributes.Wrap(err.Error())
	}

	err = c.ValidateForwarding(ctx, p.TransferAttributes, attr)
	if err != nil {
		return core.ErrValidation.Wrap(err.Error())
	}

	err = c.executeForwarding(ctx, p.TransferAttributes, attr, p.Forwarding.PassthroughPayload)
	if err != nil {
		return errorsmod.Wrap(err, "Hyperlane controller execution error")
	}

	return nil
}

// ExtractAttributes returns the hyperlane attributes from the forwarding
// or an error.
func (c *HyperlaneController) ExtractAttributes(
	f *core.Forwarding,
) (*forwardingtypes.HypAttributes, error) {
	attr, err := f.CachedAttributes()
	if err != nil {
		return nil, errorsmod.Wrap(err, "error extracting cached attributes")
	}

	hypAttr, ok := attr.(*forwardingtypes.HypAttributes)
	if !ok {
		return nil, sdkerrors.ErrInvalidType.Wrapf(
			"expected %T, got %T",
			&forwardingtypes.HypAttributes{},
			attr,
		)
	}

	return hypAttr, nil
}

// ValidateForwarding checks whether the attributes received for execute a forwarding are
// valid or not.
func (c *HyperlaneController) ValidateForwarding(
	ctx context.Context,
	tAttr *types.TransferAttributes,
	hAttr *forwardingtypes.HypAttributes,
) error {
	tokenId := hyperlaneutil.HexAddress(hAttr.GetTokenId())
	req := warptypes.QueryTokenRequest{
		Id: tokenId.String(),
	}
	resp, err := c.handler.Token(ctx, &req)
	if err != nil {
		return errorsmod.Wrap(err, "invalid Hyperlane forwarding")
	}

	if resp.Token.OriginDenom != tAttr.DestinationDenom() {
		return fmt.Errorf(
			"invalid forwarding token, wanted %s, got %s",
			resp.Token.OriginDenom, tAttr.DestinationDenom(),
		)
	}

	return nil
}

// executeForwarding initiate an Hyperlane cross-chain transfer.
func (c *HyperlaneController) executeForwarding(
	ctx context.Context,
	tAttr *types.TransferAttributes,
	hAttr *forwardingtypes.HypAttributes,
	passthroughPayload []byte,
) error {
	hookID := hyperlaneutil.HexAddress(hAttr.CustomHookId)
	_, err := c.handler.RemoteTransfer(ctx, &warptypes.MsgRemoteTransfer{
		Sender:             core.ModuleAddress.String(),
		TokenId:            hyperlaneutil.HexAddress(hAttr.GetTokenId()),
		DestinationDomain:  hAttr.DestinationDomain,
		Recipient:          hyperlaneutil.HexAddress(hAttr.GetRecipient()),
		Amount:             tAttr.DestinationAmount(),
		CustomHookId:       &hookID,
		GasLimit:           hAttr.GasLimit,
		MaxFee:             hAttr.GetMaxFee(),
		CustomHookMetadata: string(passthroughPayload),
	})
	if err != nil {
		return errorsmod.Wrap(err, "error executing Hyperlane forwarding")
	}

	return nil
}
