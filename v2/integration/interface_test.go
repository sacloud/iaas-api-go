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

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

func TestInterfaceCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	// 1. 前提の Server を作成。Interface は Server に紐付く形でしか作れない。
	// v1 test/interface_op_test.go と同様、ConnectedSwitches は指定せずに作る
	// （NIC は後で Interface Create で追加する）。
	serverResp, err := c.ServerOpCreate(ctx, &client.ServerCreateRequestEnvelope{
		Server: client.ServerCreateRequest{
			ServerPlan: client.NewOptNilServerCreateRequestServerPlan(client.ServerCreateRequestServerPlan{
				CPU:      client.NewOptInt32(1),
				MemoryMB: client.NewOptInt32(1024),
			}),
			InterfaceDriver: client.NewOptEInterfaceDriver("virtio"),
			Name:            client.NewOptString("test-server-for-interface"),
			Description:     "desc",
			Tags:            []string{"test", "integration"},
		},
	})
	require.NoError(t, err)
	serverID := serverResp.Server.ID.Value
	defer func() {
		_, _ = c.ServerOpDelete(ctx, &client.ServerDeleteRequestEnvelope{}, client.ServerOpDeleteParams{ID: serverID})
	}()

	// 前提の Switch（後で ConnectToSwitch に使う）
	swResp, err := c.SwitchOpCreate(ctx, &client.SwitchCreateRequestEnvelope{
		Switch: client.SwitchCreateRequest{
			Name: client.NewOptString("switch-for-interface"),
			Tags: []string{"test", "integration"},
		},
	})
	require.NoError(t, err)
	switchID := swResp.Switch.ID.Value
	defer func() {
		// Switch は Interface から切断されてから削除する
		_, _ = c.SwitchOpDelete(ctx, client.SwitchOpDeleteParams{ID: switchID})
	}()

	// 2. Interface Create - Server に紐付ける
	createResp, err := c.InterfaceOpCreate(ctx, &client.InterfaceCreateRequestEnvelope{
		Interface: client.InterfaceCreateRequest{
			Server: client.NewOptNilResourceRef(client.ResourceRef{ID: serverID}),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	ifaceID := createResp.Interface.ID.Value
	t.Logf("Created Interface ID: %d (on server %d)", ifaceID, serverID)
	require.Equal(t, serverID, createResp.Interface.Server.Value.ID)

	// 3. Read
	readResp, err := c.InterfaceOpRead(ctx, client.InterfaceOpReadParams{ID: ifaceID})
	require.NoError(t, err)
	require.Equal(t, ifaceID, readResp.Interface.ID.Value)

	// 4. Update - UserIPAddress を設定
	updateResp, err := c.InterfaceOpUpdate(ctx, &client.InterfaceUpdateRequestEnvelope{
		Interface: client.InterfaceUpdateRequest{
			UserIPAddress: client.NewOptString("192.2.0.1"),
		},
	}, client.InterfaceOpUpdateParams{ID: ifaceID})
	require.NoError(t, err)
	require.Equal(t, "192.2.0.1", updateResp.Interface.UserIPAddress.Value)

	// 5. Find - リストに含まれることを確認
	findResp, err := c.InterfaceOpFind(ctx, client.InterfaceOpFindParams{})
	require.NoError(t, err)
	var found bool
	for _, ifv := range findResp.Interfaces {
		if ifv.ID.Value == ifaceID {
			found = true
			break
		}
	}
	require.True(t, found, "作成した Interface がリストに含まれていること")

	// 6. ConnectToSwitch
	_, err = c.InterfaceOpConnectToSwitch(ctx, client.InterfaceOpConnectToSwitchParams{ID: ifaceID, SwitchID: switchID})
	require.NoError(t, err)

	// 接続確認
	readResp2, err := c.InterfaceOpRead(ctx, client.InterfaceOpReadParams{ID: ifaceID})
	require.NoError(t, err)
	require.Equal(t, switchID, readResp2.Interface.Switch.Value.ID.Value, "Switch が接続されていること")

	// 7. DisconnectFromSwitch
	_, err = c.InterfaceOpDisconnectFromSwitch(ctx, client.InterfaceOpDisconnectFromSwitchParams{ID: ifaceID})
	require.NoError(t, err)

	// 8. ConnectToSharedSegment（共有セグメントへ）
	_, err = c.InterfaceOpConnectToSharedSegment(ctx, client.InterfaceOpConnectToSharedSegmentParams{ID: ifaceID})
	require.NoError(t, err)

	// 9. DisconnectFromSwitch（shared から外す）
	_, err = c.InterfaceOpDisconnectFromSwitch(ctx, client.InterfaceOpDisconnectFromSwitchParams{ID: ifaceID})
	require.NoError(t, err)

	// 10. Delete
	_, err = c.InterfaceOpDelete(ctx, client.InterfaceOpDeleteParams{ID: ifaceID})
	require.NoError(t, err)

	// 削除後は 404
	_, err = c.InterfaceOpRead(ctx, client.InterfaceOpReadParams{ID: ifaceID})
	require.Error(t, err)
}
