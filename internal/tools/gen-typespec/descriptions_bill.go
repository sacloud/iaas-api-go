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
		"Bill": {
			"Amount": "請求金額 (税込、円)",
			"Date":   "請求対象月の締日",
			"Paid":   "支払済みかどうか",
		},
		"BillDetail": {
			"Amount":           "明細金額 (円)",
			"ContractEndAt":    "契約終了日時 (解約済みリソースの場合)",
			"Description":      "明細の内容説明",
			"FormattedUsage":   "表示用にフォーマットされた使用量 (例: \"123時間\")",
			"ServiceClassID":   "サービスクラス ID",
			"ServiceClassPath": "サービスクラスの階層パス",
			"ServiceUsagePath": "利用量リソースのパス",
			"Usage":            "使用量 (数値)",
			"Zone":             "サービスが提供されているゾーン名",
		},
		"BillDetailCSV": {
			"HeaderRow":   "CSV ヘッダ行",
			"BodyRows":    "CSV 本文行の配列",
			"Count":       "明細件数",
			"Filename":    "推奨ファイル名",
			"RawBody":     "CSV 本文 (エスケープ済み文字列)",
			"ResponsedAt": "CSV 生成時刻",
		},
	})
}
