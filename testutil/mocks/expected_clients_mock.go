// Code generated by MockGen. DO NOT EDIT.
// Source: finalitygadget/expected_clients.go
//
// Generated by this command:
//
//	mockgen -source=finalitygadget/expected_clients.go -package mocks -destination ./testutil/mocks/expected_clients_mock.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	big "math/big"
	reflect "reflect"

	types "github.com/babylonlabs-io/finality-gadget/types"
	chainhash "github.com/btcsuite/btcd/chaincfg/chainhash"
	wire "github.com/btcsuite/btcd/wire"
	types0 "github.com/ethereum/go-ethereum/core/types"
	gomock "go.uber.org/mock/gomock"
)

// MockIBitcoinClient is a mock of IBitcoinClient interface.
type MockIBitcoinClient struct {
	ctrl     *gomock.Controller
	recorder *MockIBitcoinClientMockRecorder
}

// MockIBitcoinClientMockRecorder is the mock recorder for MockIBitcoinClient.
type MockIBitcoinClientMockRecorder struct {
	mock *MockIBitcoinClient
}

// NewMockIBitcoinClient creates a new mock instance.
func NewMockIBitcoinClient(ctrl *gomock.Controller) *MockIBitcoinClient {
	mock := &MockIBitcoinClient{ctrl: ctrl}
	mock.recorder = &MockIBitcoinClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIBitcoinClient) EXPECT() *MockIBitcoinClientMockRecorder {
	return m.recorder
}

// GetBlockCount mocks base method.
func (m *MockIBitcoinClient) GetBlockCount() (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockCount")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockCount indicates an expected call of GetBlockCount.
func (mr *MockIBitcoinClientMockRecorder) GetBlockCount() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockCount", reflect.TypeOf((*MockIBitcoinClient)(nil).GetBlockCount))
}

// GetBlockHashByHeight mocks base method.
func (m *MockIBitcoinClient) GetBlockHashByHeight(height uint64) (*chainhash.Hash, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockHashByHeight", height)
	ret0, _ := ret[0].(*chainhash.Hash)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockHashByHeight indicates an expected call of GetBlockHashByHeight.
func (mr *MockIBitcoinClientMockRecorder) GetBlockHashByHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockHashByHeight", reflect.TypeOf((*MockIBitcoinClient)(nil).GetBlockHashByHeight), height)
}

// GetBlockHeaderByHash mocks base method.
func (m *MockIBitcoinClient) GetBlockHeaderByHash(blockHash *chainhash.Hash) (*wire.BlockHeader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockHeaderByHash", blockHash)
	ret0, _ := ret[0].(*wire.BlockHeader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockHeaderByHash indicates an expected call of GetBlockHeaderByHash.
func (mr *MockIBitcoinClientMockRecorder) GetBlockHeaderByHash(blockHash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockHeaderByHash", reflect.TypeOf((*MockIBitcoinClient)(nil).GetBlockHeaderByHash), blockHash)
}

// GetBlockHeightByTimestamp mocks base method.
func (m *MockIBitcoinClient) GetBlockHeightByTimestamp(targetTimestamp uint64) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockHeightByTimestamp", targetTimestamp)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockHeightByTimestamp indicates an expected call of GetBlockHeightByTimestamp.
func (mr *MockIBitcoinClientMockRecorder) GetBlockHeightByTimestamp(targetTimestamp any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockHeightByTimestamp", reflect.TypeOf((*MockIBitcoinClient)(nil).GetBlockHeightByTimestamp), targetTimestamp)
}

// GetBlockTimestampByHeight mocks base method.
func (m *MockIBitcoinClient) GetBlockTimestampByHeight(height uint64) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockTimestampByHeight", height)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockTimestampByHeight indicates an expected call of GetBlockTimestampByHeight.
func (mr *MockIBitcoinClientMockRecorder) GetBlockTimestampByHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockTimestampByHeight", reflect.TypeOf((*MockIBitcoinClient)(nil).GetBlockTimestampByHeight), height)
}

// MockIBabylonClient is a mock of IBabylonClient interface.
type MockIBabylonClient struct {
	ctrl     *gomock.Controller
	recorder *MockIBabylonClientMockRecorder
}

// MockIBabylonClientMockRecorder is the mock recorder for MockIBabylonClient.
type MockIBabylonClientMockRecorder struct {
	mock *MockIBabylonClient
}

// NewMockIBabylonClient creates a new mock instance.
func NewMockIBabylonClient(ctrl *gomock.Controller) *MockIBabylonClient {
	mock := &MockIBabylonClient{ctrl: ctrl}
	mock.recorder = &MockIBabylonClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIBabylonClient) EXPECT() *MockIBabylonClientMockRecorder {
	return m.recorder
}

// QueryAllFpBtcPubKeys mocks base method.
func (m *MockIBabylonClient) QueryAllFpBtcPubKeys(consumerId string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryAllFpBtcPubKeys", consumerId)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryAllFpBtcPubKeys indicates an expected call of QueryAllFpBtcPubKeys.
func (mr *MockIBabylonClientMockRecorder) QueryAllFpBtcPubKeys(consumerId any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryAllFpBtcPubKeys", reflect.TypeOf((*MockIBabylonClient)(nil).QueryAllFpBtcPubKeys), consumerId)
}

// QueryEarliestActiveDelBtcHeight mocks base method.
func (m *MockIBabylonClient) QueryEarliestActiveDelBtcHeight(fpPubkeyHexList []string) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryEarliestActiveDelBtcHeight", fpPubkeyHexList)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryEarliestActiveDelBtcHeight indicates an expected call of QueryEarliestActiveDelBtcHeight.
func (mr *MockIBabylonClientMockRecorder) QueryEarliestActiveDelBtcHeight(fpPubkeyHexList any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryEarliestActiveDelBtcHeight", reflect.TypeOf((*MockIBabylonClient)(nil).QueryEarliestActiveDelBtcHeight), fpPubkeyHexList)
}

// QueryFpPower mocks base method.
func (m *MockIBabylonClient) QueryFpPower(fpPubkeyHex string, btcHeight uint64) (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryFpPower", fpPubkeyHex, btcHeight)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryFpPower indicates an expected call of QueryFpPower.
func (mr *MockIBabylonClientMockRecorder) QueryFpPower(fpPubkeyHex, btcHeight any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryFpPower", reflect.TypeOf((*MockIBabylonClient)(nil).QueryFpPower), fpPubkeyHex, btcHeight)
}

// QueryMultiFpPower mocks base method.
func (m *MockIBabylonClient) QueryMultiFpPower(fpPubkeyHexList []string, btcHeight uint64) (map[string]uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryMultiFpPower", fpPubkeyHexList, btcHeight)
	ret0, _ := ret[0].(map[string]uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryMultiFpPower indicates an expected call of QueryMultiFpPower.
func (mr *MockIBabylonClientMockRecorder) QueryMultiFpPower(fpPubkeyHexList, btcHeight any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryMultiFpPower", reflect.TypeOf((*MockIBabylonClient)(nil).QueryMultiFpPower), fpPubkeyHexList, btcHeight)
}

// MockICosmWasmClient is a mock of ICosmWasmClient interface.
type MockICosmWasmClient struct {
	ctrl     *gomock.Controller
	recorder *MockICosmWasmClientMockRecorder
}

// MockICosmWasmClientMockRecorder is the mock recorder for MockICosmWasmClient.
type MockICosmWasmClientMockRecorder struct {
	mock *MockICosmWasmClient
}

// NewMockICosmWasmClient creates a new mock instance.
func NewMockICosmWasmClient(ctrl *gomock.Controller) *MockICosmWasmClient {
	mock := &MockICosmWasmClient{ctrl: ctrl}
	mock.recorder = &MockICosmWasmClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockICosmWasmClient) EXPECT() *MockICosmWasmClientMockRecorder {
	return m.recorder
}

// QueryConsumerId mocks base method.
func (m *MockICosmWasmClient) QueryConsumerId() (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryConsumerId")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryConsumerId indicates an expected call of QueryConsumerId.
func (mr *MockICosmWasmClientMockRecorder) QueryConsumerId() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryConsumerId", reflect.TypeOf((*MockICosmWasmClient)(nil).QueryConsumerId))
}

// QueryIsEnabled mocks base method.
func (m *MockICosmWasmClient) QueryIsEnabled() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryIsEnabled")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryIsEnabled indicates an expected call of QueryIsEnabled.
func (mr *MockICosmWasmClientMockRecorder) QueryIsEnabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryIsEnabled", reflect.TypeOf((*MockICosmWasmClient)(nil).QueryIsEnabled))
}

// QueryListOfVotedFinalityProviders mocks base method.
func (m *MockICosmWasmClient) QueryListOfVotedFinalityProviders(queryParams *types.Block) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryListOfVotedFinalityProviders", queryParams)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryListOfVotedFinalityProviders indicates an expected call of QueryListOfVotedFinalityProviders.
func (mr *MockICosmWasmClientMockRecorder) QueryListOfVotedFinalityProviders(queryParams any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryListOfVotedFinalityProviders", reflect.TypeOf((*MockICosmWasmClient)(nil).QueryListOfVotedFinalityProviders), queryParams)
}

// MockIEthL2Client is a mock of IEthL2Client interface.
type MockIEthL2Client struct {
	ctrl     *gomock.Controller
	recorder *MockIEthL2ClientMockRecorder
}

// MockIEthL2ClientMockRecorder is the mock recorder for MockIEthL2Client.
type MockIEthL2ClientMockRecorder struct {
	mock *MockIEthL2Client
}

// NewMockIEthL2Client creates a new mock instance.
func NewMockIEthL2Client(ctrl *gomock.Controller) *MockIEthL2Client {
	mock := &MockIEthL2Client{ctrl: ctrl}
	mock.recorder = &MockIEthL2ClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIEthL2Client) EXPECT() *MockIEthL2ClientMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockIEthL2Client) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close.
func (mr *MockIEthL2ClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIEthL2Client)(nil).Close))
}

// HeaderByNumber mocks base method.
func (m *MockIEthL2Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types0.Header, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HeaderByNumber", ctx, number)
	ret0, _ := ret[0].(*types0.Header)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// HeaderByNumber indicates an expected call of HeaderByNumber.
func (mr *MockIEthL2ClientMockRecorder) HeaderByNumber(ctx, number any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HeaderByNumber", reflect.TypeOf((*MockIEthL2Client)(nil).HeaderByNumber), ctx, number)
}

// TransactionReceipt mocks base method.
func (m *MockIEthL2Client) TransactionReceipt(ctx context.Context, txHash string) (*types0.Receipt, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TransactionReceipt", ctx, txHash)
	ret0, _ := ret[0].(*types0.Receipt)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// TransactionReceipt indicates an expected call of TransactionReceipt.
func (mr *MockIEthL2ClientMockRecorder) TransactionReceipt(ctx, txHash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TransactionReceipt", reflect.TypeOf((*MockIEthL2Client)(nil).TransactionReceipt), ctx, txHash)
}
