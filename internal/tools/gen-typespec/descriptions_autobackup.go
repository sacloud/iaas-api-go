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
		"AutoBackup": {
			"Settings": "自動バックアップ設定",
			"Status":   "対象ディスク情報",
		},
		"AutoBackupCreateRequest": {
			"Settings": "自動バックアップ設定",
			"Status":   "対象ディスクのID",
		},
		"AutoBackupCreateRequestSettings": {
			"Autobackup": "バックアップ本体設定",
		},
		"AutoBackupCreateRequestSettingsAutobackup": {
			"BackupSpanWeekdays":      "バックアップ取得する曜日の配列 (\"mon\", \"tue\", ...)",
			"MaximumNumberOfArchives": "保持する世代数の上限",
		},
		"AutoBackupCreateRequestStatus": {
			"DiskID": "バックアップ対象ディスクのID",
		},
		"AutoBackupSettingsAutobackup": {
			"BackupSpanWeekdays":      "バックアップ取得する曜日の配列",
			"MaximumNumberOfArchives": "保持世代数の上限",
		},
		"AutoBackupUpdateRequestSettingsAutobackup": {
			"BackupSpanWeekdays":      "バックアップ取得する曜日の配列",
			"MaximumNumberOfArchives": "保持世代数の上限",
		},
		"AutoBackupStatus": {
			"DiskID": "バックアップ対象ディスクのID",
		},
	})
}
