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
		"DNS": {
			"Records":  "DNS レコードの配列",
			"Settings": "DNS 可変設定",
			"Status":   "DNS の状態 (NS 割り当て、ゾーン名)",
		},
		"DNSStatus": {
			"Zone": "DNS ゾーン名 (対象ドメイン)",
			"NS":   "割り当てられた権威 NS サーバの配列",
		},
		"DNSRecord": {
			"Name":  "レコード名 (サブドメイン部)",
			"Type":  "レコードタイプ (\"A\", \"AAAA\", \"CNAME\", \"MX\", \"TXT\" など)",
			"RData": "レコードデータ。TTL のほか、MX なら \"10 mail.example.com.\" 形式等",
			"TTL":   "TTL (秒)。未指定ならゾーンのデフォルトが適用される",
		},
		"DNSSettings": {
			"DNS": "DNS 設定本体",
		},
		"DNSSettingsDNS": {
			"MonitoringSuiteLog": "DNS クエリログ収集設定",
		},
		"DNSCreateRequestSettings": {
			"DNS": "DNS 設定本体",
		},
		"DNSCreateRequestSettingsDNS": {
			"MonitoringSuiteLog": "DNS クエリログ収集設定",
		},
		"DNSUpdateRequestSettings": {
			"DNS": "DNS 設定本体",
		},
		"MonitoringSuiteLog": {
			"Enabled": "ログ収集機能を有効化するかどうか",
		},
	})
}
