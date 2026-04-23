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
		"ContainerRegistry": {
			"Settings": "コンテナレジストリ設定",
			"Status":   "レジストリの状態 (FQDN 等)",
		},
		"ContainerRegistryCreateRequestSettingsContainerRegistry": {
			"Public":        "アクセスレベル。\"readwrite\" / \"readonly\" / \"none\"",
			"VirtualDomain": "カスタム仮想ドメイン (任意)",
		},
		"ContainerRegistrySettingsContainerRegistry": {
			"Public":        "アクセスレベル",
			"VirtualDomain": "カスタム仮想ドメイン",
		},
		"ContainerRegistryCreateRequestStatus": {
			"RegistryName": "レジストリ名 (FQDN の先頭部分)",
		},
		"ContainerRegistryStatus": {
			"FQDN":         "レジストリの FQDN",
			"RegistryName": "レジストリ名",
		},
	})
}
