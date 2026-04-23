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
		"Internet": {
			"BandWidthMbps":  "回線帯域幅 (Mbps)。プランにより 100/250/500/1000/1500/2000/2500/3000/5000/10000 等",
			"NetworkMaskLen": "サブネットマスク長 (28/27/26)",
			"Switch":         "対応するスイッチの情報",
		},
		"InternetCreateRequest": {
			"BandWidthMbps":  "回線帯域幅 (Mbps)",
			"NetworkMaskLen": "サブネットマスク長 (28/27/26)",
		},
		"InternetUpdateBandWidthRequest": {
			"BandWidthMbps": "変更後の回線帯域幅 (Mbps)",
		},
		"InternetAddSubnetRequest": {
			"NetworkMaskLen": "追加サブネットのマスク長 (28/27/26)",
			"NextHop":        "追加サブネットへのネクストホップ",
		},
		"InternetAddSubnetRequestEnvelope": {
			"NetworkMaskLen": "追加サブネットのマスク長",
			"NextHop":        "追加サブネットへのネクストホップ",
		},
		"InternetUpdateSubnetRequest": {
			"NextHop": "サブネットのネクストホップ",
		},
		"InternetUpdateSubnetRequestEnvelope": {
			"NextHop": "サブネットのネクストホップ",
		},

		"SwitchInfo": {
			"Subnets":  "このスイッチに紐付くサブネットの配列",
			"IPv6Nets": "このスイッチに紐付く IPv6 ネットワークの配列",
		},
		"IPv6NetInfo": {
			"IPv6Prefix":    "IPv6 プレフィックス",
			"IPv6PrefixLen": "IPv6 プレフィックス長 (一般に 64)",
		},
		"InternetSubnet": {
			"IPAddresses": "サブネットで利用可能な IPv4 の配列",
		},
		"InternetSubnetOperationResult": {
			"IPAddresses": "サブネットで利用可能な IPv4 の配列",
		},
	})
}
