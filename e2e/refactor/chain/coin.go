package chain

import (
	"strings"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

var (
	UsdcMetadata = NewBankMetadata("usdc")
	IbcMetadata  = NewBankMetadata("ibc")
)

func NewBankMetadata(denom string) *banktypes.Metadata {
	microDenom := "u" + denom

	return &banktypes.Metadata{
		Description: strings.ToUpper(denom) + " Coin",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: microDenom, Exponent: 0},
			{Denom: denom, Exponent: 6},
		},
		Base:    microDenom,
		Display: denom,
		Name:    denom,
		Symbol:  denom,
	}
}
