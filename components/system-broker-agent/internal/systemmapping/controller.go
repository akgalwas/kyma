package systemmapping

import (
	"context"
	v1alpha12 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/sirupsen/logrus"
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
	r.log.Logf(logrus.InfoLevel, "Reconciling SystemMapping %s", req.NamespacedName)

	sm := &v1alpha12.SystemMapping{}
	if err := r.client.Get(context.Background(), req.NamespacedName, sm); err != nil {
		r.log.Logf(logrus.WarnLevel, "Failed to get SystemMapping %s", req.NamespacedName)
		return ctrl.Result{}, nil
	}
	// TODO: Do something!
	r.log.Logf(logrus.InfoLevel, "SystemMapping's %s ServiceIDs: %v", req.NamespacedName, sm.Spec.ServiceIDs)

	return ctrl.Result{}, nil
}
