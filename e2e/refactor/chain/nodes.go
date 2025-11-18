package chain

type NodesConfig struct {
	NumValidator int
	NumFullNodes int
	NumWallets   int
}

func DefaultNodesConfig() *NodesConfig {
	return &NodesConfig{
		NumValidator: 1,
		NumFullNodes: 0,
		NumWallets:   3,
	}
}
