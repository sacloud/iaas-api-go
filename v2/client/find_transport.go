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

package client

import (
	"net/http"
	"net/url"
	"strings"
)

// findQueryRewriteTransport は Find 系 GET リクエストの `?q=<urlencoded-json>` を
// 生 JSON のクエリストリング（`?<json>`）に書き換えるための http.RoundTripper。
//
// 背景: さくらクラウドの現行サーバー実装は GET /.../bridge?{"Count":3} のように、
// クエリストリング直下に JSON を置く非標準フォーマットを要求する。OpenAPI では
// これを直接表現できないため、TypeSpec/OpenAPI では将来形 `?q={json}` として定義し、
// クライアントはこの RoundTripper を通して `q=` プレフィックスを剥がす形で現行サーバーと
// 通信する。
//
// 将来サーバーが `?q={json}` を受けられるようになった時点で、この RoundTripper は削除し、
// ogen 生成のままで動作するようになる。
type findQueryRewriteTransport struct {
	base http.RoundTripper
}

// NewFindQueryRewriteTransport は findQueryRewriteTransport を作成する。
// base が nil のときは http.DefaultTransport を使用する。
func NewFindQueryRewriteTransport(base http.RoundTripper) http.RoundTripper {
	return &findQueryRewriteTransport{base: base}
}

// WithFindQueryRewrite は findQueryRewriteTransport で wrap された http.Client を
// ClientOption として返す。NewClient 呼び出し時にこれを渡せば Find 系 GET は
// 正しく `?{json}` 形式で送出される。
//
// 既に `WithClient` で独自の http.Client を渡している場合はこちらの option を
// 利用せず、Transport を自分で `NewFindQueryRewriteTransport` で wrap すること。
//
// 例:
//
//	c, err := client.NewClient(serverURL, sec, client.WithFindQueryRewrite())
func WithFindQueryRewrite() ClientOption {
	return WithClient(&http.Client{Transport: NewFindQueryRewriteTransport(nil)})
}

func (t *findQueryRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet && strings.HasPrefix(req.URL.RawQuery, "q=") {
		if decoded, err := url.QueryUnescape(strings.TrimPrefix(req.URL.RawQuery, "q=")); err == nil && strings.HasPrefix(decoded, "{") {
			r2 := req.Clone(req.Context())
			// url.URL.RawQuery は「エスケープ済みの値」として扱われるため、JSON の `{}` `"` を
			// そのまま文字列として詰めておけば送信時もエスケープなしで使われる。
			r2.URL.RawQuery = decoded
			req = r2
		}
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}
