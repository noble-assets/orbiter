package types

import (
	bind "github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
)

type Contract[T any] struct {
	address  common.Address
	instance *T
}

type InstanceContructor[T any] = func(common.Address, bind.ContractBackend) (*T, error)

func NewContract[T any](
	client bind.ContractBackend,
	addressHex string,
	con InstanceContructor[T],
) (*Contract[T], error) {
	address := common.HexToAddress(addressHex)
	instance, err := con(address, client)
	if err != nil {
		return nil, err
	}

	return &Contract[T]{
		address:  address,
		instance: instance,
	}, nil
}

func (c *Contract[T]) Address() common.Address {
	return c.address
}

func (c *Contract[T]) Instance() *T {
	return c.instance
}
