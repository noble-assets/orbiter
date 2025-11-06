package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var Eip712VersionBytes = []byte("\x19\x01")

var PermitTypeHash = crypto.Keccak256Hash(
	[]byte("Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"),
)

// PermitMessage contains the information associated with a Permit.
type PermitMessage struct {
	Owner    common.Address
	Spender  common.Address
	Value    *big.Int
	Nonce    *big.Int
	Deadline *big.Int
}

// NewPermitMessage returns a reference to a permit message.
func NewPermitMessage(
	owner, spender common.Address,
	value, nonce, deadline *big.Int,
) *PermitMessage {
	return &PermitMessage{
		Owner:    owner,
		Spender:  spender,
		Value:    value,
		Nonce:    nonce,
		Deadline: deadline,
	}
}

// Permit contains the information required to create the digest to be
// signed to generate a valid permit.
type Permit struct {
	DomainSeparator common.Hash
	TypeHash        common.Hash
	Message         *PermitMessage
}

// NewPermit returns an instance of a structure containing all the permit data required
// by the EIP-712 structured data hashing and signing.
//
// Ref: https://eips.ethereum.org/EIPS/eip-712
func NewPermit(
	domainSeparator common.Hash,
	owner, spender common.Address,
	value, nonce, deadline *big.Int,
) Permit {
	return Permit{
		DomainSeparator: domainSeparator,
		TypeHash:        PermitTypeHash,
		Message:         NewPermitMessage(owner, spender, value, nonce, deadline),
	}
}

// Digest returns the Keccak256 digest of the permit. The permit messages fields are
// converted into the Solidity 32 bytes slot representation before hashing.
func (p *Permit) Digest() []byte {
	hashStruct := crypto.Keccak256(
		p.TypeHash.Bytes(),
		common.LeftPadBytes(p.Message.Owner.Bytes(), 32),
		common.LeftPadBytes(p.Message.Spender.Bytes(), 32),
		common.LeftPadBytes(p.Message.Value.Bytes(), 32),
		common.LeftPadBytes(p.Message.Nonce.Bytes(), 32),
		common.LeftPadBytes(p.Message.Deadline.Bytes(), 32),
	)

	return crypto.Keccak256(
		Eip712VersionBytes,
		p.DomainSeparator.Bytes(),
		hashStruct,
	)
}
