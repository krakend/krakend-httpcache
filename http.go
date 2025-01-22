// Package httpcache introduces an in-memory-cached http client into the KrakenD stack
package httpcache

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/krakend/lru"
	"github.com/krakendio/httpcache"

	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/transport/http/client"
)

type Cache interface {
	// Get returns the []byte representation of a cached response and a bool
	// set to true if the value isn't empty
	Get(key string) (responseBytes []byte, ok bool)
	// Set stores the []byte representation of a response against a key
	Set(key string, responseBytes []byte)
	// Delete removes the value associated with the key
	Delete(key string)
}

// Namespace is the key to use to store and access the custom config data
const Namespace = "github.com/devopsfaith/krakend-httpcache"

// NewHTTPClient creates a HTTPClientFactory using an in-memory-cached http client
func NewHTTPClient(cfg *config.Backend, nextF client.HTTPClientFactory) client.HTTPClientFactory {
	raw, ok := cfg.ExtraConfig[Namespace]
	if !ok {
		return nextF
	}

	b, err := json.Marshal(raw)
	if err != nil {
		return defaultClient(nextF)
	}

	var opts options
	if err := json.Unmarshal(b, &opts); err != nil {
		return defaultClient(nextF)
	}

	return getCachedClient(nextF, selectCache(opts))
}

func defaultClient(nextF client.HTTPClientFactory) client.HTTPClientFactory {
	return getCachedClient(nextF, httpcache.NewMemoryCache())
}

func getCachedClient(nextF client.HTTPClientFactory, cache Cache) client.HTTPClientFactory {
	return func(ctx context.Context) *http.Client {
		httpClient := nextF(ctx)
		return &http.Client{
			Transport: &httpcache.Transport{
				Transport: httpClient.Transport,
				Cache:     cache,
			},
			CheckRedirect: httpClient.CheckRedirect,
			Jar:           httpClient.Jar,
			Timeout:       httpClient.Timeout,
		}
	}
}

func selectCache(opts options) Cache {
	if opts.MaxSize == 0 || opts.MaxItems == 0 {
		if opts.Shared {
			return globalCache
		}
		return httpcache.NewMemoryCache()
	}

	if !opts.Shared {
		cache, _ := lru.NewLruCache(opts.MaxSize, opts.MaxItems)
		return cache
	}

	if globalLruCache == nil {
		globalLruCache, _ = lru.NewLruCache(opts.MaxSize, opts.MaxItems)
	}
	return globalLruCache
}

var (
	globalLruCache Cache
	globalCache    = httpcache.NewMemoryCache()
)

type options struct {
	Shared   bool   `json:"shared"`
	MaxSize  uint64 `json:"max_size"`
	MaxItems uint64 `json:"max_items"`
}
