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

// Package plans は v2 リソースのプラン変更時に旧 ID を保持するヘルパーを提供する。
//
// v1 の github.com/sacloud/iaas-api-go/helper/plans の v2 相当。
package plans

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/sacloud/iaas-api-go/v2/client"
)

const (
	// PreviousIDTagName は旧 ID を保持するタグのプレフィックス。
	PreviousIDTagName = "@previous-id"
	// MaxTags はタグ配列の最大数 (API 上限)。
	MaxTags = 10
)

// AppendPreviousIDTagIfAbsent は tags に @previous-id=<currentID> タグを追加する。
// 既存の @previous-id=... タグは除去した上で 1 つだけ付与する。
// タグ上限 (MaxTags) を超える場合は tags をそのまま返す。
func AppendPreviousIDTagIfAbsent(tags []string, currentID int64) []string {
	if len(tags) > MaxTags {
		return tags
	}
	updated := make([]string, 0, len(tags)+1)
	for _, t := range tags {
		if !strings.HasPrefix(t, PreviousIDTagName) {
			updated = append(updated, t)
		}
	}
	updated = append(updated, fmt.Sprintf("%s=%d", PreviousIDTagName, currentID))
	sort.Strings(updated)
	return updated
}

// ---------- Server ----------

// ServerPlanChangeAPI は ChangeServerPlan が必要とする最小 interface。
type ServerPlanChangeAPI interface {
	Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error)
	Update(ctx context.Context, id int64, request *client.ServerUpdateRequestEnvelope) (*client.ServerUpdateResponseEnvelope, error)
	ChangePlan(ctx context.Context, id int64, request *client.ServerChangePlanRequestEnvelope) (*client.ServerChangePlanResponseEnvelope, error)
}

// ChangeServerPlan は @previous-id タグを付与してからプランを変更する。
// 返り値は ChangePlan 後の Server。
func ChangeServerPlan(ctx context.Context, op ServerPlanChangeAPI, id int64, planReq *client.ServerChangePlanRequestEnvelope) (*client.Server, error) {
	readResp, err := op.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	s := readResp.Server

	if len(s.Tags) < MaxTags {
		updatedTags := AppendPreviousIDTagIfAbsent(s.Tags, s.ID.Value)
		updateReq := &client.ServerUpdateRequestEnvelope{
			Server: client.ServerUpdateRequest{
				Name:            s.Name,
				Description:     s.Description,
				Tags:            updatedTags,
				Icon:            s.Icon,
				PrivateHost:     serverPrivateHostToRef(s.PrivateHost),
				InterfaceDriver: s.InterfaceDriver,
			},
		}
		if _, err := op.Update(ctx, id, updateReq); err != nil {
			return nil, err
		}
	}

	changeResp, err := op.ChangePlan(ctx, id, planReq)
	if err != nil {
		return nil, err
	}
	changed := changeResp.Server
	return &changed, nil
}

// serverPrivateHostToRef は Server.PrivateHost (OptNilServerPrivateHost) を
// Update リクエスト用の OptNilResourceRef に変換する。
func serverPrivateHostToRef(ph client.OptNilServerPrivateHost) client.OptNilResourceRef {
	var ref client.OptNilResourceRef
	if !ph.IsSet() {
		return ref
	}
	if ph.IsNull() {
		ref.SetToNull()
		return ref
	}
	ref.SetTo(client.ResourceRef{ID: ph.Value.ID})
	return ref
}

// ---------- Router (Internet) ----------

// RouterPlanChangeAPI は ChangeRouterPlan が必要とする最小 interface。
type RouterPlanChangeAPI interface {
	Read(ctx context.Context, id int64) (*client.InternetReadResponseEnvelope, error)
	Update(ctx context.Context, id int64, request *client.InternetUpdateRequestEnvelope) (*client.InternetUpdateResponseEnvelope, error)
	UpdateBandWidth(ctx context.Context, id int64, request *client.InternetUpdateBandWidthRequestEnvelope) (*client.InternetUpdateBandWidthResponseEnvelope, error)
}

// ChangeRouterPlan は @previous-id タグを付与してから帯域幅を変更する。
// 返り値は UpdateBandWidth 後の Internet。
func ChangeRouterPlan(ctx context.Context, op RouterPlanChangeAPI, id int64, bandWidthMbps int32) (*client.Internet, error) {
	readResp, err := op.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	r := readResp.Internet

	if len(r.Tags) < MaxTags {
		updatedTags := AppendPreviousIDTagIfAbsent(r.Tags, r.ID.Value)
		updateReq := &client.InternetUpdateRequestEnvelope{
			Internet: client.InternetUpdateRequest{
				Name:        r.Name,
				Description: r.Description,
				Tags:        updatedTags,
				Icon:        r.Icon,
			},
		}
		if _, err := op.Update(ctx, id, updateReq); err != nil {
			return nil, err
		}
	}

	bwReq := &client.InternetUpdateBandWidthRequestEnvelope{
		Internet: client.InternetUpdateBandWidthRequest{
			BandWidthMbps: client.NewOptInt32(bandWidthMbps),
		},
	}
	bwResp, err := op.UpdateBandWidth(ctx, id, bwReq)
	if err != nil {
		return nil, err
	}
	changed := bwResp.Internet
	return &changed, nil
}
