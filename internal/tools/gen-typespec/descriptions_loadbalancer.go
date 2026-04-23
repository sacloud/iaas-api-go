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
		"LoadBalancer": {
			"IPAddresses":        "実インターフェースに割り当てる IPv4 アドレスの配列 (冗長プランは 2 つ)",
			"VirtualIPAddresses": "仮想 IP (VIP) の配列。VIP ごとに振り分け先サーバを定義する",
			"Remark":             "ロードバランサの不変情報 (プランや接続スイッチ)",
		},
		"LoadBalancerCreateRequest": {
			"IPAddresses":        "実インターフェースに割り当てる IPv4 の配列 (冗長プランは 2 つ)",
			"VirtualIPAddresses": "仮想 IP (VIP) の配列",
			"Remark":             "ロードバランサの不変情報 (プラン/スイッチ/VRRP)",
		},
		"LoadBalancerCreateRequestRemark": {
			"Plan":    "ロードバランサプラン ID (1=standard, 2=highspec)",
			"Network": "ネットワーク情報 (デフォルトルート、マスク長)",
			"Switch":  "接続するスイッチ",
			"VRRP":    "VRRP 設定 (仮想ルータ ID)",
		},
		"LoadBalancerCreateRequestRemarkNetwork": {
			"DefaultRoute":   "デフォルトルートの IPv4",
			"NetworkMaskLen": "ネットワークマスク長",
		},
		"LoadBalancerCreateRequestRemarkVRRP": {
			"VRID": "仮想ルータ ID (1-255)。同一セグメントで重複禁止",
		},
		"LoadBalancerRemark": {
			"Plan":    "ロードバランサプラン ID",
			"Network": "ネットワーク情報",
			"VRRP":    "VRRP 設定",
		},
		"LoadBalancerRemarkNetwork": {
			"DefaultRoute":   "デフォルトルートの IPv4",
			"NetworkMaskLen": "ネットワークマスク長",
		},
		"LoadBalancerRemarkVRRP": {
			"VRID": "仮想ルータ ID",
		},

		"LoadBalancerServer": {
			"IPAddress":   "振り分け先サーバの IPv4 アドレス",
			"Port":        "振り分け先のポート番号",
			"Enabled":     "このサーバをバランシング対象にするか",
			"HealthCheck": "ヘルスチェック設定",
		},
		"LoadBalancerServerHealthCheck": {
			"Protocol":       "ヘルスチェックプロトコル (\"tcp\", \"http\", \"https\", \"ping\")",
			"Path":           "HTTP(S) 時のチェック用パス",
			"Status":         "HTTP(S) 時に期待するステータスコード",
			"ConnectTimeout": "接続タイムアウト (秒)",
			"Retry":          "リトライ回数",
		},

		"LoadBalancerInstance": {
			"Status": "ロードバランサの稼働状態",
			"Host":   "配置先ホスト情報",
		},
	})
}
