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
	"os"
	"testing"
	"time"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

func waitArchiveAvailable(t *testing.T, ctx context.Context, c *client.Client, zone string, id int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := c.ArchiveOpRead(ctx, client.ArchiveOpReadParams{ID: id})
		require.NoError(t, err)
		if resp.Archive.Availability.Value == "available" {
			return
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("archive %d did not become available within timeout", id)
}

// TestArchiveFindWithQuery は `?q={json}` 書き換え + FindRequest/FindFilter の動作を
// 読み取り専用で確認する。tk1v にはさくらクラウド提供の shared archive が大量にある
// ため、Count + Filter(Scope="shared", Name="CentOS") を使って q= パラメータが効いていることを検証する。
func TestArchiveFindWithQuery(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	// 1. フィルタ無し + Count=3 → 3 件以下返る
	reqCount := &client.ArchiveFindRequest{Count: 3}
	respCount, err := c.ArchiveOpFind(ctx, client.ArchiveOpFindParams{Q: reqCount.ToOptString()})
	require.NoError(t, err)
	require.LessOrEqual(t, len(respCount.Archives), 3, "Count=3 の結果は 3 件以下であること")
	t.Logf("Count=3 returned %d archives (Total=%d)", len(respCount.Archives), respCount.Total)

	// 2. Scope="shared" フィルタ → shared archive が返る
	// 注: Archive response 側の Scope フィールドは fieldmanifest allowlist で除外済みなので、
	// 返却件数のみで filter が効いていることを確認する。
	reqScope := &client.ArchiveFindRequest{
		Count:  5,
		Filter: client.ArchiveFindFilter{Scope: "shared"},
	}
	respScope, err := c.ArchiveOpFind(ctx, client.ArchiveOpFindParams{Q: reqScope.ToOptString()})
	require.NoError(t, err)
	require.Greater(t, len(respScope.Archives), 0, "shared archive が 1 件以上返ること")
	t.Logf("Scope=shared returned %d archives", len(respScope.Archives))

	// 3. Name="CentOS" 部分一致 → 全件 Name に "CentOS" を含む（大文字小文字区別しない実装あり）
	reqName := &client.ArchiveFindRequest{
		Count:  5,
		Filter: client.ArchiveFindFilter{Name: "CentOS"},
	}
	respName, err := c.ArchiveOpFind(ctx, client.ArchiveOpFindParams{Q: reqName.ToOptString()})
	require.NoError(t, err)
	// さくらクラウドが提供する CentOS archive は tk1v に少なくとも 1 つある
	require.Greater(t, len(respName.Archives), 0, "Name=CentOS にマッチする archive があること")
	for _, a := range respName.Archives {
		require.Contains(t, a.Name.Value, "CentOS", "Name=CentOS 部分一致が効いていること")
	}
	t.Logf("Name=CentOS returned %d archives (first=%s)", len(respName.Archives), respName.Archives[0].Name.Value)
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
			Name:        client.NewOptString("test-archive"),
			Description: "desc",
			Tags:        []string{"test", "integration"},
			SizeMB:      client.NewOptInt32(20 * 1024),
		},
	}

	createResp, err := c.ArchiveOpCreate(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	archiveID := createResp.Archive.ID.Value
	t.Logf("Created archive ID: %d", archiveID)
	require.Equal(t, "test-archive", createResp.Archive.Name.Value)

	// SizeMB 指定で作成するとサーバ側で FTP アップロード用セッションが開くことがある。
	// 本テストでは実データ転送は行わないのでセッションを閉じる（既に閉じている場合はエラーは無視）。
	if _, err := c.ArchiveOpCloseFTP(ctx, client.ArchiveOpCloseFTPParams{ID: archiveID}); err != nil {
		t.Logf("close FTP (ignorable if already closed): %v", err)
	}

	// 作成直後は migrating 状態なので available を待つ
	waitArchiveAvailable(t, ctx, c, zone, archiveID)

	// 2. Read
	readResp, err := c.ArchiveOpRead(ctx, client.ArchiveOpReadParams{ID: archiveID})
	require.NoError(t, err)
	require.Equal(t, "test-archive", readResp.Archive.Name.Value)
	require.Equal(t, archiveID, readResp.Archive.ID.Value)

	// 3. Update
	updateResp, err := c.ArchiveOpUpdate(ctx, &client.ArchiveUpdateRequestEnvelope{
		Archive: client.ArchiveUpdateRequest{
			Name:        client.NewOptString("test-archive-updated"),
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.ArchiveOpUpdateParams{ID: archiveID})
	require.NoError(t, err)
	require.Equal(t, "test-archive-updated", updateResp.Archive.Name.Value)

	// 4. Find
	findResp, err := c.ArchiveOpFind(ctx, client.ArchiveOpFindParams{})
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
	_, err = c.ArchiveOpDelete(ctx, client.ArchiveOpDeleteParams{ID: archiveID})
	require.NoError(t, err)

	// 削除後は 404 になることを確認
	_, err = c.ArchiveOpRead(ctx, client.ArchiveOpReadParams{ID: archiveID})
	require.Error(t, err)
}
