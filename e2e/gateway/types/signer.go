package types

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type Signer struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewSigner(privateKeyHex string) (*Signer, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to cast public key to ECDSA public key")
	}

	return &Signer{
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(*publicKeyECDSA),
	}, nil
}

func (s *Signer) Address() common.Address {
	return s.address
}

func (s *Signer) Sign(bz []byte) ([]byte, error) {
	sign, err := crypto.Sign(bz, s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign digest: %w", err)
	}

	return sign, nil
}

func (s *Signer) CreateTransactor(chainID *big.Int) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(s.privateKey, chainID)
}

// ABIEncodeSignature returns the signature encoded with the standard Solidity ABI structure.
func ABIEncodeSignature(sign []byte) []byte {
	// Update the v value of the signature according to the expected signature in ethereum.
	V := sign[64]
	if V < 27 {
		V += 27
	}

	R := sign[0:32]
	S := sign[32:64]

	var packedSign []byte
	packedSign = append(packedSign, common.LeftPadBytes([]byte{V}, 32)...)
	packedSign = append(packedSign, R...)
	packedSign = append(packedSign, S...)

	return packedSign
}
