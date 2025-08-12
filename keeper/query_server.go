package keeper

import (
	"github.com/cosmos/cosmos-sdk/types/module"

	"orbiter.dev/keeper/component/adapter"
	"orbiter.dev/keeper/component/executor"
	"orbiter.dev/keeper/component/forwarder"
	adaptertypes "orbiter.dev/types/component/adapter"
	executortypes "orbiter.dev/types/component/executor"
	forwardertypes "orbiter.dev/types/component/forwarder"
)

func RegisterQueryServers(cfg module.Configurator, k *Keeper) {
	forwardertypes.RegisterQueryServer(cfg.QueryServer(), forwarder.NewQueryServer(k.Forwarder()))
	executortypes.RegisterQueryServer(cfg.QueryServer(), executor.NewQueryServer(k.Executor()))
	adaptertypes.RegisterQueryServer(cfg.QueryServer(), adapter.NewQueryServer(k.Adapter()))
}
