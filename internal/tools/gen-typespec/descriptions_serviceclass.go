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
		"ServiceClass": {
			"DisplayName":      "画面表示用の名称",
			"IsPublic":         "公開されているサービスクラスかどうか",
			"Price":            "料金情報",
			"ServiceClassName": "サービスクラスの内部名",
			"ServiceClassPath": "サービスクラスの階層パス",
		},
		"Price": {
			"Base":          "基本料金",
			"Basic":         "ベース料金",
			"Daily":         "日額料金",
			"Hourly":        "時間料金",
			"Monthly":       "月額料金",
			"PerUse":        "使用量ベースの料金",
			"Traffic":       "トラフィック料金",
			"DocomoTraffic": "NTT docomo 網トラフィック料金 (SIM)",
			"KddiTraffic":   "KDDI 網トラフィック料金 (SIM)",
			"SbTraffic":     "SoftBank 網トラフィック料金 (SIM)",
			"SimSheet":      "SIM シート料金",
			"Zone":          "ゾーン別料金情報",
		},
	})
}
