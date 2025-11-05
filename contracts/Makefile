.PHONY: all deps compile

all: deps compile lint

#=============================================================================#
#                            Smart Contracts                                  #
#=============================================================================#

compile:
	@forge compile

clean:
	@forge clean && \
	rm -rf node_modules/

deps:
	@bun install

generate-abi:
	@jq '.abi' ./out/OrbiterGatewayCCTP.sol/OrbiterGatewayCCTP.json > ./abi/OrbiterGatewayCCTP.abi
