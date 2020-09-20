// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	versioned "github.com/kyma-project/kyma/components/system-broker-agent/pkg/clientcs/clientset/versioned"
	internalinterfaces "github.com/kyma-project/kyma/components/system-broker-agent/pkg/clientcs/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/clientcs/listers/applicationconnector/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ClusterSystemInformer provides access to a shared informer and lister for
// ClusterSystems.
type ClusterSystemInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ClusterSystemLister
}

type clusterSystemInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewClusterSystemInformer constructs a new informer for ClusterSystem type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewClusterSystemInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredClusterSystemInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredClusterSystemInformer constructs a new informer for ClusterSystem type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredClusterSystemInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApplicationconnectorV1alpha1().ClusterSystems().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.ApplicationconnectorV1alpha1().ClusterSystems().Watch(context.TODO(), options)
			},
		},
		&applicationconnectorv1alpha1.ClusterSystem{},
		resyncPeriod,
		indexers,
	)
}

func (f *clusterSystemInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredClusterSystemInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *clusterSystemInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&applicationconnectorv1alpha1.ClusterSystem{}, f.defaultInformer)
}

func (f *clusterSystemInformer) Lister() v1alpha1.ClusterSystemLister {
	return v1alpha1.NewClusterSystemLister(f.Informer().GetIndexer())
}
