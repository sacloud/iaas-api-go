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

package power

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
)

// fakeServerOp は ServerPowerAPI 最小実装。
// bootCalls / shutdownCalls で呼び出し回数を記録し、read は状態遷移を順番に返す。
type fakeServerOp struct {
	mu            sync.Mutex
	bootCalls     int
	shutdownCalls int
	// bootErrs / shutdownErrs は順番に返すエラー列。空ならデフォルトで nil。
	bootErrs     []error
	shutdownErrs []error
	// readStates は read が順番に返す (avail, instance_up) 列。末尾に達したら最後の値を返し続ける。
	readStates []struct {
		avail string
		up    string
	}
	readErrs []error
	readIdx  int
	// readOverride が非 nil なら readStates を無視してそちらで状態を返す。
	readOverride func() (avail string, up string)
}

func (f *fakeServerOp) Boot(ctx context.Context, id int64, req *client.ServerBootRequestEnvelope) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	i := f.bootCalls
	f.bootCalls++
	if i < len(f.bootErrs) {
		return f.bootErrs[i]
	}
	return nil
}

func (f *fakeServerOp) Shutdown(ctx context.Context, id int64, req *client.ServerShutdownRequestEnvelope) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	i := f.shutdownCalls
	f.shutdownCalls++
	if i < len(f.shutdownErrs) {
		return f.shutdownErrs[i]
	}
	return nil
}

func (f *fakeServerOp) Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error) {
	if f.readOverride != nil {
		avail, up := f.readOverride()
		s := client.Server{
			Availability: client.NewOptEAvailability(client.EAvailability(avail)),
		}
		if up != "" {
			s.Instance = client.NewOptNilServerInstance(client.ServerInstance{
				Status: client.NewOptEServerInstanceStatus(client.EServerInstanceStatus(up)),
			})
		}
		return &client.ServerReadResponseEnvelope{Server: s}, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	i := f.readIdx
	if i < len(f.readErrs) && f.readErrs[i] != nil {
		f.readIdx++
		return nil, f.readErrs[i]
	}
	if len(f.readStates) == 0 {
		return &client.ServerReadResponseEnvelope{}, nil
	}
	if i >= len(f.readStates) {
		i = len(f.readStates) - 1
	} else {
		f.readIdx++
	}
	st := f.readStates[i]
	s := client.Server{
		Availability: client.NewOptEAvailability(client.EAvailability(st.avail)),
	}
	if st.up != "" {
		s.Instance = client.NewOptNilServerInstance(client.ServerInstance{
			Status: client.NewOptEServerInstanceStatus(client.EServerInstanceStatus(st.up)),
		})
	}
	return &client.ServerReadResponseEnvelope{Server: s}, nil
}

func withFastTimings() func() {
	// 元の値を保存して fast なタイムアウトを設定し、テスト後に復元する。
	origBoot := BootRetrySpan
	origShutdown := ShutdownRetrySpan
	origInitTO := InitialRequestTimeout
	origInitSpan := InitialRequestRetrySpan
	origPoll := PollingInterval
	origOverall := OverallTimeout

	BootRetrySpan = 20 * time.Millisecond
	ShutdownRetrySpan = 20 * time.Millisecond
	InitialRequestTimeout = 200 * time.Millisecond
	InitialRequestRetrySpan = 5 * time.Millisecond
	PollingInterval = 2 * time.Millisecond
	OverallTimeout = 500 * time.Millisecond

	return func() {
		BootRetrySpan = origBoot
		ShutdownRetrySpan = origShutdown
		InitialRequestTimeout = origInitTO
		InitialRequestRetrySpan = origInitSpan
		PollingInterval = origPoll
		OverallTimeout = origOverall
	}
}

func TestBootServer_Success(t *testing.T) {
	defer withFastTimings()()

	op := &fakeServerOp{
		readStates: []struct{ avail, up string }{
			{"available", "down"},
			{"available", "cleaning"},
			{"available", "up"},
		},
	}
	if err := BootServer(context.Background(), op, 1); err != nil {
		t.Fatalf("BootServer failed: %v", err)
	}
	if op.bootCalls != 1 {
		t.Errorf("expected 1 boot call, got %d", op.bootCalls)
	}
}

func TestBootServer_StillCreatingRetry(t *testing.T) {
	defer withFastTimings()()

	stillCreating := iaas.NewAPIError("Server.Boot", http.StatusConflict, &client.ApiErrorStatusCode{
		StatusCode: http.StatusConflict,
		Response:   client.ApiError{ErrorCode: client.NewOptString("still_creating")},
	})

	op := &fakeServerOp{
		bootErrs: []error{stillCreating, stillCreating, nil},
		readStates: []struct{ avail, up string }{
			{"available", "up"},
		},
	}
	if err := BootServer(context.Background(), op, 1); err != nil {
		t.Fatalf("BootServer failed: %v", err)
	}
	if op.bootCalls != 3 {
		t.Errorf("expected 3 boot calls (2 still_creating + 1 success), got %d", op.bootCalls)
	}
}

func TestBootServer_StillCreatingTimeout(t *testing.T) {
	defer withFastTimings()()

	stillCreating := iaas.NewAPIError("Server.Boot", http.StatusConflict, &client.ApiErrorStatusCode{
		StatusCode: http.StatusConflict,
		Response:   client.ApiError{ErrorCode: client.NewOptString("still_creating")},
	})

	// 常に still_creating を返すので InitialRequestTimeout で打ち切り
	errs := make([]error, 1000)
	for i := range errs {
		errs[i] = stillCreating
	}
	op := &fakeServerOp{bootErrs: errs}
	err := BootServer(context.Background(), op, 1)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !contains(err.Error(), "still_creating retry timed out") {
		t.Errorf("expected still_creating timeout error, got %v", err)
	}
}

func TestBootServer_RetrySend_When_Stuck_Down(t *testing.T) {
	defer withFastTimings()()

	// 2 回目の boot 呼び出しが行われるまで "down" を返す → 再送をトリガー
	op := &fakeServerOp{}
	op.readOverride = func() (string, string) {
		op.mu.Lock()
		calls := op.bootCalls
		op.mu.Unlock()
		if calls < 2 {
			return "available", "down"
		}
		return "available", "up"
	}

	if err := BootServer(context.Background(), op, 1); err != nil {
		t.Fatalf("BootServer failed: %v", err)
	}
	if op.bootCalls < 2 {
		t.Errorf("expected boot to be retried at least once, got %d calls", op.bootCalls)
	}
}

func TestBootServer_RetrySend_409_StopsRetries(t *testing.T) {
	defer withFastTimings()()

	// 1回目 Boot 成功、以降の再送はすべて 409 を返す
	conflict := iaas.NewAPIError("Server.Boot", http.StatusConflict, &client.ApiErrorStatusCode{
		StatusCode: http.StatusConflict,
		Response:   client.ApiError{ErrorCode: client.NewOptString("already_booting")},
	})
	bootErrs := make([]error, 100)
	bootErrs[0] = nil
	for i := 1; i < len(bootErrs); i++ {
		bootErrs[i] = conflict
	}
	op := &fakeServerOp{
		bootErrs: bootErrs,
		// 6 回 down のあと up
		readStates: []struct{ avail, up string }{
			{"available", "down"}, {"available", "down"}, {"available", "down"},
			{"available", "down"}, {"available", "down"}, {"available", "down"},
			{"available", "up"},
		},
	}
	if err := BootServer(context.Background(), op, 1); err != nil {
		t.Fatalf("BootServer failed: %v", err)
	}
	// 409 検出後は再送しないので boot は 2 回まで
	if op.bootCalls > 2 {
		t.Errorf("expected boot to stop retrying after 409, got %d calls", op.bootCalls)
	}
}

func TestBootServer_Timeout(t *testing.T) {
	defer withFastTimings()()

	// 永遠に down のまま
	op := &fakeServerOp{
		readStates: []struct{ avail, up string }{
			{"available", "down"},
		},
	}
	err := BootServer(context.Background(), op, 1)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestShutdownServer_Success(t *testing.T) {
	defer withFastTimings()()

	op := &fakeServerOp{
		readStates: []struct{ avail, up string }{
			{"available", "up"},
			{"available", "cleaning"},
			{"available", "down"},
		},
	}
	if err := ShutdownServer(context.Background(), op, 1, false); err != nil {
		t.Fatalf("ShutdownServer failed: %v", err)
	}
	if op.shutdownCalls != 1 {
		t.Errorf("expected 1 shutdown call, got %d", op.shutdownCalls)
	}
}

func TestIsStillCreatingError(t *testing.T) {
	// 409 + still_creating → true
	e1 := iaas.NewAPIError("Op", http.StatusConflict, &client.ApiErrorStatusCode{
		StatusCode: http.StatusConflict,
		Response:   client.ApiError{ErrorCode: client.NewOptString("still_creating")},
	})
	if !isStillCreatingError(e1) {
		t.Error("expected still_creating to be detected")
	}

	// 409 + 別コード → false
	e2 := iaas.NewAPIError("Op", http.StatusConflict, &client.ApiErrorStatusCode{
		StatusCode: http.StatusConflict,
		Response:   client.ApiError{ErrorCode: client.NewOptString("foo")},
	})
	if isStillCreatingError(e2) {
		t.Error("should not detect non-still_creating as still_creating")
	}

	// 500 → false
	e3 := iaas.NewAPIError("Op", http.StatusInternalServerError, &client.ApiErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response:   client.ApiError{ErrorCode: client.NewOptString("still_creating")},
	})
	if isStillCreatingError(e3) {
		t.Error("should not detect non-409 as still_creating")
	}

	// nil → false
	if isStillCreatingError(nil) {
		t.Error("nil should not be still_creating")
	}

	// 通常の error → false
	if isStillCreatingError(errors.New("foo")) {
		t.Error("plain error should not be still_creating")
	}
}

func TestJoinWithNewline(t *testing.T) {
	if got := joinWithNewline(nil); got != "" {
		t.Errorf("empty: got %q", got)
	}
	if got := joinWithNewline([]string{"a"}); got != "a" {
		t.Errorf("single: got %q", got)
	}
	if got := joinWithNewline([]string{"a", "b", "c"}); got != "a\nb\nc" {
		t.Errorf("multi: got %q", got)
	}
}

func contains(haystack, needle string) bool {
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
