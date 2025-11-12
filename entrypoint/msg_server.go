package entrypoint

import (
	"context"
	"errors"
	"strconv"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"

	"cosmossdk.io/log"

	"github.com/noble-assets/orbiter/v2/types"
	adaptertypes "github.com/noble-assets/orbiter/v2/types/component/adapter"
	"github.com/noble-assets/orbiter/v2/types/core"
)

type CCTPMsgServer interface {
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

var _ MsgServer = &msgServer{}

// msgServer is the server used to handle entrypoint messages.
type msgServer struct {
	logger log.Logger

	payloadAdapter types.PayloadAdapter
	cctpServer     CCTPMsgServer
}

func NewMsgServer() msgServer {
	// TODO

	return msgServer{}
}

func ValidateCCTPNonces(transferNonce, payloadNonce uint64) error {
	isValid := transferNonce == payloadNonce+1
	if !isValid {
		return errors.New("no bueno")
	}

	return nil
}

// NOTE: the transfer could be the real message.
// ReceiveCCTPMessage implements MsgServer.
func (s *msgServer) ReceiveCCTPMessage(
	ctx context.Context,
	msg *MsgReceiveCCTPMessage,
) (*MsgReceiveCCTPMessageResponse, error) {
	transferMsg, err := new(cctptypes.Message).Parse(msg.TransferMessage)
	if err != nil {
		return nil, err
	}
	burnMessage, err := new(cctptypes.BurnMessage).Parse(transferMsg.MessageBody)

	payloadMsg, err := new(cctptypes.Message).Parse(msg.PayloadMessage)
	if err != nil {
		return nil, err
	}

	// NOTE: should we also check the destination caller is the same?
	err = ValidateCCTPNonces(transferMsg.GetNonce(), payloadMsg.GetNonce())
	if err != nil {
		return nil, err
	}

	// Verify that the payload GMP is valid.
	resp, err := s.cctpServer.ReceiveMessage(ctx, &cctptypes.MsgReceiveMessage{
		From:        msg.Signer,
		Message:     msg.PayloadMessage,
		Attestation: msg.PayloadAttestation,
	})
	if err != nil {
		return nil, err
	}
	// NOTE: resp success can only be true
	if !resp.Success {
		return nil, errors.New("error receiving the msg")
	}

	ccID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: strconv.FormatUint(uint64(transferMsg.SourceDomain), 10),
	}

	tokenPair, found := s.cctpServer.GetTokenPair(
		ctx,
		transferMsg.SourceDomain,
		burnMessage.BurnToken,
	)
	if !found {
		return nil, errors.New("does not exist")
	}

	ccPacket, err := adaptertypes.NewCCTPCrossChainPacket(
		transferMsg.GetNonce(),
		tokenPair.LocalToken,
		burnMessage.Amount,
		msg.PayloadMessage,
	)
	if err != nil {
		return nil, errors.New("something wrong bro")
	}

	orbiterPacket, err := s.payloadAdapter.AdaptPacket(ctx, ccID, ccPacket)
	if err != nil {
		return nil, errors.New("something wrong bro")
	}

	err = s.payloadAdapter.BeforeTransferHook(ctx, orbiterPacket)
	if err != nil {
		return nil, err
	}

	resp, err = s.cctpServer.ReceiveMessage(ctx, &cctptypes.MsgReceiveMessage{
		From:        msg.Signer,
		Message:     msg.TransferMessage,
		Attestation: msg.TransferAttestation,
	})
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, errors.New("error receiving the msg")
	}

	err = s.payloadAdapter.AfterTransferHook(ctx, orbiterPacket)
	if err != nil {
		return nil, err
	}

	err = s.payloadAdapter.ProcessPayload(ctx, orbiterPacket)
	if err != nil {
		return nil, err
	}

	return &MsgReceiveCCTPMessageResponse{}, nil
}
