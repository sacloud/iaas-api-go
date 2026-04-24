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

package wait

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
)

func TestStateWaiter_ImmediateMatch(t *testing.T) {
	called := 0
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			called++
			return StateResult{Availability: "available"}, nil
		},
		TargetAvailability: []string{"available"},
		Interval:           10 * time.Millisecond,
		Timeout:            1 * time.Second,
	}
	if err := w.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called != 1 {
		t.Errorf("expected 1 read, got %d", called)
	}
}

func TestStateWaiter_PendingThenTarget(t *testing.T) {
	seq := []string{"migrating", "migrating", "available"}
	idx := 0
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			r := StateResult{Availability: seq[idx]}
			idx++
			return r, nil
		},
		TargetAvailability:  []string{"available"},
		PendingAvailability: []string{"migrating"},
		Interval:            1 * time.Millisecond,
		Timeout:             1 * time.Second,
	}
	if err := w.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != len(seq) {
		t.Errorf("expected %d reads, got %d", len(seq), idx)
	}
}

func TestStateWaiter_UnexpectedAvailability(t *testing.T) {
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			return StateResult{Availability: "failed"}, nil
		},
		TargetAvailability:  []string{"available"},
		PendingAvailability: []string{"migrating"},
		Interval:            1 * time.Millisecond,
		Timeout:             1 * time.Second,
	}
	err := w.Wait(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if want := `unexpected availability: "failed"`; !contains2(err.Error(), want) {
		t.Errorf("error should mention %q: got %v", want, err)
	}
}

func TestStateWaiter_InstanceStatusCheck(t *testing.T) {
	seq := []StateResult{
		{Availability: "available", InstanceStatus: "down"},
		{Availability: "available", InstanceStatus: "cleaning"},
		{Availability: "available", InstanceStatus: "up"},
	}
	idx := 0
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			r := seq[idx]
			idx++
			return r, nil
		},
		TargetAvailability:    []string{"available"},
		TargetInstanceStatus:  []string{"up"},
		PendingInstanceStatus: []string{"down", "cleaning"},
		Interval:              1 * time.Millisecond,
		Timeout:               1 * time.Second,
	}
	if err := w.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if idx != len(seq) {
		t.Errorf("expected %d reads, got %d", len(seq), idx)
	}
}

func TestStateWaiter_NotFoundRetry_Within(t *testing.T) {
	reads := 0
	notFoundErr := newNotFoundErr()
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			reads++
			if reads < 3 {
				return StateResult{}, notFoundErr
			}
			return StateResult{Availability: "available"}, nil
		},
		TargetAvailability: []string{"available"},
		NotFoundRetry:      5,
		Interval:           1 * time.Millisecond,
		Timeout:            1 * time.Second,
	}
	if err := w.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reads != 3 {
		t.Errorf("expected 3 reads, got %d", reads)
	}
}

func TestStateWaiter_NotFoundRetry_Exceeded(t *testing.T) {
	notFoundErr := newNotFoundErr()
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			return StateResult{}, notFoundErr
		},
		TargetAvailability: []string{"available"},
		NotFoundRetry:      2,
		Interval:           1 * time.Millisecond,
		Timeout:            1 * time.Second,
	}
	err := w.Wait(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains2(err.Error(), "not found after 2 retries") {
		t.Errorf("error should mention retry count: got %v", err)
	}
}

func TestStateWaiter_ReadError(t *testing.T) {
	sentinel := errors.New("network error")
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			return StateResult{}, sentinel
		},
		TargetAvailability: []string{"available"},
		Interval:           1 * time.Millisecond,
		Timeout:            1 * time.Second,
	}
	err := w.Wait(context.Background())
	if !errors.Is(err, sentinel) {
		t.Errorf("expected wrapping of sentinel, got %v", err)
	}
}

func TestStateWaiter_Timeout(t *testing.T) {
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			return StateResult{Availability: "migrating"}, nil
		},
		TargetAvailability:  []string{"available"},
		PendingAvailability: []string{"migrating"},
		Interval:            5 * time.Millisecond,
		Timeout:             30 * time.Millisecond,
	}
	err := w.Wait(context.Background())
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestStateWaiter_ParentCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			return StateResult{Availability: "migrating"}, nil
		},
		TargetAvailability:  []string{"available"},
		PendingAvailability: []string{"migrating"},
		Interval:            10 * time.Millisecond,
		Timeout:             10 * time.Second,
	}
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	err := w.Wait(ctx)
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected Canceled, got %v", err)
	}
}

func TestStateWaiter_NoAvailabilityCheck(t *testing.T) {
	// TargetAvailability が空なら availability チェックはスキップされる。
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			return StateResult{InstanceStatus: "up"}, nil
		},
		TargetInstanceStatus: []string{"up"},
		Interval:             1 * time.Millisecond,
		Timeout:              1 * time.Second,
	}
	if err := w.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSimpleStateWaiter_Completes(t *testing.T) {
	read := 0
	target := 5
	s := &SimpleStateWaiter{
		ReadStateFunc: func(ctx context.Context) (bool, error) {
			read++
			return read >= target, nil
		},
		Timeout:         1 * time.Second,
		PollingInterval: 1 * time.Millisecond,
	}
	if err := s.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if read != target {
		t.Errorf("expected %d reads, got %d", target, read)
	}
}

func TestSimpleStateWaiter_ErrorPropagates(t *testing.T) {
	sentinel := errors.New("boom")
	s := &SimpleStateWaiter{
		ReadStateFunc: func(ctx context.Context) (bool, error) {
			return false, sentinel
		},
		Timeout:         100 * time.Millisecond,
		PollingInterval: 1 * time.Millisecond,
	}
	err := s.Wait(context.Background())
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestSimpleStateWaiter_Timeout(t *testing.T) {
	s := &SimpleStateWaiter{
		ReadStateFunc: func(ctx context.Context) (bool, error) {
			return false, nil
		},
		Timeout:         30 * time.Millisecond,
		PollingInterval: 5 * time.Millisecond,
	}
	err := s.Wait(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

// ---------- resource-specific helper tests (with fake readers) ----------

type fakeServerReader struct {
	responses []*client.ServerReadResponseEnvelope
	errs      []error
	idx       int
}

func (f *fakeServerReader) Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error) {
	if f.idx >= len(f.responses) {
		return nil, errors.New("fakeServerReader exhausted")
	}
	r, e := f.responses[f.idx], f.errs[f.idx]
	f.idx++
	return r, e
}

func serverEnvelope(avail, status string) *client.ServerReadResponseEnvelope {
	s := client.Server{
		Availability: client.NewOptEAvailability(client.EAvailability(avail)),
	}
	if status != "" {
		s.Instance = client.NewOptNilServerInstance(client.ServerInstance{
			Status: client.NewOptEServerInstanceStatus(client.EServerInstanceStatus(status)),
		})
	}
	return &client.ServerReadResponseEnvelope{Server: s}
}

func TestUntilServerIsUp(t *testing.T) {
	r := &fakeServerReader{
		responses: []*client.ServerReadResponseEnvelope{
			serverEnvelope("migrating", ""),
			serverEnvelope("available", "cleaning"),
			serverEnvelope("available", "up"),
		},
		errs: []error{nil, nil, nil},
	}
	// 短い interval/timeout を注入するため、StateWaiter を直接組み立てずに
	// 一時的に Interval を変えたいので、wait.go の default を尊重しつつ helper を直呼び出しで検証する場合は
	// ここでは wrapper の挙動を通して確認する。Interval = DefaultInterval(5s) * 3 = 15s の poll を
	// 3 step 実行するとテスト時間がかかるので、ここでは StateWaiter を直接使って同じロジックを検証する。
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			resp, err := r.Read(ctx, 1)
			if err != nil {
				return StateResult{}, err
			}
			return StateResult{
				Availability:   string(resp.Server.Availability.Value),
				InstanceStatus: string(resp.Server.Instance.Value.Status.Value),
			}, nil
		},
		TargetAvailability:    availTargetUp,
		PendingAvailability:   availPendingUp,
		TargetInstanceStatus:  statusTargetUp,
		PendingInstanceStatus: statusPendingUp,
		Interval:              1 * time.Millisecond,
		Timeout:               1 * time.Second,
	}
	if err := w.Wait(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.idx != 3 {
		t.Errorf("expected 3 reads, got %d", r.idx)
	}
}

func TestUntilServerIsUp_ReaderWrapper(t *testing.T) {
	// UntilServerIsUp 本体 (default interval) が narrow interface で呼び出せることだけ確認。
	// 実際の poll ループは TestStateWaiter_* でカバー済み。
	r := &fakeServerReader{
		responses: []*client.ServerReadResponseEnvelope{
			serverEnvelope("available", "up"),
		},
		errs: []error{nil},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	last, err := UntilServerIsUp(ctx, r, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if last == nil || string(last.Availability.Value) != "available" {
		t.Errorf("last should reflect final state, got %+v", last)
	}
}

// ---------- helpers ----------

func newNotFoundErr() error {
	// iaas.NewAPIError で 404 を含むエラーを作る。iaas.IsNotFoundError が true を返すこと。
	return iaas.NewAPIError("Test.Read", http.StatusNotFound, errors.New("404"))
}

func contains2(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
