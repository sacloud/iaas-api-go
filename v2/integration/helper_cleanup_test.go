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

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/iaas-api-go/v2/helper/cleanup"
	"github.com/sacloud/iaas-api-go/v2/helper/query"
	"github.com/sacloud/iaas-api-go/v2/helper/wait"
	"github.com/stretchr/testify/require"
)

// TestHelperCleanupDeleteDisk は Disk を作成後に cleanup.DeleteDisk が
// 参照チェック → 削除 の流れで動作することを確認する。
func TestHelperCleanupDeleteDisk(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	// 空の Disk を作成
	createResp, err := c.DiskOpCreate(ctx, &client.DiskCreateRequestEnvelope{
		Disk: client.DiskCreateRequest{
			Plan:        client.NewOptNilResourceRef(client.ResourceRef{ID: diskPlanSSD}),
			SizeMB:      client.NewOptInt32(20 * 1024),
			Name:        client.NewOptString("helper-cleanup-disk"),
			Description: "helper cleanup ACC",
			Tags:        []string{"test", "integration", "helper-cleanup"},
		},
	})
	require.NoError(t, err)
	diskID := createResp.Disk.ID.Value
	t.Logf("Created disk ID: %d", diskID)

	diskOp := iaas.NewDiskOp(c)
	_, err = wait.UntilDiskIsReady(ctx, diskOp, diskID)
	require.NoError(t, err)

	// cleanup.DeleteDisk は参照チェック + 削除。ここでは参照がないため即削除される。
	r := query.NewReferenceFinder(c)
	err = cleanup.DeleteDisk(ctx, diskOp, r, diskID, query.CheckReferencedOption{
		Timeout: 30 * time.Second,
		Tick:    2 * time.Second,
	})
	require.NoError(t, err)

	// 削除後は 404 になる
	_, err = c.DiskOpRead(ctx, client.DiskOpReadParams{ID: diskID})
	require.Error(t, err)
}
