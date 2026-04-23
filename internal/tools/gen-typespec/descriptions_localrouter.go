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
		"LocalRouter": {
			"Peers":        "ピア接続先の配列 (他ローカルルータとの対向接続)",
			"StaticRoutes": "スタティックルートの配列",
			"Settings":     "ローカルルータ設定",
			"Status":       "接続スイッチや VIP 等のステータス",
		},
		"LocalRouterPeer": {
			"ID":        "ピア先ローカルルータのID",
			"SecretKey": "ピア間で共有するシークレットキー",
			"Enabled":   "ピア接続を有効化するか",
		},
		"LocalRouterInterface": {
			"VRID":             "VRRP 仮想ルータ ID",
			"VirtualIPAddress": "VRRP の仮想 IPv4 アドレス",
		},
		"LocalRouterHealth": {
			"Peers": "各ピアのヘルス状態",
		},
		"LocalRouterHealthPeer": {
			"ID":     "ピア先のID",
			"Status": "接続状態",
			"Routes": "このピア経由で受信しているルートの配列",
		},
	})
}
