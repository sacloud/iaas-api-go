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

package query

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
)

// ---------- FindArchiveByOSType ----------

type fakeArchiveFinder struct {
	resp *client.ArchiveFindResponseEnvelope
	err  error
	req  *client.ArchiveFindRequest
}

func (f *fakeArchiveFinder) List(ctx context.Context, req *client.ArchiveFindRequest) (*client.ArchiveFindResponseEnvelope, error) {
	f.req = req
	return f.resp, f.err
}

func TestFindArchiveByOSType_Found(t *testing.T) {
	f := &fakeArchiveFinder{
		resp: &client.ArchiveFindResponseEnvelope{
			Archives: []client.Archive{
				{ID: client.NewOptInt64(42), Name: client.NewOptString("ubuntu-24")},
			},
		},
	}
	a, err := FindArchiveByOSType(context.Background(), f, Ubuntu2404)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ID.Value != 42 {
		t.Errorf("expected ID=42, got %d", a.ID.Value)
	}
	if len(f.req.Filter.Tags) == 0 {
		t.Error("expected tags in request")
	}
	if f.req.Filter.Scope != "shared" {
		t.Errorf("expected Scope=shared, got %q", f.req.Filter.Scope)
	}
}

func TestFindArchiveByOSType_NotFound(t *testing.T) {
	f := &fakeArchiveFinder{resp: &client.ArchiveFindResponseEnvelope{}}
	_, err := FindArchiveByOSType(context.Background(), f, Ubuntu)
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestFindArchiveByOSType_UnsupportedType(t *testing.T) {
	f := &fakeArchiveFinder{}
	_, err := FindArchiveByOSType(context.Background(), f, ArchiveOSType(9999))
	if err == nil {
		t.Fatal("expected unsupported type error")
	}
}

// ---------- FindServerPlan ----------

type fakeServerPlanFinder struct {
	plans []client.ServerPlan
}

func (f *fakeServerPlanFinder) List(ctx context.Context, req *client.ServerPlanFindRequest) (*client.ServerPlanFindResponseEnvelope, error) {
	return &client.ServerPlanFindResponseEnvelope{ServerPlans: f.plans}, nil
}

func makePlan(id int64, cpu, memMB int32, gen client.EPlanGeneration, avail string) client.ServerPlan {
	return client.ServerPlan{
		ID:           client.NewOptInt64(id),
		CPU:          client.NewOptInt32(cpu),
		MemoryMB:     client.NewOptInt32(memMB),
		Generation:   client.NewOptEPlanGeneration(gen),
		Availability: client.NewOptEAvailability(client.EAvailability(avail)),
	}
}

func TestFindServerPlan_MatchesCPUAndMemory(t *testing.T) {
	f := &fakeServerPlanFinder{
		plans: []client.ServerPlan{
			makePlan(1, 1, 1024, 100, "available"),
			makePlan(2, 2, 2048, 200, "available"),
			makePlan(3, 2, 4096, 200, "available"),
		},
	}
	p, err := FindServerPlan(context.Background(), f, &FindServerPlanRequest{CPU: 2, MemoryGB: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID.Value != 2 {
		t.Errorf("expected plan ID=2, got %d", p.ID.Value)
	}
}

func TestFindServerPlan_PrefersNewerGeneration(t *testing.T) {
	// gen 100 と 200 の両方にマッチする条件 → 新世代 (200) を選ぶ
	f := &fakeServerPlanFinder{
		plans: []client.ServerPlan{
			makePlan(1, 1, 1024, 100, "available"),
			makePlan(2, 1, 1024, 200, "available"),
		},
	}
	p, err := FindServerPlan(context.Background(), f, &FindServerPlanRequest{CPU: 1, MemoryGB: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID.Value != 2 {
		t.Errorf("expected newer gen plan ID=2, got %d", p.ID.Value)
	}
}

func TestFindServerPlan_NoMatch(t *testing.T) {
	f := &fakeServerPlanFinder{
		plans: []client.ServerPlan{makePlan(1, 1, 1024, 100, "available")},
	}
	_, err := FindServerPlan(context.Background(), f, &FindServerPlanRequest{CPU: 100})
	if err == nil {
		t.Fatal("expected not-found error")
	}
}

func TestFindServerPlan_SkipsDiscontinued(t *testing.T) {
	f := &fakeServerPlanFinder{
		plans: []client.ServerPlan{
			makePlan(1, 1, 1024, 200, "discontinued"),
			makePlan(2, 1, 1024, 100, "available"),
		},
	}
	p, err := FindServerPlan(context.Background(), f, &FindServerPlanRequest{CPU: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID.Value != 2 {
		t.Errorf("expected available plan ID=2, got %d", p.ID.Value)
	}
}

// ---------- ReadServer (previous-id fallback) ----------

type fakeServerReadFinder struct {
	readResp *client.ServerReadResponseEnvelope
	readErr  error
	listResp *client.ServerFindResponseEnvelope
	listErr  error
	listReq  *client.ServerFindRequest
}

func (f *fakeServerReadFinder) Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error) {
	return f.readResp, f.readErr
}
func (f *fakeServerReadFinder) List(ctx context.Context, req *client.ServerFindRequest) (*client.ServerFindResponseEnvelope, error) {
	f.listReq = req
	return f.listResp, f.listErr
}

func notFoundErr() error {
	return iaas.NewAPIError("Server.Read", http.StatusNotFound, errors.New("404"))
}

func TestReadServer_DirectRead(t *testing.T) {
	f := &fakeServerReadFinder{
		readResp: &client.ServerReadResponseEnvelope{Server: client.Server{ID: client.NewOptInt64(123)}},
	}
	s, err := ReadServer(context.Background(), f, 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID.Value != 123 {
		t.Errorf("expected ID=123, got %d", s.ID.Value)
	}
}

func TestReadServer_PreviousIDFallback(t *testing.T) {
	f := &fakeServerReadFinder{
		readErr: notFoundErr(),
		listResp: &client.ServerFindResponseEnvelope{
			Servers: []client.Server{{ID: client.NewOptInt64(456)}},
		},
	}
	s, err := ReadServer(context.Background(), f, 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID.Value != 456 {
		t.Errorf("expected fallback ID=456, got %d", s.ID.Value)
	}
	if len(f.listReq.Filter.Tags) != 1 || f.listReq.Filter.Tags[0] != "@previous-id=123" {
		t.Errorf("expected @previous-id=123 tag, got %v", f.listReq.Filter.Tags)
	}
}

func TestReadServer_NotFoundAnywhere(t *testing.T) {
	f := &fakeServerReadFinder{
		readErr:  notFoundErr(),
		listResp: &client.ServerFindResponseEnvelope{},
	}
	_, err := ReadServer(context.Background(), f, 123)
	if !errors.Is(err, ErrNoResults) {
		t.Errorf("expected ErrNoResults, got %v", err)
	}
}

func TestReadServer_NonNotFoundError(t *testing.T) {
	other := iaas.NewAPIError("Server.Read", http.StatusInternalServerError, errors.New("boom"))
	f := &fakeServerReadFinder{readErr: other}
	_, err := ReadServer(context.Background(), f, 123)
	if err == nil || errors.Is(err, ErrNoResults) {
		t.Errorf("expected pass-through error, got %v", err)
	}
}

// ---------- waitWhileReferenced ----------

func TestWaitWhileReferenced_Completes(t *testing.T) {
	calls := 0
	err := waitWhileReferenced(context.Background(), CheckReferencedOption{Timeout: 500 * time.Millisecond, Tick: 5 * time.Millisecond}, func() (bool, error) {
		calls++
		return calls < 3, nil // 3 回目で参照なし
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWaitWhileReferenced_Timeout(t *testing.T) {
	err := waitWhileReferenced(context.Background(), CheckReferencedOption{Timeout: 30 * time.Millisecond, Tick: 5 * time.Millisecond}, func() (bool, error) {
		return true, nil
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestWaitWhileReferenced_ErrorPropagates(t *testing.T) {
	sentinel := errors.New("boom")
	err := waitWhileReferenced(context.Background(), CheckReferencedOption{Timeout: 500 * time.Millisecond, Tick: 5 * time.Millisecond}, func() (bool, error) {
		return true, sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// ---------- IsDiskReferenced ----------

type fakeServerFinder struct {
	resp *client.ServerFindResponseEnvelope
}

func (f *fakeServerFinder) List(ctx context.Context, req *client.ServerFindRequest) (*client.ServerFindResponseEnvelope, error) {
	return f.resp, nil
}

func TestIsDiskReferenced_True(t *testing.T) {
	f := &fakeServerFinder{resp: &client.ServerFindResponseEnvelope{
		Servers: []client.Server{{
			ID:    client.NewOptInt64(1),
			Disks: []client.ServerConnectedDisk{{ID: client.NewOptInt64(77)}},
		}},
	}}
	r := ReferenceFinder{Server: f}
	ref, err := IsDiskReferenced(context.Background(), r, 77)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ref {
		t.Error("expected referenced=true")
	}
}

func TestIsDiskReferenced_False(t *testing.T) {
	f := &fakeServerFinder{resp: &client.ServerFindResponseEnvelope{
		Servers: []client.Server{{
			ID:    client.NewOptInt64(1),
			Disks: []client.ServerConnectedDisk{{ID: client.NewOptInt64(99)}},
		}},
	}}
	r := ReferenceFinder{Server: f}
	ref, err := IsDiskReferenced(context.Background(), r, 77)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref {
		t.Error("expected referenced=false")
	}
}
