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

func TestMobileGatewayApplianceCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}
	// MobileGateway は法人契約必須のため、v1 の test/mobile_gateway_op_test.go と同じく
	// SIM 契約を持つ環境でしかテストできない。v1 は PreCheckEnvsFunc で SIM ICCID / PASSCODE の
	// 存在を確認して skip していたので v2 も同じ env gate に合わせる。
	if os.Getenv("SAKURACLOUD_SIM_ICCID") == "" || os.Getenv("SAKURACLOUD_SIM_PASSCODE") == "" {
		t.Skip("SAKURACLOUD_SIM_ICCID / SAKURACLOUD_SIM_PASSCODE required (法人契約の SIM 情報)")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// MobileGateway は shared switch 経由で契約されるため、事前の Switch 作成は不要。
	// Remark.Switch.Scope = "shared" を指定することで shared セグメントに接続される。
	// Plan は Standard(2) 固定（v1 DSL ConstField 準拠）。
	const mgwPlan = int64(2)

	switchRaw, _ := json.Marshal(map[string]any{"Scope": "shared"})

	// InternetConnection / InterDeviceCommunication は StringFlag (v1) で送信値は
	// "True" / "False"（大文字先頭）が実 API 仕様。
	settings := map[string]any{
		"MobileGateway": map[string]any{
			"InternetConnection": map[string]any{
				"Enabled": "False",
			},
			"InterDeviceCommunication": map[string]any{
				"Enabled": "False",
			},
		},
	}
	settingsRaw, _ := json.Marshal(settings)

	createReq := &client.ApplianceCreateRequestEnvelope{
		Appliance: client.ApplianceCreateRequest{
			Class: "mobilegateway",
			Remark: client.ApplianceCreateRequestRemark{
				Plan:    client.NewOptApplianceCreateRequestRemarkPlan(client.ApplianceCreateRequestRemarkPlan{ID: mgwPlan}),
				Switch:  jx.Raw(switchRaw),
				// MG plan=2 は Servers 配列に 1 要素必須（shared switch から自動割当のため IPAddress は空で OK）
				Servers: []client.ApplianceCreateRequestRemarkServers{{IPAddress: ""}},
			},
			Settings:    jx.Raw(settingsRaw),
			Name:        "test-mgw",
			Description: "desc",
			Tags:        []string{"test", "integration"},
		},
	}

	createResp, err := c.ApplianceOpCreate(ctx, createReq, client.ApplianceOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	mgwID := createResp.Appliance.ID.Value
	mgwIDStr := fmt.Sprintf("%d", mgwID)
	t.Logf("Created MobileGateway appliance ID: %d", mgwID)
	require.Equal(t, "test-mgw", createResp.Appliance.Name.Value)
	require.Equal(t, "mobilegateway", createResp.Appliance.Class.Value)

	waitApplianceAvailable(t, ctx, c, zone, mgwIDStr)

	// Read
	readResp, err := c.ApplianceOpRead(ctx, client.ApplianceOpReadParams{Zone: zone, ID: mgwIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-mgw", readResp.Appliance.Name.Value)
	require.Equal(t, "mobilegateway", readResp.Appliance.Class.Value)

	// Update (Name / Description / Tags)
	updateResp, err := c.ApplianceOpUpdate(ctx, &client.ApplianceUpdateRequestEnvelope{
		Appliance: client.ApplianceUpdateRequest{
			Name:        "test-mgw-updated",
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.ApplianceOpUpdateParams{Zone: zone, ID: mgwIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-mgw-updated", updateResp.Appliance.Name.Value)

	// Shutdown → Delete
	_, err = c.ApplianceOpShutdown(ctx, &client.ShutdownOption{Force: true}, client.ApplianceOpShutdownParams{Zone: zone, ID: mgwIDStr})
	require.NoError(t, err)
	waitApplianceShutdown(t, ctx, c, zone, mgwIDStr)

	_, err = c.ApplianceOpDelete(ctx, client.ApplianceOpDeleteParams{Zone: zone, ID: mgwIDStr})
	require.NoError(t, err)
}
