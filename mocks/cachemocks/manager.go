// Code generated by mockery v1.0.0. DO NOT EDIT.

package cachemocks

import (
	cache "github.com/hyperledger/firefly/internal/cache"
	mock "github.com/stretchr/testify/mock"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// GetCache provides a mock function with given fields: cc
func (_m *Manager) GetCache(cc *cache.CConfig) (cache.CInterface, error) {
	ret := _m.Called(cc)

	var r0 cache.CInterface
	if rf, ok := ret.Get(0).(func(*cache.CConfig) cache.CInterface); ok {
		r0 = rf(cc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(cache.CInterface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*cache.CConfig) error); ok {
		r1 = rf(cc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListKeys provides a mock function with given fields:
func (_m *Manager) ListKeys() []string {
	ret := _m.Called()

	var r0 []string
	if rf, ok := ret.Get(0).(func() []string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	return r0
}