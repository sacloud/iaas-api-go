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

package iaas

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/sacloud/saclient-go"
)

// stripOgenAuthMiddleware は ogen が SecuritySource から入れた Authorization
// ヘッダを除去する。iaas-api-go/v2 は認証を saclient-go 側に寄せており、
// ogen 生成コード側の SecuritySource は no-op（空 BasicAuth を返すだけ）。
// ogen は SetBasicAuth("", "") を無条件で呼ぶため "Basic Og==" が付いてしまい、
// saclient.middlewareAuthorization は「Authorization が既に設定済みならスキップ」
// という挙動をとる（authorization.go 参照）ため、そのままでは空の Basic 認証が
// サーバーへ送られ 401 になる。本ミドルウェアを chain の先頭に置いて
// ヘッダを剥がし、saclient 側が正規の認証を付与できるようにする。
func stripOgenAuthMiddleware() saclient.Middleware {
	return func(req *http.Request, pull func() (saclient.Middleware, bool)) (*http.Response, error) {
		req.Header.Del("Authorization")
		return callNext(req, pull)
	}
}

// findQueryRewriteMiddleware は Find 系 GET リクエストの
// `?q=<urlencoded-json>` を生 JSON の `?<json>` に書き換える。
// v2/client/find_transport.go と同じロジックを saclient ミドルウェアとして
// 実装したもの。背景（なぜこの書き換えが必要か）は find_transport.go 参照。
func findQueryRewriteMiddleware() saclient.Middleware {
	return func(req *http.Request, pull func() (saclient.Middleware, bool)) (*http.Response, error) {
		if req.Method == http.MethodGet && strings.HasPrefix(req.URL.RawQuery, "q=") {
			if decoded, err := url.QueryUnescape(strings.TrimPrefix(req.URL.RawQuery, "q=")); err == nil && strings.HasPrefix(decoded, "{") {
				r2 := req.Clone(req.Context())
				r2.URL.RawQuery = decoded
				req = r2
			}
		}
		return callNext(req, pull)
	}
}

// callNext は saclient の pullThenCall 相当（unexported のため同等実装）。
func callNext(req *http.Request, pull func() (saclient.Middleware, bool)) (*http.Response, error) {
	next, ok := pull()
	if !ok {
		return nil, errors.New("iaas: no next middleware")
	}
	return next(req, pull)
}
