// Code generated by MockGen. DO NOT EDIT.
// Source: finalitygadget/interface.go
//
// Generated by this command:
//
//	mockgen -source=finalitygadget/interface.go -package mocks -destination ./testutil/mocks/finalitygadget_mock.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	types "github.com/babylonlabs-io/finality-gadget/types"
	gomock "go.uber.org/mock/gomock"
)

// MockIFinalityGadget is a mock of IFinalityGadget interface.
type MockIFinalityGadget struct {
	ctrl     *gomock.Controller
	recorder *MockIFinalityGadgetMockRecorder
	isgomock struct{}
}

// MockIFinalityGadgetMockRecorder is the mock recorder for MockIFinalityGadget.
type MockIFinalityGadgetMockRecorder struct {
	mock *MockIFinalityGadget
}

// NewMockIFinalityGadget creates a new mock instance.
func NewMockIFinalityGadget(ctrl *gomock.Controller) *MockIFinalityGadget {
	mock := &MockIFinalityGadget{ctrl: ctrl}
	mock.recorder = &MockIFinalityGadgetMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIFinalityGadget) EXPECT() *MockIFinalityGadgetMockRecorder {
	return m.recorder
}

// GetBlockByHash mocks base method.
func (m *MockIFinalityGadget) GetBlockByHash(hash string) (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockByHash", hash)
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockByHash indicates an expected call of GetBlockByHash.
func (mr *MockIFinalityGadgetMockRecorder) GetBlockByHash(hash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockByHash", reflect.TypeOf((*MockIFinalityGadget)(nil).GetBlockByHash), hash)
}

// GetBlockByHeight mocks base method.
func (m *MockIFinalityGadget) GetBlockByHeight(height uint64) (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockByHeight", height)
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockByHeight indicates an expected call of GetBlockByHeight.
func (mr *MockIFinalityGadgetMockRecorder) GetBlockByHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockByHeight", reflect.TypeOf((*MockIFinalityGadget)(nil).GetBlockByHeight), height)
}

// QueryBlockRangeBabylonFinalized mocks base method.
func (m *MockIFinalityGadget) QueryBlockRangeBabylonFinalized(queryBlocks []*types.Block) (*uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryBlockRangeBabylonFinalized", queryBlocks)
	ret0, _ := ret[0].(*uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryBlockRangeBabylonFinalized indicates an expected call of QueryBlockRangeBabylonFinalized.
func (mr *MockIFinalityGadgetMockRecorder) QueryBlockRangeBabylonFinalized(queryBlocks any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryBlockRangeBabylonFinalized", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryBlockRangeBabylonFinalized), queryBlocks)
}

// QueryBtcStakingActivatedTimestamp mocks base method.
func (m *MockIFinalityGadget) QueryBtcStakingActivatedTimestamp() (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryBtcStakingActivatedTimestamp")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryBtcStakingActivatedTimestamp indicates an expected call of QueryBtcStakingActivatedTimestamp.
func (mr *MockIFinalityGadgetMockRecorder) QueryBtcStakingActivatedTimestamp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryBtcStakingActivatedTimestamp", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryBtcStakingActivatedTimestamp))
}

// QueryChainSyncStatus mocks base method.
func (m *MockIFinalityGadget) QueryChainSyncStatus() (*types.ChainSyncStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryChainSyncStatus")
	ret0, _ := ret[0].(*types.ChainSyncStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryChainSyncStatus indicates an expected call of QueryChainSyncStatus.
func (mr *MockIFinalityGadgetMockRecorder) QueryChainSyncStatus() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryChainSyncStatus", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryChainSyncStatus))
}

// QueryIsBlockBabylonFinalized mocks base method.
func (m *MockIFinalityGadget) QueryIsBlockBabylonFinalized(block *types.Block) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryIsBlockBabylonFinalized", block)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryIsBlockBabylonFinalized indicates an expected call of QueryIsBlockBabylonFinalized.
func (mr *MockIFinalityGadgetMockRecorder) QueryIsBlockBabylonFinalized(block any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryIsBlockBabylonFinalized", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryIsBlockBabylonFinalized), block)
}

// QueryIsBlockBabylonFinalizedFromBabylon mocks base method.
func (m *MockIFinalityGadget) QueryIsBlockBabylonFinalizedFromBabylon(block *types.Block) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryIsBlockBabylonFinalizedFromBabylon", block)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryIsBlockBabylonFinalizedFromBabylon indicates an expected call of QueryIsBlockBabylonFinalizedFromBabylon.
func (mr *MockIFinalityGadgetMockRecorder) QueryIsBlockBabylonFinalizedFromBabylon(block any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryIsBlockBabylonFinalizedFromBabylon", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryIsBlockBabylonFinalizedFromBabylon), block)
}

// QueryIsBlockFinalizedByHash mocks base method.
func (m *MockIFinalityGadget) QueryIsBlockFinalizedByHash(hash string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryIsBlockFinalizedByHash", hash)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryIsBlockFinalizedByHash indicates an expected call of QueryIsBlockFinalizedByHash.
func (mr *MockIFinalityGadgetMockRecorder) QueryIsBlockFinalizedByHash(hash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryIsBlockFinalizedByHash", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryIsBlockFinalizedByHash), hash)
}

// QueryIsBlockFinalizedByHeight mocks base method.
func (m *MockIFinalityGadget) QueryIsBlockFinalizedByHeight(height uint64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryIsBlockFinalizedByHeight", height)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryIsBlockFinalizedByHeight indicates an expected call of QueryIsBlockFinalizedByHeight.
func (mr *MockIFinalityGadgetMockRecorder) QueryIsBlockFinalizedByHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryIsBlockFinalizedByHeight", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryIsBlockFinalizedByHeight), height)
}

// QueryLatestFinalizedBlock mocks base method.
func (m *MockIFinalityGadget) QueryLatestFinalizedBlock() (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryLatestFinalizedBlock")
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryLatestFinalizedBlock indicates an expected call of QueryLatestFinalizedBlock.
func (mr *MockIFinalityGadgetMockRecorder) QueryLatestFinalizedBlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryLatestFinalizedBlock", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryLatestFinalizedBlock))
}

// QueryTransactionStatus mocks base method.
func (m *MockIFinalityGadget) QueryTransactionStatus(txHash string) (*types.TransactionInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryTransactionStatus", txHash)
	ret0, _ := ret[0].(*types.TransactionInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryTransactionStatus indicates an expected call of QueryTransactionStatus.
func (mr *MockIFinalityGadgetMockRecorder) QueryTransactionStatus(txHash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryTransactionStatus", reflect.TypeOf((*MockIFinalityGadget)(nil).QueryTransactionStatus), txHash)
}
