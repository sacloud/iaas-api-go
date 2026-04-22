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

func TestServerCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. Create - 最小構成のサーバ（1コア / 1GiB、共有セグメント接続、ディスクなし）。
	// ディスクを接続しないためサーバは自動起動しない（停止状態で作成される）。
	createReq := &client.ServerCreateRequestEnvelope{
		Server: client.ServerCreateRequest{
			ServerPlan: client.NewOptNilServerCreateRequestServerPlan(client.ServerCreateRequestServerPlan{
				CPU:      client.NewOptInt32(1),
				MemoryMB: client.NewOptInt32(1024),
			}),
			ConnectedSwitches: []client.ConnectedSwitch{
				{Scope: client.NewOptEScope("shared")},
			},
			InterfaceDriver: client.NewOptEInterfaceDriver("virtio"),
			Name:            client.NewOptString("test-server"),
			Description:     "desc",
			Tags:            []string{"test", "integration"},
		},
	}

	createResp, err := c.ServerOpCreate(ctx, createReq, client.ServerOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	serverID := createResp.Server.ID.Value
	serverIDStr := fmt.Sprintf("%d", serverID)
	t.Logf("Created server ID: %d", serverID)
	require.Equal(t, "test-server", createResp.Server.Name.Value)

	// 2. Read
	readResp, err := c.ServerOpRead(ctx, client.ServerOpReadParams{Zone: zone, ID: serverIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-server", readResp.Server.Name.Value)
	require.Equal(t, serverID, readResp.Server.ID.Value)

	// 3. Update
	updateResp, err := c.ServerOpUpdate(ctx, &client.ServerUpdateRequestEnvelope{
		Server: client.ServerUpdateRequest{
			Name:        client.NewOptString("test-server-updated"),
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.ServerOpUpdateParams{Zone: zone, ID: serverIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-server-updated", updateResp.Server.Name.Value)

	// 4. Find
	findResp, err := c.ServerOpFind(ctx, client.ServerOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.Servers), 0)

	var found bool
	for _, s := range findResp.Servers {
		if s.ID.Value == serverID {
			found = true
			break
		}
	}
	require.True(t, found, "作成したサーバがリストに含まれていること")

	// 5. Delete（ディスクを接続していないので WithDisk は空）
	_, err = c.ServerOpDelete(ctx, &client.ServerDeleteRequestEnvelope{
		WithDisk: []client.ID{},
	}, client.ServerOpDeleteParams{Zone: zone, ID: serverIDStr})
	require.NoError(t, err)

	// 削除後は 404 になることを確認
	_, err = c.ServerOpRead(ctx, client.ServerOpReadParams{Zone: zone, ID: serverIDStr})
	require.Error(t, err)
}
