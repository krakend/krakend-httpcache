Krakend HTTP Cache
====

A cached http client for the [KrakenD](https://krakend.io) framework

## Using it

This package exposes two simple factories capable to create a instances of the `proxy.HTTPClientFactory` and the `proxy.BackendFactory` interfaces, respectively, embedding an in-memory-cached http client using the package [github.com/gregjones/httpcache](https://github.com/gregjones/httpcache). The client will cache the responses honoring the defined Cache HTTP header.

	import 	(
		"context"
		"net/http"
		"github.com/luraproject/lura/v3/config"
		"github.com/luraproject/lura/v3/proxy"
		"github.com/luraproject/lura/v3/transport/http/client"
		"github.com/krakend/krakend-httpcache/v3"
	)

	requestExecutorFactory := func(cfg *config.Backend) proxy.HTTPRequestExecutor {
		clientFactory := httpcache.NewHTTPClient(cfg, client.NewHTTPClient)
		return func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return clientFactory(ctx).Do(req.WithContext(ctx))
		}
	}

You can create your own proxy.HTTPRequestExecutor and inject it into your BackendFactory
