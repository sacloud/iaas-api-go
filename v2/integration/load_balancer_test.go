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

func TestLoadBalancerApplianceCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. 前提の Switch
	swResp, err := c.SwitchOpCreate(ctx, &client.SwitchCreateRequestEnvelope{
		Switch: client.SwitchCreateRequest{
			Name: client.NewOptString("switch-for-lb"),
			Tags: []string{"test", "integration"},
		},
	})
	require.NoError(t, err)
	switchID := swResp.Switch.ID.Value
	switchIDStr := fmt.Sprintf("%d", switchID)
	defer func() {
		_, _ = c.SwitchOpDelete(ctx, client.SwitchOpDeleteParams{ID: switchIDStr})
	}()

	// 2. Appliance Create (Class=loadbalancer, Plan=Standard/ID=1)
	// LoadBalancer は HA 構成のため IP が 2 つ必要、VRID も必須。
	// Settings.LoadBalancer に VirtualIPAddress を最低 1 つ登録する。
	const lbPlanStandard = int64(1)

	switchRaw, _ := json.Marshal(map[string]any{"ID": switchIDStr})

	settings := map[string]any{
		"LoadBalancer": []map[string]any{
			{
				"VirtualIPAddress": "192.168.0.101",
				"Port":             "80",
				"DelayLoop":        "10",
				"SorryServer":      "192.168.0.2",
				"Description":      "vip1",
				"Servers": []map[string]any{
					{
						"IPAddress": "192.168.0.201",
						"Port":      "80",
						"Enabled":   "True",
						"HealthCheck": map[string]any{
							"Protocol": "http",
							"Path":     "/",
							"Status":   "200",
						},
					},
				},
			},
		},
	}
	settingsRaw, _ := json.Marshal(settings)

	createReq := &client.ApplianceCreateRequestEnvelope{
		Appliance: client.ApplianceCreateRequest{
			Class: "loadbalancer",
			Remark: client.ApplianceCreateRequestRemark{
				Plan:   client.NewOptApplianceCreateRequestRemarkPlan(client.ApplianceCreateRequestRemarkPlan{ID: lbPlanStandard}),
				Switch: jx.Raw(switchRaw),
				VRRP:   client.NewOptApplianceCreateRequestRemarkVRRP(client.ApplianceCreateRequestRemarkVRRP{VRID: 100}),
				Servers: []client.ApplianceCreateRequestRemarkServers{
					{IPAddress: "192.168.0.11"},
					{IPAddress: "192.168.0.12"},
				},
				Network: client.NewOptApplianceCreateRequestRemarkNetwork(client.ApplianceCreateRequestRemarkNetwork{
					NetworkMaskLen: client.NewOptInt32(24),
					DefaultRoute:   client.NewOptString("192.168.0.1"),
				}),
			},
			Settings:    jx.Raw(settingsRaw),
			Name:        "test-lb",
			Description: "desc",
			Tags:        []string{"test", "integration"},
		},
	}

	createResp, err := c.ApplianceOpCreate(ctx, createReq)
	require.NoError(t, err)
	require.NotNil(t, createResp)
	lbID := createResp.Appliance.ID.Value
	lbIDStr := fmt.Sprintf("%d", lbID)
	t.Logf("Created LoadBalancer appliance ID: %d", lbID)
	require.Equal(t, "test-lb", createResp.Appliance.Name.Value)
	require.Equal(t, "loadbalancer", createResp.Appliance.Class.Value)

	waitApplianceAvailable(t, ctx, c, zone, lbIDStr)

	// 3. Read
	readResp, err := c.ApplianceOpRead(ctx, client.ApplianceOpReadParams{ID: lbIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-lb", readResp.Appliance.Name.Value)
	require.Equal(t, "loadbalancer", readResp.Appliance.Class.Value)

	// 4. Update (Name / Description / Tags)
	updateResp, err := c.ApplianceOpUpdate(ctx, &client.ApplianceUpdateRequestEnvelope{
		Appliance: client.ApplianceUpdateRequest{
			Name:        "test-lb-updated",
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.ApplianceOpUpdateParams{ID: lbIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-lb-updated", updateResp.Appliance.Name.Value)

	// 5. Shutdown → Delete
	_, err = c.ApplianceOpShutdown(ctx, &client.ShutdownOption{Force: true}, client.ApplianceOpShutdownParams{ID: lbIDStr})
	require.NoError(t, err)
	waitApplianceShutdown(t, ctx, c, zone, lbIDStr)

	_, err = c.ApplianceOpDelete(ctx, client.ApplianceOpDeleteParams{ID: lbIDStr})
	require.NoError(t, err)
}
