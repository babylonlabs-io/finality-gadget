// Code generated by MockGen. DO NOT EDIT.
// Source: sdk/client/interface.go
//
// Generated by this command:
//
//	mockgen -source=sdk/client/interface.go -package mocks -destination /Users/zidong/Documents/Projects/babylon-finality-gadget/testutil/mocks/sdkclient_mock.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	cwclient "github.com/babylonchain/babylon-finality-gadget/sdk/cwclient"
	gomock "go.uber.org/mock/gomock"
)

// MockISdkClient is a mock of ISdkClient interface.
type MockISdkClient struct {
	ctrl     *gomock.Controller
	recorder *MockISdkClientMockRecorder
}

// MockISdkClientMockRecorder is the mock recorder for MockISdkClient.
type MockISdkClientMockRecorder struct {
	mock *MockISdkClient
}

// NewMockISdkClient creates a new mock instance.
func NewMockISdkClient(ctrl *gomock.Controller) *MockISdkClient {
	mock := &MockISdkClient{ctrl: ctrl}
	mock.recorder = &MockISdkClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockISdkClient) EXPECT() *MockISdkClientMockRecorder {
	return m.recorder
}

// QueryBlockRangeBabylonFinalized mocks base method.
func (m *MockISdkClient) QueryBlockRangeBabylonFinalized(queryBlocks []*cwclient.L2Block) (*uint64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryBlockRangeBabylonFinalized", queryBlocks)
	ret0, _ := ret[0].(*uint64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryBlockRangeBabylonFinalized indicates an expected call of QueryBlockRangeBabylonFinalized.
func (mr *MockISdkClientMockRecorder) QueryBlockRangeBabylonFinalized(queryBlocks any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryBlockRangeBabylonFinalized", reflect.TypeOf((*MockISdkClient)(nil).QueryBlockRangeBabylonFinalized), queryBlocks)
}

// QueryIsBlockBabylonFinalized mocks base method.
func (m *MockISdkClient) QueryIsBlockBabylonFinalized(queryParams *cwclient.L2Block) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "QueryIsBlockBabylonFinalized", queryParams)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// QueryIsBlockBabylonFinalized indicates an expected call of QueryIsBlockBabylonFinalized.
func (mr *MockISdkClientMockRecorder) QueryIsBlockBabylonFinalized(queryParams any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryIsBlockBabylonFinalized", reflect.TypeOf((*MockISdkClient)(nil).QueryIsBlockBabylonFinalized), queryParams)
}