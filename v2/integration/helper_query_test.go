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
	"testing"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/helper/query"
	"github.com/stretchr/testify/require"
)

// TestHelperQueryFindArchiveByOSType は実 API で Ubuntu パブリックアーカイブを
// 検索できることを確認する。
func TestHelperQueryFindArchiveByOSType(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	archive, err := query.FindArchiveByOSType(ctx, iaas.NewArchiveOp(c), query.Ubuntu)
	require.NoError(t, err)
	require.NotNil(t, archive)
	t.Logf("Found Ubuntu archive ID=%d name=%s", archive.ID.Value, archive.Name.Value)
	require.Contains(t, archive.Name.Value, "Ubuntu")
}

// TestHelperQueryFindServerPlan は実 API で 1CPU/1GB プランを引けることを確認する。
func TestHelperQueryFindServerPlan(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	plan, err := query.FindServerPlan(ctx, iaas.NewServerPlanOp(c), &query.FindServerPlanRequest{
		CPU:      1,
		MemoryGB: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.Equal(t, int32(1), plan.CPU.Value)
	require.Equal(t, int32(1024), plan.MemoryMB.Value)
	t.Logf("Found server plan ID=%d name=%s", plan.ID.Value, plan.Name.Value)
}

// TestHelperQueryReadServerFallback は存在しない ID の Read が
// 404 → previous-id Find (結果0件) → ErrNoResults の流れになることを確認する。
func TestHelperQueryReadServerFallback(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	// 存在しない適当な ID
	_, err := query.ReadServer(ctx, iaas.NewServerOp(c), 999999999999)
	require.ErrorIs(t, err, query.ErrNoResults)
}
