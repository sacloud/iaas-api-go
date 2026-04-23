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
		"MobileGateway": {
			"InterfaceSettings": "インターフェース設定",
			"StaticRoutes":      "スタティックルートの配列",
			"Settings":          "モバイルゲートウェイ可変設定",
			"Remark":            "不変情報",
		},
		"MobileGatewayCreateRequest": {
			"Settings":     "モバイルゲートウェイ設定",
			"StaticRoutes": "スタティックルートの配列",
		},
		"MobileGatewayCreateRequestSettings": {
			"MobileGateway": "モバイルゲートウェイ本体設定",
		},
		"MobileGatewayCreateRequestSettingsMobileGateway": {
			"InterDeviceCommunication": "セキュアモバイル接続機器同士の通信を許可するか",
			"InternetConnection":       "インターネット接続を許可するか",
		},
		"MobileGatewayCreateRequestSettingsMobileGatewayInterDeviceCommunication": {
			"Enabled": "機器間通信の有効化フラグ",
		},
		"MobileGatewayCreateRequestSettingsMobileGatewayInternetConnection": {
			"Enabled": "インターネット接続の有効化フラグ",
		},

		"MobileGatewayRemark": {
			"MobileGateway": "モバイルゲートウェイの不変情報",
		},
		"MobileGatewayRemarkMobileGateway": {
			"GlobalAddress": "モバイルゲートウェイのグローバル IP アドレス",
		},

		"MobileGatewayDNSSetting": {
			"DNS1": "プライマリ DNS サーバの IPv4",
			"DNS2": "セカンダリ DNS サーバの IPv4",
		},

		"MobileGatewayAddSIMRequest": {
			"ResourceID": "追加する SIM のリソース ID",
		},
		"MobileGatewayAddSIMRequestEnvelope": {
			"SIM": "追加する SIM のリソース ID",
		},

		"MobileGatewayInterface": {
			"HostName":     "インターフェースのホスト名",
			"Switch":       "接続されているスイッチ",
			"PacketFilter": "割り当てられているパケットフィルタ",
		},
		"MobileGatewayInterfaceSetting": {
			"IPAddress":      "割り当てる IPv4 の配列",
			"NetworkMaskLen": "ネットワークマスク長",
			"Index":          "インターフェース番号",
		},
		"MobileGatewayInterfacePacketFilter": {
			"RequiredHostVersionn": "このパケットフィルタが要求するホストバージョン (※ API 側のフィールド名のタイポを保持)",
		},
		"MobileGatewayInterfaceSwitch": {
			"Subnet":     "接続されているサブネット情報",
			"UserSubnet": "ユーザが指定した論理サブネット情報",
		},

		"MobileGatewayInstance": {
			"Status": "モバイルゲートウェイの稼働状態",
			"Host":   "配置先ホスト情報",
		},

		"MobileGatewaySIMInfo": {
			"Activated": "SIM が有効化されているかどうか",
		},
	})
}
