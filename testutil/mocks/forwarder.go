package mocks

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	"orbiter.dev/keeper/component/forwarder"
)

func NewForwarderComponent(tb testing.TB) (*forwarder.Forwarder, *Dependencies) {
	tb.Helper()

	deps := NewDependencies(tb)

	sb := collections.NewSchemaBuilder(deps.StoreService)

	f, err := forwarder.New(
		deps.EncCfg.Codec,
		sb,
		deps.Logger,
		NewBankKeeper(),
	)
	require.NoError(tb, err)
	_, err = sb.Build()
	require.NoError(tb, err)

	return f, &deps
}
