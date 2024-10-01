// Code generated by MockGen. DO NOT EDIT.
// Source: db/interface.go
//
// Generated by this command:
//
//	mockgen -source=db/interface.go -package mocks -destination ./testutil/mocks/db_mock.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	types "github.com/babylonlabs-io/finality-gadget/types"
	gomock "go.uber.org/mock/gomock"
)

// MockIDatabaseHandler is a mock of IDatabaseHandler interface.
type MockIDatabaseHandler struct {
	ctrl     *gomock.Controller
	recorder *MockIDatabaseHandlerMockRecorder
}

// MockIDatabaseHandlerMockRecorder is the mock recorder for MockIDatabaseHandler.
type MockIDatabaseHandlerMockRecorder struct {
	mock *MockIDatabaseHandler
}

// NewMockIDatabaseHandler creates a new mock instance.
func NewMockIDatabaseHandler(ctrl *gomock.Controller) *MockIDatabaseHandler {
	mock := &MockIDatabaseHandler{ctrl: ctrl}
	mock.recorder = &MockIDatabaseHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIDatabaseHandler) EXPECT() *MockIDatabaseHandlerMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockIDatabaseHandler) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockIDatabaseHandlerMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIDatabaseHandler)(nil).Close))
}

// CreateInitialSchema mocks base method.
func (m *MockIDatabaseHandler) CreateInitialSchema() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateInitialSchema")
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateInitialSchema indicates an expected call of CreateInitialSchema.
func (mr *MockIDatabaseHandlerMockRecorder) CreateInitialSchema() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateInitialSchema", reflect.TypeOf((*MockIDatabaseHandler)(nil).CreateInitialSchema))
}

// GetActivatedTimestamp mocks base method.
func (m *MockIDatabaseHandler) GetActivatedTimestamp() (uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetActivatedTimestamp")
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetActivatedTimestamp indicates an expected call of GetActivatedTimestamp.
func (mr *MockIDatabaseHandlerMockRecorder) GetActivatedTimestamp() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetActivatedTimestamp", reflect.TypeOf((*MockIDatabaseHandler)(nil).GetActivatedTimestamp))
}

// GetBlockByHash mocks base method.
func (m *MockIDatabaseHandler) GetBlockByHash(hash string) (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockByHash", hash)
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockByHash indicates an expected call of GetBlockByHash.
func (mr *MockIDatabaseHandlerMockRecorder) GetBlockByHash(hash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockByHash", reflect.TypeOf((*MockIDatabaseHandler)(nil).GetBlockByHash), hash)
}

// GetBlockByHeight mocks base method.
func (m *MockIDatabaseHandler) GetBlockByHeight(height uint64) (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBlockByHeight", height)
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBlockByHeight indicates an expected call of GetBlockByHeight.
func (mr *MockIDatabaseHandlerMockRecorder) GetBlockByHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBlockByHeight", reflect.TypeOf((*MockIDatabaseHandler)(nil).GetBlockByHeight), height)
}

// InsertBlock mocks base method.
func (m *MockIDatabaseHandler) InsertBlock(block *types.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InsertBlock", block)
	ret0, _ := ret[0].(error)
	return ret0
}

// InsertBlock indicates an expected call of InsertBlock.
func (mr *MockIDatabaseHandlerMockRecorder) InsertBlock(block any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertBlock", reflect.TypeOf((*MockIDatabaseHandler)(nil).InsertBlock), block)
}

// QueryEarliestConsecutivelyFinalizedBlock mocks base method.
func (m *MockIDatabaseHandler) QueryEarliestConsecutivelyFinalizedBlock() (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryEarliestConsecutivelyFinalizedBlock")
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryEarliestConsecutivelyFinalizedBlock indicates an expected call of QueryEarliestConsecutivelyFinalizedBlock.
func (mr *MockIDatabaseHandlerMockRecorder) QueryEarliestConsecutivelyFinalizedBlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryEarliestConsecutivelyFinalizedBlock", reflect.TypeOf((*MockIDatabaseHandler)(nil).QueryEarliestConsecutivelyFinalizedBlock))
}

// QueryIsBlockFinalizedByHash mocks base method.
func (m *MockIDatabaseHandler) QueryIsBlockFinalizedByHash(hash string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryIsBlockFinalizedByHash", hash)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryIsBlockFinalizedByHash indicates an expected call of QueryIsBlockFinalizedByHash.
func (mr *MockIDatabaseHandlerMockRecorder) QueryIsBlockFinalizedByHash(hash any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryIsBlockFinalizedByHash", reflect.TypeOf((*MockIDatabaseHandler)(nil).QueryIsBlockFinalizedByHash), hash)
}

// QueryIsBlockFinalizedByHeight mocks base method.
func (m *MockIDatabaseHandler) QueryIsBlockFinalizedByHeight(height uint64) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryIsBlockFinalizedByHeight", height)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryIsBlockFinalizedByHeight indicates an expected call of QueryIsBlockFinalizedByHeight.
func (mr *MockIDatabaseHandlerMockRecorder) QueryIsBlockFinalizedByHeight(height any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryIsBlockFinalizedByHeight", reflect.TypeOf((*MockIDatabaseHandler)(nil).QueryIsBlockFinalizedByHeight), height)
}

// QueryLatestFinalizedBlock mocks base method.
func (m *MockIDatabaseHandler) QueryLatestFinalizedBlock() (*types.Block, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryLatestFinalizedBlock")
	ret0, _ := ret[0].(*types.Block)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryLatestFinalizedBlock indicates an expected call of QueryLatestFinalizedBlock.
func (mr *MockIDatabaseHandlerMockRecorder) QueryLatestFinalizedBlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryLatestFinalizedBlock", reflect.TypeOf((*MockIDatabaseHandler)(nil).QueryLatestFinalizedBlock))
}

// SaveActivatedTimestamp mocks base method.
func (m *MockIDatabaseHandler) SaveActivatedTimestamp(timestamp uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveActivatedTimestamp", timestamp)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveActivatedTimestamp indicates an expected call of SaveActivatedTimestamp.
func (mr *MockIDatabaseHandlerMockRecorder) SaveActivatedTimestamp(timestamp any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveActivatedTimestamp", reflect.TypeOf((*MockIDatabaseHandler)(nil).SaveActivatedTimestamp), timestamp)
}
