package systemmapping

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization/osbapi"
	v1alpha12 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

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
		if errors.IsNotFound(err) {
			r.log.Infof("SystemMapping is deleted. Deleting the ServiceInstances...")
			if err := r.deleteServiceInstances(systemMapping); err != nil {
				r.log.Errorf("Failed to delete ServiceInstances: %v", err)
				return ctrl.Result{}, err
			}
			r.log.Infof("Reconciliation of SystemMapping %s finished successfully", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		r.log.Errorf("Failed to get SystemMapping: %v", err)
		return ctrl.Result{}, err
	}

	serviceInstancesToCreate, err := r.whichServiceInstancesAreNotProperlyCreated(systemMapping)
	if err != nil {
		r.log.Errorf("Failed to check which ServiceInstances are not properly created: %v", err)
		return ctrl.Result{}, err
	}

	if err := r.createServiceInstances(systemMapping, serviceInstancesToCreate); err != nil {
		r.log.Errorf("Failed to create ServiceInstances: %v", err)
		return ctrl.Result{}, err
	}

	r.log.Infof("Reconciliation of SystemMapping %s finished successfully", req.NamespacedName)
	return ctrl.Result{}, nil
}

func (r *reconciler) deleteServiceInstances(systemMapping *v1alpha12.SystemMapping) error {
	// TODO: Delete ServiceInstances
	//       There is a problem with ServiceInstances deletion.
	//       We cannot get SystemMapping because it's deleted
	//       and therefore we cannot also get the service IDs
	//       which are Service Plans (API Packages). Because of
	//       that there is no way to identify which ServiceInstance
	//       should be deleted
	//       There is a way to do this if ServiceInstance is
	//       somehow labelled with NamespacedName of the
	//       SystemMapping so we can tell which one to delete
	return nil
}

func (r *reconciler) whichServiceInstancesAreNotProperlyCreated(systemMapping *v1alpha12.SystemMapping) ([]string, error) {
	var serviceInstancesToCreate []string
	for _, serviceMeta := range systemMapping.Spec.Services {
		ok, err := r.isServiceInstanceProperlyCreated(serviceMeta)
		if err != nil {
			return nil, err
		}
		if !ok {
			serviceInstancesToCreate = append(serviceInstancesToCreate, serviceMeta.PlanID)
		}
	}
	return serviceInstancesToCreate, nil
}

func (r *reconciler) isServiceInstanceProperlyCreated(serviceMeta v1alpha12.ServiceMeta) (bool, error) {
	// TODO: Check if ServiceInstance is properly created
	return false, nil
}

func (r *reconciler) createServiceInstances(systemMapping *v1alpha12.SystemMapping, serviceInstancesToCreate []string) error {
	for _, serviceInstanceToCreate := range serviceInstancesToCreate {
		if err := r.createServiceInstance(systemMapping, serviceInstanceToCreate); err != nil {
			return err
		}
	}
	return nil
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
