package systemmapping

import (
	"context"
	v1alpha12 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewControllerManagedBy(mgr manager.Manager) error {
	return ctrl.
		NewControllerManagedBy(mgr).
		For(&v1alpha12.SystemMapping{}).
		Complete(&reconciler{
			client: mgr.GetClient(),
			log:    logrus.WithField("Controller", "SystemMapping"),
		})
}

type reconciler struct {
	client client.Client
	log    *logrus.Entry
}

func (r *reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.log.Infof("Reconciling SystemMapping %s...", req.NamespacedName)

	systemMapping := &v1alpha12.SystemMapping{}
	if err := r.client.Get(context.Background(), req.NamespacedName, systemMapping); err != nil {
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
	for _, serviceID := range systemMapping.Spec.ServiceIDs {
		ok, err := r.isServiceInstanceProperlyCreated(serviceID)
		if err != nil {
			return nil, err
		}
		if !ok {
			serviceInstancesToCreate = append(serviceInstancesToCreate, serviceID)
		}
	}
	return serviceInstancesToCreate, nil
}

func (r *reconciler) isServiceInstanceProperlyCreated(serviceID string) (bool, error) {
	// TODO: Check if ServiceInstance is properly created
	return false, nil
}

func (r *reconciler) createServiceInstances(systemMapping *v1alpha12.SystemMapping, serviceInstancesToCreate []string) error {
	for _, serviceInstanceToCreate := range serviceInstancesToCreate {
		if err := r.createServiceInstance(serviceInstanceToCreate); err != nil {
			return err
		}
	}
	return nil
}

func (r *reconciler) createServiceInstance(serviceInstanceToCreate string) error {
	// TODO: Create ServiceInstance
	return nil
}
