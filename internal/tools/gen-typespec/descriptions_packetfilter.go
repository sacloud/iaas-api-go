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
		"PacketFilter": {
			"Expression": "フィルタルール (Expression) の配列。先頭から順に評価される",
		},
		"PacketFilterCreateRequest": {
			"Expression": "フィルタルールの配列",
		},
		"PacketFilterUpdateRequest": {
			"Expression": "フィルタルールの配列",
		},
		"PacketFilterUpdateRequestEnvelope": {
			"OriginalExpressionHash": "楽観的排他制御用の元 Expression ハッシュ。同時更新衝突検知に利用される",
		},
		"PacketFilterExpression": {
			"Action":          "\"allow\" または \"deny\"",
			"Protocol":        "プロトコル (\"tcp\", \"udp\", \"icmp\", \"fragment\", \"ip\")",
			"SourceNetwork":   "送信元ネットワーク (CIDR、\"0.0.0.0/0\" 相当でも可)",
			"SourcePort":      "送信元ポート (単独または \"1024-65535\" のレンジ)",
			"DestinationPort": "宛先ポート (単独またはレンジ)",
		},
	})
}
