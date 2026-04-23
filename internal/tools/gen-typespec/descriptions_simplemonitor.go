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
		"SimpleMonitor": {
			"Settings": "シンプル監視設定",
			"Status":   "監視対象 (\"Target\") 情報",
		},
		"SimpleMonitorCreateRequestSettingsSimpleMonitor": {
			"Enabled":           "監視を有効化するかどうか",
			"DelayLoop":         "監視間隔 (秒)",
			"HealthCheck":       "ヘルスチェック設定 (プロトコル、ポート、期待ステータス等)",
			"MaxCheckAttempts":  "リトライ上限回数",
			"RetryInterval":     "リトライ間隔 (秒)",
			"Timeout":           "タイムアウト (秒)",
			"NotifyEmail":       "メール通知設定",
			"NotifySlack":       "Slack 通知設定",
			"NotifyInterval":    "通知を繰り返す間隔 (秒)",
			"MonitoringSuiteLog": "監視結果のログ収集設定",
		},
		"SimpleMonitorCreateRequestSettingsSimpleMonitorNotifyEmail": {
			"Enabled": "メール通知を有効化するか",
			"HTML":    "HTML メールで送信するか",
		},
		"SimpleMonitorCreateRequestSettingsSimpleMonitorNotifySlack": {
			"Enabled":             "Slack 通知を有効化するか",
			"IncomingWebhooksURL": "Slack Incoming Webhook URL",
		},
		"MonitoringSuiteLog": {
			"Enabled": "ログ収集機能を有効化するかどうか",
		},
	})
}
