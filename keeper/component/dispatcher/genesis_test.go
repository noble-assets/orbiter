package dispatcher_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"orbiter.dev/testutil/mocks"
	dispatchertypes "orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

func defaultAmounts(
	sourceProtocolID, destProtocolID core.ProtocolID,
	sourceCounterpartyID, destCounterpartyID string,
) *dispatchertypes.DispatchedAmountEntry {
	return &dispatchertypes.DispatchedAmountEntry{
		SourceId: &core.CrossChainID{
			ProtocolId:     sourceProtocolID,
			CounterpartyId: sourceCounterpartyID,
		},
		DestinationId: &core.CrossChainID{
			ProtocolId:     destProtocolID,
			CounterpartyId: destCounterpartyID,
		},
		Denom: "unoble",
		AmountDispatched: dispatchertypes.AmountDispatched{
			Incoming: math.NewInt(1),
			Outgoing: math.NewInt(1),
		},
	}
}

func defaultCounts(
	sourceProtocolID, destProtocolID core.ProtocolID,
	sourceCounterpartyID, destCounterpartyID string,
) *dispatchertypes.DispatchCountEntry {
	return &dispatchertypes.DispatchCountEntry{
		SourceId: &core.CrossChainID{
			ProtocolId:     sourceProtocolID,
			CounterpartyId: sourceCounterpartyID,
		},
		DestinationId: &core.CrossChainID{
			ProtocolId:     destProtocolID,
			CounterpartyId: destCounterpartyID,
		},
		Count: 1,
	}
}

func TestInitGenesis(t *testing.T) {
	testCases := []struct {
		name    string
		genesis func() *dispatchertypes.GenesisState
		expErr  string
	}{
		{
			name:    "error - nil genesis",
			genesis: func() *dispatchertypes.GenesisState { return nil },
			expErr:  "nil",
		},
		{
			name:    "success - default genesis",
			genesis: dispatchertypes.DefaultGenesisState,
			expErr:  "",
		},
		{
			name: "success - custom genesis",
			genesis: func() *dispatchertypes.GenesisState {
				g := dispatchertypes.DefaultGenesisState()
				g.DispatchedAmounts = append(
					g.DispatchedAmounts,
					*defaultAmounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "chain-1", "chain-2"),
					*defaultAmounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "chain-3", "chain-4"),
					*defaultAmounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "chain-11", "chain-12"),
					*defaultAmounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "chain-13", "chain-14"),
				)
				g.DispatchedCounts = append(
					g.DispatchedCounts,
					*defaultCounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "chain-1", "chain-2"),
					*defaultCounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "chain-3", "chain-4"),
					*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "chain-11", "chain-12"),
					*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "chain-13", "chain-14"),
					*defaultCounts(core.PROTOCOL_HYPERLANE, core.PROTOCOL_CCTP, "1", "2"),
					*defaultCounts(core.PROTOCOL_HYPERLANE, core.PROTOCOL_CCTP, "3", "4"),
					*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_HYPERLANE, "11", "12"),
					*defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_HYPERLANE, "13", "14"),
				)

				return g
			},
			expErr: "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			d, deps := mocks.NewDispatcherComponent(t)
			ctx := deps.SdkCtx

			g := tC.genesis()
			err := d.InitGenesis(ctx, g)

			if tC.expErr != "" {
				require.ErrorContains(t, err, tC.expErr)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, g.DispatchedAmounts, d.GetAllDispatchedAmounts(ctx))
				require.ElementsMatch(t, g.DispatchedCounts, d.GetAllDispatchedCounts(ctx))
			}
		})
	}
}

func TestExportGenesis(t *testing.T) {
	testCases := []struct {
		name              string
		dispatchedAmounts func() []*dispatchertypes.DispatchedAmountEntry
		dispatchedCounts  func() []*dispatchertypes.DispatchCountEntry
		expErr            string
	}{
		{
			name:   "success - empty state",
			expErr: "nil",
		},
		{
			name: "success - not empty state",
			dispatchedAmounts: func() []*dispatchertypes.DispatchedAmountEntry {
				return []*dispatchertypes.DispatchedAmountEntry{
					defaultAmounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "chain-1", "chain-2"),
					defaultAmounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "chain-3", "chain-4"),
					defaultAmounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "chain-11", "chain-12"),
					defaultAmounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "chain-13", "chain-14"),
				}
			},
			dispatchedCounts: func() []*dispatchertypes.DispatchCountEntry {
				return []*dispatchertypes.DispatchCountEntry{
					defaultCounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "chain-1", "chain-2"),
					defaultCounts(core.PROTOCOL_IBC, core.PROTOCOL_CCTP, "chain-3", "chain-4"),
					defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "chain-11", "chain-12"),
					defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_IBC, "chain-13", "chain-14"),
					defaultCounts(core.PROTOCOL_HYPERLANE, core.PROTOCOL_CCTP, "1", "2"),
					defaultCounts(core.PROTOCOL_HYPERLANE, core.PROTOCOL_CCTP, "3", "4"),
					defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_HYPERLANE, "11", "12"),
					defaultCounts(core.PROTOCOL_CCTP, core.PROTOCOL_HYPERLANE, "13", "14"),
				}
			},

			expErr: "",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			d, deps := mocks.NewDispatcherComponent(t)
			ctx := deps.SdkCtx

			var da []*dispatchertypes.DispatchedAmountEntry
			if tC.dispatchedAmounts != nil {
				da = tC.dispatchedAmounts()
			}
			for _, amount := range da {
				err := d.SetDispatchedAmount(
					ctx,
					amount.SourceId,
					amount.DestinationId,
					amount.Denom,
					amount.AmountDispatched,
				)
				require.NoError(t, err)
			}

			var dc []*dispatchertypes.DispatchCountEntry
			if tC.dispatchedCounts != nil {
				dc = tC.dispatchedCounts()
			}
			for _, count := range dc {
				err := d.SetDispatchedCounts(
					ctx,
					count.SourceId,
					count.DestinationId,
					count.Count,
				)
				require.NoError(t, err)
			}

			g := d.ExportGenesis(ctx)

			if tC.expErr != "" {
				require.ElementsMatch(t, da, g.DispatchedAmounts)
				require.ElementsMatch(t, dc, g.DispatchedCounts)
			}
		})
	}
}
