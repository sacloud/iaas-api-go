// Copyright 2022-2025 The sacloud/iaas-api-go Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"testing"

	"github.com/sacloud/iaas-api-go/v2/client"
)

// testUserAgent is the default User-Agent for integration tests.
var testUserAgent = "iaas-api-go-acc (Go " + runtime.Version() + ")"

// getZone returns the zone to use for testing.
func getZone() string {
	zone := os.Getenv("SAKURA_ZONE")
	if zone == "" {
		return "tk1v" // デフォルトは sandbox zone
	}
	return zone
}

// getConfig returns the client configuration.
func getConfig() (accessToken, accessTokenSecret string) {
	return os.Getenv("SAKURA_ACCESS_TOKEN"),
		os.Getenv("SAKURA_ACCESS_TOKEN_SECRET")
}

// securitySource implements client.SecuritySource for BasicAuth.
type securitySource struct {
	username string
	password string
}

func (s *securitySource) BasicAuth(ctx context.Context, operationName client.OperationName) (client.BasicAuth, error) {
	return client.BasicAuth{
		Username: s.username,
		Password: s.password,
	}, nil
}

// baseTransport wraps http.DefaultTransport and adds required headers.
type baseTransport struct {
	base http.RoundTripper
}

func (b *baseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if b.base == nil {
		b.base = http.DefaultTransport
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", testUserAgent)
	}
	req.Header.Set("X-Sakura-Bigint-As-Int", "1")
	return b.base.RoundTrip(req)
}

// dumpTransport logs HTTP requests and responses for debugging.
type dumpTransport struct {
	base http.RoundTripper
}

func (d *dumpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if d.base == nil {
		d.base = &baseTransport{base: http.DefaultTransport}
	}

	reqDump, _ := httputil.DumpRequestOut(req, true)
	fmt.Fprintf(os.Stderr, "---> REQUEST\n%s\n", reqDump)

	resp, err := d.base.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "---> ERROR: %v\n", err)
		return nil, err
	}

	respDump, _ := httputil.DumpResponse(resp, true)
	fmt.Fprintf(os.Stderr, "<--- RESPONSE\n%s\n", respDump)
	return resp, nil
}

// newClient creates an ogen client for integration tests.
func newClient(t *testing.T) *client.Client {
	t.Helper()
	accessToken, accessTokenSecret := getConfig()

	if accessToken == "" || accessTokenSecret == "" {
		t.Skip("SAKURA_ACCESS_TOKEN and SAKURA_ACCESS_TOKEN_SECRET must be set")
	}

	serverURL := "https://secure.sakura.ad.jp/cloud/zone"

	sec := &securitySource{
		username: accessToken,
		password: accessTokenSecret,
	}

	transport := http.RoundTripper(&baseTransport{base: http.DefaultTransport})
	if os.Getenv("SAKURA_TRACE") == "1" {
		transport = &dumpTransport{base: transport}
	}
	opts := []client.ClientOption{
		client.WithClient(&http.Client{Transport: transport}),
	}

	c, err := client.NewClient(serverURL, sec, opts...)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return c
}
