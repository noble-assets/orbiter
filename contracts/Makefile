.PHONY: all deps compile clean generate-abi

all: deps compile lint generate-abi

#=============================================================================#
#                            Smart Contracts                                  #
#=============================================================================#

compile:
	@forge compile

clean:
	@forge clean && \
	rm -rf node_modules/

lint:
	@forge fmt

deps:
	@bun install

generate-abi:
	@jq '.abi' ./out/OrbiterGatewayCCTP.sol/OrbiterGatewayCCTP.json > ./abi/OrbiterGatewayCCTP.abi
