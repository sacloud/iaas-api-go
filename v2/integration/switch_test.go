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

func TestSwitchCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. Create - スイッチ作成
	// 注: UserSubnet は fieldmanifest allowlist で除外済み (downstream が未指定のため)。
	createReq := &client.SwitchCreateRequestEnvelope{
		Switch: client.SwitchCreateRequest{
			Name:        client.NewOptString("test-switch"),
			Description: "desc",
			Tags:        []string{"test", "integration"},
		},
	}

	createResp, err := c.SwitchOpCreate(ctx, createReq, client.SwitchOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	switchID := createResp.Switch.ID.Value
	t.Logf("Created switch ID: %d", switchID)
	require.Equal(t, "test-switch", createResp.Switch.Name.Value)

	// 2. Read - スイッチ取得
	readResp, err := c.SwitchOpRead(ctx, client.SwitchOpReadParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", switchID),
	})
	require.NoError(t, err)
	require.Equal(t, "test-switch", readResp.Switch.Name.Value)
	require.Equal(t, switchID, readResp.Switch.ID.Value)

	// 3. Update - スイッチ更新
	updateResp, err := c.SwitchOpUpdate(ctx, &client.SwitchUpdateRequestEnvelope{
		Switch: client.SwitchUpdateRequest{
			Name:        client.NewOptString("test-switch-updated"),
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.SwitchOpUpdateParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", switchID),
	})
	require.NoError(t, err)
	require.Equal(t, "test-switch-updated", updateResp.Switch.Name.Value)

	// 4. Find - スイッチ検索
	findResp, err := c.SwitchOpFind(ctx, client.SwitchOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.Switches), 0)

	var found bool
	for _, sw := range findResp.Switches {
		if sw.ID.Value == switchID {
			found = true
			break
		}
	}
	require.True(t, found, "作成したスイッチがリストに含まれていること")

	// 5. Delete - スイッチ削除
	_, err = c.SwitchOpDelete(ctx, client.SwitchOpDeleteParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", switchID),
	})
	require.NoError(t, err)

	// 削除後は 404 になることを確認
	_, err = c.SwitchOpRead(ctx, client.SwitchOpReadParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", switchID),
	})
	require.Error(t, err)
}
