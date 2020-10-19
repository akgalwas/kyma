package systemmapping

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization/osbapi"
	v1alpha12 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

const FinalizerName = "myfinalizer"

//go:generate mockery --name=CRManager
type ClusterSystemCRManager interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*v1alpha12.ClusterSystem, error)
}

func NewControllerManagedBy(mgr manager.Manager, csClient ClusterSystemCRManager, osbApiClient osbapi.Client) error {
	return ctrl.
		NewControllerManagedBy(mgr).
		For(&v1alpha12.SystemMapping{}).
		Complete(&reconciler{
			ctrlClient:   mgr.GetClient(),
			csClient:     csClient,
			osbApiClient: osbApiClient,
			log:          logrus.WithField("Controller", "SystemMapping"),
		})
}

type reconciler struct {
	ctrlClient   ctrlClient.Client
	csClient     ClusterSystemCRManager
	osbApiClient osbapi.Client
	log          *logrus.Entry
}

func (r *reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.log.Infof("Reconciling SystemMapping %s...", req.NamespacedName)

	systemMapping := &v1alpha12.SystemMapping{}
	if err := r.ctrlClient.Get(context.Background(), req.NamespacedName, systemMapping); err != nil {
		r.log.Errorf("Failed to get SystemMapping: %v", err)
		return ctrl.Result{}, err
	}

	if systemMapping.ObjectMeta.DeletionTimestamp.IsZero() {
		systemMapping.ObjectMeta.Finalizers = append(systemMapping.ObjectMeta.Finalizers, FinalizerName)
		if err := r.ctrlClient.Update(context.Background(), systemMapping); err != nil {
			r.log.Errorf("Failed to update SystemMapping: %v", err)
			return ctrl.Result{}, err
		}

		if err := r.createServiceInstances(systemMapping); err != nil {
			r.log.Errorf("Failed to create service instances: %v", err)
			return ctrl.Result{}, err
		}
	} else {
		err := r.deleteServiceInstances(systemMapping)
		if err != nil {
			r.log.Errorf("Failed to delete service instances: %v", err)
			return ctrl.Result{}, err
		}

		removeFinalizer(systemMapping, FinalizerName)
	}

	r.log.Infof("Reconciliation of SystemMapping %s finished successfully", req.NamespacedName)
	return ctrl.Result{}, nil
}

func removeFinalizer(systemMapping *v1alpha12.SystemMapping, finalizer string) {
	finalizers := make([]string, 0)
	for _, f := range systemMapping.ObjectMeta.Finalizers {
		if f == finalizer {
			continue
		}
		finalizers = append(finalizers, f)
	}

	systemMapping.Finalizers = finalizers
}

func (r *reconciler) deleteServiceInstances(systemMapping *v1alpha12.SystemMapping) error {

	return nil
}

func (r *reconciler) createServiceInstances(systemMapping *v1alpha12.SystemMapping) error {
	serviceInstancesToCreate, err := r.getServiceInstancesToCreate(systemMapping)
	if err != nil {
		return err
	}

	for _, serviceInstanceToCreate := range serviceInstancesToCreate {
		if err := r.createServiceInstance(systemMapping, serviceInstanceToCreate); err != nil {
			return err
		}
	}
	return nil
}

func (r *reconciler) isServiceInstanceCreated(serviceMeta v1alpha12.ServiceMeta) (bool, error) {
	return r.osbApiClient.InstanceExists(*serviceMeta.InstanceId)
}

func (r *reconciler) getServiceInstancesToCreate(systemMapping *v1alpha12.SystemMapping) ([]string, error) {
	var serviceInstancesToCreate []string
	for _, serviceMeta := range systemMapping.Spec.Services {
		ok, err := r.isServiceInstanceCreated(serviceMeta)
		if err != nil {
			return nil, err
		}
		if !ok {
			serviceInstancesToCreate = append(serviceInstancesToCreate, serviceMeta.PlanID)
		}
	}
	return serviceInstancesToCreate, nil
}

func (r *reconciler) createServiceInstance(systemMapping *v1alpha12.SystemMapping, servicePlanID string) error {
	serviceServiceID, err := r.getClusterSystemApplicationID(systemMapping.Name)
	if err != nil {
		return err
	}

	serviceInstanceID := uuid.New().String()

	provisionRequest := &osb.ProvisionRequest{
		InstanceID:       serviceInstanceID,
		ServiceID:        serviceServiceID,
		PlanID:           servicePlanID,
		OrganizationGUID: "organization_guid",
		SpaceGUID:        "space_guid",
	}
	if err := r.osbApiClient.ProvisionInstance(provisionRequest); err != nil {
		return err
	}

	if err := r.updateSystemMappingWithServiceInstance(systemMapping, servicePlanID, serviceInstanceID); err != nil {
		return err
	}

	return nil
}

func (r *reconciler) getClusterSystemApplicationID(clusterSystemName string) (string, error) {
	clusterSystem, err := r.csClient.Get(context.Background(), clusterSystemName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if clusterSystem.Spec.CompassMetadata == nil {
		return "", fmt.Errorf("cluster system %s has no compass metadata", clusterSystem.Name)
	}
	return clusterSystem.Spec.CompassMetadata.ApplicationID, nil
}

func (r *reconciler) updateSystemMappingWithServiceInstance(systemMapping *v1alpha12.SystemMapping, servicePlanID, serviceInstanceID string) error {
	updatedServiceMetasCounter := 0
	for _, serviceMeta := range systemMapping.Spec.Services {
		if serviceMeta.PlanID == servicePlanID {
			serviceMeta.InstanceId = &serviceInstanceID
			updatedServiceMetasCounter++
		}
	}
	if updatedServiceMetasCounter != 1 {
		return fmt.Errorf("updating system mapping malformed")
	}
	return r.ctrlClient.Update(context.Background(), systemMapping)
}
