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

package cleanup

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/iaas-api-go/v2/helper/query"
)

// ---------- DeleteServer ----------

type fakeServerCleanupOp struct {
	instanceStatus string
	disks          []client.ServerConnectedDisk

	readCalls     int
	bootCalls     int
	shutdownCalls int
	deleteReq     *client.ServerDeleteRequestEnvelope
	deleteCalls   int
}

func (f *fakeServerCleanupOp) Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error) {
	f.readCalls++
	s := client.Server{
		ID:           client.NewOptInt64(id),
		Availability: client.NewOptEAvailability(client.EAvailability("available")),
		Disks:        f.disks,
	}
	if f.instanceStatus != "" {
		s.Instance = client.NewOptNilServerInstance(client.ServerInstance{
			Status: client.NewOptEServerInstanceStatus(client.EServerInstanceStatus(f.instanceStatus)),
		})
	}
	return &client.ServerReadResponseEnvelope{Server: s}, nil
}

func (f *fakeServerCleanupOp) Boot(ctx context.Context, id int64, req *client.ServerBootRequestEnvelope) error {
	f.bootCalls++
	return nil
}

func (f *fakeServerCleanupOp) Shutdown(ctx context.Context, id int64, req *client.ServerShutdownRequestEnvelope) error {
	f.shutdownCalls++
	f.instanceStatus = "down" // shutdown 成功をシミュレート
	return nil
}

func (f *fakeServerCleanupOp) Delete(ctx context.Context, id int64, req *client.ServerDeleteRequestEnvelope) error {
	f.deleteCalls++
	f.deleteReq = req
	return nil
}

func TestDeleteServer_Down_WithoutDisks(t *testing.T) {
	f := &fakeServerCleanupOp{instanceStatus: "down"}
	if err := DeleteServer(context.Background(), f, 1, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.shutdownCalls != 0 {
		t.Errorf("expected no shutdown, got %d", f.shutdownCalls)
	}
	if f.deleteCalls != 1 {
		t.Errorf("expected 1 delete, got %d", f.deleteCalls)
	}
	if len(f.deleteReq.WithDisk) != 0 {
		t.Errorf("expected empty WithDisk, got %v", f.deleteReq.WithDisk)
	}
}

func TestDeleteServer_Up_ShutdownsFirst(t *testing.T) {
	// Server が Up の場合、Shutdown が先に呼ばれることを確認する。
	// power.ShutdownServer は default の OverallTimeout (30分) を使うが、
	// fake は Shutdown 呼び出しで即座に status を "down" に変え、poll の初回で完了するため
	// テストは速く終わる。
	f := &fakeServerCleanupOp{instanceStatus: "up"}
	// power.Shutdown* は初期化で defaults を設定するが、poll 前に status=down に書き換わる
	// ので loop の初回で完了する見込み。
	if err := DeleteServer(context.Background(), f, 1, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.shutdownCalls == 0 {
		t.Error("expected at least 1 shutdown call")
	}
	if f.deleteCalls != 1 {
		t.Errorf("expected 1 delete, got %d", f.deleteCalls)
	}
}

func TestDeleteServer_WithDisks(t *testing.T) {
	f := &fakeServerCleanupOp{
		instanceStatus: "down",
		disks: []client.ServerConnectedDisk{
			{ID: client.NewOptInt64(10)},
			{ID: client.NewOptInt64(20)},
		},
	}
	if err := DeleteServer(context.Background(), f, 1, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.deleteReq.WithDisk) != 2 {
		t.Fatalf("expected 2 disks in WithDisk, got %v", f.deleteReq.WithDisk)
	}
	if string(f.deleteReq.WithDisk[0]) != "10" || string(f.deleteReq.WithDisk[1]) != "20" {
		t.Errorf("unexpected WithDisk: %v", f.deleteReq.WithDisk)
	}
}

// ---------- DeleteDisk ----------

type fakeDiskOp struct {
	deleteCalls int
}

func (f *fakeDiskOp) Delete(ctx context.Context, id int64) error {
	f.deleteCalls++
	return nil
}

type fakeServerFinderForRef struct {
	servers []client.Server
}

func (f *fakeServerFinderForRef) List(ctx context.Context, req *client.ServerFindRequest) (*client.ServerFindResponseEnvelope, error) {
	return &client.ServerFindResponseEnvelope{Servers: f.servers}, nil
}

func TestDeleteDisk_NotReferenced(t *testing.T) {
	r := query.ReferenceFinder{Server: &fakeServerFinderForRef{}}
	op := &fakeDiskOp{}
	err := DeleteDisk(context.Background(), op, r, 77, query.CheckReferencedOption{Timeout: 100 * time.Millisecond, Tick: 5 * time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if op.deleteCalls != 1 {
		t.Errorf("expected 1 delete, got %d", op.deleteCalls)
	}
}

func TestDeleteDisk_StillReferenced_Timeout(t *testing.T) {
	r := query.ReferenceFinder{Server: &fakeServerFinderForRef{
		servers: []client.Server{
			{Disks: []client.ServerConnectedDisk{{ID: client.NewOptInt64(77)}}},
		},
	}}
	op := &fakeDiskOp{}
	err := DeleteDisk(context.Background(), op, r, 77, query.CheckReferencedOption{Timeout: 30 * time.Millisecond, Tick: 5 * time.Millisecond})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if op.deleteCalls != 0 {
		t.Errorf("expected no delete, got %d", op.deleteCalls)
	}
}

// ---------- DeleteBridge ----------

type fakeSwitchFinderForBridge struct {
	switches []client.Switch
}

func (f *fakeSwitchFinderForBridge) List(ctx context.Context, req *client.SwitchFindRequest) (*client.SwitchFindResponseEnvelope, error) {
	return &client.SwitchFindResponseEnvelope{Switches: f.switches}, nil
}

type fakeBridgeOp struct {
	deleteCalls int
}

func (f *fakeBridgeOp) Delete(ctx context.Context, id int64) error {
	f.deleteCalls++
	return nil
}

func TestDeleteBridge_NotReferenced(t *testing.T) {
	sf := &fakeSwitchFinderForBridge{}
	op := &fakeBridgeOp{}
	err := DeleteBridge(context.Background(), op, sf, 55, query.CheckReferencedOption{Timeout: 100 * time.Millisecond, Tick: 5 * time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if op.deleteCalls != 1 {
		t.Errorf("expected 1 delete, got %d", op.deleteCalls)
	}
}

func TestDeleteBridge_ReferencedByOtherBridge_Unaffected(t *testing.T) {
	// 他 Bridge ID (99) に接続された Switch があっても 55 には影響なし
	sf := &fakeSwitchFinderForBridge{
		switches: []client.Switch{
			{Bridge: client.NewOptNilResourceRef(client.ResourceRef{ID: 99})},
		},
	}
	op := &fakeBridgeOp{}
	err := DeleteBridge(context.Background(), op, sf, 55, query.CheckReferencedOption{Timeout: 100 * time.Millisecond, Tick: 5 * time.Millisecond})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if op.deleteCalls != 1 {
		t.Errorf("expected 1 delete, got %d", op.deleteCalls)
	}
}

func TestDeleteBridge_Referenced_Timeout(t *testing.T) {
	sf := &fakeSwitchFinderForBridge{
		switches: []client.Switch{
			{Bridge: client.NewOptNilResourceRef(client.ResourceRef{ID: 55})},
		},
	}
	op := &fakeBridgeOp{}
	err := DeleteBridge(context.Background(), op, sf, 55, query.CheckReferencedOption{Timeout: 30 * time.Millisecond, Tick: 5 * time.Millisecond})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if op.deleteCalls != 0 {
		t.Errorf("expected no delete, got %d", op.deleteCalls)
	}
}

// ---------- DeleteInternet (trivial) ----------

type fakeInternetOp struct {
	deleteCalls int
}

func (f *fakeInternetOp) Delete(ctx context.Context, id int64) error {
	f.deleteCalls++
	return nil
}

func TestDeleteInternet(t *testing.T) {
	op := &fakeInternetOp{}
	if err := DeleteInternet(context.Background(), op, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if op.deleteCalls != 1 {
		t.Errorf("expected 1 delete, got %d", op.deleteCalls)
	}
}

// ---------- Disk ID encoding ----------

func TestDiskIDEncoding(t *testing.T) {
	// ServerDeleteRequestEnvelope.WithDisk は []client.ID (string)。
	// サーバが返す Disk.ID は int64 なので strconv で string に変換される。
	want := "1234567890"
	got := strconv.FormatInt(1234567890, 10)
	if got != want {
		t.Errorf("ID formatting: got %q, want %q", got, want)
	}
}
