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

// Region / Zone は read-only なリソース。v1 DSL は Find + Read のみ実装している。
// どちらも global 扱いだが API path は `/{zone}/api/...` 形式なので zone パラメータは必要。

func TestRegionFindRead(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	findResp, err := c.RegionOpFind(ctx, &client.RegionFindRequestEnvelope{}, client.RegionOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.Regions), 0, "Region が 1 件以上返ること")

	first := findResp.Regions[0]
	t.Logf("First region: id=%d name=%s", first.ID.Value, first.Name.Value)
	require.NotZero(t, first.ID.Value)
	require.NotEmpty(t, first.Name.Value)

	readResp, err := c.RegionOpRead(ctx, client.RegionOpReadParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", first.ID.Value),
	})
	require.NoError(t, err)
	require.Equal(t, first.ID.Value, readResp.Region.ID.Value)
	require.Equal(t, first.Name.Value, readResp.Region.Name.Value)
}

func TestZoneFindRead(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	findResp, err := c.ZoneOpFind(ctx, &client.ZoneFindRequestEnvelope{}, client.ZoneOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.Zones), 0, "Zone が 1 件以上返ること")

	// 指定した zone (tk1v 等) が結果に含まれることを確認
	var target client.Zone
	for _, z := range findResp.Zones {
		if z.Name.Value == zone {
			target = z
			break
		}
	}
	require.Equal(t, zone, target.Name.Value, "指定ゾーン %s が Find 結果に含まれること", zone)
	require.NotZero(t, target.ID.Value)
	t.Logf("Current zone: id=%d name=%s description=%s", target.ID.Value, target.Name.Value, target.Description.Value)

	readResp, err := c.ZoneOpRead(ctx, client.ZoneOpReadParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", target.ID.Value),
	})
	require.NoError(t, err)
	require.Equal(t, target.ID.Value, readResp.Zone.ID.Value)
	require.Equal(t, zone, readResp.Zone.Name.Value)
	// Region の参照先 ID が入っていること
	require.True(t, readResp.Zone.Region.Set, "Zone.Region が含まれること")
}
