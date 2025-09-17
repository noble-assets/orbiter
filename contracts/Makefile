all: deps compile

compile: compile-go-bindings

SOLIDITY_VERSION := "0.8.13"
SOLC_IMAGE := "ethereum/solc:$(SOLIDITY_VERSION)"
compile-go-bindings:
	docker run --rm \
		-v $(PWD):/workspace \
		-w /workspace \
		$(SOLC_IMAGE) \
		--abi --bin --overwrite \
		-o /workspace/build \
		/workspace/HyperlaneEntrypoint.sol
	docker run --rm \
		-v $(PWD):/workspace \
		-w /workspace \
		ethereum/client-go:alltools-v1.13.0 \
		abigen --abi=/workspace/build/HyperlaneEntrypoint.abi \
		--bin=/workspace/build/HyperlaneEntrypoint.bin \
		--pkg=contracts \
		--out=/workspace/HyperlaneEntrypoint.go

deps:
	bun install
