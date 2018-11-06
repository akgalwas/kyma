package authorization

import (
	"testing"
	"net/http"
	"github.com/stretchr/testify/require"
	"github.com/kyma-project/kyma/components/proxy-service/internal/httpconsts"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuthStrategy(t *testing.T) {

	t.Run("should add Authorization header", func(t *testing.T) {
		// given
		basicAuthStrategy := newBasicAuthStrategy("username", "password")

		request, err := http.NewRequest("GET", "www.example.com", nil)
		require.NoError(t, err)

		// when
		err = basicAuthStrategy.Setup(request)

		// then
		require.NoError(t, err)
		authHeader := request.Header.Get(httpconsts.HeaderAuthorization)
	   	assert.Equal(t, "Basic dXNlcm5hbWU6cGFzc3dvcmQ=", authHeader)
	})
}
