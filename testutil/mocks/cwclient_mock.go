// Code generated by MockGen. DO NOT EDIT.
// Source: cwclient/interface.go
//
// Generated by this command:
//
//	mockgen -source=cwclient/interface.go -package mocks -destination ./testutil/mocks/cwclient_mock.go
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	types "github.com/babylonlabs-io/finality-gadget/types"
	gomock "go.uber.org/mock/gomock"
)

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