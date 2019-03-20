// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"
import storage "github.com/kyma-project/kyma/components/console-backend-service/internal/domain/content/storage"

// minioOpenApiSpecGetter is an autogenerated mock type for the minioOpenApiSpecGetter type
type minioOpenApiSpecGetter struct {
	mock.Mock
}

// OpenApiSpec provides a mock function with given fields: id
func (_m *minioOpenApiSpecGetter) OpenApiSpec(id string) (*storage.OpenApiSpec, bool, error) {
	ret := _m.Called(id)

	var r0 *storage.OpenApiSpec
	if rf, ok := ret.Get(0).(func(string) *storage.OpenApiSpec); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.OpenApiSpec)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(id)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
