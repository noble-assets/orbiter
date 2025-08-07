package orbiter

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	orbiterv1 "orbiter.dev/api/v1"
	"orbiter.dev/types/component/executor"
	"orbiter.dev/types/component/forwarder"
)

func (AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: orbiterv1.Msg_ServiceDesc.ServiceName,
			SubCommands: map[string]*autocliv1.ServiceCommandDescriptor{
				"executor": {
					Service:              executor.Msg_serviceDesc.ServiceName,
					Short:                "Execute actions management commands",
					SubCommands:          map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions:    getExecutorRPCOptions(),
					EnhanceCustomCommand: true,
				},
				"forwarder": {
					Service:              forwarder.Msg_serviceDesc.ServiceName,
					Short:                "Protocol forwarding management commands",
					SubCommands:          map[string]*autocliv1.ServiceCommandDescriptor{},
					RpcCommandOptions:    getForwarderRPCOptions(),
					EnhanceCustomCommand: true,
				},
			},
		},
	}
}

func getExecutorRPCOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "PauseAction",
			Use:       "pause-action [id]",
			Short:     "Pause a specific action by ID",
			Long:      "Pause a specific action by its identifier. This will prevent the action from being executed until it is unpaused.",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "action_id"},
			},
		},
		{
			RpcMethod: "UnpauseAction",
			Use:       "unpause-action [id]",
			Short:     "Resume a paused action by ID",
			Long:      "Resume execution of a previously paused action by its identifier. This will allow the action to be executed again.",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "action_id"},
			},
		},
	}
}

func getForwarderRPCOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "PauseProtocol",
			Use:       "pause-protocol [id]",
			Short:     "Pause an entire forwarding protocol",
			Long:      "Pause an entire forwarding protocol by its identifier. This will disable all forwarding operations for the specified protocol until it is unpaused.",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "UnpauseProtocol",
			Use:       "unpause-protocol [id]",
			Short:     "Resume a paused forwarding protocol",
			Long:      "Resume operations for a previously paused forwarding protocol by its identifier. This will re-enable all forwarding operations for the specified protocol.",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "PauseCounterparties",
			Use:       "pause-counterparties [protocol_id] [counterparty_ids]",
			Short:     "Pause specific counterparty pairs for a protocol",
			Long:      "Pause specific counterparty pairs for a forwarding protocol. This allows selective pausing of forwarding operations to specific destinations while keeping other counterparties active.",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
				{ProtoField: "counterparty_ids", Varargs: true},
			},
		},
		{
			RpcMethod: "UnpauseCounterparties",
			Use:       "unpause-counterparties [protocol_id] [counterparty_ids]",
			Short:     "Resume specific counterparty pairs for a protocol",
			Long:      "Resume forwarding operations for specific counterparty pairs that were previously paused. This allows selective resumption of forwarding to specific destinations.",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
				{ProtoField: "counterparty_ids", Varargs: true},
			},
		},
		{
			RpcMethod: "ReplaceDepositForBurn",
			Use:       "replace-deposit-for-burn [original_message] [original_attestation] [new_destination_caller] [new_mint_recipient]",
			Short:     "Replace a CCTP deposit for burn message",
			Long:      "Replace a sent deposit for burn message in the CCTP protocol with updated destination caller and mint recipient. This is an administrative function for correcting failed or misdirected burns.",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "original_message"},
				{ProtoField: "original_attestation"},
				{ProtoField: "new_destination_caller"},
				{ProtoField: "new_mint_recipient"},
			},
		},
	}
}
