package httpcache

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/krakendio/httpcache"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/encoding"
	"github.com/luraproject/lura/v2/proxy"
	"github.com/luraproject/lura/v2/transport/http/client"
)

var maxRequests = 100

func TestClient_shared(t *testing.T) {
	globalCache = httpcache.NewMemoryCache()
	defer func() { globalCache = nil }()

	cfg := &config.Backend{
		Decoder: encoding.JSONDecoder,
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				"shared": true,
			},
		},
	}

	b := newBackend(300)
	defer b.Close()

	testClient(t, cfg, b.URL())
	testClient(t, cfg, b.URL())
	testClient(t, cfg, b.URL())
	testClient(t, cfg, b.URL())

	if hits := b.Count(); hits != 1 {
		t.Errorf("unexpected number of hits. got: %d, want: 1", hits)
	}
}

func TestClient_refresh(t *testing.T) {
	globalCache = httpcache.NewMemoryCache()
	defer func() { globalCache = nil }()

	cfg := &config.Backend{
		Decoder: encoding.JSONDecoder,
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{
				"shared": true,
			},
		},
	}

	b := newBackend(1)
	defer b.Close()

	testClient(t, cfg, b.URL())
	<-time.After(1500 * time.Millisecond)
	testClient(t, cfg, b.URL())
	<-time.After(1500 * time.Millisecond)
	testClient(t, cfg, b.URL())

	if hits := b.Count(); hits != 3 {
		t.Errorf("unexpected number of hits. got: %d, want: 3", hits)
	}
}

func TestClient_dedicated(t *testing.T) {
	globalCache = httpcache.NewMemoryCache()
	defer func() { globalCache = nil }()

	b := newBackend(300)
	defer b.Close()

	{
		cfg := &config.Backend{
			Decoder: encoding.JSONDecoder,
			ExtraConfig: map[string]interface{}{
				Namespace: false,
			},
		}

		testClient(t, cfg, b.URL())

		if hits := b.Count(); hits != 1 {
			t.Errorf("unexpected number of hits. got: %d, want: 1", hits)
		}
	}

	{
		cfg := &config.Backend{
			Decoder: encoding.JSONDecoder,
			ExtraConfig: map[string]interface{}{
				Namespace: map[string]interface{}{},
			},
		}

		testClient(t, cfg, b.URL())

		if hits := b.Count(); hits != 2 {
			t.Errorf("unexpected number of hits. got: %d, want: 2", hits)
		}
	}
}

func TestClient_noCache(t *testing.T) {
	globalCache = httpcache.NewMemoryCache()
	defer func() { globalCache = nil }()

	b := newBackend(300)
	defer b.Close()

	cfg := &config.Backend{
		Decoder:     encoding.JSONDecoder,
		ExtraConfig: map[string]interface{}{},
	}

	testClient(t, cfg, b.URL())

	if hits := b.Count(); hits != uint64(maxRequests) {
		t.Errorf("unexpected number of hits. got: %d, want: %d", hits, uint64(maxRequests))
	}
}

func TestClient_backendFactory(t *testing.T) {
	globalCache = httpcache.NewMemoryCache()
	defer func() { globalCache = nil }()

	b := newBackend(300)
	defer b.Close()

	sampleCfg := &config.Backend{
		Decoder: encoding.JSONDecoder,
		ExtraConfig: map[string]interface{}{
			Namespace: map[string]interface{}{},
		},
	}

	httpClientFactory := NewHTTPClient(sampleCfg, client.NewHTTPClient)
	backendFactory := proxy.CustomHTTPProxyFactory(httpClientFactory)
	backendProxy := backendFactory(sampleCfg)
	ctx := context.Background()
	URL, _ := url.Parse(b.URL())

	for i := 0; i < maxRequests; i++ {
		req := &proxy.Request{
			Method: "GET",
			URL:    URL,
			Body:   ioutil.NopCloser(bytes.NewBufferString("")),
		}
		resp, err := backendProxy(ctx, req)
		if err != nil {
			t.Error(err)
			return
		}
		if !resp.IsComplete {
			t.Error("incomplete response:", *resp)
			return
		}
	}

	if hits := b.Count(); hits != 1 {
		t.Errorf("unexpected number of hits. got: %d, want: %d", hits, 1)
	}
}

func testClient(t *testing.T, cfg *config.Backend, URL string) {
	c := NewHTTPClient(cfg, client.NewHTTPClient)(context.Background())

	for i := 0; i < maxRequests; i++ {
		resp, err := c.Get(URL)
		if err != nil {
			t.Error(err)
			return
		}
		response, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Error(err)
			return
		}
		if string(response) != statusOKMsg {
			t.Error("unexpected body:", string(response))
		}
	}
}

const statusOKMsg = `{"status": "ok"}`

func newBackend(ttl int) backend {
	var ops uint64
	return backend{
		server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint64(&ops, 1)
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", ttl))
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, statusOKMsg)
		})),
		ops: &ops,
	}
}

type backend struct {
	server *httptest.Server
	ops    *uint64
}

func (b backend) Close() {
	b.server.Close()
}

func (b backend) Count() uint64 {
	return atomic.LoadUint64(b.ops)
}

func (b backend) URL() string {
	return b.server.URL
}
