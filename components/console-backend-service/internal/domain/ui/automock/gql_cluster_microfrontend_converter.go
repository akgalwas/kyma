// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import gqlschema "github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
import mock "github.com/stretchr/testify/mock"

import v1alpha1 "github.com/kyma-project/kyma/common/microfrontend-client/pkg/apis/ui/v1alpha1"

// gqlClusterMicrofrontendConverter is an autogenerated mock type for the gqlClusterMicrofrontendConverter type
type gqlClusterMicrofrontendConverter struct {
	mock.Mock
}

// ToGQL provides a mock function with given fields: in
func (_m *gqlClusterMicrofrontendConverter) ToGQL(in *v1alpha1.ClusterMicroFrontend) (*gqlschema.ClusterMicrofrontend, error) {
	ret := _m.Called(in)

	var r0 *gqlschema.ClusterMicrofrontend
	if rf, ok := ret.Get(0).(func(*v1alpha1.ClusterMicroFrontend) *gqlschema.ClusterMicrofrontend); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gqlschema.ClusterMicrofrontend)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*v1alpha1.ClusterMicroFrontend) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGQLs provides a mock function with given fields: in
func (_m *gqlClusterMicrofrontendConverter) ToGQLs(in []*v1alpha1.ClusterMicroFrontend) ([]gqlschema.ClusterMicrofrontend, error) {
	ret := _m.Called(in)

	var r0 []gqlschema.ClusterMicrofrontend
	if rf, ok := ret.Get(0).(func([]*v1alpha1.ClusterMicroFrontend) []gqlschema.ClusterMicrofrontend); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]gqlschema.ClusterMicrofrontend)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*v1alpha1.ClusterMicroFrontend) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
