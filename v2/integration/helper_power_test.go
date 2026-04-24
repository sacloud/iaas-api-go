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

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/iaas-api-go/v2/helper/power"
	"github.com/sacloud/iaas-api-go/v2/helper/wait"
	"github.com/stretchr/testify/require"
)

// TestHelperPowerBootShutdownServer は Disk を作って Server に接続し、
// BootServer -> ShutdownServer(force) が実 API で動作することを確認する。
//
// 実 VM 起動には公開アーカイブが必要なので、archive_test.go と同じ OS アーカイブを使う。
func TestHelperPowerBootShutdownServer(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	// 1. 公開アーカイブを探す (Scope=shared + Name=Ubuntu)
	reqArchive := &client.ArchiveFindRequest{
		Count:  1,
		Filter: client.ArchiveFindFilter{Scope: "shared", Name: "Ubuntu"},
	}
	archiveResp, err := c.ArchiveOpFind(ctx, client.ArchiveOpFindParams{Q: reqArchive.ToOptString()})
	require.NoError(t, err)
	require.Greater(t, len(archiveResp.Archives), 0, "Ubuntu shared archive が見つからない")
	sourceArchiveID := archiveResp.Archives[0].ID.Value
	t.Logf("Using source archive ID: %d (%s)", sourceArchiveID, archiveResp.Archives[0].Name.Value)

	// 2. Disk 作成 (アーカイブから)
	diskResp, err := c.DiskOpCreate(ctx, &client.DiskCreateRequestEnvelope{
		Disk: client.DiskCreateRequest{
			Plan:               client.NewOptNilResourceRef(client.ResourceRef{ID: diskPlanSSD}),
			SizeMB:             client.NewOptInt32(20 * 1024),
			SourceArchive:      client.NewOptNilResourceRef(client.ResourceRef{ID: sourceArchiveID}),
			Name:               client.NewOptString("helper-power-disk"),
			Description:        "helper power ACC",
			Tags:               []string{"test", "integration", "helper-power"},
		},
	})
	require.NoError(t, err)
	diskID := diskResp.Disk.ID.Value
	t.Logf("Created disk ID: %d", diskID)
	defer func() {
		_, _ = c.DiskOpDelete(ctx, client.DiskOpDeleteParams{ID: diskID})
	}()

	// disk が available になるまで待つ
	_, err = wait.UntilDiskIsReady(ctx, iaas.NewDiskOp(c), diskID)
	require.NoError(t, err)

	// 3. Server 作成
	srvResp, err := c.ServerOpCreate(ctx, &client.ServerCreateRequestEnvelope{
		Server: client.ServerCreateRequest{
			ServerPlan: client.NewOptNilServerCreateRequestServerPlan(client.ServerCreateRequestServerPlan{
				CPU:      client.NewOptInt32(1),
				MemoryMB: client.NewOptInt32(1024),
			}),
			ConnectedSwitches: []client.ConnectedSwitch{
				{Scope: client.NewOptEScope("shared")},
			},
			InterfaceDriver: client.NewOptEInterfaceDriver("virtio"),
			Name:            client.NewOptString("helper-power-server"),
			Description:     "helper power ACC",
			Tags:            []string{"test", "integration", "helper-power"},
		},
	})
	require.NoError(t, err)
	serverID := srvResp.Server.ID.Value
	t.Logf("Created server ID: %d", serverID)
	defer func() {
		// Disk を接続したまま削除
		_, _ = c.ServerOpDelete(ctx, &client.ServerDeleteRequestEnvelope{}, client.ServerOpDeleteParams{ID: serverID})
	}()

	// 4. Disk を Server に接続
	_, err = c.DiskOpConnectToServer(ctx, client.DiskOpConnectToServerParams{ID: diskID, ServerID: serverID})
	require.NoError(t, err)

	// 5. BootServer
	op := iaas.NewServerOp(c)
	require.NoError(t, power.BootServer(ctx, op, serverID))
	t.Logf("Server %d booted", serverID)

	// 状態確認: available + up
	readResp, err := c.ServerOpRead(ctx, client.ServerOpReadParams{ID: serverID})
	require.NoError(t, err)
	require.Equal(t, "up", string(readResp.Server.Instance.Value.Status.Value))

	// 6. ShutdownServer (force)
	require.NoError(t, power.ShutdownServer(ctx, op, serverID, true))
	t.Logf("Server %d shut down", serverID)

	// 状態確認: available + down
	readResp2, err := c.ServerOpRead(ctx, client.ServerOpReadParams{ID: serverID})
	require.NoError(t, err)
	require.Equal(t, "down", string(readResp2.Server.Instance.Value.Status.Value))

	// 7. Disk 切断
	_, _ = c.DiskOpDisconnectFromServer(ctx, client.DiskOpDisconnectFromServerParams{ID: diskID})
}
