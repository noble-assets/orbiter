
# ------------------------------
# Build & Deps
# ------------------------------
.PHONY: all deps compile
all: deps compile lint

compile:
	@forge compile

clean:
	@forge clean && \
	rm -rf node_modules/

deps:
	@bun install

# ------------------------------
# Linting
# ------------------------------
.PHONY: lint lint-solidity
lint: lint-solidity

SOLHINT_VERSION := "latest"
SOLHINT_IMAGE := "protodb/protofire-solhint:$(SOLHINT_VERSION)"
lint-solidity:
	@docker run -v $(PWD):/workdir -w /workdir \
		-it $(SOLHINT_IMAGE) \
		solhint 'src/*.sol'

# ------------------------------
# Testing
# ------------------------------
.PHONY: test test-forge
test: test-forge

test-forge:
	@forge test -vv
