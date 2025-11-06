package service

import (
	"math/big"

	bind "github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	gethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/noble-assets/orbiter/e2e/gateway/types"
)

func (s *Service) DepositForBurnWithOrbiter(
	txOpts *bind.TransactOpts,
	amount, blocktimeDeadline *big.Int,
	permitSignature, orbiterPayload []byte,
) (*gethtypes.Transaction, error) {
	return s.gateway.Instance().
		DepositForBurnWithOrbiter(txOpts, amount, blocktimeDeadline, permitSignature, orbiterPayload)
}

func (s *Service) ParseDepositForBurnEvents(
	receipt *gethtypes.Receipt,
) ([]*types.OrbiterGatewayCCTPDepositForBurnWithOrbiter, error) {
	var events []*types.OrbiterGatewayCCTPDepositForBurnWithOrbiter

	for _, log := range receipt.Logs {
		if log.Address != s.GatewayAddress() {
			continue
		}

		event, err := s.gateway.Instance().ParseDepositForBurnWithOrbiter(*log)
		if err != nil {
			continue
		}

		events = append(events, event)
	}

	return events, nil
}
