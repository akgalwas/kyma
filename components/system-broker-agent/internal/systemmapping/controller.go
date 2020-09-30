package systemmapping

import (
	"context"
	"fmt"
	v1alpha12 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func NewControllerManagedBy(mgr manager.Manager) error {
	return ctrl.
		NewControllerManagedBy(mgr).
		For(&v1alpha12.SystemMapping{}).
		Complete(&reconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
			log:    logrus.WithField("Controller", "SystemMapping"),
		})
}

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
	log    *logrus.Entry
}

func (r *reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.log.Logf(logrus.InfoLevel, "Reconciling SystemMapping %s", req.NamespacedName)

	sm := &v1alpha12.SystemMapping{}
	if err := r.Client.Get(context.Background(), req.NamespacedName, sm); err != nil {
		return ctrl.Result{}, errors.Wrap(err, fmt.Sprintf("failed to get SystemMapping %s", req.NamespacedName))
	}
	// TODO: Do something!
	r.log.Logf(logrus.InfoLevel, "SystemMapping's %s ServiceIDs: %v", req.NamespacedName, sm.Spec.ServiceIDs)

	return ctrl.Result{}, nil
}
