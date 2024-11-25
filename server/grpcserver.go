package server

import (
	"context"

	"google.golang.org/grpc"

	"github.com/babylonlabs-io/finality-gadget/proto"
	"github.com/babylonlabs-io/finality-gadget/types"
)

// RegisterWithGrpcServer registers the rpcServer with the passed root gRPC
// server.
func (r *Server) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	// Register the main RPC server.
	proto.RegisterFinalityGadgetServer(grpcServer, r)
	return nil
}

// QueryIsBlockBabylonFinalized is an RPC method that returns the finality status of a block by querying the internal db.
func (r *Server) QueryIsBlockBabylonFinalized(ctx context.Context, req *proto.QueryIsBlockBabylonFinalizedRequest) (*proto.QueryIsBlockFinalizedResponse, error) {
	isFinalized, err := r.fg.QueryIsBlockBabylonFinalized(&types.Block{
		BlockHash:      req.Block.BlockHash,
		BlockHeight:    req.Block.BlockHeight,
		BlockTimestamp: req.Block.BlockTimestamp,
	})
	if err != nil {
		return nil, err
	}

	return &proto.QueryIsBlockFinalizedResponse{IsFinalized: isFinalized}, nil
}

// QueryIsBlockBabylonFinalizedFromBabylon is an RPC method that returns the finality status of a block by querying Babylon chain.
func (r *Server) QueryIsBlockBabylonFinalizedFromBabylon(ctx context.Context, req *proto.QueryIsBlockBabylonFinalizedRequest) (*proto.QueryIsBlockFinalizedResponse, error) {
	isFinalized, err := r.fg.QueryIsBlockBabylonFinalizedFromBabylon(&types.Block{
		BlockHash:      req.Block.BlockHash,
		BlockHeight:    req.Block.BlockHeight,
		BlockTimestamp: req.Block.BlockTimestamp,
	})
	if err != nil {
		return nil, err
	}

	return &proto.QueryIsBlockFinalizedResponse{IsFinalized: isFinalized}, nil
}

// QueryBlockRangeBabylonFinalized is an RPC method that returns the latest Babylon finalized block in a range by querying Babylon chain.
func (r *Server) QueryBlockRangeBabylonFinalized(ctx context.Context, req *proto.QueryBlockRangeBabylonFinalizedRequest) (*proto.QueryBlockRangeBabylonFinalizedResponse, error) {
	blocks := make([]*types.Block, 0, len(req.Blocks))

	for _, block := range req.Blocks {
		blocks = append(blocks, &types.Block{
			BlockHash:      block.BlockHash,
			BlockHeight:    block.BlockHeight,
			BlockTimestamp: block.BlockTimestamp,
		})
	}

	blockHeight, err := r.fg.QueryBlockRangeBabylonFinalized(blocks)
	if err != nil {
		return nil, err
	}

	response := &proto.QueryBlockRangeBabylonFinalizedResponse{}
	if blockHeight == nil {
		response.LastFinalizedBlockHeight = 0
	} else {
		response.LastFinalizedBlockHeight = *blockHeight
	}
	return response, nil
}

// QueryBtcStakingActivatedTimestamp is an RPC method that returns the timestamp when BTC staking was activated.
func (r *Server) QueryBtcStakingActivatedTimestamp(ctx context.Context, req *proto.QueryBtcStakingActivatedTimestampRequest) (*proto.QueryBtcStakingActivatedTimestampResponse, error) {
	timestamp, err := r.fg.QueryBtcStakingActivatedTimestamp()
	if err != nil {
		return nil, err
	}

	return &proto.QueryBtcStakingActivatedTimestampResponse{ActivatedTimestamp: timestamp}, nil
}

// QueryIsBlockFinalizedByHeight is an RPC method that returns the status of a block at a given height.
func (r *Server) QueryIsBlockFinalizedByHeight(ctx context.Context, req *proto.QueryIsBlockFinalizedByHeightRequest) (*proto.QueryIsBlockFinalizedResponse, error) {
	isFinalized, err := r.fg.QueryIsBlockFinalizedByHeight(req.BlockHeight)

	if err != nil {
		return nil, err
	}

	return &proto.QueryIsBlockFinalizedResponse{IsFinalized: isFinalized}, nil
}

// QueryIsBlockFinalizedByHeight is an RPC method that returns the status of a block at a given height.
func (r *Server) QueryIsBlockFinalizedByHash(ctx context.Context, req *proto.QueryIsBlockFinalizedByHashRequest) (*proto.QueryIsBlockFinalizedResponse, error) {
	isFinalized, err := r.fg.QueryIsBlockFinalizedByHash(req.BlockHash)

	if err != nil {
		return nil, err
	}

	return &proto.QueryIsBlockFinalizedResponse{IsFinalized: isFinalized}, nil
}

// QueryLatestFinalizedBlock is an RPC method that returns the latest consecutively finalized block.
func (r *Server) QueryLatestFinalizedBlock(ctx context.Context, req *proto.QueryLatestFinalizedBlockRequest) (*proto.QueryBlockResponse, error) {
	block, err := r.fg.QueryLatestFinalizedBlock()

	if block == nil {
		return nil, types.ErrBlockNotFound
	}

	if err != nil {
		return nil, err
	}

	return &proto.QueryBlockResponse{
		Block: &proto.BlockInfo{
			BlockHash:      block.BlockHash,
			BlockHeight:    block.BlockHeight,
			BlockTimestamp: block.BlockTimestamp,
		},
	}, nil
}
