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
		"GSLB": {
			"DestinationServers": "振り分け先サーバの配列",
			"Settings":           "GSLB 設定",
			"Status":             "GSLB の状態 (割り当てられた FQDN 等)",
		},
		"GSLBSettingsGSLB": {
			"DelayLoop":          "ヘルスチェック間隔 (秒)",
			"HealthCheck":        "ヘルスチェック設定",
			"SorryServer":        "全サーバがダウンした場合に返すフォールバックサーバ IP",
			"Weighted":           "重み付けラウンドロビンを有効化するかどうか",
			"MonitoringSuiteLog": "GSLB ヘルスチェック結果のログ収集設定",
		},
		"GSLBCreateRequestSettingsGSLB": {
			"DelayLoop":          "ヘルスチェック間隔 (秒)",
			"HealthCheck":        "ヘルスチェック設定",
			"SorryServer":        "全サーバダウン時のフォールバック IP",
			"Weighted":           "重み付けラウンドロビン有効化",
			"MonitoringSuiteLog": "ログ収集設定",
		},
		"GSLBHealthCheck": {
			"Protocol": "プロトコル (\"http\", \"https\", \"tcp\", \"ping\")",
			"Port":     "HTTP(S)/TCP 時のチェックポート",
			"Path":     "HTTP(S) 時のチェックパス",
			"Host":     "HTTP(S) 時の Host ヘッダに設定する FQDN",
			"Status":   "HTTP(S) 時に期待するステータスコード",
		},
		"GSLBServer": {
			"IPAddress": "振り分け先サーバの IPv4 アドレス",
			"Enabled":   "振り分け対象にするか",
			"Weight":    "重み (1-10000)。大きいほど多く振り分けられる",
		},
	})
}
