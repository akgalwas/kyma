package synchronization

import (
	"context"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization/osbapi"
	"github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

//go:generate mockery --name=CRManager
type ClusterSystemCRManager interface {
	Create(ctx context.Context, cc *v1alpha1.ClusterSystem, options metav1.CreateOptions) (*v1alpha1.ClusterSystem, error)
	Update(ctx context.Context, cc *v1alpha1.ClusterSystem, options metav1.UpdateOptions) (*v1alpha1.ClusterSystem, error)
	Delete(ctx context.Context, name string, options metav1.DeleteOptions) error
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha1.ClusterSystem, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1alpha1.ClusterSystemList, error)
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
	ServiceClassName string
	ServiceClassID   string
	Operation        Operation
	Error            apperrors.AppError
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

	clusterSystemList, err := s.clusterSystemClient.List(context.Background(), metav1.ListOptions{})
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

	results := make([]Result, 0)
	for _, service := range services {
		if !clusterSystemExists(service.Name, clusterSystems) {
			results = append(results, s.createClusterSystem(service))
		}
	}

	return results
}

func (s synchronizer) createClusterSystem(service osb.Service) Result {
	clusterSystem := toClusterSystem(service)

	var appErr apperrors.AppError
	_, err := s.clusterSystemClient.Create(context.Background(), &clusterSystem, metav1.CreateOptions{})

	if err != nil {
		appErr = apperrors.Internal("failed to create Cluster System %s: %s", service.Name, err.Error())
	}

	return Result{
		ServiceClassName: service.Name,
		ServiceClassID:   service.ID,
		Operation:        Create,
		Error:            appErr,
	}
}

func (s synchronizer) deleteClusterSystems(services []osb.Service, clusterSystems []v1alpha1.ClusterSystem) []Result {

	results := make([]Result, 0)
	for _, clusterSystem := range clusterSystems {
		if !serviceClassExists(clusterSystem.Name, services) {
			results = append(results, s.deleteClusterSystem(clusterSystem))
		}
	}

	return nil
}

func (s synchronizer) deleteClusterSystem(clusterSystem v1alpha1.ClusterSystem) Result {
	var appErr apperrors.AppError

	err := s.clusterSystemClient.Delete(context.Background(), clusterSystem.Name, metav1.DeleteOptions{})
	if err != nil {
		appErr = apperrors.Internal("failed to delete Cluster System %s: %s", clusterSystem.Name, err.Error())
	}

	return Result{
		ServiceClassName: clusterSystem.Name,
		Operation:        Create,
		Error:            appErr,
	}
}

func (s synchronizer) updateClusterSystems(services []osb.Service, clusterSystems []v1alpha1.ClusterSystem) []Result {
	results := make([]Result, 0)
	for _, service := range services {
		if clusterSystemExists(service.Name, clusterSystems) {
			results = append(results, s.updateClusterSystem(service))
		}
	}

	return results
}

func (s synchronizer) updateClusterSystem(service osb.Service) Result {
	clusterSystem := toClusterSystem(service)

	var appErr apperrors.AppError
	_, err := s.clusterSystemClient.Update(context.Background(), &clusterSystem, metav1.UpdateOptions{})

	if err != nil {
		appErr = apperrors.Internal("failed to update Cluster System %s: %s", service.Name, err.Error())
	}

	return Result{
		ServiceClassName: service.Name,
		ServiceClassID:   service.ID,
		Operation:        Create,
		Error:            appErr,
	}
}

func clusterSystemExists(name string, clusterSystems []v1alpha1.ClusterSystem) bool {

	for _, clusterSystem := range clusterSystems {
		if clusterSystem.Name == name {
			return true
		}
	}

	return false
}

func serviceClassExists(name string, services []osb.Service) bool {
	for _, service := range services {
		if service.Name == name {
			return true
		}
	}

	return false
}
