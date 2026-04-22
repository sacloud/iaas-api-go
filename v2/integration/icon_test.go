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

func TestIconCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	// 1. Create - アイコン作成
	// 1x1 pixel transparent PNG (base64エンコード済み)
	base64Image := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

	createReq := &client.IconCreateRequestEnvelope{
		Icon: client.IconCreateRequest{
			Name:  client.NewOptString("test-icon"),
			Tags:  []string{"test", "integration"},
			Image: client.NewOptString(base64Image),
		},
	}

	createResp, err := c.IconOpCreate(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	iconID := createResp.Icon.ID
	t.Logf("Created icon ID: %d", iconID)
	require.Equal(t, "test-icon", createResp.Icon.Name.Value)

	// 2. Read - アイコン取得
	readParams := client.IconOpReadParams{ID: iconID}

	readResp, err := c.IconOpRead(ctx, readParams)
	require.NoError(t, err)
	require.NotNil(t, readResp)
	require.Equal(t, "test-icon", readResp.Icon.Name.Value)
	require.Equal(t, iconID, readResp.Icon.ID)

	// 3. Update - アイコン更新
	updateReq := &client.IconUpdateRequestEnvelope{
		Icon: client.IconUpdateRequest{
			Name: client.NewOptString("test-icon-updated"),
			Tags: []string{"test", "integration", "updated"},
		},
	}
	updateParams := client.IconOpUpdateParams{ID: iconID}

	updateResp, err := c.IconOpUpdate(ctx, updateReq, updateParams)
	require.NoError(t, err)
	require.NotNil(t, updateResp)
	require.Equal(t, "test-icon-updated", updateResp.Icon.Name.Value)

	// 4. Find - アイコン検索
	findParams := client.IconOpFindParams{}

	findResp, err := c.IconOpFind(ctx, findParams)
	require.NoError(t, err)
	require.NotNil(t, findResp)
	require.Greater(t, len(findResp.Icons), 0)

	// 作成したアイコンが含まれていることを確認
	var found bool
	for _, icon := range findResp.Icons {
		if icon.ID == iconID {
			found = true
			break
		}
	}
	require.True(t, found, "作成したアイコンがリストに含まれていること")

	// 5. Delete - アイコン削除
	deleteParams := client.IconOpDeleteParams{ID: iconID}

	_, err = c.IconOpDelete(ctx, deleteParams) //nolint:errcheck
	require.NoError(t, err)

	// 削除後の取得でエラーになることを確認（404 Not Found）
	_, err = c.IconOpRead(ctx, readParams)
	require.Error(t, err)
}
