// Code generated by MockGen. DO NOT EDIT.
// Source: types/finalitygadget.go
//
// Generated by this command:
//
//	mockgen -source=types/finalitygadget.go -package mocks -destination ./testutil/mocks/finalitygadget_mock.go
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

// GetBlockStatusByHash mocks base method.
func (m *MockIFinalityGadget) GetBlockStatusByHash(hash string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockStatusByHash", hash)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockStatusByHash indicates an expected call of GetBlockStatusByHash.
func (mr *MockIFinalityGadgetMockRecorder) GetBlockStatusByHash(hash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockStatusByHash", reflect.TypeOf((*MockIFinalityGadget)(nil).GetBlockStatusByHash), hash)
}

// GetBlockStatusByHeight mocks base method.
func (m *MockIFinalityGadget) GetBlockStatusByHeight(height uint64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockStatusByHeight", height)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockStatusByHeight indicates an expected call of GetBlockStatusByHeight.
func (mr *MockIFinalityGadgetMockRecorder) GetBlockStatusByHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockStatusByHeight", reflect.TypeOf((*MockIFinalityGadget)(nil).GetBlockStatusByHeight), height)
}

// GetLatestBlock mocks base method.
func (m *MockIFinalityGadget) GetLatestBlock() (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLatestBlock")
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLatestBlock indicates an expected call of GetLatestBlock.
func (mr *MockIFinalityGadgetMockRecorder) GetLatestBlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLatestBlock", reflect.TypeOf((*MockIFinalityGadget)(nil).GetLatestBlock))
}

// InsertBlock mocks base method.
func (m *MockIFinalityGadget) InsertBlock(block *types.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertBlock", block)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertBlock indicates an expected call of InsertBlock.
func (mr *MockIFinalityGadgetMockRecorder) InsertBlock(block any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertBlock", reflect.TypeOf((*MockIFinalityGadget)(nil).InsertBlock), block)
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
