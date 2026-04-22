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

// Account グループ: AuthStatus (Read) と License (CRUD)。
// AuthStatus の v2 envelope は `{is_ok, AuthStatus: {...}}` だが実 API はフラット（AuthStatus の
// フィールドが envelope 直下に並ぶ）で decode できない既知の差分があるため、ここでは扱わない。
// 詳細は AGENTS.md の「実装しないエンドポイント」表を参照。

// License Create は tk1v sandbox で `dont_create_in_sandbox` (403) を返すため tk1a 固定で走る。
// v1 の挙動を踏襲した上でさくら社員前提の運用に合わせる（project_test_acc_context.md 参照）。
const licenseTestZone = "tk1a"

func TestLicenseCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	c := newClientForZone(t, licenseTestZone)
	ctx := context.Background()

	// 1. LicenseInfo を Find して有効な LicenseInfo.ID を得る
	infoResp, err := c.LicenseInfoOpFind(ctx, client.LicenseInfoOpFindParams{})
	require.NoError(t, err)
	require.Greater(t, len(infoResp.LicenseInfo), 0)
	licenseInfoID := infoResp.LicenseInfo[0].ID.Value
	require.NotZero(t, licenseInfoID)
	t.Logf("Using LicenseInfo ID: %d (name=%s)", licenseInfoID, infoResp.LicenseInfo[0].Name.Value)

	// 2. License Create
	createResp, err := c.LicenseOpCreate(ctx, &client.LicenseCreateRequestEnvelope{
		License: client.LicenseCreateRequest{
			Name:        client.NewOptString("test-license"),
			LicenseInfo: client.NewOptNilResourceRef(client.ResourceRef{ID: licenseInfoID}),
		},
	})
	require.NoError(t, err)
	licenseID := createResp.License.ID.Value
	t.Logf("Created License ID: %d", licenseID)
	require.Equal(t, "test-license", createResp.License.Name.Value)

	// 3. Read
	readResp, err := c.LicenseOpRead(ctx, client.LicenseOpReadParams{ID: licenseID})
	require.NoError(t, err)
	require.Equal(t, licenseID, readResp.License.ID.Value)
	require.Equal(t, "test-license", readResp.License.Name.Value)

	// 4. Update (Name 変更)
	updateResp, err := c.LicenseOpUpdate(ctx, &client.LicenseUpdateRequestEnvelope{
		License: client.LicenseUpdateRequest{
			Name: client.NewOptString("test-license-updated"),
		},
	}, client.LicenseOpUpdateParams{ID: licenseID})
	require.NoError(t, err)
	require.Equal(t, "test-license-updated", updateResp.License.Name.Value)

	// 5. Find - リストに含まれることを確認
	findResp, err := c.LicenseOpFind(ctx, client.LicenseOpFindParams{})
	require.NoError(t, err)
	var found bool
	for _, l := range findResp.Licenses {
		if l.ID.Value == licenseID {
			found = true
			break
		}
	}
	require.True(t, found, "作成した License がリストに含まれていること")

	// 6. Delete
	_, err = c.LicenseOpDelete(ctx, client.LicenseOpDeleteParams{ID: licenseID})
	require.NoError(t, err)

	// 削除後は 404
	_, err = c.LicenseOpRead(ctx, client.LicenseOpReadParams{ID: licenseID})
	require.Error(t, err)
}
