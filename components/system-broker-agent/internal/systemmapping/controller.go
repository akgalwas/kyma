package systemmapping

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization/osbapi"
	v1alpha12 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const FinalizerName = "service-instance-finalizer"

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
		if k8serrors.IsNotFound(err) {
			r.log.Infof("SystemMapping %s deleted.", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		r.log.Errorf("Failed to get SystemMapping: %v", err)
		return ctrl.Result{}, err
	}

	if systemMapping.ObjectMeta.DeletionTimestamp.IsZero() {
		r.log.Infof("Processing SystemMapping %s creation", req.NamespacedName)
		if !hasFinalizer(systemMapping, FinalizerName) {
			r.log.Infof("Adding SystemMapping %s finalizer", req.NamespacedName)
			systemMapping.ObjectMeta.Finalizers = append(systemMapping.ObjectMeta.Finalizers, FinalizerName)
			if err := r.ctrlClient.Update(context.Background(), systemMapping); err != nil {
				r.log.Errorf("Failed to add finalizer: %v", err)
				return ctrl.Result{}, err
			}

			r.log.Infof("Creating ServiceInstances for SystemMapping %s", req.NamespacedName)
			if err := r.createAndBindServiceInstances(systemMapping); err != nil {
				r.log.Errorf("Failed to create service instances: %v", err)
				return ctrl.Result{}, err
			}
		}
	} else {
		err := r.unbindAndDeleteServiceInstances(systemMapping)
		if err != nil {
			r.log.Errorf("Failed to delete service instances: %v", err)
			return ctrl.Result{}, err
		}

		err = r.removeFinalizer(systemMapping, FinalizerName)
		if err != nil {
			r.log.Errorf("Failed to remove finalizer: %v", err)
			return ctrl.Result{}, err
		}
	}

	r.log.Infof("Reconciliation of SystemMapping %s finished successfully", req.NamespacedName)
	return ctrl.Result{}, nil
}

func hasFinalizer(systemMapping *v1alpha12.SystemMapping, finalizer string) bool {
	for _, f := range systemMapping.ObjectMeta.Finalizers {
		if f == finalizer {
			return true
		}
	}

	return false
}

func (r *reconciler) removeFinalizer(systemMapping *v1alpha12.SystemMapping, finalizer string) error {
	finalizers := make([]string, 0)
	for _, f := range systemMapping.ObjectMeta.Finalizers {
		if f == finalizer {
			continue
		}
		finalizers = append(finalizers, f)
	}

	systemMapping.Finalizers = finalizers

	return r.ctrlClient.Update(context.Background(), systemMapping)
}

func (r *reconciler) unbindAndDeleteServiceInstances(systemMapping *v1alpha12.SystemMapping) error {
	serviceID, err := r.getClusterSystemApplicationID(systemMapping.Name)
	if err != nil {
		return err
	}

	for _, service := range systemMapping.Spec.Services {
		if service.BindingID != nil {
			r.log.Infof("Unbinding Service Instance ", *service.InstanceID)
			r.osbApiClient.Unbind(serviceID, service.PlanID, *service.InstanceID, *service.BindingID)
		}

		if service.InstanceID != nil {
			r.log.Infof("Unbinding Service Instance ", *service.InstanceID)
			r.osbApiClient.DeprovisionInstance(serviceID, service.PlanID, *service.InstanceID)
		}
	}
	return nil
}

func (r *reconciler) createAndBindServiceInstances(systemMapping *v1alpha12.SystemMapping) error {
	serviceInstances, err := r.getServiceInstancesToCreate(systemMapping)
	if err != nil {
		return err
	}

	for _, serviceInstanceToCreate := range serviceInstances {
		if err := r.createAndBindServiceInstance(systemMapping, serviceInstanceToCreate); err != nil {
			return err
		}
	}
	return nil
}

func (r *reconciler) getServiceInstancesToCreate(systemMapping *v1alpha12.SystemMapping) ([]string, error) {
	var serviceInstancesToCreate []string
	for _, serviceMeta := range systemMapping.Spec.Services {
		exists, err := r.osbApiClient.InstanceExists(serviceMeta.InstanceID)
		if err != nil {
			return nil, err
		}
		if !exists {
			serviceInstancesToCreate = append(serviceInstancesToCreate, serviceMeta.PlanID)
		}
	}
	return serviceInstancesToCreate, nil
}

func (r *reconciler) createAndBindServiceInstance(systemMapping *v1alpha12.SystemMapping, servicePlanID string) error {
	serviceServiceID, err := r.getClusterSystemApplicationID(systemMapping.Name)
	if err != nil {
		return err
	}

	serviceInstanceID := uuid.New().String()

	if err := r.osbApiClient.ProvisionInstance(serviceServiceID, servicePlanID, serviceInstanceID); err != nil {
		return err
	}

	bindingID := uuid.New().String()
	// TODO: use credentials to create a secret
	_, err = r.osbApiClient.Bind(serviceServiceID, servicePlanID, serviceInstanceID, bindingID)

	if err != nil {
		return err
	}

	if err := r.updateSystemMapping(systemMapping, servicePlanID, serviceInstanceID, bindingID); err != nil {
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

func (r *reconciler) updateSystemMapping(systemMapping *v1alpha12.SystemMapping, servicePlanID, serviceInstanceID string, bindingID string) error {
	updatedServiceMetasCounter := 0
	for i, service := range systemMapping.Spec.Services {
		if service.PlanID == servicePlanID {
			newService := service
			newService.InstanceID = &serviceInstanceID
			newService.BindingID = &bindingID

			systemMapping.Spec.Services[i] = newService

			updatedServiceMetasCounter++
		}
	}
	if updatedServiceMetasCounter != 1 {
		return fmt.Errorf("updating system mapping malformed")
	}
	return r.ctrlClient.Update(context.Background(), systemMapping)
}
