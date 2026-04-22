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
		id := ii.ID.Value
		t.Logf("Deleting internet %d (name=%s)", id, ii.Name.Value)
		// Internet 配下の IPv6Net を先に外す（そうしないと delete が 409 になる）
		if ii.Switch.Set && !ii.Switch.Null {
			for _, ipv6 := range ii.Switch.Value.IPv6Nets {
				ipv6ID := ipv6.ID.Value
				if _, err := c.InternetOpDisableIPv6(ctx, client.InternetOpDisableIPv6Params{ID: id,
					Ipv6netID: ipv6ID,
				}); err != nil {
					t.Logf("disable ipv6 %d on internet %d failed: %v", ipv6ID, id, err)
				}
			}
		}
		if _, err := c.InternetOpDelete(ctx, client.InternetOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete internet %d failed: %v", id, err)
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
		id := sw.ID.Value
		t.Logf("Deleting switch %d (name=%s)", id, sw.Name.Value)
		if _, err := c.SwitchOpDelete(ctx, client.SwitchOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete switch %d failed: %v", id, err)
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
		id := b.ID.Value
		t.Logf("Deleting bridge %d (name=%s)", id, b.Name.Value)
		// 接続中の Switch があれば先に disconnect する必要があるが、このテストでは defer で済ませているので不要
		if _, err := c.BridgeOpDelete(ctx, client.BridgeOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete bridge %d failed: %v", id, err)
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
		id := app.ID.Value
		t.Logf("Deleting appliance %d (name=%s class=%s status=%s)", id, app.Name.Value, app.Class.Value, app.Instance.Value.Status.Value)
		if app.Instance.Value.Status.Value == "up" {
			if _, err := c.ApplianceOpShutdown(ctx, &client.ShutdownOption{Force: true}, client.ApplianceOpShutdownParams{ID: id}); err != nil {
				t.Logf("force shutdown %d failed: %v", id, err)
			}
			// wait a bit for status to flip
			time.Sleep(10 * time.Second)
		}
		if _, err := c.ApplianceOpDelete(ctx, client.ApplianceOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete appliance %d failed: %v", id, err)
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
		id := ph.ID
		t.Logf("Deleting privatehost %d (name=%s)", id, ph.Name.Value)
		if _, err := c.PrivateHostOpDelete(ctx, client.PrivateHostOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete privatehost %d failed: %v", id, err)
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
		id := cd.ID.Value
		t.Logf("Deleting cdrom %d (name=%s avail=%s)", id, cd.Name.Value, cd.Availability.Value)
		if cd.Availability.Value == "uploading" {
			if _, err := c.CDROMOpCloseFTP(ctx, client.CDROMOpCloseFTPParams{ID: id}); err != nil {
				t.Logf("close FTP on cdrom %d failed: %v", id, err)
			}
		}
		if _, err := c.CDROMOpDelete(ctx, client.CDROMOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete cdrom %d failed: %v", id, err)
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
		id := s.ID.Value
		status := ""
		if s.Instance.Set {
			status = string(s.Instance.Value.Status.Value)
		}
		if status != "" && status != "down" {
			t.Logf("Skipping running server %d (name=%s status=%s)", id, s.Name.Value, status)
			continue
		}
		t.Logf("Deleting server %d (name=%s)", id, s.Name.Value)
		if _, err := c.ServerOpDelete(ctx, &client.ServerDeleteRequestEnvelope{}, client.ServerOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete server %d failed: %v", id, err)
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
		id := n.ID.Value
		t.Logf("Deleting note %d (name=%s)", id, n.Name.Value)
		if _, err := c.NoteOpDelete(ctx, client.NoteOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete note %d failed: %v", id, err)
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
		id := k.ID.Value
		t.Logf("Deleting sshkey %d (name=%s)", id, k.Name.Value)
		if _, err := c.SSHKeyOpDelete(ctx, client.SSHKeyOpDeleteParams{ID: id}); err != nil {
			t.Logf("delete sshkey %d failed: %v", id, err)
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
