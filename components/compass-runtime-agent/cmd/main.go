package main

import (
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compass"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/graphql"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/k8sconsts"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/applications"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/model"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"

	"os"
	"time"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/compassconnection"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	apis "github.com/kyma-project/kyma/components/compass-runtime-agent/pkg/apis/compass/v1alpha1"
	log "github.com/sirupsen/logrus"
	restclient "k8s.io/client-go/rest"
)

func main() {
	// TODO - wait for Istio sidecar or do not inject at all?

	log.Infoln("Starting Runtime Agent")
	options := parseArgs()
	log.Infof("Options: %s", options)

	// Get a config to talk to the apiserver
	log.Info("Setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "unable to set up client config")
		os.Exit(1)
	}

	syncPeriod := time.Second * time.Duration(options.controllerSyncPeriod)

	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: &syncPeriod})
	if err != nil {
		log.Error(err, "unable to set up overall controller manager")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	log.Info("Setting up scheme")
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "Unable add APIs to scheme")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	compassConnectionCRClient, err := v1alpha1.NewForConfig(cfg)
	if err != nil {
		log.Error("Unable to setup Compass Connection CR client")
		os.Exit(1)
	}

	certManager := certificates.NewCredentialsManager()
	compassConfigClient := compass.NewConfigurationClient(options.tenant, options.runtimeId, graphql.New)
	syncService := createNewSynchronizationService()

	compassConnector := compass.NewCompassConnector(options.tokenURLConfigFile)
	connectionSupervisor := compassconnection.NewSupervisor(
		compassConnector,
		compassConnectionCRClient.CompassConnections(),
		certManager,
		compassConfigClient,
		syncService)

	minimalConfigSyncTime := time.Duration(options.minimalConfigSyncTime) * time.Second

	// Setup all Controllers
	log.Info("Setting up controller")
	if err := compassconnection.InitCompassConnectionController(mgr, connectionSupervisor, minimalConfigSyncTime); err != nil {
		log.Error(err, "Unable to register controllers to the manager")
		os.Exit(1)
	}

	// Initialize Compass Connection CR
	log.Infoln("Initializing Compass Connection CR")
	_, err = connectionSupervisor.InitializeCompassConnection()
	if err != nil {
		log.Error("Unable to initialize Compass Connection CR")
	}

	// Start the Cmd
	log.Info("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Unable to run the manager")
		os.Exit(1)
	}
}

func createNewSynchronizationService() kyma.Service {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Errorf("Failed to read k8s in-cluster configuration, %s", err)

		return uninitializedKymaService{}
	}

	applicationManager, err := newApplicationManager(k8sConfig)
	if err != nil {
		log.Errorf("Failed to initialize Applications manager, %s", err)
		return uninitializedKymaService{}
	}

	resourcesService := apiresources.NewService()
	// TODO: pass the namespace name in parameters
	nameResolver := k8sconsts.NewNameResolver("kyma-integration")
	converter := applications.NewConverter(nameResolver)

	return kyma.NewService(applicationManager, converter, resourcesService)
}

func newApplicationManager(config *restclient.Config) (applications.Repository, apperrors.AppError) {
	applicationEnvironmentClientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, apperrors.Internal("Failed to create k8s application client, %s", err)
	}

	appInterface := applicationEnvironmentClientset.ApplicationconnectorV1alpha1().Applications()

	return applications.NewRepository(appInterface), nil
}

type uninitializedKymaService struct {
}

func (u uninitializedKymaService) Apply(applications []model.Application) ([]kyma.Result, apperrors.AppError) {
	return nil, apperrors.Internal("Service not initialized")
}
