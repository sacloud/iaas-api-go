// Copyright 2022-2026 The sacloud/iaas-api-go Authors
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

package query

import (
	"context"
	"errors"

	"github.com/sacloud/iaas-api-go/v2/client"
)

// FindServerPlanRequest は ServerPlan 検索の条件。
// v2 の ServerPlanFindFilter は Name しか持たないため、その他の条件は
// サーバー取得後クライアント側で絞り込む。
type FindServerPlanRequest struct {
	CPU            int32
	MemoryGB       int32
	GPU            int32
	GPUModel       string
	CPUModel       string
	Commitment     string // "standard" / "dedicatedcpu" 等。空は指定なし。
	Generation     int32  // 100 / 200 等。0 は指定なし。
	ConfidentialVM bool
}

// FindServerPlan は希望条件に最もマッチするサーバプランを 1 件返す。
// Generation 降順で並べ、条件にマッチする最初のプランを選ぶ (v1 互換)。
func FindServerPlan(ctx context.Context, finder ServerPlanFinder, req *FindServerPlanRequest) (*client.ServerPlan, error) {
	resp, err := finder.List(ctx, &client.ServerPlanFindRequest{Count: 1000})
	if err != nil {
		return nil, err
	}
	plans := resp.ServerPlans

	// Generation 降順でソート (新しい世代を優先)
	sortServerPlansByGenerationDesc(plans)

	if req == nil {
		if len(plans) == 0 {
			return nil, errors.New("server plan not found")
		}
		return &plans[0], nil
	}

	for _, p := range plans {
		if !matchServerPlan(&p, req) {
			continue
		}
		return &p, nil
	}
	return nil, errors.New("server plan not found")
}

func sortServerPlansByGenerationDesc(plans []client.ServerPlan) {
	// 単純な挿入ソート。件数が少ない (< 100) ため十分。
	for i := 1; i < len(plans); i++ {
		for j := i; j > 0 && int32(plans[j].Generation.Value) > int32(plans[j-1].Generation.Value); j-- {
			plans[j], plans[j-1] = plans[j-1], plans[j]
		}
	}
}

func matchServerPlan(p *client.ServerPlan, req *FindServerPlanRequest) bool {
	if req.CPU > 0 && p.CPU.Value != req.CPU {
		return false
	}
	if req.MemoryGB > 0 && p.MemoryMB.Value != req.MemoryGB*1024 {
		return false
	}
	if req.GPU > 0 && p.GPU.Value != req.GPU {
		return false
	}
	if req.GPUModel != "" && p.GPUModel.Value != req.GPUModel {
		return false
	}
	if req.CPUModel != "" && p.CPUModel.Value != req.CPUModel {
		return false
	}
	if req.Commitment != "" && string(p.Commitment.Value) != req.Commitment {
		return false
	}
	if req.Generation > 0 && int32(p.Generation.Value) != req.Generation {
		return false
	}
	// Availability が available のものに絞る
	if string(p.Availability.Value) != "available" {
		return false
	}
	return true
}
