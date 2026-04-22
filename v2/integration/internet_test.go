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

// waitInternetSwitchReady は ルータ+スイッチ作成後に Switch 側のサブネット割当が完了するのを待つ。
// 作成直後は Switch そのものがまだ 404 を返すことがあるので、エラーは握りつぶしてリトライする。
func waitInternetSwitchReady(t *testing.T, ctx context.Context, c *client.Client, zone string, switchID int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := c.SwitchOpRead(ctx, client.SwitchOpReadParams{ID: switchID})
		if err == nil {
			if len(resp.Switch.Subnets) > 0 && resp.Switch.Subnets[0].IPAddresses.Value.Min != "" {
				return
			}
		}
		time.Sleep(3 * time.Second)
	}
	t.Fatalf("switch %d did not get subnet assigned within timeout", switchID)
}

func TestInternetCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. Create - 最小構成のルータ+スイッチ（/28, 100Mbps）
	createReq := &client.InternetCreateRequestEnvelope{
		Internet: client.InternetCreateRequest{
			Name:           client.NewOptString("test-internet"),
			Description:    "desc",
			Tags:           []string{"test", "integration"},
			NetworkMaskLen: client.NewOptInt32(28),
			BandWidthMbps:  client.NewOptInt32(100),
		},
	}

	createResp, err := c.InternetOpCreate(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	internetID := createResp.Internet.ID.Value
	t.Logf("Created internet ID: %d", internetID)
	require.Equal(t, "test-internet", createResp.Internet.Name.Value)
	require.Equal(t, int32(28), createResp.Internet.NetworkMaskLen.Value)
	require.Equal(t, int32(100), createResp.Internet.BandWidthMbps.Value)

	// 作成直後はサブネット割当がまだ完了していないため待機
	switchID := createResp.Internet.Switch.Value.ID.Value
	waitInternetSwitchReady(t, ctx, c, zone, switchID)

	// 2. Read
	readResp, err := c.InternetOpRead(ctx, client.InternetOpReadParams{ID: internetID})
	require.NoError(t, err)
	require.Equal(t, "test-internet", readResp.Internet.Name.Value)
	require.Equal(t, internetID, readResp.Internet.ID.Value)

	// 3. Update - 名前・タグ・説明の更新
	updateResp, err := c.InternetOpUpdate(ctx, &client.InternetUpdateRequestEnvelope{
		Internet: client.InternetUpdateRequest{
			Name:        client.NewOptString("test-internet-updated"),
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.InternetOpUpdateParams{ID: internetID})
	require.NoError(t, err)
	require.Equal(t, "test-internet-updated", updateResp.Internet.Name.Value)

	// 4. Find
	findResp, err := c.InternetOpFind(ctx, client.InternetOpFindParams{})
	require.NoError(t, err)
	require.Greater(t, len(findResp.Internet), 0)
	var found bool
	for _, ii := range findResp.Internet {
		if ii.ID.Value == internetID {
			found = true
			break
		}
	}
	require.True(t, found, "作成したルータがリストに含まれていること")

	// 5. EnableIPv6
	ipv6Resp, err := c.InternetOpEnableIPv6(ctx, client.InternetOpEnableIPv6Params{ID: internetID})
	require.NoError(t, err)
	ipv6NetID := ipv6Resp.IPv6Net.ID.Value
	require.NotZero(t, ipv6NetID)
	t.Logf("Enabled IPv6 net ID: %d", ipv6NetID)

	// 6. DisableIPv6
	_, err = c.InternetOpDisableIPv6(ctx, client.InternetOpDisableIPv6Params{ID: internetID, Ipv6netID: ipv6NetID})
	require.NoError(t, err)

	// 7. Delete
	_, err = c.InternetOpDelete(ctx, client.InternetOpDeleteParams{ID: internetID})
	require.NoError(t, err)

	// 削除後は 404
	_, err = c.InternetOpRead(ctx, client.InternetOpReadParams{ID: internetID})
	require.Error(t, err)
}
