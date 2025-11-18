package noble_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/noble-assets/orbiter/v2/e2e/refactor/noble"
)

func TestBuilder(t *testing.T) {
	ctx := context.Background()
	builder := noble.NewBuilder(ctx)
	require.NotNil(t, builder)

	builder = noble.NewBuilder(
		ctx,
		noble.WithCircleAccounts(),
		noble.WithOrbiterLogLevel("trace"),
	)
	require.NotNil(t, builder)
}
