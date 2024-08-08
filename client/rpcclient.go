package client

import (
	"context"
	"fmt"

	"github.com/babylonlabs-io/finality-gadget/proto"
	"github.com/babylonlabs-io/finality-gadget/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FinalityGadgetGrpcClient struct {
	client proto.FinalityGadgetClient
	conn   *grpc.ClientConn
}

func NewFinalityGadgetGrpcClient(
	remoteAddr string,
) (*FinalityGadgetGrpcClient, error) {
	conn, err := grpc.NewClient(remoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to build gRPC connection to %s: %w", remoteAddr, err)
	}

	gClient := &FinalityGadgetGrpcClient{
		client: proto.NewFinalityGadgetClient(conn),
		conn:   conn,
	}

	return gClient, nil
}

func (c *FinalityGadgetGrpcClient) QueryIsBlockFinalizedByHeight(height uint64) (bool, error) {
	req := &proto.QueryIsBlockFinalizedByHeightRequest{
		BlockHeight: height,
	}

	res, err := c.client.QueryIsBlockFinalizedByHeight(context.Background(), req)
	if err != nil {
		return false, err
	}

	return res.IsFinalized, nil
}

func (c *FinalityGadgetGrpcClient) QueryIsBlockFinalizedByHash(hash string) (bool, error) {
	req := &proto.QueryIsBlockFinalizedByHashRequest{
		BlockHash: hash,
	}

	res, err := c.client.QueryIsBlockFinalizedByHash(context.Background(), req)
	if err != nil {
		return false, err
	}

	return res.IsFinalized, nil
}

func (c *FinalityGadgetGrpcClient) QueryLatestFinalizedBlock() (*types.Block, error) {
	req := &proto.QueryLatestFinalizedBlockRequest{}

	res, err := c.client.QueryLatestFinalizedBlock(context.Background(), req)
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
