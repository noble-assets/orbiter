package main

import (
	errorsmod "cosmossdk.io/errors"

	orbiter "orbiter.dev"
	"orbiter.dev/testutil"
	"orbiter.dev/types"
	"orbiter.dev/types/core"
)

func buildFinalPayload(forwarding *core.Forwarding, actions []*core.Action) (string, error) {
	payload, err := core.NewPayloadWrapper(forwarding, actions)
	if err != nil {
		return "", errorsmod.Wrap(err, "failed to create payload wrapper")
	}

	encCfg := testutil.MakeTestEncodingConfig("noble")
	orbiter.RegisterInterfaces(encCfg.InterfaceRegistry)

	payloadBz, err := types.MarshalJSON(encCfg.Codec, payload)
	if err != nil {
		return "", errorsmod.Wrap(err, "failed to marshal payload")
	}

	return string(payloadBz), nil
}
