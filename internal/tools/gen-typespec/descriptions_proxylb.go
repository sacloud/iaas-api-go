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

package main

func init() {
	registerFieldDescriptions(map[string]map[string]string{
		"ProxyLB": {
			"Plan":      "プラン (最大同時接続数による区分)",
			"BindPorts": "待ち受けポート設定の配列 (HTTP/HTTPS)",
			"Servers":   "振り分け先サーバの配列",
			"Rules":     "振り分けルール (条件付きのバックエンド選択)",
			"Settings":  "エンハンスド LB 設定",
			"Status":    "状態情報 (FQDN、VIP、証明書有効期限 等)",
		},
		"ProxyLBBindPort": {
			"ProxyMode":         "プロキシモード (\"http\", \"https\", \"tcp\")",
			"Port":              "待ち受けポート番号",
			"RedirectToHTTPS":   "HTTP→HTTPS リダイレクトを有効化するか",
			"AddResponseHeader": "レスポンスに追加するヘッダの配列",
		},
		"ProxyLBBackendHttpKeepAlive": {
			"Mode": "バックエンドへの HTTP Keep-Alive モード",
		},
		"ProxyLBACMESetting": {
			"Enabled":         "Let's Encrypt (ACME) を利用するかどうか",
			"CommonName":      "発行する証明書の CommonName",
			"SubjectAltNames": "発行する証明書の SAN の配列",
		},
		"ProxyLBAdditionalCert": {
			"ServerCertificate":       "サーバ証明書 (PEM)",
			"IntermediateCertificate": "中間 CA 証明書 (PEM)",
			"PrivateKey":              "秘密鍵 (PEM)",
			"CertificateCommonName":   "証明書の CommonName",
			"CertificateAltNames":     "証明書の SAN",
			"CertificateEndDate":      "証明書の有効期限",
		},
		"LoadBalancerServerStatus": {
			"IPAddress":  "振り分け先サーバの IPv4",
			"Port":       "振り分け先ポート",
			"Status":     "振り分け先サーバの状態 (\"up\" / \"down\")",
			"ActiveConn": "現在のアクティブ接続数",
			"CPS":        "秒間接続数",
		},
		"MonitoringSuiteLog": {
			"Enabled": "ログ収集機能を有効化するかどうか",
		},
	})
}
