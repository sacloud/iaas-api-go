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

// Package iaas は iaas-api-go v2 の公開ラッパー。
//
// saclient-go（認証・プロファイル解決・ヘッダ付与・リトライ等の共通土台）を
// 受け取り、ogen 生成の v2/client パッケージを組み立てる。
// リソースごとの CRUD は NewNoteOp のような Op コンストラクタを経由して呼び出す。
// 使用感は sacloud/simple-notification-api-go 等の sacloud-sdk-go ファミリに揃えている。
package iaas

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/saclient-go"
)

const (
	// APIRootURLTemplate デフォルトの API ルート URL テンプレート。
	// OpenAPI の servers 変数と同じく {zone} プレースホルダを含む。
	// NewClient / NewClientWithAPIRootURL では zone を埋め込んで使用する。
	APIRootURLTemplate = "https://secure.sakura.ad.jp/cloud/zone/{zone}/api/cloud/1.1"

	// ServiceKey SDKの種別を示すキー、プロファイルでのエンドポイント取得に利用する。
	// SAKURA_ENDPOINTS_iaas 環境変数などで上書き可能。
	// エンドポイント値は {zone} プレースホルダを含むテンプレートで指定する。
	ServiceKey = "iaas"
)

// UserAgent APIリクエスト時のユーザーエージェント。
var UserAgent = fmt.Sprintf(
	"iaas-api-go/%s (%s/%s; +https://github.com/sacloud/iaas-api-go)",
	Version,
	runtime.GOOS,
	runtime.GOARCH,
)

// noopSecuritySource は ogen の SecuritySource インターフェースを満たすだけの
// no-op 実装。認証は saclient 側のミドルウェアに委譲するため、ここでは
// 空の BasicAuth を返す（付与された空 Authorization ヘッダは
// stripOgenAuthMiddleware が削除する）。
type noopSecuritySource struct{}

func (noopSecuritySource) BasicAuth(_ context.Context, _ client.OperationName) (client.BasicAuth, error) {
	return client.BasicAuth{}, nil
}

// NewClient は zone と saclient.ClientAPI を受け取って ogen クライアントを組み立てる。
// URL テンプレートは saclient.EndpointConfig の "iaas" キーを優先し、
// 見つからなければ APIRootURLTemplate にフォールバックする。いずれも {zone}
// プレースホルダを含み、本関数内で zone が埋め込まれる。
func NewClient(c saclient.ClientAPI, zone string) (*client.Client, error) {
	if zone == "" {
		return nil, NewError("zone must not be empty", nil)
	}
	endpointConfig, err := c.EndpointConfig()
	if err != nil {
		return nil, NewError("unable to load endpoint configuration", err)
	}
	template := APIRootURLTemplate
	if ep, ok := endpointConfig.Endpoints[ServiceKey]; ok && ep != "" {
		template = ep
	}
	return NewClientWithAPIRootURL(c, resolveAPIRootURL(template, zone))
}

// NewClientWithAPIRootURL は zone 解決済みの完全な API ルート URL を指定して
// クライアントを組み立てる。テストや非標準環境向け。通常は NewClient を使うこと。
func NewClientWithAPIRootURL(c saclient.ClientAPI, apiRootURL string) (*client.Client, error) {
	dupable, ok := c.(saclient.ClientOptionAPI)
	if !ok {
		return nil, NewError("client does not implement saclient.ClientOptionAPI", nil)
	}
	augmented, err := dupable.DupWith(
		saclient.WithUserAgent(UserAgent),
		saclient.WithMiddleware(
			stripOgenAuthMiddleware(),
			findQueryRewriteMiddleware(),
		),
	)
	if err != nil {
		return nil, err
	}
	return client.NewClient(apiRootURL, noopSecuritySource{}, client.WithClient(augmented))
}

// resolveAPIRootURL は URL テンプレート中の {zone} を指定ゾーンで置換する。
// テンプレートに {zone} が含まれない（zone を含めない形で設定された）場合は
// そのまま返す。
func resolveAPIRootURL(template, zone string) string {
	return strings.ReplaceAll(template, "{zone}", zone)
}
