// Package httpcache introduces an in-memory-cached http client into the KrakenD stack
package httpcache

import (
	"context"
	"net/http"

	"github.com/gregjones/httpcache"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/transport/http/client"
)

// Namespace is the key to use to store and access the custom config data
const Namespace = "github.com/devopsfaith/krakend-httpcache"

var (
	memTransport = httpcache.NewMemoryCacheTransport()
	memClient    = http.Client{Transport: memTransport}
)

// NewHTTPClient creates a HTTPClientFactory using an in-memory-cached http client
func NewHTTPClient(cfg *config.Backend) client.HTTPClientFactory {
	_, ok := cfg.ExtraConfig[Namespace]
	if !ok {
		return client.NewHTTPClient
	}
	return func(_ context.Context) *http.Client {
		return &memClient
	}
}

// BackendFactory returns a proxy.BackendFactory that creates backend proxies using
// an in-memory-cached http client
func BackendFactory(cfg *config.Backend) proxy.BackendFactory {
	return proxy.CustomHTTPProxyFactory(NewHTTPClient(cfg))
}
