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
		"NFS": {
			"IPAddresses": "実インターフェースに割り当てる IPv4 の配列",
			"Switch":      "接続されているスイッチ情報",
			"Remark":      "不変情報 (プラン ID / 接続スイッチ / ネットワーク)",
		},
		"NFSCreateRequest": {
			"IPAddresses": "実インターフェースに割り当てる IPv4 の配列",
			"Remark":      "不変情報",
		},
		"NFSCreateRequestRemark": {
			"Plan":    "NFS プラン ID (サイズとストレージ種別を特定)",
			"Switch":  "接続するスイッチ",
			"Network": "ネットワーク情報",
		},
		"NFSCreateRequestRemarkNetwork": {
			"DefaultRoute":   "デフォルトルートの IPv4",
			"NetworkMaskLen": "ネットワークマスク長",
		},
		"NFSRemark": {
			"Plan":    "NFS プラン ID",
			"Network": "ネットワーク情報",
		},
		"NFSRemarkNetwork": {
			"DefaultRoute":   "デフォルトルートの IPv4",
			"NetworkMaskLen": "ネットワークマスク長",
		},
		"NFSSwitch": {
			"Name": "接続先スイッチ名",
		},
		"NFSInstance": {
			"Status": "NFS の稼働状態",
			"Host":   "配置先ホスト情報",
		},
	})
}
