// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import apperrors "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
import sync "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/sync"

// Reconciler is an autogenerated mock type for the Reconciler type
type Reconciler struct {
	mock.Mock
}

// Do provides a mock function with given fields: applications
func (_m *Reconciler) Do(applications []model.Application) ([]sync.ApplicationAction, apperrors.AppError) {
	ret := _m.Called(applications)

	var r0 []sync.ApplicationAction
	if rf, ok := ret.Get(0).(func([]model.Application) []sync.ApplicationAction); ok {
		r0 = rf(applications)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]sync.ApplicationAction)
		}
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func([]model.Application) apperrors.AppError); ok {
		r1 = rf(applications)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
