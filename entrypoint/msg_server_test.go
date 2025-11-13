package entrypoint_test

import (
	"testing"

	"cosmossdk.io/log"

	"github.com/noble-assets/orbiter/v2/entrypoint"
	"github.com/noble-assets/orbiter/v2/testutil/mocks"
)

func TestMsgHandlerCCTPMessage(t *testing.T) {
	testCases := []struct{}

	adapter, deps := mocks.NewAdapterComponent(t)
	cctpHandler := mocks.NewCCTPServer()

	server := entrypoint.NewMsgServer(deps.Logger, adapter, cctpHandler)

	server.HandleCCTPMessage(v, signer string, message []byte, attestation []byte)
}
