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

// Package power は v2 リソースの起動 / シャットダウン制御を行う。
//
// v1 の github.com/sacloud/iaas-api-go/helper/power の v2 相当。
//
// 主な責務:
//   - Boot / Shutdown API の呼び出し (409 still_creating に対するリトライ含む)
//   - 目標状態 (Up / Down) への遷移を待機
//   - 途中で遷移しない場合の Boot / Shutdown 再送 (v1 互換動作)
package power

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
)

var (
	// BootRetrySpan は Boot API 呼出後、Instance が Down のままだった場合に再送するまでの間隔。
	BootRetrySpan time.Duration
	// ShutdownRetrySpan は Shutdown API 呼出後、Instance が Up のままだった場合に再送するまでの間隔。
	ShutdownRetrySpan time.Duration
	// InitialRequestTimeout は初回の Boot / Shutdown リクエストが still_creating で弾かれ続ける場合の全体タイムアウト。
	InitialRequestTimeout time.Duration
	// InitialRequestRetrySpan は初回 Boot / Shutdown が still_creating で弾かれたときのリトライ間隔。
	InitialRequestRetrySpan time.Duration
	// PollingInterval は Boot / Shutdown 後の状態 poll 間隔。
	PollingInterval time.Duration
	// OverallTimeout は Boot / Shutdown 全体 (初回 API + 状態遷移待ち) のタイムアウト。
	OverallTimeout time.Duration
)

const (
	defaultBootRetrySpan           = 20 * time.Second
	defaultShutdownRetrySpan       = 20 * time.Second
	defaultInitialRequestTimeout   = 30 * time.Minute
	defaultInitialRequestRetrySpan = 10 * time.Second
	defaultPollingInterval         = 5 * time.Second
	defaultOverallTimeout          = 30 * time.Minute
)

var defaultsMu sync.Mutex

func initDefaults() {
	defaultsMu.Lock()
	defer defaultsMu.Unlock()
	if BootRetrySpan == 0 {
		BootRetrySpan = defaultBootRetrySpan
	}
	if ShutdownRetrySpan == 0 {
		ShutdownRetrySpan = defaultShutdownRetrySpan
	}
	if InitialRequestTimeout == 0 {
		InitialRequestTimeout = defaultInitialRequestTimeout
	}
	if InitialRequestRetrySpan == 0 {
		InitialRequestRetrySpan = defaultInitialRequestRetrySpan
	}
	if PollingInterval == 0 {
		PollingInterval = defaultPollingInterval
	}
	if OverallTimeout == 0 {
		OverallTimeout = defaultOverallTimeout
	}
}

// handler は Boot / Shutdown / Read を抽象化する内部 interface。
// Server と Appliance で Boot / Shutdown の引数型が異なるため、ここで吸収する。
type handler interface {
	boot(ctx context.Context) error
	shutdown(ctx context.Context) error
	read(ctx context.Context) (available bool, instanceUp bool, err error)
}

// ---------- Server ----------

// ServerPowerAPI は BootServer / ShutdownServer が必要とする最小 interface。
// iaas.ServerAPI はこれを満たす。
type ServerPowerAPI interface {
	Boot(ctx context.Context, id int64, request *client.ServerBootRequestEnvelope) error
	Shutdown(ctx context.Context, id int64, request *client.ServerShutdownRequestEnvelope) error
	Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error)
}

// BootServer は Server を起動する。
// variables が指定された場合は CloudInit UserData として Boot リクエストに乗せる。
func BootServer(ctx context.Context, op ServerPowerAPI, id int64, variables ...string) error {
	req := &client.ServerBootRequestEnvelope{}
	if len(variables) > 0 {
		userData := joinWithNewline(variables)
		req.UserBootVariables = client.NewOptServerBootVariables(client.ServerBootVariables{
			CloudInit: client.ServerBootVariablesCloudInit{UserData: userData},
		})
	}
	h := &serverHandler{op: op, id: id, bootReq: req}
	return boot(ctx, h)
}

// ShutdownServer は Server をシャットダウンする。force=true で強制停止。
func ShutdownServer(ctx context.Context, op ServerPowerAPI, id int64, force bool) error {
	req := &client.ServerShutdownRequestEnvelope{Force: force}
	h := &serverHandler{op: op, id: id, shutdownReq: req}
	return shutdown(ctx, h)
}

type serverHandler struct {
	op          ServerPowerAPI
	id          int64
	bootReq     *client.ServerBootRequestEnvelope
	shutdownReq *client.ServerShutdownRequestEnvelope
}

func (h *serverHandler) boot(ctx context.Context) error {
	req := h.bootReq
	if req == nil {
		req = &client.ServerBootRequestEnvelope{}
	}
	return h.op.Boot(ctx, h.id, req)
}

func (h *serverHandler) shutdown(ctx context.Context) error {
	req := h.shutdownReq
	if req == nil {
		req = &client.ServerShutdownRequestEnvelope{}
	}
	return h.op.Shutdown(ctx, h.id, req)
}

func (h *serverHandler) read(ctx context.Context) (bool, bool, error) {
	resp, err := h.op.Read(ctx, h.id)
	if err != nil {
		return false, false, err
	}
	avail := string(resp.Server.Availability.Value) == "available"
	up := string(resp.Server.Instance.Value.Status.Value) == "up"
	return avail, up, nil
}

// ---------- Appliance (Database / LoadBalancer / NFS / MobileGateway / VPCRouter) ----------

// AppliancePowerAPI は BootAppliance / ShutdownAppliance が必要とする最小 interface。
// iaas.ApplianceAPI はこれを満たす。
type AppliancePowerAPI interface {
	Boot(ctx context.Context, id int64) error
	Shutdown(ctx context.Context, id int64, request *client.ShutdownOption) error
	Read(ctx context.Context, id int64) (*client.DatabaseReadResponseEnvelope, error)
}

// BootAppliance は Appliance (Database / LoadBalancer / NFS / MobileGateway / VPCRouter) を起動する。
func BootAppliance(ctx context.Context, op AppliancePowerAPI, id int64) error {
	h := &applianceHandler{op: op, id: id}
	return boot(ctx, h)
}

// ShutdownAppliance は Appliance をシャットダウンする。force=true で強制停止。
func ShutdownAppliance(ctx context.Context, op AppliancePowerAPI, id int64, force bool) error {
	h := &applianceHandler{op: op, id: id, force: force}
	return shutdown(ctx, h)
}

type applianceHandler struct {
	op    AppliancePowerAPI
	id    int64
	force bool
}

func (h *applianceHandler) boot(ctx context.Context) error {
	return h.op.Boot(ctx, h.id)
}

func (h *applianceHandler) shutdown(ctx context.Context) error {
	return h.op.Shutdown(ctx, h.id, &client.ShutdownOption{Force: h.force})
}

func (h *applianceHandler) read(ctx context.Context) (bool, bool, error) {
	resp, err := h.op.Read(ctx, h.id)
	if err != nil {
		return false, false, err
	}
	avail := string(resp.Appliance.Availability.Value) == "available"
	up := string(resp.Appliance.Instance.Value.Status.Value) == "up"
	return avail, up, nil
}

// ---------- internal boot / shutdown logic ----------

// boot は h.boot を呼んだ後、InstanceStatus=up になるまで poll する。
// 途中で Down のまま時間が経過したら boot を再送する (409 が返ったら以降は再送を止める)。
func boot(ctx context.Context, h handler) error {
	initDefaults()

	if err := callWithStillCreatingRetry(ctx, h.boot); err != nil {
		return err
	}

	return pollUntil(ctx, h, targetUp, BootRetrySpan)
}

// shutdown は h.shutdown を呼んだ後、InstanceStatus=down になるまで poll する。
func shutdown(ctx context.Context, h handler) error {
	initDefaults()

	if err := callWithStillCreatingRetry(ctx, h.shutdown); err != nil {
		return err
	}

	return pollUntil(ctx, h, targetDown, ShutdownRetrySpan)
}

type targetState int

const (
	targetUp targetState = iota
	targetDown
)

// pollUntil は目標状態に達するまで read し続ける。
// 定期的に (retrySpan ごとに) 状態が逆方向のままであれば boot/shutdown を再送する。
func pollUntil(ctx context.Context, h handler, target targetState, retrySpan time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, OverallTimeout)
	defer cancel()

	lastRetry := time.Now()
	retryDisabled := false

	for {
		avail, up, err := h.read(ctx)
		if err != nil {
			if !iaas.IsNotFoundError(err) {
				return fmt.Errorf("power: read failed: %w", err)
			}
			// 404 はアプライアンスで create 直後に発生し得る。poll を継続。
		} else {
			if avail && reachedTarget(target, up) {
				return nil
			}
			// 逆方向状態 (Boot 中に down / Shutdown 中に up) が retrySpan 以上続いたら再送
			if !retryDisabled && avail && needRetry(target, up) && time.Since(lastRetry) >= retrySpan {
				retryErr := retrySend(ctx, h, target)
				if retryErr != nil {
					var e *iaas.Error
					if errors.As(retryErr, &e) && e.ResponseCode() == http.StatusConflict {
						// 409 = API 側は受け入れ済、以降は再送せず poll のみ
						retryDisabled = true
					} else {
						return fmt.Errorf("power: retry failed: %w", retryErr)
					}
				}
				lastRetry = time.Now()
			}
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("power: %w", ctx.Err())
		case <-time.After(PollingInterval):
		}
	}
}

func reachedTarget(target targetState, up bool) bool {
	switch target {
	case targetUp:
		return up
	case targetDown:
		return !up
	}
	return false
}

// needRetry は「目標と逆方向の状態のまま」の判定。
// Boot (targetUp) なのに Down → retry 送る。Shutdown (targetDown) なのに Up → retry。
func needRetry(target targetState, up bool) bool {
	switch target {
	case targetUp:
		return !up
	case targetDown:
		return up
	}
	return false
}

func retrySend(ctx context.Context, h handler, target targetState) error {
	if target == targetUp {
		return h.boot(ctx)
	}
	return h.shutdown(ctx)
}

// callWithStillCreatingRetry は fn を呼び、409+still_creating の場合は InitialRequestRetrySpan
// で再試行する。InitialRequestTimeout を超えるとエラー。
func callWithStillCreatingRetry(ctx context.Context, fn func(context.Context) error) error {
	deadline := time.Now().Add(InitialRequestTimeout)
	for {
		err := fn(ctx)
		if err == nil {
			return nil
		}
		if !isStillCreatingError(err) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("power: still_creating retry timed out after %s", InitialRequestTimeout)
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("power: %w", ctx.Err())
		case <-time.After(InitialRequestRetrySpan):
		}
	}
}

// isStillCreatingError は 409 + error_code="still_creating" を検出する。
// v1 iaas.IsStillCreatingError の v2 相当だが、downstream には公開せず内部ヘルパとする。
func isStillCreatingError(err error) bool {
	var e *iaas.Error
	if !errors.As(err, &e) {
		return false
	}
	return e.ResponseCode() == http.StatusConflict && e.Code() == "still_creating"
}

// joinWithNewline は variables を改行で連結する。CloudInit UserData 向け。
func joinWithNewline(variables []string) string {
	return strings.Join(variables, "\n")
}
