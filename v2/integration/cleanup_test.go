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
	"strings"
	"testing"
	"time"

	"github.com/sacloud/iaas-api-go/v2/client"
)

// TestCleanupInternet は "test" タグが付いた Internet リソースを一括削除する。
// TEST_ACC_CLEANUP=1 が設定されたときだけ動作する。
func TestCleanupInternet(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.InternetOpFind(ctx, client.InternetOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, ii := range findResp.Internet {
		if !hasTestTag(ii.Tags) && !strings.HasPrefix(ii.Name.Value, "test-internet") {
			continue
		}
		idStr := fmt.Sprintf("%d", ii.ID.Value)
		t.Logf("Deleting internet %s (name=%s)", idStr, ii.Name.Value)
		// Internet 配下の IPv6Net を先に外す（そうしないと delete が 409 になる）
		if ii.Switch.Set && !ii.Switch.Null {
			for _, ipv6 := range ii.Switch.Value.IPv6Nets {
				ipv6IDStr := fmt.Sprintf("%d", ipv6.ID.Value)
				if _, err := c.InternetOpDisableIPv6(ctx, client.InternetOpDisableIPv6Params{ID:        idStr,
					Ipv6netID: ipv6IDStr,
				}); err != nil {
					t.Logf("disable ipv6 %s on internet %s failed: %v", ipv6IDStr, idStr, err)
				}
			}
		}
		if _, err := c.InternetOpDelete(ctx, client.InternetOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete internet %s failed: %v", idStr, err)
		}
	}
}

// TestCleanupSwitchTK1a は tk1a に取り残された test switch を削除する。
// Bridge 連携テスト（TestSwitchBridgeConnect）が tk1a で動くため。
func TestCleanupSwitchTK1a(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClientForZone(t, bridgeTestZone)
	ctx := context.Background()

	findResp, err := c.SwitchOpFind(ctx, client.SwitchOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, sw := range findResp.Switches {
		if !hasTestTag(sw.Tags) && !strings.HasPrefix(sw.Name.Value, "test-switch") && !strings.HasPrefix(sw.Name.Value, "switch-for-") {
			continue
		}
		idStr := fmt.Sprintf("%d", sw.ID.Value)
		t.Logf("Deleting switch %s (name=%s)", idStr, sw.Name.Value)
		if _, err := c.SwitchOpDelete(ctx, client.SwitchOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete switch %s failed: %v", idStr, err)
		}
	}
}

// TestCleanupBridge は test-bridge* の Bridge を tk1a 固定で一括削除する。
// Bridge 自体は Tags を持たないので Name prefix で判定する。
func TestCleanupBridge(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClientForZone(t, bridgeTestZone)
	ctx := context.Background()

	findResp, err := c.BridgeOpFind(ctx, client.BridgeOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, b := range findResp.Bridges {
		if !strings.HasPrefix(b.Name.Value, "test-bridge") {
			continue
		}
		idStr := fmt.Sprintf("%d", b.ID.Value)
		t.Logf("Deleting bridge %s (name=%s)", idStr, b.Name.Value)
		// 接続中の Switch があれば先に disconnect する必要があるが、このテストでは defer で済ませているので不要
		if _, err := c.BridgeOpDelete(ctx, client.BridgeOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete bridge %s failed: %v", idStr, err)
		}
	}
}

// TestCleanupAppliance は "test" タグが付いた Appliance（NFS / DB / LB / VPC Router 等）を一括削除する。
// shutdown が必要な状態なら force shutdown してから削除する。
func TestCleanupAppliance(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.ApplianceOpFind(ctx, client.ApplianceOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, app := range findResp.Appliances {
		if !hasTestTag(app.Tags) && !strings.HasPrefix(app.Name.Value, "test-") {
			continue
		}
		idStr := fmt.Sprintf("%d", app.ID.Value)
		t.Logf("Deleting appliance %s (name=%s class=%s status=%s)", idStr, app.Name.Value, app.Class.Value, app.Instance.Value.Status.Value)
		if app.Instance.Value.Status.Value == "up" {
			if _, err := c.ApplianceOpShutdown(ctx, &client.ShutdownOption{Force: true}, client.ApplianceOpShutdownParams{ID: idStr}); err != nil {
				t.Logf("force shutdown %s failed: %v", idStr, err)
			}
			// wait a bit for status to flip
			time.Sleep(10 * time.Second)
		}
		if _, err := c.ApplianceOpDelete(ctx, client.ApplianceOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete appliance %s failed: %v", idStr, err)
		}
	}
}

// TestCleanupPrivateHost は tk1a に取り残された "test" タグの PrivateHost を一括削除する。
// PrivateHost は sandbox では Plan が無いためテストは tk1a 固定で走る（private_host_test.go 参照）。
func TestCleanupPrivateHost(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClientForZone(t, privateHostTestZone)
	ctx := context.Background()

	findResp, err := c.PrivateHostOpFind(ctx, client.PrivateHostOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, ph := range findResp.PrivateHosts {
		if !hasTestTag(ph.Tags) && !strings.HasPrefix(ph.Name.Value, "test-private-host") {
			continue
		}
		idStr := fmt.Sprintf("%d", ph.ID)
		t.Logf("Deleting privatehost %s (name=%s)", idStr, ph.Name.Value)
		if _, err := c.PrivateHostOpDelete(ctx, client.PrivateHostOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete privatehost %s failed: %v", idStr, err)
		}
	}
}

// TestCleanupCDROM は "test" タグが付いた CDROM リソースを一括削除する。
// FTP 共有中（uploading）の CDROM は先に CloseFTP してから削除する。
func TestCleanupCDROM(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.CDROMOpFind(ctx, client.CDROMOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, cd := range findResp.CDROMs {
		if !hasTestTag(cd.Tags) && !strings.HasPrefix(cd.Name.Value, "test-cdrom") {
			continue
		}
		idStr := fmt.Sprintf("%d", cd.ID.Value)
		t.Logf("Deleting cdrom %s (name=%s avail=%s)", idStr, cd.Name.Value, cd.Availability.Value)
		if cd.Availability.Value == "uploading" {
			if _, err := c.CDROMOpCloseFTP(ctx, client.CDROMOpCloseFTPParams{ID: idStr}); err != nil {
				t.Logf("close FTP on cdrom %s failed: %v", idStr, err)
			}
		}
		if _, err := c.CDROMOpDelete(ctx, client.CDROMOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete cdrom %s failed: %v", idStr, err)
		}
	}
}

// TestCleanupServer は "test" タグ or test-server* の Server リソースを一括削除する。
// 停止中のみ削除可能。起動中の場合はログだけ出してスキップする。
func TestCleanupServer(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.ServerOpFind(ctx, client.ServerOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, s := range findResp.Servers {
		if !hasTestTag(s.Tags) && !strings.HasPrefix(s.Name.Value, "test-server") {
			continue
		}
		idStr := fmt.Sprintf("%d", s.ID.Value)
		status := ""
		if s.Instance.Set {
			status = string(s.Instance.Value.Status.Value)
		}
		if status != "" && status != "down" {
			t.Logf("Skipping running server %s (name=%s status=%s)", idStr, s.Name.Value, status)
			continue
		}
		t.Logf("Deleting server %s (name=%s)", idStr, s.Name.Value)
		if _, err := c.ServerOpDelete(ctx, &client.ServerDeleteRequestEnvelope{}, client.ServerOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete server %s failed: %v", idStr, err)
		}
	}
}

// TestCleanupNote は "test" タグ or test-note* の Note リソースを一括削除する。
// TestIaasNoteCRUD の List が Count=50 で検索するため、アカウント内の Note が
// 50 件を超えると新規作成分が先頭ページに出なくなってテストが落ちる。
func TestCleanupNote(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.NoteOpFind(ctx, client.NoteOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, n := range findResp.Notes {
		if !hasTestTag(n.Tags) && !strings.HasPrefix(n.Name.Value, "test-note") {
			continue
		}
		idStr := fmt.Sprintf("%d", n.ID.Value)
		t.Logf("Deleting note %s (name=%s)", idStr, n.Name.Value)
		if _, err := c.NoteOpDelete(ctx, client.NoteOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete note %s failed: %v", idStr, err)
		}
	}
}

// TestCleanupSSHKey は test-sshkey* 系の SSHKey リソースを一括削除する。
// SSHKey には Tags が無いので Name prefix で対象を判定する。
func TestCleanupSSHKey(t *testing.T) {
	if os.Getenv("TEST_ACC_CLEANUP") == "" {
		t.Skip("TEST_ACC_CLEANUP=1 env var required")
	}
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.SSHKeyOpFind(ctx, client.SSHKeyOpFindParams{})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, k := range findResp.SSHKeys {
		if !strings.HasPrefix(k.Name.Value, "test-sshkey") {
			continue
		}
		idStr := fmt.Sprintf("%d", k.ID.Value)
		t.Logf("Deleting sshkey %s (name=%s)", idStr, k.Name.Value)
		if _, err := c.SSHKeyOpDelete(ctx, client.SSHKeyOpDeleteParams{ID: idStr}); err != nil {
			t.Logf("delete sshkey %s failed: %v", idStr, err)
		}
	}
}

func hasTestTag(tags []string) bool {
	for _, tag := range tags {
		if tag == "test" || tag == "integration" {
			return true
		}
	}
	return false
}
