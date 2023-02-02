// Code generated by mockery v2.16.0. DO NOT EDIT.

package spieventsmocks

import (
	http "net/http"

	core "github.com/hyperledger/firefly/pkg/core"

	mock "github.com/stretchr/testify/mock"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Dispatch provides a mock function with given fields: changeEvent
func (_m *Manager) Dispatch(changeEvent *core.ChangeEvent) {
	_m.Called(changeEvent)
}

// ServeHTTPWebSocketListener provides a mock function with given fields: res, req
func (_m *Manager) ServeHTTPWebSocketListener(res http.ResponseWriter, req *http.Request) {
	_m.Called(res, req)
}

// WaitStop provides a mock function with given fields:
func (_m *Manager) WaitStop() {
	_m.Called()
}

type mockConstructorTestingTNewManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewManager creates a new instance of Manager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewManager(t mockConstructorTestingTNewManager) *Manager {
	mock := &Manager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}