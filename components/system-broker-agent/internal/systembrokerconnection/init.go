package systembrokerconnection

import (
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/compass"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/config"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization"
	"time"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/compass/cache"

	"github.com/pkg/errors"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/certificates"

	"github.com/kyma-project/kyma/components/system-broker-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"
	"k8s.io/client-go/rest"
)

type DependencyConfig struct {
	K8sConfig         *rest.Config
	ControllerManager manager.Manager

	CredentialsManager  certificates.Manager
	ConfigProvider      config.Provider
	ConnectionDataCache cache.ConnectionDataCache

	CertValidityRenewalThreshold float64
	MinimalCompassSyncTime       time.Duration

	ClientsProvider compass.ClientsProvider

	Synchronizer synchronization.Synchronizer
}

func (config DependencyConfig) InitializeController() (Supervisor, error) {
	compassConnectionCRClient, err := v1alpha1.NewForConfig(config.K8sConfig)

	if err != nil {
		return nil, errors.Wrap(err, "Unable to setup Compass Connection CR client")
	}

	csrProvider := certificates.NewCSRProvider()
	compassConnector := NewCompassConnector(csrProvider, config.ClientsProvider)

	connectionSupervisor := NewSupervisor(
		compassConnector,
		compassConnectionCRClient.SystemBrokerConnections(),
		config.CredentialsManager,
		config.ConfigProvider,
		config.CertValidityRenewalThreshold,
		config.MinimalCompassSyncTime,
		config.ConnectionDataCache,
		config.Synchronizer)

	if err := InitCompassConnectionController(config.ControllerManager, connectionSupervisor, config.MinimalCompassSyncTime); err != nil {
		return nil, errors.Wrap(err, "Unable to register controllers to the manager")
	}

	return connectionSupervisor, nil
}
