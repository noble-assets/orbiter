package dispatcher

import (
	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func WithCollectionPaginationQuadPrefix[K1, K2, K3, K4 any](
	prefix K1,
) func(o *query.CollectionsPaginateOptions[collections.Quad[K1, K2, K3, K4]]) {
	return func(o *query.CollectionsPaginateOptions[collections.Quad[K1, K2, K3, K4]]) {
		prefix := collections.QuadPrefix[K1, K2, K3, K4](prefix)
		o.Prefix = &prefix
	}
}
