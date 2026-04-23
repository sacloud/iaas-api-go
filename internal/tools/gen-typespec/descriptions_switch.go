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
		"Switch": {
			"ServerCount": "このスイッチに接続されているサーバ台数",
			"Subnets":     "このスイッチに紐付くサブネットの配列 (ルータ+スイッチ時のみ)",
			"Bridge":      "スイッチ間接続用ブリッジ情報 (接続されている場合)",
		},
		"SwitchSubnet": {
			"IPAddresses": "このサブネットで割り当て可能な IPv4 範囲",
			"Internet":    "ルータ+スイッチの場合の回線情報",
		},
		"SwitchSubnetIPAddresses": {
			"Max": "割り当て可能な IPv4 範囲の上限",
			"Min": "割り当て可能な IPv4 範囲の下限",
		},
		"InternetInfo": {
			"BandWidthMbps": "ルータの帯域幅 (Mbps)",
		},
	})
}
