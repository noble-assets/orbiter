package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterInterfaces is used to register in the chain codec
// all interfaces and associated implementations defined in
// the Orbiter module.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgPauseProtocol{},
		&MsgPauseCounterparties{},
		&MsgUnpauseProtocol{},
		&MsgUnpauseCounterparties{},
		&MsgPauseAction{},
		&MsgUnpauseAction{},
	)

	registry.RegisterInterface(
		"noble.orbiter.v1.OrbitAttributes",
		(*OrbitAttributes)(nil),
	)

	registry.RegisterInterface(
		"noble.orbiter.v1.ActionAttributes",
		(*ActionAttributes)(nil),
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
