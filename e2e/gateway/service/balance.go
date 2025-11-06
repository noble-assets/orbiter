package service

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// BalanceAt retrieves the balance of an account at the current block.
func (s *Service) Balance(
	ctx context.Context,
	account common.Address,
) (*big.Int, error) {
	return s.client.BalanceAt(ctx, account, nil)
}

func (s *Service) SignerBalance(ctx context.Context) (*big.Int, error) {
	return s.Balance(ctx, s.signer.Address())
}

func (s *Service) USDCBalance(ctx context.Context, address common.Address) (*big.Int, error) {
	return s.usdc.Instance().BalanceOf(nil, address)
}

func (s *Service) SignerUSDCBalance(ctx context.Context) (*big.Int, error) {
	return s.usdc.Instance().BalanceOf(nil, s.signer.Address())
}
