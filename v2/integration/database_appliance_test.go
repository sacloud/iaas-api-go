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
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/go-faster/jx"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

func TestDatabaseApplianceCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. 前提の Switch を作成
	swResp, err := c.SwitchOpCreate(ctx, &client.SwitchCreateRequestEnvelope{
		Switch: client.SwitchCreateRequest{
			Name: client.NewOptString("switch-for-db"),
			Tags: []string{"test", "integration"},
		},
	})
	require.NoError(t, err)
	switchID := swResp.Switch.ID.Value
	switchIDStr := fmt.Sprintf("%d", switchID)
	defer func() {
		_, _ = c.SwitchOpDelete(ctx, client.SwitchOpDeleteParams{ID: switchIDStr})
	}()

	// 2. Appliance Create (Class=database, Plan=DB10GB / ID=10)
	// v1 と同じ最小構成:
	// - Remark.Plan.ID, Remark.Switch.ID, Remark.Servers[].IPAddress, Remark.Network.*, Remark.DBConf.Common
	// - Settings.DBConf.Common (ServicePort 等), Settings.Replication, Settings.MonitoringSuite
	const dbPlanID = int64(10)

	switchRaw, _ := json.Marshal(map[string]any{"ID": switchIDStr})

	// v1 test (test/database_op_test.go) と同じ Settings。
	settings := map[string]any{
		"DBConf": map[string]any{
			"Common": map[string]any{
				"ServicePort":     5432,
				"DefaultUser":     "exa.mple",
				"UserPassword":    "LibsacloudExamplePassword01",
				"ReplicaUser":     "replica",
				"ReplicaPassword": "replica-user-password",
			},
			"Replication": map[string]any{
				"Model": "Master-Slave",
			},
		},
		"MonitoringSuite": map[string]any{
			"Enabled": true,
		},
	}
	settingsRaw, _ := json.Marshal(settings)

	createReq := &client.ApplianceCreateRequestEnvelope{
		Appliance: client.ApplianceCreateRequest{
			Class: "database",
			// 実 API は Plan は Remark.Plan.ID に置く。top-level Plan は送らない。
			Remark: client.ApplianceCreateRequestRemark{
				Plan:   client.NewOptApplianceCreateRequestRemarkPlan(client.ApplianceCreateRequestRemarkPlan{ID: dbPlanID}),
				Switch: jx.Raw(switchRaw),
				Servers: []client.ApplianceCreateRequestRemarkServers{
					{IPAddress: "192.168.0.11"},
				},
				Network: client.NewOptApplianceCreateRequestRemarkNetwork(client.ApplianceCreateRequestRemarkNetwork{
					NetworkMaskLen: client.NewOptInt32(24),
					DefaultRoute:   client.NewOptString("192.168.0.1"),
				}),
				DBConf: client.NewOptApplianceCreateRequestRemarkDBConf(client.ApplianceCreateRequestRemarkDBConf{
					Common: client.DatabaseRemarkDBConfCommon{
						DatabaseName: client.NewOptString("MariaDB"),
						DefaultUser:  client.NewOptString("exa.mple"),
						UserPassword: client.NewOptString("LibsacloudExamplePassword01"),
					},
				}),
			},
			Settings:    jx.Raw(settingsRaw),
			Name:        "test-db",
			Description: "desc",
			Tags:        []string{"test", "integration"},
		},
	}

	createResp, err := c.ApplianceOpCreate(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	dbID := createResp.Appliance.ID.Value
	dbIDStr := fmt.Sprintf("%d", dbID)
	t.Logf("Created DB appliance ID: %d", dbID)
	require.Equal(t, "test-db", createResp.Appliance.Name.Value)
	require.Equal(t, "database", createResp.Appliance.Class.Value)

	// DB の up までは数分かかる。ここでは available 到達は待たず、すぐ shutdown/delete する
	// （mapconv 確認が目的なので round-trip が取れればよい）。
	// ただし作成直後だと shutdown 不可なことがあるので少しだけ待つ。
	waitApplianceAvailable(t, ctx, c, zone, dbIDStr)

	// 3. Read
	readResp, err := c.ApplianceOpRead(ctx, client.ApplianceOpReadParams{ID: dbIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-db", readResp.Appliance.Name.Value)
	require.Equal(t, "database", readResp.Appliance.Class.Value)

	// 4. Shutdown → Delete
	_, err = c.ApplianceOpShutdown(ctx, &client.ShutdownOption{Force: true}, client.ApplianceOpShutdownParams{ID: dbIDStr})
	require.NoError(t, err)
	waitApplianceShutdown(t, ctx, c, zone, dbIDStr)

	_, err = c.ApplianceOpDelete(ctx, client.ApplianceOpDeleteParams{ID: dbIDStr})
	require.NoError(t, err)
}
