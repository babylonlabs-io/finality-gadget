package server

import (
	"context"

	"google.golang.org/grpc"

	"github.com/babylonlabs-io/finality-gadget/finalitygadget"
	"github.com/babylonlabs-io/finality-gadget/proto"
)

// rpcServer is the main RPC server for the finality gadget daemon that handles
// gRPC incoming requests.
type rpcServer struct {
	proto.UnimplementedFinalityGadgetServer

	fg finalitygadget.IFinalityGadget
}

// newRPCServer creates a new RPC sever from the set of input dependencies.
func newRPCServer(
	fg finalitygadget.IFinalityGadget,
) *rpcServer {
	return &rpcServer{
		fg: fg,
	}
}

// RegisterWithGrpcServer registers the rpcServer with the passed root gRPC
// server.
func (r *rpcServer) RegisterWithGrpcServer(grpcServer *grpc.Server) error {
	// Register the main RPC server.
	proto.RegisterFinalityGadgetServer(grpcServer, r)
	return nil
}

// QueryIsBlockFinalizedByHeight is an RPC method that returns the status of a block at a given height.
func (r *rpcServer) QueryIsBlockFinalizedByHeight(ctx context.Context, req *proto.QueryIsBlockFinalizedByHeightRequest) (*proto.QueryIsBlockFinalizedResponse, error) {
	isFinalized, err := r.fg.QueryIsBlockFinalizedByHeight(req.BlockHeight)

	if err != nil {
		return nil, err
	}

	return &proto.QueryIsBlockFinalizedResponse{IsFinalized: isFinalized}, nil
}

// QueryIsBlockFinalizedByHeight is an RPC method that returns the status of a block at a given height.
func (r *rpcServer) QueryIsBlockFinalizedByHash(ctx context.Context, req *proto.QueryIsBlockFinalizedByHashRequest) (*proto.QueryIsBlockFinalizedResponse, error) {
	isFinalized, err := r.fg.QueryIsBlockFinalizedByHash(req.BlockHash)

	if err != nil {
		return nil, err
	}

	return &proto.QueryIsBlockFinalizedResponse{IsFinalized: isFinalized}, nil
}

// QueryLatestFinalizedBlock is an RPC method that returns the latest consecutively finalized block.
func (r *rpcServer) QueryLatestFinalizedBlock(ctx context.Context, req *proto.QueryLatestFinalizedBlockRequest) (*proto.QueryBlockResponse, error) {
	block, err := r.fg.QueryLatestFinalizedBlock()

	if err != nil {
		return nil, err
	}

	return &proto.QueryBlockResponse{
		BlockHash:      block.BlockHash,
		BlockHeight:    block.BlockHeight,
		BlockTimestamp: block.BlockTimestamp,
	}, nil
}
