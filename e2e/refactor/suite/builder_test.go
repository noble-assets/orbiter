package suite_test

import (
	"context"
	"testing"

	"github.com/noble-assets/orbiter/v2/e2e/refactor/noble"
	"github.com/noble-assets/orbiter/v2/e2e/refactor/suite"
)

func TestNewBuilder(t *testing.T) {
	ctx := context.Background()

	nobleBuilder := noble.NewBuilder(
		ctx,
		noble.WithCircleAccounts(),
		noble.WithOrbiterLogLevel("trace"),
	)

	suiteBuilder := suite.NewBuilder(t, nobleBuilder.BuilderOpt())
}
