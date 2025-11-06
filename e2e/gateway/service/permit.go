package service

import "math/big"

func (s *Service) SignerUSDCNonce() (*big.Int, error) {
	return s.usdc.Instance().Nonces(nil, s.signer.Address())
}

func (s *Service) USDCDomainSeparator() ([32]byte, error) {
	return s.usdc.Instance().DOMAINSEPARATOR(nil)
}
