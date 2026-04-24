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

// Package wait は v2 リソースの状態遷移を待機する。
//
// v1 の github.com/sacloud/iaas-api-go/helper/wait の v2 相当。v1 のような
// iaas.StatePollingWaiter には依存せず、本パッケージ内で StateWaiter を実装する。
//
// resource 向けヘルパー (UntilServerIsUp 等) は narrow な reader interface を引数に取るため、
// テスト時はその interface を満たす fake を渡せばよい。iaas.ServerAPI / iaas.DiskAPI 等の
// 公開 interface は自動的に narrow interface を満たす。
package wait

import (
	"context"
	"fmt"
	"time"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
)

const (
	// DefaultInterval は poll 間隔のデフォルト値。
	DefaultInterval = 5 * time.Second
	// DefaultTimeout は全体タイムアウトのデフォルト値。
	DefaultTimeout = 30 * time.Minute

	// ApplianceNotFoundRetryCount はアプライアンス待機時に 404 を許容する回数。
	// create 直後はしばらく 404 を返す実サーバー挙動に対応。
	ApplianceNotFoundRetryCount = 30
)

// StateResult は 1 回の poll で読み取ったリソース状態スナップショット。
type StateResult struct {
	// Availability は API が返した availability 文字列 ("available" / "migrating" 等)。
	// 未設定の場合は空文字列。
	Availability string
	// InstanceStatus は Instance.Status 文字列 ("up" / "down" / "cleaning" 等)。
	// Instance が無いリソース (Archive / Disk 等) では空文字列を渡す。
	InstanceStatus string
}

// StateReadFunc は 1 回の poll で現在状態を読む関数。
// 404 のときは iaas パッケージ由来のエラーをそのまま返すこと。StateWaiter が
// iaas.IsNotFoundError で判定する。
type StateReadFunc func(ctx context.Context) (StateResult, error)

// StateWaiter は availability / instance status に基づいて待機する汎用ポーラー。
type StateWaiter struct {
	ReadFunc StateReadFunc

	// TargetAvailability は目標とする availability の集合。空の場合は availability チェックをスキップ。
	TargetAvailability []string
	// PendingAvailability は待機中として扱う availability の集合。
	// TargetAvailability でも PendingAvailability でもない値を観測したらエラー終了する。
	PendingAvailability []string

	// TargetInstanceStatus は目標とする Instance.Status の集合。空の場合はスキップ。
	TargetInstanceStatus []string
	// PendingInstanceStatus は待機中として扱う Instance.Status の集合。
	PendingInstanceStatus []string

	// NotFoundRetry は 404 を許容する連続回数。超えたらエラー。0 なら即エラー。
	NotFoundRetry int

	// Interval は poll 間隔 (未指定時 DefaultInterval)。
	Interval time.Duration
	// Timeout は全体タイムアウト (未指定時 DefaultTimeout)。
	Timeout time.Duration
}

// Wait は目標状態に到達するまで poll する。ctx 取消しや Timeout 経過で error を返す。
func (w *StateWaiter) Wait(ctx context.Context) error {
	interval := w.Interval
	if interval <= 0 {
		interval = DefaultInterval
	}
	timeout := w.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	notFoundCount := 0
	for {
		result, err := w.ReadFunc(ctx)
		if err != nil {
			if iaas.IsNotFoundError(err) {
				notFoundCount++
				if notFoundCount > w.NotFoundRetry {
					return fmt.Errorf("wait: resource not found after %d retries: %w", notFoundCount-1, err)
				}
			} else {
				return fmt.Errorf("wait: read failed: %w", err)
			}
		} else {
			notFoundCount = 0
			matched, checkErr := w.check(result)
			if checkErr != nil {
				return fmt.Errorf("wait: %w", checkErr)
			}
			if matched {
				return nil
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("wait: %w", ctx.Err())
		case <-time.After(interval):
		}
	}
}

func (w *StateWaiter) check(r StateResult) (bool, error) {
	availMatched, err := matchState("availability", r.Availability, w.TargetAvailability, w.PendingAvailability)
	if err != nil {
		return false, err
	}
	statusMatched, err := matchState("instance status", r.InstanceStatus, w.TargetInstanceStatus, w.PendingInstanceStatus)
	if err != nil {
		return false, err
	}
	return availMatched && statusMatched, nil
}

// matchState は current が target/pending のどちらに該当するかを返す。
// targets が空のときはチェックをスキップして true を返す。
func matchState(label, current string, targets, pending []string) (bool, error) {
	if len(targets) == 0 {
		return true, nil
	}
	if contains(targets, current) {
		return true, nil
	}
	if contains(pending, current) {
		return false, nil
	}
	return false, fmt.Errorf("unexpected %s: %q", label, current)
}

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

// SimpleStateWaiter は bool を返す ReadStateFunc ベースの単純な待機器。
// 「特定条件が満たされるまで poll する」用途の最小 API。
type SimpleStateWaiter struct {
	// ReadStateFunc は完了判定 func。true を返せば完了。false なら待機継続。
	ReadStateFunc func(ctx context.Context) (bool, error)
	// Timeout は全体タイムアウト (未指定時 DefaultTimeout)。
	Timeout time.Duration
	// PollingInterval は poll 間隔 (未指定時 DefaultInterval)。
	PollingInterval time.Duration
}

// Wait は ReadStateFunc が true を返すまで poll する。
func (s *SimpleStateWaiter) Wait(ctx context.Context) error {
	interval := s.PollingInterval
	if interval <= 0 {
		interval = DefaultInterval
	}
	timeout := s.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		done, err := s.ReadStateFunc(ctx)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

// ---------- resource-specific helpers ----------

// 目標/待機集合は v1 helper/wait の semantics に揃えている。
var (
	availTargetUp     = []string{"available"}
	availPendingUp    = []string{"", "unknown", "migrating", "uploading", "transferring", "discontinued"}
	availTargetDown   = []string{"available"}
	availPendingDown  = []string{"", "unknown"}
	statusTargetUp    = []string{"up"}
	statusPendingUp   = []string{"", "unknown", "cleaning", "down"}
	statusTargetDown  = []string{"down"}
	statusPendingDown = []string{"up", "cleaning", "unknown"}
	availTargetReady  = []string{"available"}
	availPendingReady = []string{"", "unknown", "migrating", "uploading", "transferring", "discontinued"}
)

// ArchiveReader は UntilArchiveIsReady が必要とする最小 interface。
// iaas.ArchiveAPI はこれを満たす。
type ArchiveReader interface {
	Read(ctx context.Context, id int64) (*client.ArchiveReadResponseEnvelope, error)
}

// UntilArchiveIsReady は Archive のコピー完了まで待機する。
func UntilArchiveIsReady(ctx context.Context, op ArchiveReader, id int64) (*client.Archive, error) {
	var last *client.Archive
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			resp, err := op.Read(ctx, id)
			if err != nil {
				return StateResult{}, err
			}
			last = &resp.Archive
			return StateResult{Availability: string(resp.Archive.Availability.Value)}, nil
		},
		TargetAvailability:  availTargetReady,
		PendingAvailability: availPendingReady,
	}
	return last, w.Wait(ctx)
}

// DiskReader は UntilDiskIsReady が必要とする最小 interface。
type DiskReader interface {
	Read(ctx context.Context, id int64) (*client.DiskReadResponseEnvelope, error)
}

// UntilDiskIsReady は Disk のコピー完了 / ディスク修正完了まで待機する。
func UntilDiskIsReady(ctx context.Context, op DiskReader, id int64) (*client.Disk, error) {
	var last *client.Disk
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			resp, err := op.Read(ctx, id)
			if err != nil {
				return StateResult{}, err
			}
			last = &resp.Disk
			return StateResult{Availability: string(resp.Disk.Availability.Value)}, nil
		},
		TargetAvailability:  availTargetReady,
		PendingAvailability: availPendingReady,
	}
	return last, w.Wait(ctx)
}

// ServerReader は UntilServerIs* が必要とする最小 interface。
type ServerReader interface {
	Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error)
}

// UntilServerIsUp は Server の起動完了まで待機する。
func UntilServerIsUp(ctx context.Context, op ServerReader, id int64) (*client.Server, error) {
	var last *client.Server
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			resp, err := op.Read(ctx, id)
			if err != nil {
				return StateResult{}, err
			}
			last = &resp.Server
			return StateResult{
				Availability:   string(resp.Server.Availability.Value),
				InstanceStatus: string(resp.Server.Instance.Value.Status.Value),
			}, nil
		},
		TargetAvailability:    availTargetUp,
		PendingAvailability:   availPendingUp,
		TargetInstanceStatus:  statusTargetUp,
		PendingInstanceStatus: statusPendingUp,
	}
	return last, w.Wait(ctx)
}

// UntilServerIsDown は Server のシャットダウン完了まで待機する。
func UntilServerIsDown(ctx context.Context, op ServerReader, id int64) (*client.Server, error) {
	var last *client.Server
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			resp, err := op.Read(ctx, id)
			if err != nil {
				return StateResult{}, err
			}
			last = &resp.Server
			return StateResult{
				Availability:   string(resp.Server.Availability.Value),
				InstanceStatus: string(resp.Server.Instance.Value.Status.Value),
			}, nil
		},
		TargetAvailability:    availTargetDown,
		PendingAvailability:   availPendingDown,
		TargetInstanceStatus:  statusTargetDown,
		PendingInstanceStatus: statusPendingDown,
	}
	return last, w.Wait(ctx)
}

// ApplianceReader は UntilApplianceIs* が必要とする最小 interface。
// iaas.ApplianceAPI はこれを満たす (Database / LoadBalancer / NFS / MobileGateway / VPCRouter
// の共通 Read を持つ)。
type ApplianceReader interface {
	Read(ctx context.Context, id int64) (*client.DatabaseReadResponseEnvelope, error)
}

func untilApplianceIs(ctx context.Context, op ApplianceReader, id int64, availTarget, availPending, statusTarget, statusPending []string, notFoundRetry int) (*client.Database, error) {
	var last *client.Database
	w := &StateWaiter{
		ReadFunc: func(ctx context.Context) (StateResult, error) {
			resp, err := op.Read(ctx, id)
			if err != nil {
				return StateResult{}, err
			}
			last = &resp.Appliance
			return StateResult{
				Availability:   string(resp.Appliance.Availability.Value),
				InstanceStatus: string(resp.Appliance.Instance.Value.Status.Value),
			}, nil
		},
		TargetAvailability:    availTarget,
		PendingAvailability:   availPending,
		TargetInstanceStatus:  statusTarget,
		PendingInstanceStatus: statusPending,
		NotFoundRetry:         notFoundRetry,
	}
	return last, w.Wait(ctx)
}

// UntilApplianceIsUp は Appliance (Database/LoadBalancer/NFS/MobileGateway/VPCRouter)
// の起動完了まで待機する。404 リトライを含む。
func UntilApplianceIsUp(ctx context.Context, op ApplianceReader, id int64) (*client.Database, error) {
	return untilApplianceIs(ctx, op, id, availTargetUp, availPendingUp, statusTargetUp, statusPendingUp, ApplianceNotFoundRetryCount)
}

// UntilApplianceIsDown は Appliance のシャットダウン完了まで待機する。
func UntilApplianceIsDown(ctx context.Context, op ApplianceReader, id int64) (*client.Database, error) {
	return untilApplianceIs(ctx, op, id, availTargetDown, availPendingDown, statusTargetDown, statusPendingDown, 0)
}

// UntilApplianceIsReady は Appliance の作成完了 (available かつ Instance 未起動を許容) まで待機する。
// VPCRouter のように create 後に Instance が自動起動しないアプライアンス向け。
func UntilApplianceIsReady(ctx context.Context, op ApplianceReader, id int64) (*client.Database, error) {
	return untilApplianceIs(ctx, op, id, availTargetReady, availPendingReady, nil, nil, ApplianceNotFoundRetryCount)
}
