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
	"github.com/sacloud/iaas-api-go/v2/helper/wait"
	"github.com/stretchr/testify/require"
)

// TestHelperWaitUntilDiskIsReady は Disk 作成後に
// wait.UntilDiskIsReady が "available" 到達を待機できることを確認する。
func TestHelperWaitUntilDiskIsReady(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	createReq := &client.DiskCreateRequestEnvelope{
		Disk: client.DiskCreateRequest{
			Plan:        client.NewOptNilResourceRef(client.ResourceRef{ID: diskPlanSSD}),
			SizeMB:      client.NewOptInt32(20 * 1024),
			Name:        client.NewOptString("helper-wait-disk"),
			Description: "helper wait ACC",
			Tags:        []string{"test", "integration", "helper-wait"},
		},
	}
	createResp, err := c.DiskOpCreate(ctx, createReq)
	require.NoError(t, err)
	diskID := createResp.Disk.ID.Value
	t.Logf("Created disk ID: %d", diskID)
	defer func() {
		_, _ = c.DiskOpDelete(ctx, client.DiskOpDeleteParams{ID: diskID})
	}()

	op := iaas.NewDiskOp(c)
	disk, err := wait.UntilDiskIsReady(ctx, op, diskID)
	require.NoError(t, err)
	require.NotNil(t, disk)
	require.Equal(t, "available", string(disk.Availability.Value))
}

// TestHelperWaitUntilServerIsDown は Server 作成→Boot→Shutdown を経て
// wait.UntilServerIsDown が Down 到達を待機できることを確認する。
// Archive 依存を避けるため、Disk を作らずに CDROM ブートで確認する。
func TestHelperWaitUntilServerIsDown(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	// 最小構成のサーバー (1コア/1GB、CDROM/ディスク無し)。起動不能のまま放置し、
	// UntilServerIsDown が Availability=available + Instance=down を観測できることを確認する。
	createReq := &client.ServerCreateRequestEnvelope{
		Server: client.ServerCreateRequest{
			Name:        client.NewOptString("helper-wait-server"),
			Description: "helper wait ACC",
			Tags:        []string{"test", "integration", "helper-wait"},
			ServerPlan: client.NewOptNilServerCreateRequestServerPlan(client.ServerCreateRequestServerPlan{
				Generation: client.NewOptEPlanGeneration(client.EPlanGeneration(200)),
				CPU:        client.NewOptInt32(1),
				MemoryMB:   client.NewOptInt32(1024),
			}),
			InterfaceDriver: client.NewOptEInterfaceDriver(client.EInterfaceDriver("virtio")),
		},
	}
	createResp, err := c.ServerOpCreate(ctx, createReq)
	require.NoError(t, err)
	serverID := createResp.Server.ID.Value
	t.Logf("Created server ID: %d", serverID)
	defer func() {
		_, _ = c.ServerOpDelete(ctx, &client.ServerDeleteRequestEnvelope{}, client.ServerOpDeleteParams{ID: serverID})
	}()

	op := iaas.NewServerOp(c)
	server, err := wait.UntilServerIsDown(ctx, op, serverID)
	require.NoError(t, err)
	require.NotNil(t, server)
}
