package osbapi

import (
	"github.com/pkg/errors"
	"net/http"
	osb "sigs.k8s.io/go-open-service-broker-client/v2"
)

type client struct {
	osbAPIClient osb.Client
}

type Client interface {
	GetCatalog() ([]osb.Service, error)
	ProvisionInstance(*osb.ProvisionRequest) error
}

func NewClient(url string) (Client, error) {
	config := osb.DefaultClientConfiguration()
	config.URL = url
	config.Insecure = true

	// TODO: this is a workaround done on my fork. Default OSB API client doesn't have such facilities.
	config.DoRequestFunc = func(client *http.Client, req *http.Request) (*http.Response, error) {
		req.Header.Set("Tenant", "3e64ebae-38b5-46a0-b1ed-9ccee153a0ae")

		return client.Do(req)
	}

	osbAPIClient, err := osb.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Service Broker API client")
	}

	return &client{
		osbAPIClient,
	}, nil
}

func (c client) GetCatalog() ([]osb.Service, error) {
	response, err := c.osbAPIClient.GetCatalog()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service catalog form System Broker")
	}

	return response.Services, nil
}

func (c client) ProvisionInstance(provisionRequest *osb.ProvisionRequest) error {
	_, err := c.osbAPIClient.ProvisionInstance(provisionRequest)
	if err != nil {
		return errors.Wrap(err, "failed to provision service instance in System Broker")
	}
	return nil
}
