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
	"net/http"
	"os"
	"testing"

	"github.com/sacloud/iaas-api-go/v2/client"
)

// getZone returns the zone to use for testing.
func getZone() string {
	zone := os.Getenv("SAKURA_ZONE")
	if zone == "" {
		return "tk1v" // デフォルトは sandbox zone
	}
	return zone
}

// getConfig returns the client configuration.
func getConfig() (accessToken, accessTokenSecret, zone string) {
	return os.Getenv("SAKURA_ACCESS_TOKEN"),
		os.Getenv("SAKURA_ACCESS_TOKEN_SECRET"),
		getZone()
}

// newClient creates an ogen client for integration tests.
func newClient(t *testing.T) *client.Client {
	t.Helper()
	accessToken, accessTokenSecret, zone := getConfig()

	if accessToken == "" || accessTokenSecret == "" {
		t.Skip("SAKURA_ACCESS_TOKEN and SAKURA_ACCESS_TOKEN_SECRET must be set")
	}

	serverURL := "https://secure.sakura.ad.jp/cloud/zone/" + zone + "/api/cloud/1.1"

	httpClient := &http.Client{
		Transport: &authTransport{
			AccessToken:       accessToken,
			AccessTokenSecret: accessTokenSecret,
			Base:              http.DefaultTransport,
		},
	}

	c, err := client.NewClient(serverURL, client.WithClient(httpClient))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return c
}

// authTransport adds authentication headers to requests.
type authTransport struct {
	AccessToken       string
	AccessTokenSecret string
	Base              http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(t.AccessToken, t.AccessTokenSecret)
	req.Header.Set("X-Sakura-Bigint-As-Int", "1")
	return t.Base.RoundTrip(req)
}
