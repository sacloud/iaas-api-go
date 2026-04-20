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
	zone := getZone()

	findResp, err := c.InternetOpFind(ctx, &client.InternetFindRequestEnvelope{}, client.InternetOpFindParams{Zone: zone})
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
			for _, ipv6 := range ii.Switch.Value.IPv6Nets.Value {
				ipv6IDStr := fmt.Sprintf("%d", ipv6.ID.Value)
				if _, err := c.InternetOpDisableIPv6(ctx, client.InternetOpDisableIPv6Params{
					Zone:      zone,
					ID:        idStr,
					Ipv6netID: ipv6IDStr,
				}); err != nil {
					t.Logf("disable ipv6 %s on internet %s failed: %v", ipv6IDStr, idStr, err)
				}
			}
		}
		if _, err := c.InternetOpDelete(ctx, client.InternetOpDeleteParams{Zone: zone, ID: idStr}); err != nil {
			t.Logf("delete internet %s failed: %v", idStr, err)
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
	zone := getZone()

	findResp, err := c.ApplianceOpFind(ctx, &client.FindCondition{}, client.ApplianceOpFindParams{Zone: zone})
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
			if _, err := c.ApplianceOpShutdown(ctx, &client.ShutdownOption{Force: true}, client.ApplianceOpShutdownParams{Zone: zone, ID: idStr}); err != nil {
				t.Logf("force shutdown %s failed: %v", idStr, err)
			}
			// wait a bit for status to flip
			time.Sleep(10 * time.Second)
		}
		if _, err := c.ApplianceOpDelete(ctx, client.ApplianceOpDeleteParams{Zone: zone, ID: idStr}); err != nil {
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
	c := newClient(t)
	ctx := context.Background()

	findResp, err := c.PrivateHostOpFind(ctx, &client.PrivateHostFindRequestEnvelope{}, client.PrivateHostOpFindParams{Zone: privateHostTestZone})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, ph := range findResp.PrivateHosts {
		if !hasTestTag(ph.Tags) && !strings.HasPrefix(ph.Name.Value, "test-private-host") {
			continue
		}
		idStr := fmt.Sprintf("%d", ph.ID)
		t.Logf("Deleting privatehost %s (name=%s)", idStr, ph.Name.Value)
		if _, err := c.PrivateHostOpDelete(ctx, client.PrivateHostOpDeleteParams{Zone: privateHostTestZone, ID: idStr}); err != nil {
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
	zone := getZone()

	findResp, err := c.CDROMOpFind(ctx, &client.CDROMFindRequestEnvelope{}, client.CDROMOpFindParams{Zone: zone})
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
			if _, err := c.CDROMOpCloseFTP(ctx, client.CDROMOpCloseFTPParams{Zone: zone, ID: idStr}); err != nil {
				t.Logf("close FTP on cdrom %s failed: %v", idStr, err)
			}
		}
		if _, err := c.CDROMOpDelete(ctx, client.CDROMOpDeleteParams{Zone: zone, ID: idStr}); err != nil {
			t.Logf("delete cdrom %s failed: %v", idStr, err)
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
	zone := getZone()

	findResp, err := c.SSHKeyOpFind(ctx, &client.SSHKeyFindRequestEnvelope{}, client.SSHKeyOpFindParams{Zone: zone})
	if err != nil {
		t.Fatalf("find failed: %v", err)
	}
	for _, k := range findResp.SSHKeys {
		if !strings.HasPrefix(k.Name.Value, "test-sshkey") {
			continue
		}
		idStr := fmt.Sprintf("%d", k.ID.Value)
		t.Logf("Deleting sshkey %s (name=%s)", idStr, k.Name.Value)
		if _, err := c.SSHKeyOpDelete(ctx, client.SSHKeyOpDeleteParams{Zone: zone, ID: idStr}); err != nil {
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
