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
// このマップは、単純な合流では API 仕様と整合が取れない op（wire 形式が根本的に違うなど）を
// 例外的に除外したい場合の最終手段として残している。現時点で除外対象なし。
var excludedOps = map[string]map[string]bool{}

func opIsExcluded(api *dsl.Resource, op *dsl.Operation) bool {
	if s, ok := excludedOps[api.Name]; ok {
		return s[op.Name]
	}
	return false
}
