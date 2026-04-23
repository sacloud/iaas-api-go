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
		"ESMESendMessageWithGeneratedOTPRequest": {
			"Destination": "SMS 送信先電話番号 (E.164 形式、\"+81901234567\" 等)",
			"Sender":      "SMS 送信元名 (英数字)",
			"DomainName":  "OTP 送信時に利用するドメイン名 (任意)",
		},
		"ESMESendMessageWithInputtedOTPRequest": {
			"Destination": "SMS 送信先電話番号",
			"Sender":      "SMS 送信元名",
			"DomainName":  "OTP 送信時に利用するドメイン名 (任意)",
			"OTP":         "送信したいワンタイムパスワード文字列",
		},
		"ESMESendMessageResult": {
			"MessageID": "送信メッセージのID",
			"OTP":       "自動発行された OTP (generated モードのみ)",
			"Status":    "送信リクエストの受付ステータス",
		},
		"ESMELogs": {
			"Destination": "送信先電話番号",
			"MessageID":   "メッセージID",
			"OTP":         "OTP 値",
			"Status":      "最新の送信ステータス",
			"RetryCount":  "送信リトライ回数",
			"SentAt":      "送信実行日時",
			"DoneAt":      "最終ステータス確定日時",
		},
	})
}
