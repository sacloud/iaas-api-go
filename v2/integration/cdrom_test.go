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

// waitCDROMAvailable は CDROM が uploading → available に遷移するのを待つ。
// create 直後は FTP アップロード待ち（uploading）状態なので、update/delete する前に
// CloseFTP を呼んで available に落とす必要がある。
func waitCDROMAvailable(t *testing.T, ctx context.Context, c *client.Client, zone, id string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := c.CDROMOpRead(ctx, client.CDROMOpReadParams{Zone: zone, ID: id})
		if err == nil && resp.CDROM.Availability.Value == "available" {
			return
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("cdrom %s did not become available within timeout", id)
}

func TestCDROMCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. Create - 5 GiB のブランク ISO を作成
	// create 時点で FTP セッションが開き、レスポンスに FTPServer（HostName/User/Password）が含まれる。
	createReq := &client.CDROMCreateRequestEnvelope{
		CDROM: client.CDROMCreateRequest{
			SizeMB:      client.NewOptInt32(5 * 1024),
			Name:        client.NewOptString("test-cdrom"),
			Description: "desc",
			Tags:        []string{"test", "integration"},
		},
	}

	createResp, err := c.CDROMOpCreate(ctx, createReq, client.CDROMOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	cdromID := createResp.CDROM.ID.Value
	cdromIDStr := fmt.Sprintf("%d", cdromID)
	t.Logf("Created CDROM ID: %d", cdromID)
	require.Equal(t, "test-cdrom", createResp.CDROM.Name.Value)
	require.NotEmpty(t, createResp.FTPServer.HostName.Value, "FTPServer.HostName must be set on create")
	require.NotEmpty(t, createResp.FTPServer.User.Value)
	require.NotEmpty(t, createResp.FTPServer.Password.Value)

	// create 後は FTP 共有が開いた状態。実データ転送は行わないので閉じる。
	if _, err := c.CDROMOpCloseFTP(ctx, client.CDROMOpCloseFTPParams{Zone: zone, ID: cdromIDStr}); err != nil {
		t.Logf("close FTP (ignorable if already closed): %v", err)
	}
	waitCDROMAvailable(t, ctx, c, zone, cdromIDStr)

	// 2. Read
	readResp, err := c.CDROMOpRead(ctx, client.CDROMOpReadParams{Zone: zone, ID: cdromIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-cdrom", readResp.CDROM.Name.Value)
	require.Equal(t, cdromID, readResp.CDROM.ID.Value)

	// 3. Update - 名前・タグ更新
	updateResp, err := c.CDROMOpUpdate(ctx, &client.CDROMUpdateRequestEnvelope{
		CDROM: client.CDROMUpdateRequest{
			Name:        client.NewOptString("test-cdrom-updated"),
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.CDROMOpUpdateParams{Zone: zone, ID: cdromIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-cdrom-updated", updateResp.CDROM.Name.Value)

	// 4. Find
	findResp, err := c.CDROMOpFind(ctx, client.CDROMOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.CDROMs), 0)

	var found bool
	for _, cd := range findResp.CDROMs {
		if cd.ID.Value == cdromID {
			found = true
			break
		}
	}
	require.True(t, found, "作成した CDROM がリストに含まれていること")

	// 5. Delete
	// 注: 明示的な OpenFTP / CloseFTP の再呼び出しはテストに含めない。
	// tk1v サンドボックスではブランク ISO に対し CloseFTP 後の OpenFTP が
	// `ftp_is_already_close` (409) を返すため、envelope の decode は create 時に
	// FTPServer を受け取った時点 + create 直後の CloseFTP で検証済みとする。
	_, err = c.CDROMOpDelete(ctx, client.CDROMOpDeleteParams{Zone: zone, ID: cdromIDStr})
	require.NoError(t, err)

	// 削除後は 404
	_, err = c.CDROMOpRead(ctx, client.CDROMOpReadParams{Zone: zone, ID: cdromIDStr})
	require.Error(t, err)
}
