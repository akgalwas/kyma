package main

import (
	"fmt"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/compass"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/compass/cache"
	confProvider "github.com/kyma-project/kyma/components/system-broker-agent/internal/config"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/graphql"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/secrets"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization/osbapi"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/systembrokerconnection"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/systemmapping"
	applicationconnectorv1alpha1 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/applicationconnector/v1alpha1"
	compassv1alpha1 "github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/compass/v1alpha1"
	"github.com/kyma-project/kyma/components/system-broker-agent/pkg/client/applicationconnector/clientset/versioned/typed/applicationconnector/v1alpha1"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/vrischmann/envconfig"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	fmt.Println("Starting System Broker Agent")

	var options Config
	err := envconfig.InitWithPrefix(&options, "APP")
	exitOnError(err, "Failed to process environment variables")

	log.Infof("Env config: %s", options.String())

	// Get a config to talk to the apiserver
	log.Info("Setting up client for manager")
	cfg, err := getk8sConfig()
	exitOnError(err, "Failed to set up client config")

	synchronizer, err := createSynchronizer(cfg)
	exitOnError(err, "Failed to create synchronizer")

	log.Info("Setting up manager")
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: &options.ControllerSyncPeriod})
	exitOnError(err, "Failed to set up overall controller manager")

	// Setup Scheme for all resources
	log.Info("Setting up scheme")
	err = compassv1alpha1.AddToScheme(mgr.GetScheme())
	exitOnError(err, "Failed to add Compass APIs to scheme")
	err = applicationconnectorv1alpha1.AddToScheme(mgr.GetScheme())
	exitOnError(err, "Failed to add ApplicationConnector APIs to scheme")

	log.Info("Registering Components.")

	//k8sResourceClientSets, err := k8sResourceClients(cfg)
	//exitOnError(err, "Failed to initialize K8s resource clients")

	//secretsManagerConstructor := func(namespace string) secrets.Manager {
	//	return k8sResourceClientSets.core.CoreV1().Secrets(namespace)
	//}

	secretsManagerConstructor := func(namespace string) secrets.Manager {
		return nil
	}

	secretsRepository := secrets.NewRepository(secretsManagerConstructor)

	clusterCertSecret := parseNamespacedName(options.ClusterCertificatesSecret)
	caCertSecret := parseNamespacedName(options.CaCertificatesSecret)

	certManager := certificates.NewCredentialsManager(clusterCertSecret, caCertSecret, secretsRepository)

	agentConfigSecretNamespacedName := parseNamespacedName(options.AgentConfigurationSecret)

	connectionDataCache := cache.NewConnectionDataCache()

	configProvider := confProvider.NewConfigProvider(agentConfigSecretNamespacedName, secretsRepository)
	clientsProvider := compass.NewClientsProvider(graphql.New, options.SkipCompassTLSVerify, options.QueryLogging)
	connectionDataCache.AddSubscriber(clientsProvider.UpdateConnectionData)

	//log.Infoln("Setting up Director Proxy Service")
	//directorProxy := director.NewProxy(options.DirectorProxy)
	//err = mgr.Add(directorProxy)
	//exitOnError(err, "Failed to create director proxy")
	//connectionDataCache.AddSubscriber(directorProxy.SetURLAndCerts)

	log.Infoln("Setting up Controller")
	controllerDependencies := systembrokerconnection.DependencyConfig{
		K8sConfig:                    cfg,
		ControllerManager:            mgr,
		ClientsProvider:              clientsProvider,
		CredentialsManager:           certManager,
		ConfigProvider:               configProvider,
		ConnectionDataCache:          connectionDataCache,
		CertValidityRenewalThreshold: options.CertValidityRenewalThreshold,
		MinimalCompassSyncTime:       options.MinimalCompassSyncTime,
		Synchronizer:                 synchronizer,
	}

	compassConnectionSupervisor, err := controllerDependencies.InitializeController()
	exitOnError(err, "Failed to initialize CompassConnection Controller")

	log.Infoln("Initializing Compass Connection CR")
	_, err = compassConnectionSupervisor.InitializeCompassConnection()
	exitOnError(err, "Failed to initialize Compass Connection CR")

	err = systemmapping.NewControllerManagedBy(mgr)
	exitOnError(err, "Failed to initialize SystemMapping Controller")

	log.Info("Starting the Cmd.")
	err = mgr.Start(signals.SetupSignalHandler())
	exitOnError(err, "Failed to run the manager")

}

func createSynchronizer(restConfig *restclient.Config) (synchronization.Synchronizer, error) {
	osbAPIClient, err := osbapi.NewClient("https://compass-gateway.cmp-test.dev.kyma.cloud.sap/broker")
	if err != nil {
		return nil, errors.Wrap(err, "failed to init OSB API client")
	}

	crClient, err := v1alpha1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to init Cluster System client")
	}

	synchronizer := synchronization.New(osbAPIClient, crClient.ClusterSystems())

	return synchronizer, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatal(errors.Wrap(err, context))
	}
}

func getk8sConfig() (*restclient.Config, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.Warnf("Failed to read in cluster config: %s", err.Error())
		log.Info("Trying to initialize with local config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, errors.Errorf("failed to read k8s in-cluster configuration, %s", err.Error())
		}
	}

	return k8sConfig, nil
}
