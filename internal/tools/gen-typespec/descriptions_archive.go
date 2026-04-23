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
		"Archive": {
			"SizeMB": "アーカイブサイズ (MB)",
		},
		"ArchiveCreateRequest": {
			"SizeMB":        "アーカイブサイズ (MB)。ブランクアーカイブ作成時に指定する",
			"SourceArchive": "複製元アーカイブ。SourceDisk と排他",
			"SourceDisk":    "複製元ディスク。SourceArchive と排他",
		},
		"ArchiveTransferRequest": {
			"SizeMB": "転送先アーカイブのサイズ (MB)",
		},

		"ArchiveShareRequestEnvelope": {
			"Shared":         "アーカイブの共有を有効にするかどうか",
			"ChangePassword": "共有キーを再発行するかどうか",
		},
		"ArchiveShareInfo": {
			"SharedKey": "他アカウント/ゾーンへアーカイブを転送する際に利用する共有キー",
		},

		"FTPServer": {
			"HostName":  "FTP サーバのホスト名",
			"IPAddress": "FTP サーバの IPv4 アドレス",
			"User":      "FTP 接続ユーザ名",
			"Password":  "FTP 接続パスワード",
		},
		"OpenFTPRequest": {
			"ChangePassword": "FTP のパスワードを再発行するかどうか",
		},

		"SourceArchiveInfo": {
			"ArchiveUnderZone": "他ゾーンからアーカイブ共有転送された場合の情報",
		},
		"SourceArchiveInfoArchiveUnderZone": {
			"Account": "共有元のアカウント情報",
			"Zone":    "共有元のゾーン情報",
		},
		"SourceArchiveInfoArchiveUnderZoneZone": {
			"ID":   "共有元ゾーンのID",
			"Name": "共有元ゾーンの名前",
		},

		"Storage": {
			"Class":                    "ストレージクラス",
			"Generation":               "ストレージ世代",
			"DedicatedStorageContract": "専有ストレージ契約情報",
		},
	})
}
