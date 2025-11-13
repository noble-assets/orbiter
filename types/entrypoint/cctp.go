package entrypoint

import (
	"context"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
)

type CCTPHandler interface {
	ReceiveMessage(
		context.Context,
		*cctptypes.MsgReceiveMessage,
	) (*cctptypes.MsgReceiveMessageResponse, error)
	GetTokenPair(
		context.Context,
		uint32,
		[]byte,
	) (cctptypes.TokenPair, bool)
}
