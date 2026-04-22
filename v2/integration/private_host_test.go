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
	"fmt"
	"os"
	"testing"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

// PrivateHost はサンドボックス tk1v には Plan が無いので、v1 の test/private_host_op_test.go に
// 倣って本番ゾーン tk1a をハードコードする。さくら社員向けテストなので料金面は気にしない運用。
const privateHostTestZone = "tk1a"

func TestPrivateHostPlanFind(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	req := &client.PrivateHostPlanFindRequest{Count: 1}
	findResp, err := c.PrivateHostPlanOpFind(ctx, client.PrivateHostPlanOpFindParams{Zone: privateHostTestZone, Q: req.ToOptString()})
	require.NoError(t, err)
	require.Greater(t, len(findResp.PrivateHostPlans), 0, "PrivateHostPlan が 1 件以上返ること")
	require.LessOrEqual(t, len(findResp.PrivateHostPlans), 1, "Count=1 が反映されていること")

	plan := findResp.PrivateHostPlans[0]
	planIDStr := fmt.Sprintf("%d", plan.ID.Value)
	t.Logf("First plan: id=%s name=%s class=%s dedicated=%v", planIDStr, plan.Name.Value, plan.Class.Value, plan.Dedicated)
	require.NotZero(t, plan.ID.Value)
	require.NotEmpty(t, plan.Name.Value)
	require.NotEmpty(t, plan.Class.Value)
	require.Greater(t, plan.CPU.Value, int32(0))
	require.Greater(t, plan.MemoryMB.Value, int32(0))

	readResp, err := c.PrivateHostPlanOpRead(ctx, client.PrivateHostPlanOpReadParams{Zone: privateHostTestZone, ID: planIDStr})
	require.NoError(t, err)
	require.Equal(t, plan.ID.Value, readResp.PrivateHostPlan.ID.Value)
}

func TestPrivateHostCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	// PrivateHostPlan を Class=dynamic, Dedicated=false で検索して ID を得る
	planReq := &client.PrivateHostPlanFindRequest{
		Filter: client.PrivateHostPlanFindFilter{Class: "dynamic"},
	}
	planFindResp, err := c.PrivateHostPlanOpFind(ctx, client.PrivateHostPlanOpFindParams{Zone: privateHostTestZone, Q: planReq.ToOptString()})
	require.NoError(t, err)
	require.Greater(t, len(planFindResp.PrivateHostPlans), 0, "PrivateHostPlan (Class=dynamic) が見つかること")
	var planID int64
	for _, p := range planFindResp.PrivateHostPlans {
		if !p.Dedicated {
			planID = p.ID.Value
			break
		}
	}
	require.NotZero(t, planID, "非 Dedicated な PrivateHostPlan が見つかること")
	t.Logf("Using PrivateHostPlan ID: %d", planID)

	// 1. Create
	createReq := &client.PrivateHostCreateRequestEnvelope{
		PrivateHost: client.PrivateHostCreateRequest{
			Name:        client.NewOptString("test-private-host"),
			Description: "desc",
			Tags:        []string{"test", "integration"},
			Plan:        client.NewOptNilResourceRef(client.ResourceRef{ID: planID}),
		},
	}
	createResp, err := c.PrivateHostOpCreate(ctx, createReq, client.PrivateHostOpCreateParams{Zone: privateHostTestZone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	phID := createResp.PrivateHost.ID
	phIDStr := fmt.Sprintf("%d", phID)
	t.Logf("Created PrivateHost ID: %d", phID)
	require.Equal(t, "test-private-host", createResp.PrivateHost.Name.Value)

	// 2. Read
	readResp, err := c.PrivateHostOpRead(ctx, client.PrivateHostOpReadParams{Zone: privateHostTestZone, ID: phIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-private-host", readResp.PrivateHost.Name.Value)
	require.Equal(t, phID, readResp.PrivateHost.ID)

	// 3. Update
	updateResp, err := c.PrivateHostOpUpdate(ctx, &client.PrivateHostUpdateRequestEnvelope{
		PrivateHost: client.PrivateHostUpdateRequest{
			Name:        client.NewOptString("test-private-host-updated"),
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.PrivateHostOpUpdateParams{Zone: privateHostTestZone, ID: phIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-private-host-updated", updateResp.PrivateHost.Name.Value)

	// 4. Find
	findResp, err := c.PrivateHostOpFind(ctx, client.PrivateHostOpFindParams{Zone: privateHostTestZone})
	require.NoError(t, err)
	var found bool
	for _, ph := range findResp.PrivateHosts {
		if ph.ID == phID {
			found = true
			break
		}
	}
	require.True(t, found, "作成した PrivateHost がリストに含まれていること")

	// 5. Delete
	_, err = c.PrivateHostOpDelete(ctx, client.PrivateHostOpDeleteParams{Zone: privateHostTestZone, ID: phIDStr})
	require.NoError(t, err)

	// 削除後は 404
	_, err = c.PrivateHostOpRead(ctx, client.PrivateHostOpReadParams{Zone: privateHostTestZone, ID: phIDStr})
	require.Error(t, err)
}
