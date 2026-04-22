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
	"time"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

// DiskPlans.SSD の ID（types.DiskPlans.SSD = 4）
const diskPlanSSD int64 = 4

// waitDiskAvailable は Disk が migrating → available に遷移するまでポーリングする。
// create 直後は migrating 状態のため、update/delete する前に available を待つ。
func waitDiskAvailable(t *testing.T, ctx context.Context, c *client.Client, zone, id string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := c.DiskOpRead(ctx, client.DiskOpReadParams{Zone: zone, ID: id})
		require.NoError(t, err)
		if resp.Disk.Availability.Value == "available" {
			return
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("disk %s did not become available within timeout", id)
}

func TestDiskCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. Create - ブランクディスク（OS なし、20GiB SSD）を作成
	// KMSKey / DistantFrom は optional のため省略（暗号化なし・distant from 指定なしでディスク作成）
	createReq := &client.DiskCreateRequestEnvelope{
		Disk: client.DiskCreateRequest{
			Plan:        client.NewOptNilResourceRef(client.ResourceRef{ID: diskPlanSSD}),
			SizeMB:      client.NewOptInt32(20 * 1024), // 20 GiB
			Name:        client.NewOptString("test-disk"),
			Description: "desc",
			Tags:        []string{"test", "integration"},
		},
	}

	createResp, err := c.DiskOpCreate(ctx, createReq, client.DiskOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	diskID := createResp.Disk.ID.Value
	diskIDStr := fmt.Sprintf("%d", diskID)
	t.Logf("Created disk ID: %d", diskID)
	require.Equal(t, "test-disk", createResp.Disk.Name.Value)

	// 作成直後は migrating 状態なので available を待つ
	waitDiskAvailable(t, ctx, c, zone, diskIDStr)

	// 2. Read - ディスク取得
	readResp, err := c.DiskOpRead(ctx, client.DiskOpReadParams{Zone: zone, ID: diskIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-disk", readResp.Disk.Name.Value)
	require.Equal(t, diskID, readResp.Disk.ID.Value)

	// 3. Update - 名前・タグ更新
	updateResp, err := c.DiskOpUpdate(ctx, &client.DiskUpdateRequestEnvelope{
		Disk: client.DiskUpdateRequest{
			Name:        client.NewOptString("test-disk-updated"),
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.DiskOpUpdateParams{Zone: zone, ID: diskIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-disk-updated", updateResp.Disk.Name.Value)

	// 4. Find - リストに含まれることを確認
	findResp, err := c.DiskOpFind(ctx, client.DiskOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.Disks), 0)

	var found bool
	for _, d := range findResp.Disks {
		if d.ID.Value == diskID {
			found = true
			break
		}
	}
	require.True(t, found, "作成したディスクがリストに含まれていること")

	// 5. Delete
	_, err = c.DiskOpDelete(ctx, client.DiskOpDeleteParams{Zone: zone, ID: diskIDStr})
	require.NoError(t, err)

	// 削除後は 404 になることを確認
	_, err = c.DiskOpRead(ctx, client.DiskOpReadParams{Zone: zone, ID: diskIDStr})
	require.Error(t, err)
}
