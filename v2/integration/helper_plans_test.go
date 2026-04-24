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
	"strconv"
	"testing"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/iaas-api-go/v2/helper/plans"
	"github.com/stretchr/testify/require"
)

// TestHelperPlansChangeRouterPlan は Internet (ルータ+スイッチ) の帯域幅変更時に
// @previous-id タグが付与され、帯域幅更新が反映されることを確認する。
func TestHelperPlansChangeRouterPlan(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	createResp, err := c.InternetOpCreate(ctx, &client.InternetCreateRequestEnvelope{
		Internet: client.InternetCreateRequest{
			Name:           client.NewOptString("helper-plans-router"),
			Description:    "helper plans ACC",
			Tags:           []string{"test", "integration", "helper-plans"},
			NetworkMaskLen: client.NewOptInt32(28),
			BandWidthMbps:  client.NewOptInt32(100),
		},
	})
	require.NoError(t, err)
	internetID := createResp.Internet.ID.Value
	t.Logf("Created internet ID: %d", internetID)
	defer func() {
		_, _ = c.InternetOpDelete(ctx, client.InternetOpDeleteParams{ID: internetID})
	}()

	// プラン変更 100 → 250 Mbps
	op := iaas.NewInternetOp(c)
	updated, err := plans.ChangeRouterPlan(ctx, op, internetID, 250)
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.Equal(t, int32(250), updated.BandWidthMbps.Value)

	// タグに @previous-id が付いているか確認
	readResp, err := c.InternetOpRead(ctx, client.InternetOpReadParams{ID: updated.ID.Value})
	require.NoError(t, err)
	wantTag := "@previous-id=" + strconv.FormatInt(internetID, 10)
	found := false
	for _, tag := range readResp.Internet.Tags {
		if tag == wantTag {
			found = true
			break
		}
	}
	require.True(t, found, "expected %s in tags, got %v", wantTag, readResp.Internet.Tags)
}
