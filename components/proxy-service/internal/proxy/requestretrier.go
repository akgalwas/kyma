package proxy

import (
	"context"
	"github.com/kyma-project/kyma/components/proxy-service/internal/proxy/proxycache"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type retrier struct {
	id      string
	request *http.Request
	retried bool
	timeout int
	proxyCacheEntry   *proxycache.CacheEntry
}

func newRequestRetrier(id string, request *http.Request, proxyCacheEntry *proxycache.CacheEntry, timeout int) *retrier {
	return &retrier{id: id, request: request, retried: false, proxyCacheEntry: proxyCacheEntry, timeout: timeout}
}

func (rr *retrier) CheckResponse(r *http.Response) error {
	if rr.retried {
		return nil
	}

	rr.retried = true

	if r.StatusCode == 403 {
		log.Infof("Request from service with id %s failed with 403 status, invalidating proxy and retrying.", rr.id)

		res, err := rr.retry()
		if err != nil {
			return err
		}

		if res != nil {
			r.Body.Close()
			*r = *res
		}

	}

	return nil
}

func (rr *retrier) retry() (*http.Response, error) {
	request, cancel := rr.prepareRequest()
	defer cancel()

	err := rr.addAuthorization(request)
	if err != nil {
		return nil, err
	}

	return rr.performRequest(request)
}

func (rr *retrier) prepareRequest() (*http.Request, context.CancelFunc) {
	rr.request.RequestURI = ""
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(rr.timeout)*time.Second)

	return rr.request.WithContext(ctx), cancel
}

func (rr *retrier) addAuthorization(r *http.Request) error {
	authorizationStrategy := rr.proxyCacheEntry.AuthorizationStrategy
	authorizationStrategy.Reset()

	return authorizationStrategy.Setup(r)
}

func (rr *retrier) performRequest(r *http.Request) (*http.Response, error) {
	reverseProxy := rr.proxyCacheEntry.Proxy
	reverseProxy.Director(r)

	client := &http.Client{
		Transport: reverseProxy.Transport,
	}

	res, err := client.Do(r)

	return res, err
}
