package entrypoint

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	cctptypes "github.com/circlefin/noble-cctp/x/cctp/types"
	"github.com/ethereum/go-ethereum/common"

	"cosmossdk.io/log"

	"github.com/noble-assets/orbiter/v2/types"
	adaptertypes "github.com/noble-assets/orbiter/v2/types/component/adapter"
	"github.com/noble-assets/orbiter/v2/types/core"
	entrypointtypes "github.com/noble-assets/orbiter/v2/types/entrypoint"
)

var _ entrypointtypes.MsgServer = &msgServer{}

// msgServer is the server used to handle entrypoint messages for the CCTP protocol.
type msgServer struct {
	logger log.Logger

	payloadAdapter types.PayloadAdapter
	cctpHandler    entrypointtypes.CCTPHandler
}

// NewMsgServer returns a new CCTP entrypoints message server.
func NewMsgServer(
	logger log.Logger,
	payloadAdapter types.PayloadAdapter,
	cctpHandler entrypointtypes.CCTPHandler,
) msgServer {
	if logger == nil {
		panic(core.ErrNilPointer.Wrap("logger is not set"))
	}

	if payloadAdapter == nil {
		panic(core.ErrNilPointer.Wrap("payload adapter is not set"))
	}

	if cctpHandler == nil {
		panic(core.ErrNilPointer.Wrap("CCTP handler is not set"))
	}

	return msgServer{
		logger:         logger,
		payloadAdapter: payloadAdapter,
		cctpHandler:    cctpHandler,
	}
}

// NOTE: the transfer could be the real message.
//
// ReceiveCCTPMessage is the server method that allows to initiate an Orbiter execution for
// the CCTP protocol.
func (s *msgServer) ReceiveCCTPMessage(
	ctx context.Context,
	msg *entrypointtypes.MsgReceiveCCTPMessage,
) (*entrypointtypes.MsgReceiveCCTPMessageResponse, error) {
	transferMsg, payloadMsg, err := ParseCCTPMessages(msg.TransferMessage, msg.PayloadMessage)
	if err != nil {
		return nil, core.ErrParsing.Wrapf("failed to parse CCTP messages: %s", err.Error())
	}

	burnMessage, err := new(cctptypes.BurnMessage).Parse(transferMsg.MessageBody)
	if err != nil {
		return nil, core.ErrParsing.Wrapf("burn message is not valid: %s", err.Error())
	}

	err = ValidateCCTPMessages(transferMsg, payloadMsg)
	if err != nil {
		return nil, core.ErrValidation.Wrap(err.Error())
	}

	// Verify that the payload GMP is valid.
	err = s.HandleCCTPMessage(ctx, msg.Signer, msg.PayloadMessage, msg.PayloadAttestation)
	if err != nil {
		return nil, core.ErrValidation.Wrapf("failed to validate payload message: %s", err.Error())
	}

	ccID := core.CrossChainID{
		ProtocolId:     core.PROTOCOL_CCTP,
		CounterpartyId: strconv.FormatUint(uint64(transferMsg.SourceDomain), 10),
	}

	ccPacket, err := adaptertypes.NewCCTPCrossChainPacket(
		transferMsg.GetNonce(),
		tokenPair.LocalToken,
		burnMessage.Amount,
		msg.PayloadMessage,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create the CCTP cross-chain packet: %w", err)
	}

	orbiterPacket, err := s.payloadAdapter.AdaptPacket(ctx, ccID, ccPacket)
	if err != nil {
		return nil, fmt.Errorf("failed to adapt CCTP packet: %w", err)
	}

	err = s.payloadAdapter.BeforeTransferHook(ctx, orbiterPacket)
	if err != nil {
		return nil, fmt.Errorf("failed to execute before transfer hook: %w", err)
	}

	err = s.HandleCCTPMessage(ctx, msg.Signer, msg.TransferMessage, msg.TransferAttestation)
	if err != nil {
		return nil, core.ErrValidation.Wrapf("failed to execute transfer message: %s", err.Error())
	}

	err = s.payloadAdapter.AfterTransferHook(ctx, orbiterPacket)
	if err != nil {
		return nil, fmt.Errorf("failed to executing after transfer hook: %w", err)
	}

	err = s.payloadAdapter.ProcessPayload(ctx, orbiterPacket)
	if err != nil {
		return nil, fmt.Errorf("failed to process orbiter payload: %w", err)
	}

	return &entrypointtypes.MsgReceiveCCTPMessageResponse{}, nil
}

func ParseCCTPMessages(
	transferMsgBz, payloadMsgBz []byte,
) (*cctptypes.Message, *cctptypes.Message, error) {
	transferMsg, err := new(cctptypes.Message).Parse(transferMsgBz)
	if err != nil {
		return nil, nil, fmt.Errorf("transfer message is not a valid CCTP message: %w", err)
	}

	payloadMsg, err := new(cctptypes.Message).Parse(payloadMsgBz)
	if err != nil {
		return nil, nil, fmt.Errorf("payload message is not a valid CCTP message: %w", err)
	}

	return transferMsg, payloadMsg, nil
}

func ValidateCCTPMessages(transferMsg, payloadMsg *cctptypes.Message) error {
	if transferMsg.GetNonce() != payloadMsg.GetNonce()+1 {
		return errors.New("messages nonces are not valid")
	}

	return nil
}

func (s msgServer) GetTransferCoin(ctx context.Context, sourceDomain uint32, burnToken []byte) {
	// Retrieve the token transferred information.
	tokenPair, found := s.cctpHandler.GetTokenPair(
		ctx,
		transferMsg.SourceDomain,
		burnMessage.BurnToken,
	)
	if !found {
		return nil, fmt.Errorf(
			"token pair for burn token %s and source domain %d does not exist",
			common.Bytes2Hex(burnMessage.BurnToken),
			transferMsg.SourceDomain,
		)
	}
}

// HandleCCTPMessage is an utility method to call into the CCTP module and handle the response.
func (s msgServer) HandleCCTPMessage(
	ctx context.Context,
	signer string,
	message, attestation []byte,
) error {
	resp, err := s.cctpHandler.ReceiveMessage(ctx, &cctptypes.MsgReceiveMessage{
		From:        signer,
		Message:     message,
		Attestation: attestation,
	})
	if err != nil {
		return fmt.Errorf("received an error handling CCTP message: %w", err)
	}

	if resp == nil || !resp.Success {
		return errors.New("received unexpected response from CCTP")
	}

	return nil
}
