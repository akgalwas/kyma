// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

import v1alpha1 "github.com/kyma-project/kyma/components/event-bus/generated/push/clientset/versioned/typed/eventing.kyma-project.io/v1alpha1"

// SubscriptionsGetter is an autogenerated mock type for the SubscriptionsGetter type
type SubscriptionsGetter struct {
	mock.Mock
}

// Subscriptions provides a mock function with given fields: namespace
func (_m *SubscriptionsGetter) Subscriptions(namespace string) v1alpha1.SubscriptionInterface {
	ret := _m.Called(namespace)

	var r0 v1alpha1.SubscriptionInterface
	if rf, ok := ret.Get(0).(func(string) v1alpha1.SubscriptionInterface); ok {
		r0 = rf(namespace)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.SubscriptionInterface)
		}
	}

	return r0
}
