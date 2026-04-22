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

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/go-faster/jx"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

func TestVPCRouterApplianceCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// VPCRouter Standard プラン(1) は shared switch 接続のシングル構成で、事前の Switch 作成は不要。
	// Standard plan は `Plan.ID` を TOP-LEVEL に置く（他のアプライアンスは Remark.Plan.ID だが
	// VPCRouter だけ PlanID() mapconv が `Plan.ID` であり v1 も top-level に送っている）。
	const vpcPlanStandard = int64(1)

	// Remark.Switch は {Scope: "shared"} のみ。Version は 2（v1 DSL のデフォルト）。
	switchRaw, _ := json.Marshal(map[string]any{"Scope": "shared"})

	createReq := &client.ApplianceCreateRequestEnvelope{
		Appliance: client.ApplianceCreateRequest{
			Class: "vpcrouter",
			Plan:  client.NewOptApplianceCreateRequestPlan(client.ApplianceCreateRequestPlan{ID: vpcPlanStandard}),
			Remark: client.ApplianceCreateRequestRemark{
				Switch: jx.Raw(switchRaw),
				Router: client.NewOptApplianceCreateRequestRemarkRouter(client.ApplianceCreateRequestRemarkRouter{VPCRouterVersion: 2}),
				// Standard plan は shared switch 経由で IP 自動割当、1 server 必須
				Servers: []client.ApplianceCreateRequestRemarkServers{{IPAddress: ""}},
			},
			Name:        "test-vpc-router",
			Description: "desc",
			Tags:        []string{"test", "integration"},
		},
	}

	createResp, err := c.ApplianceOpCreate(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	vpcID := createResp.Appliance.ID.Value
	vpcIDStr := fmt.Sprintf("%d", vpcID)
	t.Logf("Created VPCRouter appliance ID: %d", vpcID)
	require.Equal(t, "test-vpc-router", createResp.Appliance.Name.Value)
	require.Equal(t, "vpcrouter", createResp.Appliance.Class.Value)

	// VPCRouter は create 直後 Instance.Status=="down" のまま。明示的に Boot しない限り
	// 起動しない。本テストは CRUD round-trip の検証が目的なので down のまま Availability のみ待つ。
	waitApplianceAvailableOpt(t, ctx, c, zone, vpcIDStr, false)

	// Read
	readResp, err := c.ApplianceOpRead(ctx, client.ApplianceOpReadParams{ID: vpcIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-vpc-router", readResp.Appliance.Name.Value)
	require.Equal(t, "vpcrouter", readResp.Appliance.Class.Value)

	// Update (Name / Description / Tags)
	updateResp, err := c.ApplianceOpUpdate(ctx, &client.ApplianceUpdateRequestEnvelope{
		Appliance: client.ApplianceUpdateRequest{
			Name:        "test-vpc-router-updated",
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.ApplianceOpUpdateParams{ID: vpcIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-vpc-router-updated", updateResp.Appliance.Name.Value)

	// Delete（既に down なので shutdown 不要）
	_, err = c.ApplianceOpDelete(ctx, client.ApplianceOpDeleteParams{ID: vpcIDStr})
	require.NoError(t, err)
}
