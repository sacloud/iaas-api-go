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

import "github.com/sacloud/iaas-api-go/internal/dsl"

// excludedOps は v2 TypeSpec の生成対象から除外する DSL オペレーションの集合。
// キー = API（リソース）名、値 = 除外する op 名のセット。
//
// 除外基準: v1 DSL が同一エンドポイントに対して複数の Go メソッドを用意しているケースは、
// 通常 `buildMergedEnvelopeInfos` と `computeRequestModelMerges` で primary に合流させて
// 1 API 定義にまとめるため、このマップへ載せる必要はない。
//
// さらに、downstream (usacloud / terraform-provider-sakuracloud / terraform-provider-sakura) と
// iaas-service-go いずれからも呼ばれていない Monitor/Log/Status 系オペレーションをここで除外する。
// 除外するとそのオペレーションに紐づくリクエスト/レスポンスモデル (例: VPCRouterLog, DatabaseLog) も
// 他オペレーションから参照されない限り emit されなくなる。
var excludedOps = map[string]map[string]bool{
	// 未使用 Monitor 系 (iaas-service-go に対応する monitor_*_service.go が存在しない)
	"NFS":          {"MonitorCPU": true},
	"LoadBalancer": {"MonitorCPU": true, "Status": true},

	// 未使用 Logs 系 (iaas-service-go に対応する logs_service.go が存在せず、downstream も呼ばない)
	"Database":  {"Logs": true},
	"VPCRouter": {"Logs": true},

	// 未使用 Status 系
	"SIM":                           {"Status": true},
	"SimpleNotificationDestination": {"Status": true},
}

func opIsExcluded(api *dsl.Resource, op *dsl.Operation) bool {
	if s, ok := excludedOps[api.Name]; ok {
		return s[op.Name]
	}
	return false
}
