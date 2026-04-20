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

func waitArchiveAvailable(t *testing.T, ctx context.Context, c *client.Client, zone, id string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := c.ArchiveOpRead(ctx, client.ArchiveOpReadParams{Zone: zone, ID: id})
		require.NoError(t, err)
		if resp.Archive.Availability.Value == "available" {
			return
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("archive %s did not become available within timeout", id)
}

func TestArchiveCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. Create - 20GiB の空アーカイブを作成（SizeMB 指定）
	createReq := &client.ArchiveCreateRequestEnvelope{
		Archive: client.ArchiveCreateRequest{
			Name:        client.NewOptNilString("test-archive"),
			Description: "desc",
			Tags:        []string{"test", "integration"},
			SizeMB:      client.NewOptNilInt32(20 * 1024),
		},
	}

	createResp, err := c.ArchiveOpCreate(ctx, createReq, client.ArchiveOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	archiveID := createResp.Archive.ID.Value
	archiveIDStr := fmt.Sprintf("%d", archiveID)
	t.Logf("Created archive ID: %d", archiveID)
	require.Equal(t, "test-archive", createResp.Archive.Name.Value)

	// SizeMB 指定で作成するとサーバ側で FTP アップロード用セッションが開くことがある。
	// 本テストでは実データ転送は行わないのでセッションを閉じる（既に閉じている場合はエラーは無視）。
	if _, err := c.ArchiveOpCloseFTP(ctx, client.ArchiveOpCloseFTPParams{Zone: zone, ID: archiveIDStr}); err != nil {
		t.Logf("close FTP (ignorable if already closed): %v", err)
	}

	// 作成直後は migrating 状態なので available を待つ
	waitArchiveAvailable(t, ctx, c, zone, archiveIDStr)

	// 2. Read
	readResp, err := c.ArchiveOpRead(ctx, client.ArchiveOpReadParams{Zone: zone, ID: archiveIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-archive", readResp.Archive.Name.Value)
	require.Equal(t, archiveID, readResp.Archive.ID.Value)

	// 3. Update
	updateResp, err := c.ArchiveOpUpdate(ctx, &client.ArchiveUpdateRequestEnvelope{
		Archive: client.ArchiveUpdateRequest{
			Name:        client.NewOptNilString("test-archive-updated"),
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.ArchiveOpUpdateParams{Zone: zone, ID: archiveIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-archive-updated", updateResp.Archive.Name.Value)

	// 4. Find
	findResp, err := c.ArchiveOpFind(ctx, &client.ArchiveFindRequestEnvelope{}, client.ArchiveOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.Archives), 0)

	var found bool
	for _, a := range findResp.Archives {
		if a.ID.Value == archiveID {
			found = true
			break
		}
	}
	require.True(t, found, "作成したアーカイブがリストに含まれていること")

	// 5. Delete
	_, err = c.ArchiveOpDelete(ctx, client.ArchiveOpDeleteParams{Zone: zone, ID: archiveIDStr})
	require.NoError(t, err)

	// 削除後は 404 になることを確認
	_, err = c.ArchiveOpRead(ctx, client.ArchiveOpReadParams{Zone: zone, ID: archiveIDStr})
	require.Error(t, err)
}
