package systembrokerconnection

import (
	"github.com/kyma-project/kyma/components/system-broker-agent/internal/compass"
	"github.com/kyma-project/kyma/components/system-broker-agent/pkg/apis/compass/v1alpha1"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/certificates"
)

type EstablishedConnection struct {
	Credentials    certificates.Credentials
	ManagementInfo v1alpha1.ManagementInfo
}

const (
	ConnectorTokenHeader = "Connector-Token"
)

//go:generate mockery --name=Connector
type Connector interface {
	EstablishConnection(connectorURL, token string) (EstablishedConnection, error)
	MaintainConnection(renewCert bool) (*certificates.Credentials, v1alpha1.ManagementInfo, error)
}

func NewCompassConnector(
	csrProvider certificates.CSRProvider,
	clientsProvider compass.ClientsProvider,
) Connector {
	return &compassConnector{
		csrProvider:     csrProvider,
		clientsProvider: clientsProvider,
	}
}

type compassConnector struct {
	csrProvider     certificates.CSRProvider
	clientsProvider compass.ClientsProvider
}

func (cc *compassConnector) EstablishConnection(connectorURL, token string) (EstablishedConnection, error) {

	return EstablishedConnection{
		Credentials:    certificates.Credentials{},
		ManagementInfo: v1alpha1.ManagementInfo{},
	}, nil
}

func (cc *compassConnector) MaintainConnection(renewCert bool) (*certificates.Credentials, v1alpha1.ManagementInfo, error) {
	return &certificates.Credentials{}, v1alpha1.ManagementInfo{}, nil
}
