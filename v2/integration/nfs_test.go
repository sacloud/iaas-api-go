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
	"time"

	"github.com/go-faster/jx"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

// findNFSPlanID は `sys-nfs` という Note のコンテンツ（JSON）を解析し、
// 指定の DiskPlan ("HDD"=2 / "SSD"=4) とサイズ (100, 500, 1024 ...) に対応する NFSPlan ID を返す。
// v1 helper/query/nfs_plan.go と同じロジック。
func findNFSPlanID(t *testing.T, ctx context.Context, c *client.Client, zone string, diskPlanID int64, sizeGB int) int64 {
	t.Helper()

	// v2 の Filter syntax は v1 search.Filter と異なるため、全 Note を舐めて Name=="sys-nfs" & Class=="json" を探す。
	resp, err := c.NoteOpFind(ctx, client.NoteOpFindParams{Zone: zone})
	require.NoError(t, err)

	var content string
	for _, n := range resp.Notes {
		if n.Name.Value == "sys-nfs" && n.Class.Value == "json" {
			content = n.Content.Value
			break
		}
	}
	require.NotEmpty(t, content, "sys-nfs note must exist")

	type nfsPlanRow struct {
		PlanID       int64  `json:"planId"`
		Size         int    `json:"size"`
		Availability string `json:"availability"`
	}
	var envelope struct {
		Plans struct {
			HDD []nfsPlanRow `json:"HDD"`
			SSD []nfsPlanRow `json:"SSD"`
		} `json:"plans"`
	}
	require.NoError(t, json.Unmarshal([]byte(content), &envelope))

	pickFrom := func(plans []nfsPlanRow) int64 {
		for _, p := range plans {
			if p.Availability == "available" && p.Size == sizeGB {
				return p.PlanID
			}
		}
		return 0
	}

	var planID int64
	switch diskPlanID {
	case 2:
		planID = pickFrom(envelope.Plans.HDD)
	case 4:
		planID = pickFrom(envelope.Plans.SSD)
	}
	require.NotZero(t, planID, "matching NFS plan not found")
	return planID
}

// waitApplianceAvailable は Appliance が利用可能かつ Instance が up まで達したのを待つ。
// create 直後は Availability が "migrating"、その後 "available" になっても Instance.Status は
// しばらく "" のまま（VM 起動中）。shutdown/update を走らせる前に Instance が "up" に
// 達していないと 423 Locked で弾かれる。
// VPCRouter のように create 後 Instance が自動起動しない appliance は useStatusUp=false で呼ぶ。
func waitApplianceAvailable(t *testing.T, ctx context.Context, c *client.Client, zone, id string) {
	t.Helper()
	waitApplianceAvailableOpt(t, ctx, c, zone, id, true)
}

func waitApplianceAvailableOpt(t *testing.T, ctx context.Context, c *client.Client, zone, id string, requireUp bool) {
	t.Helper()
	deadline := time.Now().Add(15 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := c.ApplianceOpRead(ctx, client.ApplianceOpReadParams{Zone: zone, ID: id})
		if err == nil && resp.Appliance.Availability.Value == "available" {
			if !requireUp {
				return
			}
			if resp.Appliance.Instance.Value.Status.Value == "up" {
				return
			}
		}
		time.Sleep(10 * time.Second)
	}
	t.Fatalf("appliance %s did not become available(+up=%v) within timeout", id, requireUp)
}

// waitApplianceShutdown は Appliance の Instance.Status が "down" になるまで待つ。
func waitApplianceShutdown(t *testing.T, ctx context.Context, c *client.Client, zone, id string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := c.ApplianceOpRead(ctx, client.ApplianceOpReadParams{Zone: zone, ID: id})
		if err == nil && resp.Appliance.Instance.Value.Status.Value == "down" {
			return
		}
		time.Sleep(5 * time.Second)
	}
	t.Fatalf("appliance %s did not shut down within timeout", id)
}

func TestNFSApplianceCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. 前提の Switch を作成
	swResp, err := c.SwitchOpCreate(ctx, &client.SwitchCreateRequestEnvelope{
		Switch: client.SwitchCreateRequest{
			Name: client.NewOptNilString("switch-for-nfs"),
			Tags: []string{"test", "integration"},
		},
	}, client.SwitchOpCreateParams{Zone: zone})
	require.NoError(t, err)
	switchID := swResp.Switch.ID.Value
	switchIDStr := fmt.Sprintf("%d", switchID)
	defer func() {
		_, _ = c.SwitchOpDelete(ctx, client.SwitchOpDeleteParams{Zone: zone, ID: switchIDStr})
	}()

	// 2. NFSPlan ID を sys-nfs Note から検索。SSD/100GB を使う。
	const diskPlanSSD = int64(4)
	planID := findNFSPlanID(t, ctx, c, zone, diskPlanSSD, 100)
	t.Logf("Using NFSPlan ID: %d (SSD/100GB)", planID)

	// 3. Appliance Create (Class=nfs)
	//    実 API は `{"Appliance": {"Class":"nfs", "Remark":{"Plan":{"ID":...}, "Switch":{"ID":"..."}, "Servers":[{"IPAddress":"..."}], "Network":{...}}, "Name":...}}` を要求する。
	//    v2 fat model は Remark.Switch が jx.Raw なので手で JSON を作る。
	switchRaw, _ := json.Marshal(map[string]any{"ID": switchIDStr})
	createReq := &client.ApplianceCreateRequestEnvelope{
		Appliance: client.ApplianceCreateRequest{
			Class: "nfs",
			Remark: client.ApplianceCreateRequestRemark{
				Plan:   client.NewOptApplianceCreateRequestRemarkPlan(client.ApplianceCreateRequestRemarkPlan{ID: planID}),
				Switch: jx.Raw(switchRaw),
				Servers: []client.ApplianceCreateRequestRemarkServers{
					{IPAddress: "192.168.0.11"},
				},
				Network: client.NewOptApplianceCreateRequestRemarkNetwork(client.ApplianceCreateRequestRemarkNetwork{
					NetworkMaskLen: client.NewOptInt32(24),
					DefaultRoute:   client.NewOptString("192.168.0.1"),
				}),
			},
			// Settings / Icon / Plan / Disk は NFS では不要（optional）
			Name:        "test-nfs",
			Description: "desc",
			Tags:        []string{"test", "integration"},
		},
	}
	createResp, err := c.ApplianceOpCreate(ctx, createReq, client.ApplianceOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	nfsID := createResp.Appliance.ID.Value
	nfsIDStr := fmt.Sprintf("%d", nfsID)
	t.Logf("Created NFS appliance ID: %d", nfsID)
	require.Equal(t, "test-nfs", createResp.Appliance.Name.Value)
	require.Equal(t, "nfs", createResp.Appliance.Class.Value)

	waitApplianceAvailable(t, ctx, c, zone, nfsIDStr)

	// 4. Read
	readResp, err := c.ApplianceOpRead(ctx, client.ApplianceOpReadParams{Zone: zone, ID: nfsIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-nfs", readResp.Appliance.Name.Value)
	require.Equal(t, "nfs", readResp.Appliance.Class.Value)

	// 5. Update (Name / Description / Tags)
	updateResp, err := c.ApplianceOpUpdate(ctx, &client.ApplianceUpdateRequestEnvelope{
		Appliance: client.ApplianceUpdateRequest{
			Name:        "test-nfs-updated",
			Description: "desc-updated",
			Tags:        []string{"test", "integration", "updated"},
		},
	}, client.ApplianceOpUpdateParams{Zone: zone, ID: nfsIDStr})
	require.NoError(t, err)
	require.Equal(t, "test-nfs-updated", updateResp.Appliance.Name.Value)

	// 6. Find
	findResp, err := c.ApplianceOpFind(ctx, client.ApplianceOpFindParams{Zone: zone})
	require.NoError(t, err)
	var found bool
	for _, app := range findResp.Appliances {
		if app.ID.Value == nfsID {
			found = true
			break
		}
	}
	require.True(t, found, "作成した NFS がリストに含まれていること")

	// 7. Shutdown (force)
	_, err = c.ApplianceOpShutdown(ctx, &client.ShutdownOption{Force: true}, client.ApplianceOpShutdownParams{Zone: zone, ID: nfsIDStr})
	require.NoError(t, err)
	waitApplianceShutdown(t, ctx, c, zone, nfsIDStr)

	// 8. Delete
	_, err = c.ApplianceOpDelete(ctx, client.ApplianceOpDeleteParams{Zone: zone, ID: nfsIDStr})
	require.NoError(t, err)

	// 削除後は 404
	_, err = c.ApplianceOpRead(ctx, client.ApplianceOpReadParams{Zone: zone, ID: nfsIDStr})
	require.Error(t, err)
}
