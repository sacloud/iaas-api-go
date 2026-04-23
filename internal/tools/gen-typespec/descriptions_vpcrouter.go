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
		"VPCRouter": {
			"Remark":   "VPC ルータのプランや接続スイッチなど不変的な情報 (作成時に決定)",
			"Settings": "VPC ルータの可変設定 (ファイアウォール、VPN、DHCP 等)",
		},
		"VPCRouterCreateRequest": {
			"Plan":        "VPC ルータプラン ID (1=standard, 2=premium, 3=highspec, 4=highspec4000)",
			"IPAddresses": "プレミアムプラン以上で必須となる実インターフェースの IPv4 アドレスの配列",
			"Remark":      "VPC ルータの不変情報 (プランや接続スイッチ)",
			"Settings":    "VPC ルータの設定",
		},
		"VPCRouterCreateRequestRemark": {
			"Router": "VPC ルータのバージョン情報",
			"Switch": "接続するスイッチの情報",
		},
		"VPCRouterCreateRequestRemarkRouter": {
			"VPCRouterVersion": "VPC ルータの機能世代番号 (1 = 標準、2 = プレミアム版)",
		},

		"VPCRouterInstance": {
			"Status":          "VPC ルータの稼働状態",
			"StatusChangedAt": "稼働状態が最後に変化した日時",
			"Host":            "配置先ホスト情報",
		},

		"VPCRouterInterface": {
			"Switch":       "接続先スイッチ",
			"PacketFilter": "割り当てられているパケットフィルタ",
		},
		"VPCRouterInterfaceSetting": {
			"IPAddress":        "インターフェースに割り当てる IPv4 の配列 (プレミアムプランは冗長構成のため複数)",
			"IPAliases":        "同一インターフェースに追加で割り当てる IPv4 エイリアス",
			"VirtualIPAddress": "VRRP の仮想 IPv4 アドレス (冗長化プラン用)",
			"NetworkMaskLen":   "インターフェースに設定するネットワークマスク長",
			"Index":            "インターフェース番号 (0 = eth0 / 実回線側、1 以降 = ユーザスイッチ側)",
		},

		"VPCRouterDHCPServer": {
			"Interface":  "DHCP を有効化するインターフェース名 (例: \"eth1\")",
			"RangeStart": "DHCP で配布する IP レンジの開始アドレス",
			"RangeStop":  "DHCP で配布する IP レンジの終了アドレス",
			"DNSServers": "クライアントに通知する DNS サーバの配列",
		},
		"VPCRouterDHCPServerLease": {
			"IPAddress":  "リース中の IPv4 アドレス",
			"MACAddress": "クライアントの MAC アドレス",
		},
		"VPCRouterDHCPStaticMapping": {
			"IPAddress":  "静的に割り当てる IPv4 アドレス",
			"MACAddress": "対象クライアントの MAC アドレス",
		},
		"VPCRouterDNSForwarding": {
			"Interface":  "DNS フォワーディングを有効化するインターフェース",
			"DNSServers": "問い合わせを転送する DNS サーバ",
		},

		"VPCRouterFirewall": {
			"Receive": "受信方向 (inbound) のフィルタルール",
			"Send":    "送信方向 (outbound) のフィルタルール",
		},
		"VPCRouterFirewallRule": {
			"Action":             "\"allow\" または \"deny\"",
			"Protocol":           "プロトコル (\"tcp\", \"udp\", \"icmp\", \"fragment\", \"ip\")",
			"SourceNetwork":      "送信元ネットワーク (CIDR)",
			"SourcePort":         "送信元ポート (単独/レンジ)",
			"DestinationNetwork": "宛先ネットワーク (CIDR)",
			"DestinationPort":    "宛先ポート (単独/レンジ)",
			"Logging":            "ルール一致時のロギング有無",
		},

		"ApplianceConnectedSwitch": {
			"ID":    "接続先スイッチのID",
			"Scope": "\"shared\" = 共有セグメント、\"user\" = ユーザスイッチ",
		},
		"MonitoringSuite": {
			"Enabled": "アクティビティモニタリングの有効化フラグ",
		},
	})
}
