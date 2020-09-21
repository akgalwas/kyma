package main

import (
	appclient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/apperrors"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type k8sResourceClientSets struct {
	core        *kubernetes.Clientset
	application *appclient.Clientset
	dynamic     dynamic.Interface
}

func k8sResourceClients(k8sConfig *restclient.Config) (*k8sResourceClientSets, error) {
	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create k8s core client")
	}

	applicationClientset, err := appclient.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create k8s application client, %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("Failed to create dynamic client, %s", err)
	}

	return &k8sResourceClientSets{
		core:        coreClientset,
		application: applicationClientset,
		dynamic:     dynamicClient,
	}, nil
}

//func newMetricsLogger(loggingTimeInterval time.Duration) (metrics.Logger, error) {
//	config, err := restclient.InClusterConfig()
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to get cluster config")
//	}
//
//	resourcesClientset, err := kubernetes.NewForConfig(config)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to create resources clientset for config")
//	}
//
//	metricsClientset, err := clientset.NewForConfig(config)
//	if err != nil {
//		return nil, errors.Wrap(err, "failed to create metrics clientset for config")
//	}
//
//	return metrics.NewMetricsLogger(resourcesClientset, metricsClientset, loggingTimeInterval), nil
//}
