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
		"Disk": {
			"SizeMB":              "ディスクサイズ (MB)",
			"Connection":          "サーバへの接続方式。\"virtio\" (推奨) または \"ide\"",
			"EncryptionAlgorithm": "ディスク暗号化アルゴリズム。\"none\" または \"aes256_xts\"",
			"EncryptionKey":       "暗号化ディスクに紐付く KMS キー情報",
			"Plan":                "ディスクプラン (SSD / 標準 など)",
			"SourceArchive":       "複製元アーカイブ。SourceDisk と排他。どちらか片方のみ指定する",
			"SourceDisk":          "複製元ディスク。SourceArchive と排他。どちらか片方のみ指定する",
			"Storage":             "配置先ストレージ情報",
			"Server":              "接続先サーバ (未接続なら null)",
		},
		"DiskCreateRequest": {
			"SizeMB":              "ディスクサイズ (MB)",
			"Connection":          "サーバへの接続方式。\"virtio\" (推奨) または \"ide\"",
			"EncryptionAlgorithm": "ディスク暗号化アルゴリズム。\"none\" または \"aes256_xts\"",
			"Plan":                "ディスクプラン (ResourceRef で ID 指定)",
			"SourceArchive":       "複製元アーカイブ。SourceDisk と排他",
			"SourceDisk":          "複製元ディスク。SourceArchive と排他",
		},
		"DiskCreateRequestEnvelope": {
			"Disk":                           "ディスク作成パラメータ",
			"Config":                         "作成と同時に適用するディスクの修正パラメータ (ホスト名/SSHキー/IPアドレス等)",
			"BootAtAvailable":                "ディスク作成完了後に接続サーバを起動するかどうか",
			"DistantFrom":                    "他ディスクの物理配置と分散させたい場合の対向ディスクIDの配列",
			"KMSKey":                         "暗号化に利用する KMS キー情報",
			"TargetDedicatedStorageContract": "配置先の専有ストレージ契約ID",
		},

		"DiskUpdateRequest": {
			"Connection": "サーバへの接続方式 (\"virtio\" / \"ide\")",
		},

		"DiskEditRequest": {
			"HostName":            "ゲスト OS に設定するホスト名",
			"Password":            "ゲスト OS の administrator/root パスワード",
			"SSHKey":              "登録する SSH 公開鍵 (単一)",
			"SSHKeys":             "登録する SSH 公開鍵の配列",
			"UserIPAddress":       "ゲスト OS に設定する IPv4 アドレス",
			"UserSubnet":          "ゲスト OS に設定するサブネット情報 (マスク長/デフォルトルート)",
			"Notes":               "起動時に適用するスタートアップスクリプト (Note) の配列",
			"DisablePWAuth":       "SSH パスワード認証を無効化するかどうか",
			"EnableDHCP":          "DHCP を有効化するかどうか",
			"ChangePartitionUUID": "ディスクのパーティション UUID を再生成するかどうか",
			"Background":          "ディスク修正を非同期で実行するかどうか",
		},
		"DiskConfigRequestEnvelope": {
			"HostName":            "ゲスト OS に設定するホスト名",
			"Password":            "ゲスト OS の administrator/root パスワード",
			"SSHKey":              "登録する SSH 公開鍵 (単一)",
			"SSHKeys":             "登録する SSH 公開鍵の配列",
			"UserIPAddress":       "ゲスト OS に設定する IPv4 アドレス",
			"UserSubnet":          "ゲスト OS に設定するサブネット情報",
			"Notes":               "起動時に適用するスタートアップスクリプト (Note) の配列",
			"DisablePWAuth":       "SSH パスワード認証を無効化するかどうか",
			"EnableDHCP":          "DHCP を有効化するかどうか",
			"ChangePartitionUUID": "ディスクのパーティション UUID を再生成するかどうか",
			"Background":          "ディスク修正を非同期で実行するかどうか",
		},
		"DiskResizePartitionRequest": {
			"Background": "パーティションリサイズを非同期で実行するかどうか",
		},
		"DiskResizePartitionRequestEnvelope": {
			"Background": "パーティションリサイズを非同期で実行するかどうか",
		},

		"DiskEditSSHKey": {
			"PublicKey": "登録する SSH 公開鍵文字列",
		},
		"DiskEditNote": {
			"APIKey":    "Note 実行時に渡す API キー (ResourceRef で ID 指定)",
			"Variables": "Note に渡す変数 (key-value マップ)",
		},
		"DiskEditUserSubnet": {
			"DefaultRoute":   "ゲスト OS に設定するデフォルトルート",
			"NetworkMaskLen": "ゲスト OS に設定するネットワークマスク長",
		},
		"DiskEncryptionKey": {
			"KMSKeyID": "KMS キーのID",
		},

		"DiskSourceArchive": {
			"Availability": "複製元アーカイブの利用状態",
		},
		"DiskSourceDisk": {
			"Availability": "複製元ディスクの利用状態",
		},
		"DiskServer": {
			"ID":   "接続先サーバのID",
			"Name": "接続先サーバの名前",
		},

		"JobStatus": {
			"Status":      "非同期ジョブの状態。\"done\" / \"running\" / \"failed\" 等",
			"ConfigError": "ディスク修正ジョブで発生したエラー詳細 (正常終了時は null)",
		},
		"JobConfigError": {
			"ErrorCode": "エラーコード",
			"ErrorMsg":  "エラーメッセージ",
			"Status":    "エラー発生時のジョブ状態",
		},
		"Storage": {
			"Class":                    "ストレージクラス",
			"Generation":               "ストレージ世代",
			"DedicatedStorageContract": "専有ストレージ契約情報",
		},
	})
}
