package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"path"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/noble-assets/orbiter/e2e/gateway/config"
	"github.com/noble-assets/orbiter/e2e/gateway/service"
	gatewaytypes "github.com/noble-assets/orbiter/e2e/gateway/types"
)

const (
	permitDuration = 60 * 60 * 24
	payload        = "1b8e010064ca0e8337064bf0cca84a2a084f4c2f787a482707a2373d5ce4a45f40d23c818c9a071eb209427ec0ba1eb86f6861e878cb1024c4d14c1709c5327c1726427c6c37438b0b50fa93a9b172426c290415880d86e0b86008e241343a99a0c99a26c189491a4c4ab89448e1a93487c069282bd36a4a06afb26c466a7abc7d0f48a0ee401ca4f1f1781249e16a624a7a3292989cc2a29d647c2aa51028cfa6a1929cdc697fc5d01f4c43ba537849a793923a295cd1043a3d351e575945e5e99454012550928074c8feb091e359430b63c713862025991a334623e9b05c63f7aac05f1fdc99a27f942eb7d8ad0e7e59723c1bf9c2f8ecef379df71cbb0abf3b47ececf7f49f79fc6dba28fb3f5ab77461b523e1fbf040ffe9850be76e0c791b879d7f3ea00c06fa9f13cf64ecfb4d6948d5ec87ade6918be3d74053409ecba90658f2bb5e5dfb35d138213e3ed30a9848424fb7a371427c3c"
)

func main() {
	ctx := context.Background()

	file := path.Join(config.ConfigDir, "./example/config_sepolia_local.toml")
	configuration, err := config.Load(file)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	service, err := service.NewService(ctx, &configuration.Ethereum)
	if err != nil {
		log.Fatalf("failed to start service: %v", err)
	}
	defer service.Close()

	signer := service.Signer()

	fmt.Println("=================================================================================")
	fmt.Println("Connected to Ethereum node")
	fmt.Printf("-  Signer address: %s\n", signer.Address().Hex())
	fmt.Printf("-  Gateway address: %s\n", service.GatewayAddress().Hex())
	fmt.Printf("-  USDC address: %s\n\n", service.USDCAddress().Hex())

	ethBalance, err := service.SignerBalance(ctx)
	if err != nil {
		log.Fatalf("failed to get sender ETH balance: %v", err)
	}

	usdcBalance, err := service.SignerUSDCBalance(ctx)
	if err != nil {
		log.Fatalf("failed to get sender USDC balance: %v", err)
	}

	fmt.Println("=================================================================================")
	fmt.Println("Signer account balances")
	fmt.Printf("- USDC: %suusdc\n", usdcBalance.String())
	fmt.Printf("- ETH: %swei\n\n", ethBalance.String())

	// Get current block to calculate deadline
	blockTime, err := service.BlockTime(ctx)
	if err != nil {
		log.Fatalf("Failed to query current block number: %v", err)
	}

	amount := big.NewInt(
		1_000_000,
	) // 1 USDC (6 decimals)
	deadline := big.NewInt(
		int64(blockTime) + permitDuration,
	) // Block deadline

	fmt.Println("=================================================================================")
	fmt.Println("Orbiter transfer parameters")
	fmt.Printf("- Amount: %s\n", amount.String())
	fmt.Printf("- Current block time: %d\n", blockTime)
	fmt.Printf("- Deadline block time: %s\n", deadline.String())
	fmt.Printf("- Orbiter payload hex: %s\n\n", payload)

	signerNonce, err := service.SignerUSDCNonce()
	if err != nil {
		log.Fatalf("Failed to get signer USDC nonce: %v", err)
	}

	domainSeparator, err := service.USDCDomainSeparator()
	if err != nil {
		log.Fatalf("Failed to get USDC domain separator: %v", err)
	}

	permit := gatewaytypes.NewPermit(
		domainSeparator,
		signer.Address(),
		service.GatewayAddress(),
		amount,
		signerNonce,
		deadline,
	)

	sign, err := signer.Sign(permit.Digest())
	if err != nil {
		log.Fatalf("Failed to sign permit: %v", err)
	}

	abiSign := gatewaytypes.ABIEncodeSignature(sign)

	txOpts, err := service.TxOpts(ctx)
	if err != nil {
		log.Fatalf("Failed to create tx options: %v", err)
	}

	tx, err := service.DepositForBurnWithOrbiter(
		txOpts,
		amount,
		deadline,
		abiSign,
		common.Hex2Bytes(payload),
	)
	if err != nil {
		log.Fatalf("Failed to send tx: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	receipt, err := service.WaitForTransaction(ctx, tx.Hash())
	if err != nil {
		log.Fatalf("Failed to execute the transaction: %v", err)
	}

	events, err := service.ParseDepositForBurnEvents(receipt)
	if err != nil {
		log.Fatalf("Failed to parse events: %v", err)
	}

	var eventLogs string
	for i, event := range events {
		eventLogs += fmt.Sprintf("-  Event %d: TransferNonce=%d, PayloadNonce=%d\n",
			i+1, event.TransferNonce, event.PayloadNonce)
	}

	fmt.Println("=================================================================================")
	fmt.Println("Deposit transaction")
	fmt.Printf("Transaction hash: %s\n", tx.Hash().Hex())
	fmt.Printf("Emitted %d DepositForBurnWithOrbiter event(s)\n", len(events))
	fmt.Printf("%s", eventLogs)
}
