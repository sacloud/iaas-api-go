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
	"strings"
	"testing"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

// Bridge は tk1v sandbox で Create が `dont_create_in_sandbox` (403) を返すため tk1a 固定で走る。
// Switch-Bridge 接続系の相互テストも同じゾーンで実施する。
const bridgeTestZone = "tk1a"

func isLimitCountError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "limit_count_")
}

func TestBridgeCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()
	zone := bridgeTestZone

	// 1. Create
	createResp, err := c.BridgeOpCreate(ctx, &client.BridgeCreateRequestEnvelope{
		Bridge: client.BridgeCreateRequest{
			Name:        client.NewOptNilString("test-bridge"),
			Description: "desc",
		},
	}, client.BridgeOpCreateParams{Zone: zone})
	require.NoError(t, err)
	bridgeID := createResp.Bridge.ID.Value
	bridgeIDStr := fmt.Sprintf("%d", bridgeID)
	t.Logf("Created Bridge ID: %d", bridgeID)
	require.Equal(t, "test-bridge", createResp.Bridge.Name.Value)

	// 2. Read
	readResp, err := c.BridgeOpRead(ctx, client.BridgeOpReadParams{Zone: zone, ID: bridgeIDStr})
	require.NoError(t, err)
	require.Equal(t, bridgeID, readResp.Bridge.ID.Value)
	require.Equal(t, "test-bridge", readResp.Bridge.Name.Value)

	// 3. Update
	updateResp, err := c.BridgeOpUpdate(ctx, &client.BridgeUpdateRequestEnvelope{
		Bridge: client.BridgeUpdateRequest{
			Name:        client.NewOptNilString("test-bridge-updated"),
			Description: "desc-updated",
		},
	}, client.BridgeOpUpdateParams{Zone: zone, ID: bridgeIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-bridge-updated", updateResp.Bridge.Name.Value)

	// 4. Find
	findResp, err := c.BridgeOpFind(ctx, &client.BridgeFindRequestEnvelope{}, client.BridgeOpFindParams{Zone: zone})
	require.NoError(t, err)
	var found bool
	for _, b := range findResp.Bridges {
		if b.ID.Value == bridgeID {
			found = true
			break
		}
	}
	require.True(t, found, "作成した Bridge がリストに含まれていること")

	// 5. Delete
	_, err = c.BridgeOpDelete(ctx, client.BridgeOpDeleteParams{Zone: zone, ID: bridgeIDStr})
	require.NoError(t, err)

	// 削除後は 404
	_, err = c.BridgeOpRead(ctx, client.BridgeOpReadParams{Zone: zone, ID: bridgeIDStr})
	require.Error(t, err)
}

// TestSwitchBridgeConnect は Switch↔Bridge の connect/disconnect ラウンドトリップを検証する。
// 保留中（Bridge 側のテストが揃うまで）だった Switch の ConnectToBridge / DisconnectFromBridge を
// Bridge CRUD が整ったタイミングでここで対応する。
func TestSwitchBridgeConnect(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()
	zone := bridgeTestZone

	// Bridge を作成
	bridgeResp, err := c.BridgeOpCreate(ctx, &client.BridgeCreateRequestEnvelope{
		Bridge: client.BridgeCreateRequest{
			Name: client.NewOptNilString("test-bridge-for-switch"),
		},
	}, client.BridgeOpCreateParams{Zone: zone})
	require.NoError(t, err)
	bridgeID := bridgeResp.Bridge.ID.Value
	bridgeIDStr := fmt.Sprintf("%d", bridgeID)
	defer func() {
		_, _ = c.BridgeOpDelete(ctx, client.BridgeOpDeleteParams{Zone: zone, ID: bridgeIDStr})
	}()

	// Switch を作成。tk1a は `switch: 1` per zone quota なので既存の switch があると 409 になる。
	// その場合は envelope 部分までは動作確認できているので skip してラップアップする。
	swResp, err := c.SwitchOpCreate(ctx, &client.SwitchCreateRequestEnvelope{
		Switch: client.SwitchCreateRequest{
			Name: client.NewOptNilString("test-switch-for-bridge"),
			Tags: []string{"test", "integration"},
		},
	}, client.SwitchOpCreateParams{Zone: zone})
	if err != nil {
		// 他のリソース（v1 テスト残骸など）に食われていた場合は skip
		if isLimitCountError(err) {
			t.Skipf("Switch quota exhausted in %s: %v", zone, err)
		}
		t.Fatalf("unexpected SwitchOpCreate error: %v", err)
	}
	switchID := swResp.Switch.ID.Value
	switchIDStr := fmt.Sprintf("%d", switchID)
	defer func() {
		_, _ = c.SwitchOpDelete(ctx, client.SwitchOpDeleteParams{Zone: zone, ID: switchIDStr})
	}()

	// ConnectToBridge
	_, err = c.SwitchOpConnectToBridge(ctx, client.SwitchOpConnectToBridgeParams{
		Zone: zone, ID: switchIDStr, BridgeID: bridgeIDStr,
	})
	require.NoError(t, err)

	// Switch 側 Bridge 参照確認
	readSwitch, err := c.SwitchOpRead(ctx, client.SwitchOpReadParams{Zone: zone, ID: switchIDStr})
	require.NoError(t, err)
	require.Equal(t, bridgeID, readSwitch.Switch.Bridge.Value.ID, "Switch.Bridge.ID が接続先の Bridge を指すこと")

	// Bridge 側 SwitchInZone 参照確認
	readBridge, err := c.BridgeOpRead(ctx, client.BridgeOpReadParams{Zone: zone, ID: bridgeIDStr})
	require.NoError(t, err)
	require.Equal(t, switchID, readBridge.Bridge.SwitchInZone.Value.ID.Value, "Bridge.SwitchInZone.ID が接続された Switch を指すこと")

	// DisconnectFromBridge
	_, err = c.SwitchOpDisconnectFromBridge(ctx, client.SwitchOpDisconnectFromBridgeParams{Zone: zone, ID: switchIDStr})
	require.NoError(t, err)
}
