package client

import (
	"context"
	"fmt"

	"github.com/babylonlabs-io/finality-gadget/db"
	"github.com/babylonlabs-io/finality-gadget/proto"
	"github.com/babylonlabs-io/finality-gadget/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FinalityGadgetGrpcClient struct {
	client proto.FinalityGadgetClient
	conn   *grpc.ClientConn
	db     db.IDatabaseHandler
}

func NewFinalityGadgetGrpcClient(
	db db.IDatabaseHandler,
	remoteAddr string,
) (*FinalityGadgetGrpcClient, error) {
	conn, err := grpc.NewClient(remoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to build gRPC connection to %s: %w", remoteAddr, err)
	}

	gClient := &FinalityGadgetGrpcClient{
		client: proto.NewFinalityGadgetClient(conn),
		conn:   conn,
		db:     db,
	}

	return gClient, nil
}

func (c *FinalityGadgetGrpcClient) InsertBlock(block *types.Block) (bool, error) {
	req := &proto.BlockInfo{
		BlockHash:      block.BlockHash,
		BlockHeight:    block.BlockHeight,
		BlockTimestamp: block.BlockTimestamp,
	}

	res, err := c.client.InsertBlock(context.Background(), req)
	if err != nil {
		return false, err
	}

	return res.Success, nil
}

func (c *FinalityGadgetGrpcClient) GetBlockStatusByHeight(height uint64) (bool, error) {
	req := &proto.GetBlockStatusByHeightRequest{
		BlockHeight: height,
	}

	res, err := c.client.GetBlockStatusByHeight(context.Background(), req)
	if err != nil {
		return false, err
	}

	return res.IsFinalized, nil
}

func (c *FinalityGadgetGrpcClient) GetBlockStatusByHash(hash string) (bool, error) {
	req := &proto.GetBlockStatusByHashRequest{
		BlockHash: hash,
	}

	res, err := c.client.GetBlockStatusByHash(context.Background(), req)
	if err != nil {
		return false, err
	}

	return res.IsFinalized, nil
}

func (c *FinalityGadgetGrpcClient) GetLatestBlock() (*types.Block, error) {
	req := &proto.GetLatestBlockRequest{}

	res, err := c.client.GetLatestBlock(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return &types.Block{
		BlockHash:      res.BlockHash,
		BlockHeight:    res.BlockHeight,
		BlockTimestamp: res.BlockTimestamp,
	}, nil
}

func (c *FinalityGadgetGrpcClient) Close() error {
	return c.conn.Close()
}
