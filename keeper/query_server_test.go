package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/noble-assets/orbiter/testutil/mocks/orbiter"
	orbitertypes "github.com/noble-assets/orbiter/types"
)

func TestPendingPayloads(t *testing.T) {
	examplePayload := createTestPendingPayloadWithSequence(t, 0).Payload

	testcases := []struct {
		name       string
		nPayloads  int
		pagination *query.PageRequest
		expLen     int
	}{
		{
			name:      "success - no hashes stored",
			nPayloads: 0,
			expLen:    0,
		},
		{
			name:      "success - 1 hashes stored",
			nPayloads: 1,
			expLen:    1,
		},
		{
			name:      "success - 5 hashes stored",
			nPayloads: 5,
			expLen:    5,
		},
		{
			name:       "success - 5 hashes stored with 2 pagination",
			nPayloads:  5,
			pagination: &query.PageRequest{Offset: 1, Limit: 2},
			expLen:     2,
		},
		{
			name:       "success - 5 hashes stored with offset out of range so no results returned",
			nPayloads:  5,
			pagination: &query.PageRequest{Offset: 6, Limit: 2},
			expLen:     0,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _, k := orbiter.OrbiterKeeper(t)

			for range tc.nPayloads {
				_, err := k.AcceptPayload(ctx, examplePayload)
				require.NoError(t, err, "failed to setup payloads")
			}

			res, err := k.PendingPayloads(
				ctx,
				&orbitertypes.QueryPendingPayloadsRequest{
					Pagination: tc.pagination,
				},
			)
			require.NoError(t, err, "failed to query pending payloads")
			require.Equal(
				t,
				tc.expLen,
				len(res.Hashes),
				"expected different number of hashes returned",
			)
		})
	}
}
