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

// `/product/*` のプラン情報 API（read-only）。v1 DSL は全て Find + Read のみ実装。
// PrivateHostPlan は private_host_test.go で検証済みなのでここでは省略する。

func TestDiskPlanFindRead(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.DiskPlanOpFind(ctx, client.DiskPlanOpFindParams{})
	require.NoError(t, err)
	require.Greater(t, len(findResp.DiskPlans), 0, "DiskPlan が 1 件以上返ること")

	first := findResp.DiskPlans[0]
	t.Logf("First disk plan: id=%d name=%s class=%s", first.ID.Value, first.Name.Value, first.StorageClass.Value)
	require.NotZero(t, first.ID.Value)
	require.NotEmpty(t, first.Name.Value)

	readResp, err := c.DiskPlanOpRead(ctx, client.DiskPlanOpReadParams{ID: first.ID.Value})
	require.NoError(t, err)
	require.Equal(t, first.ID.Value, readResp.DiskPlan.ID.Value)
}

func TestInternetPlanFindRead(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.InternetPlanOpFind(ctx, client.InternetPlanOpFindParams{})
	require.NoError(t, err)
	require.Greater(t, len(findResp.InternetPlans), 0, "InternetPlan が 1 件以上返ること")

	first := findResp.InternetPlans[0]
	t.Logf("First internet plan: id=%d name=%s bandwidth=%d", first.ID.Value, first.Name.Value, first.BandWidthMbps.Value)
	require.NotZero(t, first.ID.Value)
	require.Greater(t, first.BandWidthMbps.Value, int32(0))

	readResp, err := c.InternetPlanOpRead(ctx, client.InternetPlanOpReadParams{ID: first.ID.Value})
	require.NoError(t, err)
	require.Equal(t, first.ID.Value, readResp.InternetPlan.ID.Value)
}

func TestServerPlanFindRead(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.ServerPlanOpFind(ctx, client.ServerPlanOpFindParams{})
	require.NoError(t, err)
	require.Greater(t, len(findResp.ServerPlans), 0, "ServerPlan が 1 件以上返ること")

	first := findResp.ServerPlans[0]
	t.Logf("First server plan: id=%d name=%s cpu=%d memMB=%d", first.ID.Value, first.Name.Value, first.CPU.Value, first.MemoryMB.Value)
	require.NotZero(t, first.ID.Value)
	require.Greater(t, first.CPU.Value, int32(0))
	require.Greater(t, first.MemoryMB.Value, int32(0))

	readResp, err := c.ServerPlanOpRead(ctx, client.ServerPlanOpReadParams{ID: first.ID.Value})
	require.NoError(t, err)
	require.Equal(t, first.ID.Value, readResp.ServerPlan.ID.Value)
}

// NOTE: `/public/price` の ServiceClass は v2 ジェネレータでは未対応。
// 理由は AGENTS.md の「実装しないエンドポイント」表を参照。

func TestLicenseInfoFindRead(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.LicenseInfoOpFind(ctx, client.LicenseInfoOpFindParams{})
	require.NoError(t, err)
	require.Greater(t, len(findResp.LicenseInfo), 0, "LicenseInfo が 1 件以上返ること")

	first := findResp.LicenseInfo[0]
	t.Logf("First license info: id=%d name=%s", first.ID.Value, first.Name.Value)
	require.NotZero(t, first.ID.Value)
	require.NotEmpty(t, first.Name.Value)

	readResp, err := c.LicenseInfoOpRead(ctx, client.LicenseInfoOpReadParams{ID: first.ID.Value})
	require.NoError(t, err)
	require.Equal(t, first.ID.Value, readResp.LicenseInfo.ID.Value)
}
