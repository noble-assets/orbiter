package forwarder

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

func TxCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "PauseProtocol",
			Use:       "pause-protocol [protocol_id]",
			Short:     "Pause a specific forwarding protocol by ID",
			Long: `Pause a specific forwarding protocol by its identifier. This will 
disable all forwarding operations for the specified protocol until it is unpaused.

Note: This operation does not affect the granular pausing system based on protocol 
and counterparty ID combinations.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "UnpauseProtocol",
			Use:       "unpause-protocol [id]",
			Short:     "Resume a paused forwarding protocol by ID",
			Long: `Resume operations for a previously paused forwarding protocol by its 
identifier. This will re-enable all forwarding operations for the specified protocol. 

Note: This operation does not affect the granular pausing system based on protocol 
and counterparty ID combinations.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "PauseCrossChains",
			Use:       "pause-cross-chains [protocol_id] [counterparty_ids]",
			Short:     "Pause specific counterparties for a protocol",
			Long: `Pause specific counterparties for a forwarding protocol. This 
allows selective pausing of forwarding operations to specific destinations while 
keeping other counterparties active.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
				{ProtoField: "counterparty_ids", Varargs: true},
			},
		},
		{
			RpcMethod: "UnpauseCrossChains",
			Use:       "unpause-cross-chains [protocol_id] [counterparty_ids]",
			Short:     "Resume specific counterparty pairs for a protocol",
			Long: `Resume forwarding operations for specific counterparty pairs that 
were previously paused. This allows selective resumption of forwarding to 
specific destinations.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
				{ProtoField: "counterparty_ids", Varargs: true},
			},
		},
		{
			RpcMethod: "ReplaceDepositForBurn",
			Use:       "replace-deposit-for-burn [original_message] [original_attestation] [new_destination_caller] [new_mint_recipient]",
			Short:     "Replace a CCTP deposit for burn message",
			Long: `Replace a sent deposit for burn message in the CCTP protocol with updated 
destination caller and mint recipient.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "original_message"},
				{ProtoField: "original_attestation"},
				{ProtoField: "new_destination_caller"},
				{ProtoField: "new_mint_recipient"},
			},
		},
	}
}

func QueryCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "IsProtocolPaused",
			Use:       "is-protocol-paused [protocol_id]",
			Short:     "Check if a protocol is paused",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
		{
			RpcMethod: "PausedProtocols",
			Use:       "paused-protocols",
			Short:     "Get all paused protocols IDs",
		},
		{
			RpcMethod: "IsCrossChainsPaused",
			Use:       "is-cross-chain-paused [protocol_id] [counterparty_id]",
			Short:     "Check if the counterparty is paused for the specific protocol",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
				{ProtoField: "counterparty_id"},
			},
		},
		{
			RpcMethod: "PausedCrossChains",
			Use:       "paused-cross-chains [protocol_id]",
			Short:     "Get all paused counterparties for the specific protocol ID",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "protocol_id"},
			},
		},
	}
}
