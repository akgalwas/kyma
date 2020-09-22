package synchronization

import (
	"context"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization/osbapi"
	"github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

//go:generate mockery --name=CRManager
type ClusterSystemCRManager interface {
	Create(ctx context.Context, cc *v1alpha1.ClusterSystem, options v1.CreateOptions) (*v1alpha1.ClusterSystem, error)
	Update(ctx context.Context, cc *v1alpha1.ClusterSystem, options v1.UpdateOptions) (*v1alpha1.ClusterSystem, error)
	Delete(ctx context.Context, name string, options v1.DeleteOptions) error
	Get(ctx context.Context, name string, options v1.GetOptions) (*v1alpha1.ClusterSystem, error)
	List(opts v1.ListOptions) (*v1alpha1.ClusterSystemList, error)
}

type synchronizer struct {
	osbAPIClient        osbapi.Client
	clusterSystemClient ClusterSystemCRManager
}

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

type Result struct {
	ApplicationName string
	ApplicationID   string
	Operation       Operation
	Error           apperrors.AppError
}

type Synchronizer interface {
	Do() ([]Result, error)
}

func New(osbAPIClient osbapi.Client, clusterSystemManager ClusterSystemCRManager) Synchronizer {
	return &synchronizer{
		osbAPIClient:        osbAPIClient,
		clusterSystemClient: clusterSystemManager,
	}
}

func (s synchronizer) Do() ([]Result, error) {

	results := make([]Result, 0)

	services, err := s.osbAPIClient.GetCatalog()
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch OSB API Services")
	}

	clusterSystemList, err := s.clusterSystemClient.List(v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch ClusterSystems")
	}

	clusterSystems := clusterSystemList.Items
	created := s.createClusterSystems(services, clusterSystems)
	deleted := s.deleteClusterSystems(services, clusterSystems)
	updated := s.updateClusterSystems(services, clusterSystems)

	results = append(results, created...)
	results = append(results, deleted...)
	results = append(results, updated...)

	return nil, nil
}

func (s synchronizer) createClusterSystems(services []osb.Service, clusterSystems []v1alpha1.ClusterSystem) []Result {
	return nil
}

func (s synchronizer) deleteClusterSystems(services []osb.Service, clusterSystems []v1alpha1.ClusterSystem) []Result {
	return nil
}

func (s synchronizer) updateClusterSystems(services []osb.Service, clusterSystems []v1alpha1.ClusterSystem) []Result {
	return nil
}
