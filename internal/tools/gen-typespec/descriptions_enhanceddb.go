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
		"EnhancedDB": {
			"Status": "TiDB 情報 (データベース名、接続情報、リージョン)",
		},
		"EnhancedDBConfig": {
			"AllowedNetworks": "接続を許可する CIDR の配列",
			"MaxConnections":  "最大同時接続数",
		},
		"EnhancedDBCreateRequestStatus": {
			"DatabaseName": "論理データベース名",
			"DatabaseType": "データベース種別 (\"tidb\" または \"mariadb\")",
			"Region":       "TiDB 配置リージョン (\"is1\" 等)",
		},
		"EnhancedDBStatus": {
			"DatabaseName": "論理データベース名",
			"DatabaseType": "データベース種別",
			"Region":       "配置リージョン",
			"HostName":     "接続用ホスト名",
			"Port":         "接続ポート",
		},
		"EnhancedDBSetConfigRequest": {
			"EnhancedDB": "設定値のラッパー",
		},
		"EnhancedDBSetConfigRequestEnhancedDB": {
			"AllowedNetworks": "接続を許可する CIDR の配列",
		},
		"EnhancedDBSetPasswordRequest": {
			"EnhancedDB": "パスワード再設定のラッパー",
		},
		"EnhancedDBSetPasswordRequestEnhancedDB": {
			"Password": "新しいパスワード",
		},
	})
}
