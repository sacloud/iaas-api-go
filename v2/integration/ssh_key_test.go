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

// 固定の公開鍵（テスト専用、対応する秘密鍵は保持しない）
const testSSHPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAEzCRoP3i4CwfvwoWNbAmX0T4fUA2CWzohbkXbHyE7x iaas-api-go-test"

func TestSSHKeyCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. Create - 公開鍵登録
	createReq := &client.SSHKeyCreateRequestEnvelope{
		SSHKey: client.SSHKeyCreateRequest{
			Name:        client.NewOptString("test-sshkey"),
			Description: "desc",
			PublicKey:   client.NewOptString(testSSHPublicKey),
		},
	}

	createResp, err := c.SSHKeyOpCreate(ctx, createReq, client.SSHKeyOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	sshKeyID := createResp.SSHKey.ID.Value
	sshKeyIDStr := fmt.Sprintf("%d", sshKeyID)
	t.Logf("Created SSH key ID: %d", sshKeyID)
	require.Equal(t, "test-sshkey", createResp.SSHKey.Name.Value)
	require.NotEmpty(t, createResp.SSHKey.Fingerprint.Value, "fingerprint must be returned on create")

	// 2. Read
	readResp, err := c.SSHKeyOpRead(ctx, client.SSHKeyOpReadParams{Zone: zone, ID: sshKeyIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-sshkey", readResp.SSHKey.Name.Value)
	require.Equal(t, sshKeyID, readResp.SSHKey.ID.Value)
	require.Equal(t, testSSHPublicKey, readResp.SSHKey.PublicKey.Value)

	// 3. Update - 名前・説明の更新（PublicKey は update では変更不可）
	updateResp, err := c.SSHKeyOpUpdate(ctx, &client.SSHKeyUpdateRequestEnvelope{
		SSHKey: client.SSHKeyUpdateRequest{
			Name:        client.NewOptString("test-sshkey-updated"),
			Description: "desc-updated",
		},
	}, client.SSHKeyOpUpdateParams{Zone: zone, ID: sshKeyIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-sshkey-updated", updateResp.SSHKey.Name.Value)
	require.Equal(t, "desc-updated", updateResp.SSHKey.Description)

	// 4. Find
	findResp, err := c.SSHKeyOpFind(ctx, client.SSHKeyOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.SSHKeys), 0)

	var found bool
	for _, k := range findResp.SSHKeys {
		if k.ID.Value == sshKeyID {
			found = true
			break
		}
	}
	require.True(t, found, "作成した SSH キーがリストに含まれていること")

	// 5. Delete
	_, err = c.SSHKeyOpDelete(ctx, client.SSHKeyOpDeleteParams{Zone: zone, ID: sshKeyIDStr})
	require.NoError(t, err)

	// 削除後は 404 になることを確認
	_, err = c.SSHKeyOpRead(ctx, client.SSHKeyOpReadParams{Zone: zone, ID: sshKeyIDStr})
	require.Error(t, err)
}
