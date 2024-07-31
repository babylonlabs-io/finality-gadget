package server

import (
	"context"

	"google.golang.org/grpc"

	"github.com/babylonchain/babylon-finality-gadget/finalitygadget"
	"github.com/babylonchain/babylon-finality-gadget/proto"
	"github.com/babylonchain/babylon-finality-gadget/types"
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

func (r *rpcServer) Ping(ctx context.Context, req *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{}, nil
}

// InsertBlock is an RPC method that inserts a block into the database.
func (r *rpcServer) InsertBlock(ctx context.Context, req *proto.BlockInfo) (*proto.InsertBlockResponse, error) {
	err := r.fg.InsertBlock(&types.Block{
		BlockHash:      req.BlockHash,
		BlockHeight:    req.BlockHeight,
		BlockTimestamp: req.BlockTimestamp,
	})

	if err != nil {
		return &proto.InsertBlockResponse{Success: false}, err
	}

	return &proto.InsertBlockResponse{Success: true}, nil
}

// GetBlockStatusByHeight is an RPC method that returns the status of a block at a given height.
func (r *rpcServer) GetBlockStatusByHeight(ctx context.Context, req *proto.GetBlockStatusByHeightRequest) (*proto.GetBlockStatusResponse, error) {
	isFinalized, err := r.fg.GetBlockStatusByHeight(req.BlockHeight)

	if err != nil {
		return nil, err
	}

	return &proto.GetBlockStatusResponse{IsFinalized: isFinalized}, nil
}

// GetBlockStatusByHeight is an RPC method that returns the status of a block at a given height.
func (r *rpcServer) GetBlockStatusByHash(ctx context.Context, req *proto.GetBlockStatusByHashRequest) (*proto.GetBlockStatusResponse, error) {
	isFinalized, err := r.fg.GetBlockStatusByHash(req.BlockHash)

	if err != nil {
		return nil, err
	}

	return &proto.GetBlockStatusResponse{IsFinalized: isFinalized}, nil
}

// GetLatestBlock is an RPC method that returns the latest consecutively finalized block.
func (r *rpcServer) GetLatestBlock(ctx context.Context, req *proto.GetLatestBlockRequest) (*proto.BlockInfo, error) {
	block, err := r.fg.GetLatestBlock()

	if err != nil {
		return nil, err
	}

	return &proto.BlockInfo{
		BlockHash:      block.BlockHash,
		BlockHeight:    block.BlockHeight,
		BlockTimestamp: block.BlockTimestamp,
	}, nil
}
