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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFindQueryRewriteTransport(t *testing.T) {
	var gotRawQuery string
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotRawQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	t.Run("rewrites q= prefix to raw JSON", func(t *testing.T) {
		hc := &http.Client{Transport: NewFindQueryRewriteTransport(nil)}
		req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		// クライアントが組み立てた URL は `?q=%7B%22Count%22%3A3%7D` 相当
		req.URL.RawQuery = "q=" + urlEncode(`{"Count":3}`)
		resp, err := hc.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		_ = resp.Body.Close()

		if gotMethod != http.MethodGet {
			t.Errorf("method = %q, want GET", gotMethod)
		}
		if gotRawQuery != `{"Count":3}` {
			t.Errorf("RawQuery = %q, want %q", gotRawQuery, `{"Count":3}`)
		}
	})

	t.Run("leaves non-GET requests untouched", func(t *testing.T) {
		hc := &http.Client{Transport: NewFindQueryRewriteTransport(nil)}
		req, err := http.NewRequest(http.MethodPost, srv.URL+"?q="+urlEncode(`{"Count":3}`), strings.NewReader("body"))
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		resp, err := hc.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		_ = resp.Body.Close()
		// POST は書き換えない (URL-encoded のまま)
		if gotRawQuery != "q="+urlEncode(`{"Count":3}`) {
			t.Errorf("POST RawQuery = %q, want URL-encoded passthrough", gotRawQuery)
		}
	})

	t.Run("leaves non-JSON q= untouched", func(t *testing.T) {
		hc := &http.Client{Transport: NewFindQueryRewriteTransport(nil)}
		req, err := http.NewRequest(http.MethodGet, srv.URL+"?q=foo", nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		resp, err := hc.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		_ = resp.Body.Close()
		if gotRawQuery != "q=foo" {
			t.Errorf("non-JSON RawQuery = %q, want unchanged", gotRawQuery)
		}
	})

	t.Run("leaves empty query untouched", func(t *testing.T) {
		hc := &http.Client{Transport: NewFindQueryRewriteTransport(nil)}
		req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
		if err != nil {
			t.Fatalf("NewRequest: %v", err)
		}
		resp, err := hc.Do(req)
		if err != nil {
			t.Fatalf("Do: %v", err)
		}
		_ = resp.Body.Close()
		if gotRawQuery != "" {
			t.Errorf("empty RawQuery = %q, want empty", gotRawQuery)
		}
	})
}

// urlEncode は url.QueryEscape 相当だが、テストケースの可読性のために
// 明示的にヘルパーとして切り出している。
func urlEncode(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '{':
			b.WriteString("%7B")
		case '}':
			b.WriteString("%7D")
		case '"':
			b.WriteString("%22")
		case ':':
			b.WriteString("%3A")
		case ',':
			b.WriteString("%2C")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
