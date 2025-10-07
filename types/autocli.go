package types

import autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

func TxCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "SubmitPayload",
			Use:       "submit [payload]",
			Short:     "Submit a payload to be handled by the orbiter",
			Long:      `Submit a payload to be handled by the orbiter.`,
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{
				{ProtoField: "payload"},
			},
		},
	}
}

func QueryCommandOptions() []*autocliv1.RpcCommandOptions {
	return []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "PendingPayloads",
			Use:       "pending",
			Short:     "Query pending payloads",
		},
	}
}
