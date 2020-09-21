package compass

import (
	"errors"
	"net/http"
	"testing"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/compass/cache"

	"github.com/stretchr/testify/require"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/graphql/mocks"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-project/kyma/components/system-broker-agent/internal/graphql"
)

func newMockGQLConstructor(
	t *testing.T,
	returnedError error,
	expectedEndpoint string,
	expectedLogging bool) graphql.ClientConstructor {
	return func(httpClient *http.Client, graphqlEndpoint string, enableLogging bool) (graphql.Client, error) {
		assert.Equal(t, expectedEndpoint, graphqlEndpoint)
		assert.Equal(t, expectedLogging, enableLogging)

		return &mocks.Client{}, returnedError
	}
}

func TestClientsProvider_GetConnectorTokenSecuredClient(t *testing.T) {

	url := "http://api.io"
	enableLogging := true
	insecureFetch := true

	t.Run("should create new Connector token-secured Client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, nil, url, enableLogging)

		provider := NewClientsProvider(constructor, insecureFetch, enableLogging)

		// when
		configClient, err := provider.GetConnectorTokensClient(url)

		// then
		require.NoError(t, err)
		assert.NotNil(t, configClient)
	})

	t.Run("should return error when failed to create GraphQL client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, errors.New("error"), url, enableLogging)

		provider := NewClientsProvider(constructor, insecureFetch, enableLogging)

		// when
		_, err := provider.GetConnectorTokensClient(url)

		// then
		require.Error(t, err)
	})
}

func TestClientsProvider_GetConnectorCertSecuredClient(t *testing.T) {

	url := "http://api.io"
	enableLogging := true
	insecureFetch := true

	t.Run("should create new Connector cert-secured Client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, nil, url, enableLogging)

		provider := NewClientsProvider(constructor, insecureFetch, enableLogging)
		_ = provider.UpdateConnectionData(cache.ConnectionData{ConnectorURL: url})

		// when
		configClient, err := provider.GetConnectorCertSecuredClient()

		// then
		require.NoError(t, err)
		assert.NotNil(t, configClient)
	})

	t.Run("should return error when failed to create GraphQL client", func(t *testing.T) {
		// given
		constructor := newMockGQLConstructor(t, errors.New("error"), url, enableLogging)

		provider := NewClientsProvider(constructor, insecureFetch, enableLogging)
		_ = provider.UpdateConnectionData(cache.ConnectionData{ConnectorURL: url})

		// when
		_, err := provider.GetConnectorCertSecuredClient()

		// then
		require.Error(t, err)
	})

}
