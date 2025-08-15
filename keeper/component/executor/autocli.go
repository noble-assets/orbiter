package executor

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func TxCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "PauseAction",
			Use:       "pause-action [action_id]",
			Short:     "Pause a specific action by ID",
			Long: `Pause a specific action by its identifier. This will prevent 
the action from being executed until it is unpaused.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "action_id"},
			},
		},
		{
			RpcMethod: "UnpauseAction",
			Use:       "unpause-action [action_id]",
			Short:     "Resume a paused action by ID",
			Long: `Resume execution of a previously paused action by its identifier. This 
will allow the action to be executed again.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "action_id"},
			},
		},
	}
}

func QueryCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "IsActionPaused",
			Use:       "is-action-paused [action_id]",
			Short:     "Check if an action is paused",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "action_id"},
			},
		},
		{
			RpcMethod: "PausedActions",
			Use:       "paused-actions",
			Short:     "Get all paused action IDs",
		},
	}
}
