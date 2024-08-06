// Code generated by MockGen. DO NOT EDIT.
// Source: btcclient/interface.go
//
// Generated by this command:
//
//	mockgen -source=btcclient/interface.go -package mocks -destination ./testutil/mocks/btcclient_mock.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	chainhash "github.com/btcsuite/btcd/chaincfg/chainhash"
	wire "github.com/btcsuite/btcd/wire"
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