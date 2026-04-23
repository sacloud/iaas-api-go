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
		"AutoScale": {
			"Settings": "オートスケール設定",
			"Status":   "オートスケールに紐付く API キー情報",
		},
		"AutoScaleCreateRequestSettings": {
			"Zones":                  "オートスケール対象となるゾーン名の配列",
			"Config":                 "スケーリング設定本体 (YAML 文字列)",
			"TriggerType":            "発火条件種別 (\"cpu\", \"router\", \"schedule\")",
			"Disabled":               "オートスケールを一時停止するかどうか",
			"CPUThresholdScaling":    "CPU 使用率ベースのスケーリング設定",
			"RouterThresholdScaling": "ルータトラフィックベースのスケーリング設定",
			"ScheduleScaling":        "スケジュールベースのスケーリング設定",
		},
		"AutoScaleCreateRequestStatus": {
			"APIKey": "オートスケール実行用 API キー (ResourceRef で ID 指定)",
		},
		"AutoScaleCPUThresholdScaling": {
			"ServerPrefix": "対象サーバ名の接頭辞",
			"Up":           "スケールアウト閾値 (CPU 使用率 %)",
			"Down":         "スケールイン閾値 (CPU 使用率 %)",
		},
		"AutoScaleRouterThresholdScaling": {
			"Direction":    "監視方向 (\"in\" / \"out\")",
			"Mbps":         "閾値 (Mbps)",
			"RouterPrefix": "対象ルータ名の接頭辞",
		},
		"AutoScaleScheduleScaling": {
			"Action":    "スケール動作 (\"up\" / \"down\")",
			"DayOfWeek": "動作させる曜日の配列",
			"Hour":      "動作時刻 (時)",
			"Minute":    "動作時刻 (分)",
		},
	})
}
