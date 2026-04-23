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
		"Interface": {
			"Server":        "接続先サーバ情報 (未接続なら null)",
			"Switch":        "接続先スイッチ情報 (未接続なら null)",
			"UserIPAddress": "ユーザが手動で割り当てた IPv4 アドレス",
		},
		"InterfaceCreateRequest": {
			"Server": "NIC を作成して接続するサーバ (ResourceRef で ID 指定)",
		},
		"InterfaceUpdateRequest": {
			"UserIPAddress": "ユーザが手動で設定する IPv4 アドレス",
		},
		"InterfaceSwitch": {
			"ID":    "接続先スイッチのID",
			"Scope": "\"shared\" または \"user\"",
		},
	})
}
