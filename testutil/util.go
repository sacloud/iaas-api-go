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

package testutil

import (
	"fmt"
	"os"
	"sync"

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/api"
	"github.com/sacloud/packages-go/testutil"
)

// TestResourcePrefix テスト時に作成するリソースの名称に付与するプレフィックス
//
// このプレフィックスを持つリソースは受入テスト実行後に削除される
const TestResourcePrefix = "iaas-api-go-test-"

// ResourceName テスト時に作成するリソースの名称
func ResourceName(name string) string {
	return fmt.Sprintf("%s%s", TestResourcePrefix, name)
}

// RandomPrefix テスト時に作成するリソースに付与するランダムなプレフィックスを生成する
func RandomPrefix() string {
	return testutil.RandomName(TestResourcePrefix, 5, testutil.CharSetAlpha) + "-"
}

// WithRandomPrefix ランダムなプレフィックスをつけて返す
func WithRandomPrefix(name string) string {
	return fmt.Sprintf("%s%s", RandomPrefix(), name)
}

var apiCaller iaas.APICaller

var accTestOnce sync.Once
var accTestMu sync.Mutex

// SingletonAPICaller 環境変数からシングルトンAPICallerを作成する
func SingletonAPICaller() iaas.APICaller {
	accTestMu.Lock()
	defer accTestMu.Unlock()
	accTestOnce.Do(func() {
		if !IsAccTest() {
			os.Setenv("SAKURACLOUD_ACCESS_TOKEN", "dummy")
			os.Setenv("SAKURACLOUD_ACCESS_TOKEN_SECRET", "dummy")
		}

		accessToken := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
		accessTokenSecret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
		apiCaller = api.NewCallerWithOptions(&api.CallerOptions{
			Options: &client.Options{
				AccessToken:       accessToken,
				AccessTokenSecret: accessTokenSecret,
				UserAgent:         "test-" + iaas.DefaultUserAgent,
				AcceptLanguage:    "en-US,en;q=0.9",
				RetryMax:          20,
				Trace:             IsEnableTrace() || IsEnableHTTPTrace(),
			},
			TraceAPI: IsEnableTrace() || IsEnableAPITrace(),
			FakeMode: !IsAccTest(),
		})
	})
	return apiCaller
}

// TestZone SAKURACLOUD_ZONE環境変数からテスト対象のゾーンを取得 デフォルトはtk1v
func TestZone() string {
	testZone := os.Getenv("SAKURACLOUD_ZONE")
	if testZone == "" {
		testZone = "tk1v"
	}
	return testZone
}

// IsAccTest TESTACC環境変数が指定されているか
func IsAccTest() bool {
	return testutil.IsAccTest()
}

// IsEnableTrace SAKURACLOUD_TRACE環境変数が指定されているか
func IsEnableTrace() bool {
	return os.Getenv("SAKURACLOUD_TRACE") != ""
}

// IsEnableAPITrace SAKURACLOUD_TRACE_API環境変数が指定されているか
func IsEnableAPITrace() bool {
	return os.Getenv("SAKURACLOUD_TRACE_API") != ""
}

// IsEnableHTTPTrace SAKURACLOUD_TRACE_HTTP環境変数が指定されているか
func IsEnableHTTPTrace() bool {
	return os.Getenv("SAKURACLOUD_TRACE_HTTP") != ""
}
