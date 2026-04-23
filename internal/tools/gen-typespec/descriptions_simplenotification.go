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
		"SimpleNotificationDestination": {
			"Settings": "通知先設定",
		},
		"SimpleNotificationDestinationCreateRequestSettings": {
			"Type":     "通知先種別 (\"slack\" / \"email\" / \"webhook\" 等)",
			"Value":    "通知先エンドポイント (Slack Webhook URL / メールアドレス等)",
			"Disabled": "この通知先を一時的に無効化するか",
		},
		"SimpleNotificationDestinationSettings": {
			"Type":     "通知先種別",
			"Value":    "通知先エンドポイント",
			"Disabled": "無効化フラグ",
		},

		"SimpleNotificationGroup": {
			"Settings": "通知グループ設定",
		},
		"SimpleNotificationGroupCreateRequestSettings": {
			"Sources":      "通知元となるリソースID (もしくはグループ ID) の配列",
			"Destinations": "通知先 (Destination) のID 配列",
			"Disabled":     "このグループを一時的に無効化するか",
		},
		"SimpleNotificationGroupSettings": {
			"Sources":      "通知元となるリソースIDの配列",
			"Destinations": "通知先のID配列",
			"Disabled":     "無効化フラグ",
		},
	})
}
