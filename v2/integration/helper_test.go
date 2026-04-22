// Copyright 2022-2026 The sacloud/iaas-api-go Authors
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
	"os"
	"testing"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/saclient-go"
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
func getConfig() (accessToken, accessTokenSecret string) {
	return os.Getenv("SAKURA_ACCESS_TOKEN"),
		os.Getenv("SAKURA_ACCESS_TOKEN_SECRET")
}

// newClient は integration テスト用の ogen クライアントを生成する。
//
// saclient-go のプロファイル / 環境変数解決をそのまま使い、
// iaas.NewClient 経由で認証 / User-Agent / X-Sakura-Bigint-As-Int ヘッダ /
// find query 書き換え / ogen 空 Authorization の除去を一括適用する。
//
// SAKURA_TRACE=1 が設定されていれば saclient 内蔵 tracer を "all" モードで
// 有効化し、リクエスト / レスポンスを log パッケージ経由で stderr にダンプする。
func newClient(t *testing.T) *client.Client {
	t.Helper()
	accessToken, accessTokenSecret := getConfig()
	if accessToken == "" || accessTokenSecret == "" {
		t.Skip("SAKURA_ACCESS_TOKEN and SAKURA_ACCESS_TOKEN_SECRET must be set")
	}

	var sc saclient.Client
	if err := sc.SetEnviron(os.Environ()); err != nil {
		t.Fatalf("saclient SetEnviron: %v", err)
	}

	var scAPI saclient.ClientAPI = &sc
	if os.Getenv("SAKURA_TRACE") == "1" {
		dupped, err := sc.DupWith(saclient.WithTraceMode("all"))
		if err != nil {
			t.Fatalf("saclient DupWith(trace): %v", err)
		}
		scAPI = dupped
	}

	c, err := iaas.NewClient(scAPI, getZone())
	if err != nil {
		t.Fatalf("iaas.NewClient: %v", err)
	}
	return c
}

// newClientForZone は明示的にゾーンを指定してクライアントを生成する。
// archive の他ゾーン転送のようにテスト中で複数ゾーンを切り替える場合に使う。
func newClientForZone(t *testing.T, zone string) *client.Client {
	t.Helper()
	accessToken, accessTokenSecret := getConfig()
	if accessToken == "" || accessTokenSecret == "" {
		t.Skip("SAKURA_ACCESS_TOKEN and SAKURA_ACCESS_TOKEN_SECRET must be set")
	}

	var sc saclient.Client
	if err := sc.SetEnviron(os.Environ()); err != nil {
		t.Fatalf("saclient SetEnviron: %v", err)
	}

	var scAPI saclient.ClientAPI = &sc
	if os.Getenv("SAKURA_TRACE") == "1" {
		dupped, err := sc.DupWith(saclient.WithTraceMode("all"))
		if err != nil {
			t.Fatalf("saclient DupWith(trace): %v", err)
		}
		scAPI = dupped
	}

	c, err := iaas.NewClient(scAPI, zone)
	if err != nil {
		t.Fatalf("iaas.NewClient: %v", err)
	}
	return c
}
