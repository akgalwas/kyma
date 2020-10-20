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
	ProvisionInstance(serviceID, planID, instanceID string) error
	InstanceExists(serviceID *string) (bool, error)
	DeprovisionInstance(serviceID, planID, instanceID string) error
	Unbind(serviceID, planID, instanceID, bindingID string) error
	Bind(serviceID, planID, instanceID, bindingID string) (map[string]interface{}, error)
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

func (c client) ProvisionInstance(serviceID, planID, instanceID string) error {
	provisionRequest := &osb.ProvisionRequest{
		InstanceID:        instanceID,
		ServiceID:         serviceID,
		PlanID:            planID,
		OrganizationGUID:  "organization_guid",
		SpaceGUID:         "space_guid",
		AcceptsIncomplete: true,
	}

	_, err := c.osbAPIClient.ProvisionInstance(provisionRequest)
	if err != nil {
		return errors.Wrap(err, "failed to provision service instance in System Broker")
	}
	return nil
}

func (c client) Bind(serviceID, planID, instanceID, bindingID string) (map[string]interface{}, error) {
	bindRequest := osb.BindRequest{
		ServiceID:  serviceID,
		PlanID:     planID,
		InstanceID: instanceID,
		BindingID:  bindingID,
	}

	res, err := c.osbAPIClient.Bind(&bindRequest)
	if err != nil {
		return nil, err
	}

	return res.Credentials, nil
}

func (c client) Unbind(serviceID, planID, instanceID, bindingID string) error {
	unbindRequest := osb.UnbindRequest{
		ServiceID:  serviceID,
		PlanID:     planID,
		InstanceID: instanceID,
		BindingID:  bindingID,
	}

	_, err := c.osbAPIClient.Unbind(&unbindRequest)

	if err != nil {
		isHttpError, httpErr := asHTTPError(err)
		if isHttpError {
			if httpErr.StatusCode == http.StatusGone {
				return nil
			}
		}

		return err
	}

	return nil
}

func (c client) DeprovisionInstance(serviceID, planID, instanceID string) error {
	deprovisioningRequest := osb.DeprovisionRequest{
		ServiceID:         serviceID,
		PlanID:            planID,
		InstanceID:        instanceID,
		AcceptsIncomplete: true,
	}
	_, err := c.osbAPIClient.DeprovisionInstance(&deprovisioningRequest)

	if err != nil {
		isHttpError, httpErr := asHTTPError(err)
		if isHttpError {
			if httpErr.StatusCode == http.StatusGone {
				return nil
			}
		}

		return err
	}

	return nil
}

func (c client) InstanceExists(instanceID *string) (bool, error) {
	if instanceID == nil {
		return false, nil
	}

	request := osb.GetInstanceRequest{
		InstanceID: *instanceID,
	}
	_, err := c.osbAPIClient.GetInstance(&request)

	if err != nil {
		isHttpError, httpErr := asHTTPError(err)
		if isHttpError {
			if httpErr.StatusCode == http.StatusNotFound {
				return false, nil
			}
		}

		return false, err
	}

	return true, nil
}

func asHTTPError(err error) (bool, *osb.HTTPStatusCodeError) {
	if err != nil {
		httpErr, ok := osb.IsHTTPError(err)
		if ok {
			return true, httpErr
		}
		return false, nil
	}

	return false, nil
}
