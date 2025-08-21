package forwarding

import (
	"context"

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

type HyperlaneController struct {
	*controller.BaseController[core.ProtocolID]

	logger log.Logger
	server forwardingtypes.HyperlaneMsgServer
}

func NewHyperlaneController(
	l log.Logger,
	msgServer forwardingtypes.HyperlaneMsgServer,
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
		server:         msgServer,
	}

	return &c, c.Validate()
}

func (c *HyperlaneController) Validate() error {
	if c.logger == nil {
		return core.ErrNilPointer.Wrap("logger is required for the Hyperlane controller")
	}
	if c.BaseController == nil {
		return core.ErrNilPointer.Wrap("base controller is required for the Hyperlane controller")
	}
	if c.server == nil {
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

	err = c.ValidateAttributes(attr)
	if err != nil {
		return core.ErrValidation.Wrap(err.Error())
	}

	err = c.executeForwarding(ctx, p.TransferAttributes, attr)
	if err != nil {
		return errorsmod.Wrap(err, "Hyperlane controller execution error")
	}

	return nil
}

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

func (c *HyperlaneController) ValidateAttributes(
	a *forwardingtypes.HypAttributes,
) error {
	return nil
}

func (c *HyperlaneController) executeForwarding(
	ctx context.Context,
	tAttr *types.TransferAttributes,
	hAttr *forwardingtypes.HypAttributes,
) error {
	hookID := hyperlaneutil.HexAddress(hAttr.CustomHookId)
	msg := warptypes.MsgRemoteTransfer{
		Sender:             core.ModuleAddress.String(),
		TokenId:            hyperlaneutil.HexAddress{},
		DestinationDomain:  hAttr.DestinationDomain,
		Recipient:          hyperlaneutil.HexAddress(hAttr.ExternalRecipient),
		Amount:             tAttr.DestinationAmount(),
		CustomHookId:       &hookID,
		GasLimit:           hAttr.GasLimit,
		MaxFee:             hAttr.GetMaxFee(),
		CustomHookMetadata: "",
	}
	_, err := c.server.RemoteTransfer(ctx, &msg)
	if err != nil {
		return err
	}

	return nil
}
