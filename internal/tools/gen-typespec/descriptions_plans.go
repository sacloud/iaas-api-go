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

// descriptions_plans.go は "*Plan" リソース (ServerPlan / DiskPlan / InternetPlan / PrivateHostPlan)
// に共通する説明を登録する。
func init() {
	registerFieldDescriptions(map[string]map[string]string{
		"ServerPlan": {
			"CPU":            "CPU コア数",
			"MemoryMB":       "メモリ容量 (MB)",
			"GPU":            "GPU 数 (GPU プランの場合)",
			"GPUModel":       "GPU モデル名",
			"CPUModel":       "CPU モデル (\"uncategorized\" など)",
			"Commitment":     "CPU 割当コミット方式 (\"standard\" / \"dedicatedcpu\")",
			"Generation":     "プラン世代 (100=第1、200=第2(新) 等)",
		},

		"DiskPlan": {
			"Size":         "このプランで選択可能なサイズ情報の配列",
			"StorageClass": "配置ストレージクラス (\"iscsi1204\" 等)",
		},
		"DiskPlanSizeInfo": {
			"SizeMB":        "実サイズ (MB)",
			"DisplaySize":   "表示用のサイズ数値",
			"DisplaySuffix": "表示用の単位 (\"GB\" 等)",
		},

		"InternetPlan": {
			"BandWidthMbps": "帯域幅 (Mbps)",
		},

		"PrivateHostPlan": {
			"CPU":       "このプランで提供される CPU コア数",
			"MemoryMB":  "このプランで提供されるメモリ容量 (MB)",
			"Class":     "専用ホストのクラス (\"dynamic\" 等)",
			"Dedicated": "専用ホスト (占有) かどうか",
		},
	})
}
