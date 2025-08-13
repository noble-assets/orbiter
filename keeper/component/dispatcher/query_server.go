package dispatcher

import (
	"context"

	"cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	dispatchertypes "orbiter.dev/types/component/dispatcher"
	"orbiter.dev/types/core"
)

var _ dispatchertypes.QueryServer = &queryServer{}

type queryServer struct {
	*Dispatcher
}

func NewQueryServer(d *Dispatcher) dispatchertypes.QueryServer {
	return queryServer{Dispatcher: d}
}

func (q queryServer) DispatchedCounts(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedCountsRequest,
) (*dispatchertypes.QueryDispatchedCountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	sourceID, err := core.NewCrossChainID(req.SourceProtocolId, req.SourceCounterpartyId)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating source cross-chain ID")
	}

	destID, err := core.NewCrossChainID(req.DestinationProtocolId, req.DestinationCounterpartyId)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating destination cross-chain ID")
	}

	counts := q.GetDispatchedCounts(ctx, &sourceID, &destID)

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts: []*dispatchertypes.DispatchCountEntry{counts},
	}, nil
}

func (q queryServer) DispatchedCountsByDestinationProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedCountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedCountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	if err := req.ProtocolId.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid protocol ID")
	}

	counts := q.GetDispatchedCountsByDestinationProtocolID(ctx, req.ProtocolId)

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts: counts,
	}, nil
}

func (q queryServer) DispatchedCountsBySourceProtocolID(
	ctx context.Context,
	req *dispatchertypes.QueryDispatchedCountsByProtocolIDRequest,
) (*dispatchertypes.QueryDispatchedCountsResponse, error) {
	if req == nil {
		return nil, sdkerrors.ErrInvalidRequest
	}

	if err := req.ProtocolId.Validate(); err != nil {
		return nil, errors.Wrapf(err, "invalid protocol ID")
	}

	counts := q.GetDispatchedCountsBySourceProtocolID(ctx, req.ProtocolId)

	return &dispatchertypes.QueryDispatchedCountsResponse{
		Counts: counts,
	}, nil
}
