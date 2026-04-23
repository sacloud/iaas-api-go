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
		"Database": {
			"IPAddresses":       "実インターフェースに割り当てる IPv4 の配列",
			"InterfaceSettings": "インターフェース設定の配列 (IPv4/マスク長等)",
			"Remark":            "不変情報 (プラン/接続スイッチ/DB 種別/ソース・アプライアンス)",
			"Settings":          "データベースの可変設定 (DB 設定、バックアップ、レプリケーション、監視)",
			"Disk":              "データ格納ディスク情報 (暗号化情報含む)",
		},
		"DatabaseCreateRequest": {
			"IPAddresses":       "実インターフェースに割り当てる IPv4 の配列",
			"InterfaceSettings": "インターフェース設定の配列",
			"Remark":            "不変情報",
			"Settings":          "データベース設定",
			"Disk":              "データ格納ディスク設定",
		},
		"DatabaseCreateRequestRemark": {
			"Plan":            "データベースプラン ID",
			"Network":         "ネットワーク情報",
			"Switch":          "接続するスイッチ",
			"DBConf":          "DB エンジン情報 (MariaDB / PostgreSQL の種別)",
			"SourceAppliance": "マネージドレプリカのソース DB アプライアンス ID",
		},
		"DatabaseCreateRequestRemarkDBConf": {
			"Common": "DB 共通設定 (バージョン等)",
		},
		"DatabaseCreateRequestRemarkNetwork": {
			"DefaultRoute":   "デフォルトルートの IPv4",
			"NetworkMaskLen": "ネットワークマスク長",
		},
		"DatabaseCreateRequestSettings": {
			"DBConf":          "DB 設定詳細",
			"MonitoringSuite": "アクティビティモニタリング設定",
		},
		"DatabaseCreateRequestSettingsDBConf": {
			"Common":      "DB 共通設定 (管理ユーザ、パスワード、ポート、ソースネットワーク等)",
			"Backup":      "従来バックアップ設定 (v1)",
			"Backupv2":    "バックアップ設定 (v2)",
			"Replication": "レプリケーション設定",
		},

		"DatabaseDisk": {
			"EncryptionAlgorithm": "ディスク暗号化アルゴリズム",
			"EncryptionKey":       "KMS 暗号化キー情報",
		},
		"DatabaseDiskEncryptionKey": {
			"KMSKeyID": "KMS キーID",
		},

		"DatabaseBackupHistory": {
			"RecoveredAt": "リストア (リカバリ) に利用された日時。未使用なら null",
			"Size":        "バックアップサイズ (バイト)",
		},
		"DatabaseLog": {
			"Name": "ログ名 (エラーログ / スローログ等)",
			"Data": "ログ本文",
			"Size": "ログサイズ",
		},
		"DatabaseParameter": {
			"MetaInfo":  "DB パラメータメタ情報 (利用可能な範囲や例)",
			"Parameter": "現在適用中のパラメータ",
		},
		"DatabaseParameterMeta": {
			"Label":   "パラメータのラベル",
			"Name":    "パラメータ名",
			"Options": "取り得る値の制約 (最大値、例示)",
		},
		"DatabaseParameterMetaOptions": {
			"Example": "例示値",
			"Max":     "最大値",
		},

		"DatabaseInstance": {
			"Status": "データベースアプライアンスの稼働状態",
			"Host":   "配置先ホスト情報",
		},
	})
}
