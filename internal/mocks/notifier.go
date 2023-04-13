// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/authelia/authelia/v4/internal/notification (interfaces: Notifier)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	mail "net/mail"
	reflect "reflect"

	templates "github.com/authelia/authelia/v4/internal/templates"
	gomock "github.com/golang/mock/gomock"
)

// MockNotifier is a mock of Notifier interface.
type MockNotifier struct {
	ctrl     *gomock.Controller
	recorder *MockNotifierMockRecorder
}

// MockNotifierMockRecorder is the mock recorder for MockNotifier.
type MockNotifierMockRecorder struct {
	mock *MockNotifier
}

// NewMockNotifier creates a new mock instance.
func NewMockNotifier(ctrl *gomock.Controller) *MockNotifier {
	mock := &MockNotifier{ctrl: ctrl}
	mock.recorder = &MockNotifierMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNotifier) EXPECT() *MockNotifierMockRecorder {
	return m.recorder
}

// Send mocks base method.
func (m *MockNotifier) Send(arg0 context.Context, arg1 mail.Address, arg2 string, arg3 *templates.EmailTemplate, arg4 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Send", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(error)
	return ret0
}

// Send indicates an expected call of Send.
func (mr *MockNotifierMockRecorder) Send(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Send", reflect.TypeOf((*MockNotifier)(nil).Send), arg0, arg1, arg2, arg3, arg4)
}

// StartupCheck mocks base method.
func (m *MockNotifier) StartupCheck() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StartupCheck")
	ret0, _ := ret[0].(error)
	return ret0
}

// StartupCheck indicates an expected call of StartupCheck.
func (mr *MockNotifierMockRecorder) StartupCheck() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartupCheck", reflect.TypeOf((*MockNotifier)(nil).StartupCheck))
}
