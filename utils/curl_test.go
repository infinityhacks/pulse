package pulse

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

//Tests if we can fetch a file from S3 or not...
func TestCurlS3(t *testing.T) {
	req := &CurlRequest{
		Path:     "/tb-minion/latest",
		Endpoint: "s3.amazonaws.com",
		Host:     "s3.amazonaws.com",
		Ssl:      true,
	}
	resp := CurlImpl(context.Background(), req)
	if resp.Err != "" {
		t.Error(resp.Err)
	}
	if resp.Status != 200 {
		t.Error("Status should be 200... got ", resp.Status)
	}
}

//Tests if we can override host header...
func TestCurlInvalidS3(t *testing.T) {
	req := &CurlRequest{
		Path:     "/tb-minion/latest",
		Endpoint: "s3.amazonaws.com",
		Host:     "www.turbobytes.com", //Bogus Host header not configured with S3
		Ssl:      false,
	}
	resp := CurlImpl(context.Background(), req)
	if resp.Err != "" {
		t.Error(resp.Err)
	}
	if resp.Status != 404 {
		t.Error("Status should be 404... got ", resp.Status)
	}
}

//Tests if a local url is being blocked correctly or not...
func TestCurlLocalBlock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()
	url, _ := url.Parse(ts.URL)
	req := &CurlRequest{
		Path:     "/tb-minion/latest",
		Endpoint: url.Host,
		Host:     url.Host,
		Ssl:      false,
	}
	resp := CurlImpl(context.Background(), req)
	if !strings.Contains(resp.Err, securityerr.Error()) {
		t.Error("Security err should have been raised")
	}
}

//Tests if fixipv6endpoint() is behaving correctly or not.
func TestFixipv6endpoint(t *testing.T) {
	//Test endpoints that should *not* be changed
	valid_endpoints := []string{
		"bar.foo.com",
		"[bar.foo.com]",
		"bar.foo.com:123",
		"1.1.1.1",
		"1.1.1.1:432",
		"[2400:cb00:2048:1::c629:d7a2]:443",
		"2400:cb00:2048:1::c629:d7a2",
	}
	for _, ep := range valid_endpoints {
		fixed := fixipv6endpoint(ep)
		if fixed != ep {
			t.Errorf("%s should remain the same, got %s", ep, fixed)
		}
	}
	//Test endpoints that should be changed
	invalid_endpoints := map[string]string{"[2400:cb00:2048:1::c629:d7a2]": "[2400:cb00:2048:1::c629:d7a2]:443"}
	for ep, expected := range invalid_endpoints {
		fixed := fixipv6endpoint(ep)
		if fixed != expected {
			t.Errorf("%s should become %s, got %s", ep, expected, fixed)
		}
	}
}

//Tests if CurlImpl honors deadlines
func TestCurlHardTimeout(t *testing.T) {
	timeout := time.Millisecond
	req := &CurlRequest{
		Path:     "/",
		Endpoint: "www.google.com",
		Host:     "www.google.com",
		Ssl:      true,
	}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	resp := CurlImpl(ctx, req)
	cancel()
	if !strings.Contains(resp.Err, "request canceled") {
		t.Errorf("unexpected error: %s", resp.Err)
	}
	if resp.Status != 0 {
		t.Error("Status should be 0... got ", resp.Status)
	}
}

// Timing tests require hitting an external url towards TurboBytes mock server.
// Maybe this is not a good idea. Perhaps we need a local mocker.
// Perhaps wait for https://go-review.googlesource.com/#/c/29440/ to land in release to help with local mock server.
// Following URLs will give unpredictable results for people behind a transparent proxy.
// example url: http://<random>.imagetest.100msdelay.mock.turbobytes.com:8100/static/rum/image-15kb.jpg
// no DNS delay: <random>.imagetest.npdelay.mock.turbobytes.com
// 100ms DNS delay: <random>.imagetest.100msdelay.mock.turbobytes.com
// 200ms DNS delay: <random>.imagetest.200msdelay.mock.turbobytes.com
// HTTP port 80: no latency
// HTTP port 8100: Additional 100ms network latency
// TODO: Endpoint to simulate delay during TTFB
// TODO: Blackhole test - perhaps use WPT's endpoint - blackhole.webpagetest.org
