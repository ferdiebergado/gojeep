// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ferdiebergado/gojeep/internal/service (interfaces: Service)
//
// Generated by this command:
//
//	mockgen -destination=mock/service_mock.go -package=mock . Service
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
	isgomock struct{}
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// PingDB mocks base method.
func (m *MockService) PingDB(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingDB", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// PingDB indicates an expected call of PingDB.
func (mr *MockServiceMockRecorder) PingDB(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingDB", reflect.TypeOf((*MockService)(nil).PingDB), ctx)
}
