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

func RegisterMsgServers(cfg module.Configurator, k *Keeper) {
	forwardertypes.RegisterMsgServer(cfg.MsgServer(), forwarder.NewMsgServer(k.Forwarder(), k))
	executortypes.RegisterMsgServer(cfg.MsgServer(), executor.NewMsgServer(k.Executor(), k))
	adaptertypes.RegisterMsgServer(cfg.MsgServer(), adapter.NewMsgServer(k.Adapter(), k))
}
