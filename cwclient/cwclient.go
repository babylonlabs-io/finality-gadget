package cwclient

import (
	"context"
	"encoding/json"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/babylonlabs-io/finality-gadget/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
)

type CosmWasmClient struct {
	rpcclient.Client
	contractAddr string
}

const (
	// hardcode the timeout to 20 seconds. We can expose it to the params once needed
	DefaultTimeout = 20 * time.Second
)

//////////////////////////////
// CONSTRUCTOR
//////////////////////////////

func NewCosmWasmClient(rpcClient rpcclient.Client, contractAddr string) *CosmWasmClient {
	return &CosmWasmClient{
		Client:       rpcClient,
		contractAddr: contractAddr,
	}
}

//////////////////////////////
// METHODS
//////////////////////////////

func (cwClient *CosmWasmClient) QueryListOfVotedFinalityProviders(
	queryParams *types.Block,
) ([]string, error) {
	queryData, err := createBlockVotersQueryData(queryParams)
	if err != nil {
		return nil, err
	}

	resp, err := cwClient.querySmartContractState(queryData)
	if err != nil {
		return nil, err
	}
	// BlockVoters's return type is Option<HashSet<String>> in contract
	// Check empty response before unmarshaling
	if len(resp.Data) == 0 {
		return nil, nil
	}

	votedFpPkHexList := &[]string{}
	if err := json.Unmarshal(resp.Data, votedFpPkHexList); err != nil {
		return nil, err
	}

	return *votedFpPkHexList, nil
}

func (cwClient *CosmWasmClient) QueryConsumerId() (string, error) {
	queryData, err := createConfigQueryData()
	if err != nil {
		return "", err
	}

	resp, err := cwClient.querySmartContractState(queryData)
	if err != nil {
		return "", err
	}

	var data contractConfigResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return "", err
	}

	return data.ConsumerId, nil
}

func (cwClient *CosmWasmClient) QueryIsEnabled() (bool, error) {
	queryData, err := createIsEnabledQueryData()
	if err != nil {
		return false, err
	}

	resp, err := cwClient.querySmartContractState(queryData)
	if err != nil {
		return false, err
	}

	var isEnabled bool
	if err := json.Unmarshal(resp.Data, &isEnabled); err != nil {
		return false, err
	}

	return isEnabled, nil
}

func (cwClient *CosmWasmClient) QueryIsBlockForked(blockHeight uint64) (bool, error) {
	queryData, err := createIsForkedBlockQueryData(blockHeight)
	if err != nil {
		return false, err
	}

	resp, err := cwClient.querySmartContractState(queryData)
	if err != nil {
		return false, err
	}

	var isForked bool
	if err := json.Unmarshal(resp.Data, &isForked); err != nil {
		return false, err
	}

	return isForked, nil
}

// TODO: for temp testing only, to remove if not used
func (cwClient *CosmWasmClient) QueryForkedBlocksInRange(startHeight, endHeight uint64) ([]blockRange, error) {
	queryData, err := createForkedBlocksInRangeQueryData(startHeight, endHeight)
	if err != nil {
		return nil, err
	}

	resp, err := cwClient.querySmartContractState(queryData)
	if err != nil {
		return nil, err
	}

	var forkedBlocks []blockRange
	if err := json.Unmarshal(resp.Data, &forkedBlocks); err != nil {
		return nil, err
	}

	return forkedBlocks, nil
}

//////////////////////////////
// INTERNAL
//////////////////////////////

func createBlockVotersQueryData(queryParams *types.Block) ([]byte, error) {
	queryData := ContractQueryMsgs{
		BlockVoters: &blockVotersQuery{
			Height: queryParams.BlockHeight,
			Hash:   queryParams.BlockHash,
		},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type contractConfigResponse struct {
	ConsumerId string `json:"consumer_id"`
}

type ContractQueryMsgs struct {
	Config              *contractConfig     `json:"config,omitempty"`
	BlockVoters         *blockVotersQuery   `json:"block_voters,omitempty"`
	IsBlockForked       *isBlockForkedQuery `json:"is_block_forked,omitempty"`
	ForkedBlocksInRange *blockRange         `json:"forked_blocks_in_range,omitempty"`
	IsEnabled           *isEnabledQuery     `json:"is_enabled,omitempty"`
}

type blockRange struct {
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
}

type blockVotersQuery struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`
}

type isBlockForkedQuery struct {
	Height uint64 `json:"height"`
}

type isEnabledQuery struct{}

type contractConfig struct{}

func createConfigQueryData() ([]byte, error) {
	queryData := ContractQueryMsgs{
		Config: &contractConfig{},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createIsEnabledQueryData() ([]byte, error) {
	queryData := ContractQueryMsgs{
		IsEnabled: &isEnabledQuery{},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createIsForkedBlockQueryData(blockHeight uint64) ([]byte, error) {
	queryData := ContractQueryMsgs{
		IsBlockForked: &isBlockForkedQuery{
			Height: blockHeight,
		},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createForkedBlocksInRangeQueryData(startHeight, endHeight uint64) ([]byte, error) {
	queryData := ContractQueryMsgs{
		ForkedBlocksInRange: &blockRange{
			Start: startHeight,
			End:   endHeight,
		},
	}
	data, err := json.Marshal(queryData)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// querySmartContractState queries the smart contract state given the contract address and query data
func (cwClient *CosmWasmClient) querySmartContractState(
	queryData []byte,
) (*wasmtypes.QuerySmartContractStateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	sdkClientCtx := cosmosclient.Context{Client: cwClient.Client}
	wasmQueryClient := wasmtypes.NewQueryClient(sdkClientCtx)

	req := &wasmtypes.QuerySmartContractStateRequest{
		Address:   cwClient.contractAddr,
		QueryData: queryData,
	}
	return wasmQueryClient.SmartContractState(ctx, req)
}
