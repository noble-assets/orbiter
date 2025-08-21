package forwarding

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/noble-assets/orbiter/types/core"
)

type HypInputs struct {
	destinationDomain uint32
	externalRecipient []byte
	customHookID      []byte
	gasLimit          math.Int
	maxFee            sdk.Coin
}

func NewHyperlaneAttributes(i *HypInputs) (*HypAttributes, error) {
	if i == nil {
		return nil, errorsmod.Wrap(core.ErrNilPointer, "hyperlane inputs are not set")
	}

	attr := HypAttributes{
		DestinationDomain: i.destinationDomain,
		ExternalRecipient: i.externalRecipient,
		CustomHookId:      i.customHookID,
		GasLimit:          i.gasLimit,
		MaxFee:            i.maxFee,
	}

	return &attr, attr.Validate()
}

func (a *HypAttributes) Validate() error {
	return nil
}

var _ core.ForwardingAttributes = &HypAttributes{}

func (a *HypAttributes) CounterpartyID() string {
	return fmt.Sprintf("%d", a.GetDestinationDomain())
}

// NewHyperlaneForwarding returns a reference to a validated Hyperlane forwarding.
func NewHyperlaneForwarding(i *HypInputs, passthroughPayload []byte) (*core.Forwarding, error) {
	attributes, err := NewHyperlaneAttributes(i)
	if err != nil {
		return nil, err
	}

	return core.NewForwarding(core.PROTOCOL_HYPERLANE, attributes, passthroughPayload)
}
