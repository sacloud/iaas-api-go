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

package plans

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/sacloud/iaas-api-go/v2/client"
)

func TestAppendPreviousIDTagIfAbsent_Empty(t *testing.T) {
	got := AppendPreviousIDTagIfAbsent(nil, 123)
	want := []string{"@previous-id=123"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAppendPreviousIDTagIfAbsent_ReplacesExisting(t *testing.T) {
	got := AppendPreviousIDTagIfAbsent([]string{"@previous-id=100", "foo"}, 200)
	want := []string{"@previous-id=200", "foo"}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAppendPreviousIDTagIfAbsent_SkipsWhenMaxTags(t *testing.T) {
	// 既に 10 個を超えて付与されている場合はそのまま返す
	base := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	got := AppendPreviousIDTagIfAbsent(base, 99)
	if !reflect.DeepEqual(got, base) {
		t.Errorf("expected unchanged, got %v", got)
	}
}

func TestAppendPreviousIDTagIfAbsent_Sorted(t *testing.T) {
	got := AppendPreviousIDTagIfAbsent([]string{"zeta", "alpha"}, 5)
	want := []string{"@previous-id=5", "alpha", "zeta"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// ---------- ChangeServerPlan ----------

type fakeServerPlanOp struct {
	readServer    client.Server
	updatedReq    *client.ServerUpdateRequestEnvelope
	updateCalls   int
	changePlanReq *client.ServerChangePlanRequestEnvelope
	changeCalls   int
}

func (f *fakeServerPlanOp) Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error) {
	return &client.ServerReadResponseEnvelope{Server: f.readServer}, nil
}
func (f *fakeServerPlanOp) Update(ctx context.Context, id int64, req *client.ServerUpdateRequestEnvelope) (*client.ServerUpdateResponseEnvelope, error) {
	f.updateCalls++
	f.updatedReq = req
	return &client.ServerUpdateResponseEnvelope{Server: f.readServer}, nil
}
func (f *fakeServerPlanOp) ChangePlan(ctx context.Context, id int64, req *client.ServerChangePlanRequestEnvelope) (*client.ServerChangePlanResponseEnvelope, error) {
	f.changeCalls++
	f.changePlanReq = req
	return &client.ServerChangePlanResponseEnvelope{Server: f.readServer}, nil
}

func TestChangeServerPlan_AddsPreviousIDTag(t *testing.T) {
	f := &fakeServerPlanOp{
		readServer: client.Server{
			ID:   client.NewOptInt64(42),
			Name: client.NewOptString("srv"),
			Tags: []string{"existing"},
		},
	}
	_, err := ChangeServerPlan(context.Background(), f, 42, &client.ServerChangePlanRequestEnvelope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.updateCalls != 1 {
		t.Errorf("expected 1 Update call, got %d", f.updateCalls)
	}
	if f.changeCalls != 1 {
		t.Errorf("expected 1 ChangePlan call, got %d", f.changeCalls)
	}
	tags := f.updatedReq.Server.Tags
	found := false
	for _, t := range tags {
		if t == "@previous-id=42" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected @previous-id=42 in tags, got %v", tags)
	}
}

func TestChangeServerPlan_SkipsUpdateWhenMaxTags(t *testing.T) {
	// 10 個のタグがあれば Update をスキップ
	tags := make([]string, MaxTags)
	for i := range tags {
		tags[i] = "t" + string(rune('a'+i))
	}
	f := &fakeServerPlanOp{
		readServer: client.Server{ID: client.NewOptInt64(42), Tags: tags},
	}
	_, err := ChangeServerPlan(context.Background(), f, 42, &client.ServerChangePlanRequestEnvelope{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.updateCalls != 0 {
		t.Errorf("expected 0 Update calls when tags at max, got %d", f.updateCalls)
	}
	if f.changeCalls != 1 {
		t.Errorf("expected 1 ChangePlan call, got %d", f.changeCalls)
	}
}

// ---------- ChangeRouterPlan ----------

type fakeRouterPlanOp struct {
	readInternet  client.Internet
	updateCalls   int
	updatedReq    *client.InternetUpdateRequestEnvelope
	bwCalls       int
	bwReq         *client.InternetUpdateBandWidthRequestEnvelope
}

func (f *fakeRouterPlanOp) Read(ctx context.Context, id int64) (*client.InternetReadResponseEnvelope, error) {
	return &client.InternetReadResponseEnvelope{Internet: f.readInternet}, nil
}
func (f *fakeRouterPlanOp) Update(ctx context.Context, id int64, req *client.InternetUpdateRequestEnvelope) (*client.InternetUpdateResponseEnvelope, error) {
	f.updateCalls++
	f.updatedReq = req
	return &client.InternetUpdateResponseEnvelope{Internet: f.readInternet}, nil
}
func (f *fakeRouterPlanOp) UpdateBandWidth(ctx context.Context, id int64, req *client.InternetUpdateBandWidthRequestEnvelope) (*client.InternetUpdateBandWidthResponseEnvelope, error) {
	f.bwCalls++
	f.bwReq = req
	return &client.InternetUpdateBandWidthResponseEnvelope{Internet: f.readInternet}, nil
}

func TestChangeRouterPlan_AddsPreviousIDAndUpdatesBandwidth(t *testing.T) {
	f := &fakeRouterPlanOp{
		readInternet: client.Internet{
			ID:   client.NewOptInt64(100),
			Tags: []string{},
		},
	}
	_, err := ChangeRouterPlan(context.Background(), f, 100, 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.updateCalls != 1 {
		t.Errorf("expected 1 Update, got %d", f.updateCalls)
	}
	if f.bwCalls != 1 {
		t.Errorf("expected 1 UpdateBandWidth, got %d", f.bwCalls)
	}
	if f.bwReq.Internet.BandWidthMbps.Value != 500 {
		t.Errorf("expected bw=500, got %d", f.bwReq.Internet.BandWidthMbps.Value)
	}
}
