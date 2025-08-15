package adapter

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func TxCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "UpdateParams",
			Use:       "update-params [params]",
			Short:     "Update the component's parameters",
			Long: `Update the component's parameters. This message can only be sent 
by the authority of the module.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "params"},
			},
		},
	}
}

func QueryCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "Params",
			Use:       "params",
			Short:     "Get the component's parameters",
		},
	}
}
