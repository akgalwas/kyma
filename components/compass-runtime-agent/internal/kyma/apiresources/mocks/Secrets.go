// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import apperrors "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
import applications "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/secrets/model"
import types "k8s.io/apimachinery/pkg/types"

// Secrets is an autogenerated mock type for the Secrets type
type Secrets struct {
	mock.Mock
}

// Create provides a mock function with given fields: application, appUID, serviceID, credentials
func (_m *Secrets) Create(application string, appUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF) (applications.Credentials, apperrors.AppError) {
	ret := _m.Called(application, appUID, serviceID, credentials)

	var r0 applications.Credentials
	if rf, ok := ret.Get(0).(func(string, types.UID, string, *model.CredentialsWithCSRF) applications.Credentials); ok {
		r0 = rf(application, appUID, serviceID, credentials)
	} else {
		r0 = ret.Get(0).(applications.Credentials)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, types.UID, string, *model.CredentialsWithCSRF) apperrors.AppError); ok {
		r1 = rf(application, appUID, serviceID, credentials)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}

// Delete provides a mock function with given fields: name
func (_m *Secrets) Delete(name string) apperrors.AppError {
	ret := _m.Called(name)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(string) apperrors.AppError); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}

// Upsert provides a mock function with given fields: application, appUID, serviceID, credentials
func (_m *Secrets) Upsert(application string, appUID types.UID, serviceID string, credentials *model.CredentialsWithCSRF) (applications.Credentials, apperrors.AppError) {
	ret := _m.Called(application, appUID, serviceID, credentials)

	var r0 applications.Credentials
	if rf, ok := ret.Get(0).(func(string, types.UID, string, *model.CredentialsWithCSRF) applications.Credentials); ok {
		r0 = rf(application, appUID, serviceID, credentials)
	} else {
		r0 = ret.Get(0).(applications.Credentials)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(string, types.UID, string, *model.CredentialsWithCSRF) apperrors.AppError); ok {
		r1 = rf(application, appUID, serviceID, credentials)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
