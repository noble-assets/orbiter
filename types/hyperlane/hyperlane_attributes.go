package hyperlane

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"encoding/binary"
	"fmt"
	hyperlaneutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/noble-assets/orbiter/types/controller/forwarding"
	"github.com/noble-assets/orbiter/types/core"
	"math/big"
)

// Hyperlane attributes constants
//
// TODO: cross-check assumptions here, correct length used for everything?
// TODO: can probably be removed in favor of ABI unpacking
const (
	tokenIDLength      = hyperlaneutil.HEX_ADDRESS_LENGTH
	destDomainLength   = 4
	recipientLength    = hyperlaneutil.HEX_ADDRESS_LENGTH
	customHookIDLength = hyperlaneutil.HEX_ADDRESS_LENGTH
	gasLimitLength     = 32 // uint256
	feeAmountLength    = 32 // uint256
	feeDenomLength     = 32 // just a random assumption / requirement we can put in place?

	hypAttrMinLength = tokenIDLength + destDomainLength + recipientLength + customHookIDLength + gasLimitLength + feeAmountLength + feeDenomLength
)

// unpackHypAttributes does the manual unpacking of Hyperlane orbiter attributes from a bytes payload.
//
// TODO: move into common utils? could also be used for the CCTP adapter
func unpackHypAttributes(bz []byte) (core.ForwardingAttributes, error) {
	if len(bz) < hypAttrMinLength {
		return nil, fmt.Errorf("minimum length for hyperlane attributes is %d; got %d", hypAttrMinLength, len(bz))
	}

	offset := 0

	tokenID := bz[offset:tokenIDLength]
	offset += tokenIDLength

	destDomain := binary.BigEndian.Uint32(bz[offset : offset+destDomainLength])
	offset += destDomainLength

	recipient := bz[offset : offset+recipientLength]
	offset += recipientLength

	customHookID := bz[offset:customHookIDLength]
	offset += customHookIDLength

	gasLimit := new(big.Int).SetBytes(bz[offset : offset+gasLimitLength])
	offset += gasLimitLength

	maxFeeAmount := new(big.Int).SetBytes(bz[offset : offset+feeAmountLength])
	offset += feeAmountLength

	maxFeeDenom := string(bz[offset : offset+feeDenomLength])
	offset += feeDenomLength

	// since this is of variable length, we unpack this last
	customHookMetadata := string(bz[offset:])

	maxFee := sdk.Coin{Denom: maxFeeDenom, Amount: sdkmath.NewIntFromBigInt(maxFeeAmount)}
	if err := maxFee.Validate(); err != nil {
		return nil, errorsmod.Wrap(err, "invalid max fee provided")
	}

	return forwarding.NewHyperlaneAttributes(
		tokenID,
		destDomain,
		recipient,
		customHookID,
		customHookMetadata,
		sdkmath.NewIntFromBigInt(gasLimit),
		maxFee,
	)
}
