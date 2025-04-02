// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ferdiebergado/gojeep/internal/service (interfaces: TokenService)
//
// Generated by this command:
//
//	mockgen -destination=mock/token_service_mock.go -package=mock . TokenService
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockTokenService is a mock of TokenService interface.
type MockTokenService struct {
	ctrl     *gomock.Controller
	recorder *MockTokenServiceMockRecorder
	isgomock struct{}
}

// MockTokenServiceMockRecorder is the mock recorder for MockTokenService.
type MockTokenServiceMockRecorder struct {
	mock *MockTokenService
}

// NewMockTokenService creates a new mock instance.
func NewMockTokenService(ctrl *gomock.Controller) *MockTokenService {
	mock := &MockTokenService{ctrl: ctrl}
	mock.recorder = &MockTokenServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTokenService) EXPECT() *MockTokenServiceMockRecorder {
	return m.recorder
}

// SaveToken mocks base method.
func (m *MockTokenService) SaveToken(ctx context.Context, id, email string, ttl time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SaveToken", ctx, id, email, ttl)
	ret0, _ := ret[0].(error)
	return ret0
}

// SaveToken indicates an expected call of SaveToken.
func (mr *MockTokenServiceMockRecorder) SaveToken(ctx, id, email, ttl any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SaveToken", reflect.TypeOf((*MockTokenService)(nil).SaveToken), ctx, id, email, ttl)
}

// Sign mocks base method.
func (m *MockTokenService) Sign(email string, audience []string, ttl time.Duration) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sign", email, audience, ttl)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Sign indicates an expected call of Sign.
func (mr *MockTokenServiceMockRecorder) Sign(email, audience, ttl any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sign", reflect.TypeOf((*MockTokenService)(nil).Sign), email, audience, ttl)
}

// Verify mocks base method.
func (m *MockTokenService) Verify(tokenString string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Verify", tokenString)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Verify indicates an expected call of Verify.
func (mr *MockTokenServiceMockRecorder) Verify(tokenString any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Verify", reflect.TypeOf((*MockTokenService)(nil).Verify), tokenString)
}
