.PHONY: proto-format proto-lint proto-gen license build
all: proto-all format lint license test-unit build

#=============================================================================#
#                                  Build                                      #
#=============================================================================#

build:
	@echo "==================================================================="
	@echo "Building simd..."
	@cd simapp && GOWORK=off make build 1> /dev/null
	@echo "Completed build!"

#=============================================================================#
#                                  Protobuf                                   #
#=============================================================================#

BUF_VERSION=1.50
BUILDER_VERSION=0.15.3

proto-all: proto-format proto-lint proto-gen proto-testutil-gen

proto-format:
	@echo "==================================================================="
	@echo "Running protobuf formatter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) format --diff --write
	@echo "Completed protobuf formatting!"

proto-gen:
	@echo "==================================================================="
	@echo "Generating code from protobuf..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		ghcr.io/cosmos/proto-builder:$(BUILDER_VERSION) sh ./proto/generate.sh
	@echo "Completed code generation!"

proto-lint:
	@echo "==================================================================="
	@echo "Running protobuf linter..."
	@docker run --rm --volume "$(PWD)":/workspace --workdir /workspace \
		bufbuild/buf:$(BUF_VERSION) lint
	@echo "Completed protobuf linting!"

proto-testutil-gen:
	@echo "==================================================================="
	@echo "Generating code from testutil protobuf..."
	@cd testutil/testdata && buf generate --template buf.gen.yaml
	@echo "Completed code generation!"


#=============================================================================#
#                                 Tooling                                     #
#=============================================================================#

FILES := $(shell find . -name "*.go" -not -path "./simapp/*" -not -name "*.pb.go" -not -name "*.pb.gw.go" -not -name "*.pulsar.go")
license:
	@echo "==================================================================="
	@echo "Adding license to files..."
	@go-license --config .github/license.yaml $(FILES)
	@echo "Completed license addition!"

format:
	@echo "==================================================================="
	@echo "Running formatters..."
	@go tool golangci-lint fmt -c ./.golangci.yaml
	@echo "Completed formatting!"

lint:
	@echo "==================================================================="
	@echo "Running linter..."
	@go tool golangci-lint run -c ./.golangci.yaml
	@echo "Completed linting!"

#=============================================================================#
#                                    Test                                     #
#=============================================================================#

test-unit:
	@echo "==================================================================="
	@echo "Running unit tests for keeper package..."
	@go test -cover -coverpkg=./keeper/... -coverprofile=coverage.out -race -v ./keeper/...
	@go tool cover -html=coverage.out && go tool cover -func=coverage.out
	@echo "Running unit tests for types package..."
	@go test -v ./types/...

local-image:
	@echo "==================================================================="
	@echo "Building image..."
	@heighliner build --chain orbiter-simd --file e2e/chains.yaml --local 1> /dev/null
	@echo "Completed build!"

test-e2e:
	@echo "==================================================================="
	@echo "Running e2e tests..."
	@cd e2e && go test -timeout 15m -race -v ./...
	@echo "Completed e2e tests!"
