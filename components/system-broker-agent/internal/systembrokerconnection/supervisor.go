package systembrokerconnection

import (
	"context"
	"fmt"
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/synchronization"
	"time"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/compass/cache"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/config"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/certificates"
	"github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/compass/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	DefaultCompassConnectionName = "compass-connection"
)

//go:generate mockery --name=CRManager
type CRManager interface {
	Create(ctx context.Context, cc *v1alpha1.SystemBrokerConnection, options v1.CreateOptions) (*v1alpha1.SystemBrokerConnection, error)
	Update(ctx context.Context, cc *v1alpha1.SystemBrokerConnection, options v1.UpdateOptions) (*v1alpha1.SystemBrokerConnection, error)
	Delete(ctx context.Context, name string, options v1.DeleteOptions) error
	Get(ctx context.Context, name string, options v1.GetOptions) (*v1alpha1.SystemBrokerConnection, error)
}

//go:generate mockery --name=Supervisor
type Supervisor interface {
	InitializeCompassConnection() (*v1alpha1.SystemBrokerConnection, error)
	SynchronizeWithSystemBroker(connection *v1alpha1.SystemBrokerConnection) (*v1alpha1.SystemBrokerConnection, error)
}

func NewSupervisor(
	connector Connector,
	crManager CRManager,
	credManager certificates.Manager,
	configProvider config.Provider,
	certValidityRenewalThreshold float64,
	minimalCompassSyncTime time.Duration,
	connectionDataCache cache.ConnectionDataCache,
	synchronizer synchronization.Synchronizer,
) Supervisor {
	return &crSupervisor{
		compassConnector:             connector,
		crManager:                    crManager,
		credentialsManager:           credManager,
		configProvider:               configProvider,
		certValidityRenewalThreshold: certValidityRenewalThreshold,
		minimalCompassSyncTime:       minimalCompassSyncTime,
		connectionDataCache:          connectionDataCache,
		synchronizer:                 synchronizer,
		log:                          logrus.WithField("Supervisor", "CompassConnection"),
	}
}

type crSupervisor struct {
	compassConnector             Connector
	crManager                    CRManager
	credentialsManager           certificates.Manager
	configProvider               config.Provider
	certValidityRenewalThreshold float64
	minimalCompassSyncTime       time.Duration
	log                          *logrus.Entry
	connectionDataCache          cache.ConnectionDataCache
	synchronizer                 synchronization.Synchronizer
}

func (s *crSupervisor) InitializeCompassConnection() (*v1alpha1.SystemBrokerConnection, error) {
	compassConnectionCR, err := s.crManager.Get(context.Background(), DefaultCompassConnectionName, v1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return s.newCompassConnection()
		}

		return nil, errors.Wrap(err, "Connection failed while getting existing Compass Connection")
	}

	s.log.Infof("Compass Connection exists with state %s", compassConnectionCR.Status.State)

	if !compassConnectionCR.ShouldAttemptReconnect() {
		s.log.Infof("Connection already initialized, skipping ")

		credentials, err := s.credentialsManager.GetClientCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials while initializing Compass Connection CR: %s", err.Error())
		}

		s.connectionDataCache.UpdateConnectionData(
			credentials.AsTLSCertificate(),
			compassConnectionCR.Spec.ManagementInfo.DirectorURL,
			compassConnectionCR.Spec.ManagementInfo.ConnectorURL,
		)

		return compassConnectionCR, nil
	}

	s.establishConnection(compassConnectionCR)

	return s.updateSystemBrokerConnection(compassConnectionCR)
}

func (s *crSupervisor) SynchronizeWithSystemBroker(connection *v1alpha1.SystemBrokerConnection) (*v1alpha1.SystemBrokerConnection, error) {
	syncAttemptTime := metav1.Now()

	results, err := s.synchronizer.Do()
	if err != nil {
		connection.Status.State = v1alpha1.ResourceApplicationFailed
		connection.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
			LastAttempt:         syncAttemptTime,
			LastSuccessfulFetch: syncAttemptTime,
			Error:               fmt.Sprintf("Failed to apply configuration: %s", err.Error()),
		}
		return s.updateSystemBrokerConnection(connection)
	}

	// TODO: save result to CR and possibly log in better manner
	s.log.Infof("Config application results: ")
	for _, res := range results {
		s.log.Info(res)
	}

	s.log.Infof("Setting System Broker Connection to Synchronized state")
	connection.Status.State = v1alpha1.Synchronized
	connection.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
		LastAttempt:               syncAttemptTime,
		LastSuccessfulFetch:       syncAttemptTime,
		LastSuccessfulApplication: syncAttemptTime,
	}

	return s.updateSystemBrokerConnection(connection)
}

func (s *crSupervisor) maintainCompassConnection(compassConnection *v1alpha1.SystemBrokerConnection) error {
	shouldRenew := compassConnection.ShouldRenewCertificate(s.certValidityRenewalThreshold, s.minimalCompassSyncTime)

	s.log.Infof("Trying to maintain certificates connection... Renewal: %v", shouldRenew)
	newCreds, managementInfo, err := s.compassConnector.MaintainConnection(shouldRenew)
	if err != nil {
		return errors.Wrap(err, "Failed to connect to Compass Connector")
	}

	connectionTime := metav1.Now()

	if shouldRenew && newCreds != nil {
		s.log.Infof("Trying to save renewed certificates...")
		err = s.credentialsManager.PreserveCredentials(*newCreds)
		if err != nil {
			return errors.Wrap(err, "Failed to preserve certificate")
		}

		s.log.Infof("Successfully saved renewed certificates")
		compassConnection.SetCertificateStatus(connectionTime, newCreds.ClientCertificate)
		compassConnection.Spec.RefreshCredentialsNow = false
		compassConnection.Status.ConnectionStatus.Renewed = connectionTime

		s.connectionDataCache.UpdateConnectionData((*newCreds).AsTLSCertificate(), managementInfo.DirectorURL, managementInfo.ConnectorURL)
		s.log.Infof("Refreshed connection data cache")
	}

	if s.urlsUpdated(compassConnection, managementInfo) {
		s.log.Infof("Compass URLs modified. Updating cache. Connector: %s => %s, Director: %s => %s",
			compassConnection.Spec.ManagementInfo.ConnectorURL, managementInfo.ConnectorURL,
			compassConnection.Spec.ManagementInfo.DirectorURL, managementInfo.DirectorURL)
		s.connectionDataCache.UpdateURLs(managementInfo.DirectorURL, managementInfo.ConnectorURL)
	}

	s.log.Infof("Connection maintained. Director URL: %s , ConnectorURL: %s", managementInfo.DirectorURL, managementInfo.ConnectorURL)

	if compassConnection.Status.ConnectionStatus == nil {
		compassConnection.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{}
	}

	compassConnection.Status.ConnectionStatus.LastSync = connectionTime
	compassConnection.Status.ConnectionStatus.LastSuccess = connectionTime

	return nil
}

func (s *crSupervisor) urlsUpdated(compassConnectionCR *v1alpha1.SystemBrokerConnection, managementInfo v1alpha1.ManagementInfo) bool {
	return compassConnectionCR.Spec.ManagementInfo.ConnectorURL != managementInfo.ConnectorURL ||
		compassConnectionCR.Spec.ManagementInfo.DirectorURL != managementInfo.DirectorURL
}

func (s *crSupervisor) newCompassConnection() (*v1alpha1.SystemBrokerConnection, error) {
	connectionCR := &v1alpha1.SystemBrokerConnection{
		ObjectMeta: v1.ObjectMeta{
			Name: DefaultCompassConnectionName,
		},
		Spec: v1alpha1.SystemBrokerConnectionSpec{},
	}

	s.establishConnection(connectionCR)

	return s.crManager.Create(context.Background(), connectionCR, v1.CreateOptions{})
}

func (s *crSupervisor) establishConnection(connectionCR *v1alpha1.SystemBrokerConnection) {
	// TODO: init secrets interface
	connCfg := config.ConnectionConfig{}
	//connCfg, err := s.configProvider.GetConnectionConfig()
	//if err != nil {
	//	s.setConnectionFailedStatus(connectionCR, err, fmt.Sprintf("Failed to retrieve certificate: %s", err.Error()))
	//	return
	//}

	connection, err := s.compassConnector.EstablishConnection(connCfg.ConnectorURL, connCfg.Token)
	if err != nil {
		s.setConnectionFailedStatus(connectionCR, err, fmt.Sprintf("Failed to retrieve certificate: %s", err.Error()))
		return
	}

	connectionTime := metav1.Now()

	// TODO: store credentials
	//err = s.credentialsManager.PreserveCredentials(connection.Credentials)
	//if err != nil {
	//	s.setConnectionFailedStatus(connectionCR, err, fmt.Sprintf("Failed to preserve certificate: %s", err.Error()))
	//	return
	//}

	s.log.Infof("Connection established. Director URL: %s , ConnectorURL: %s", connection.ManagementInfo.DirectorURL, connection.ManagementInfo.ConnectorURL)

	connectionCR.Status.State = v1alpha1.Connected
	connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{
		Established: connectionTime,
		LastSync:    connectionTime,
		LastSuccess: connectionTime,
	}
	connectionCR.SetCertificateStatus(connectionTime, connection.Credentials.ClientCertificate)

	connectionCR.Spec.ManagementInfo = connection.ManagementInfo

	s.connectionDataCache.UpdateConnectionData(
		connection.Credentials.AsTLSCertificate(),
		connection.ManagementInfo.DirectorURL,
		connection.ManagementInfo.ConnectorURL,
	)
}

func (s *crSupervisor) setConnectionFailedStatus(connectionCR *v1alpha1.SystemBrokerConnection, err error, connStatusError string) {
	s.log.Errorf("Error while establishing connection with Compass: %s", err.Error())
	s.log.Infof("Setting Compass Connection to ConnectionFailed state")
	connectionCR.Status.State = v1alpha1.ConnectionFailed
	if connectionCR.Status.ConnectionStatus == nil {
		connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{}
	}
	connectionCR.Status.ConnectionStatus.LastSync = metav1.Now()
	connectionCR.Status.ConnectionStatus.Error = connStatusError
}

func (s *crSupervisor) setConnectionSynchronizedStatus(connectionCR *v1alpha1.SystemBrokerConnection, attemptTime metav1.Time) {
	s.log.Infof("Setting Compass Connection to Synchronized state")
	connectionCR.Status.State = v1alpha1.Synchronized
	connectionCR.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{
		LastAttempt:               attemptTime,
		LastSuccessfulFetch:       attemptTime,
		LastSuccessfulApplication: attemptTime,
	}
}

func (s *crSupervisor) setConnectionMaintenanceFailedStatus(connectionCR *v1alpha1.SystemBrokerConnection, attemptTime metav1.Time, errorMsg string) {
	s.log.Error(errorMsg)
	s.log.Infof("Setting Compass Connection to ConnectionMaintenanceFailed state")
	connectionCR.Status.State = v1alpha1.ConnectionMaintenanceFailed
	if connectionCR.Status.ConnectionStatus == nil {
		connectionCR.Status.ConnectionStatus = &v1alpha1.ConnectionStatus{}
	}
	connectionCR.Status.ConnectionStatus.LastSync = attemptTime
	connectionCR.Status.ConnectionStatus.Error = errorMsg
}

func (s *crSupervisor) updateSystemBrokerConnection(connectionCR *v1alpha1.SystemBrokerConnection) (*v1alpha1.SystemBrokerConnection, error) {
	// TODO: with retries

	return s.crManager.Update(context.Background(), connectionCR, v1.UpdateOptions{})
}

func (s *crSupervisor) setSyncFailedStatus(connectionCR *v1alpha1.SystemBrokerConnection, attemptTime metav1.Time, errorMsg string) {
	s.log.Error(errorMsg)
	s.log.Infof("Setting Compass Connection to SynchronizationFailed state")
	connectionCR.Status.State = v1alpha1.SynchronizationFailed
	if connectionCR.Status.SynchronizationStatus == nil {
		connectionCR.Status.SynchronizationStatus = &v1alpha1.SynchronizationStatus{}
	}
	connectionCR.Status.SynchronizationStatus.LastAttempt = attemptTime
	connectionCR.Status.SynchronizationStatus.Error = errorMsg
}
