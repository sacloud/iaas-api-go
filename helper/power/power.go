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

package power

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/accessor"
	"github.com/sacloud/iaas-api-go/defaults"
	"github.com/sacloud/iaas-api-go/types"
)

var (
	// BootRetrySpan 起動APIをコールしてからリトライするまでの待機時間
	BootRetrySpan time.Duration
	// ShutdownRetrySpan シャットダウンAPIをコールしてからリトライするまでの待機時間
	ShutdownRetrySpan time.Duration
	// InitialRequestTimeout 初回のBoot/Shutdownリクエストが受け入れられるまでのタイムアウト時間
	InitialRequestTimeout time.Duration
	// InitialRequestRetrySpan 初回のBoot/Shutdownリクエストをリトライする場合のリトライ間隔
	InitialRequestRetrySpan time.Duration
)

/************************************************
 * Server
 ***********************************************/

// BootServer 起動
//
// variablesが指定された場合、PUT /server/:id/powerのCloudInit用のパラメータとして渡される
// variablesが複数指定された場合は改行で結合される
func BootServer(ctx context.Context, client ServerAPI, zone string, id types.ID, variables ...string) error {
	return boot(ctx, &serverHandler{
		ctx:       ctx,
		client:    client,
		zone:      zone,
		id:        id,
		variables: variables,
	})
}

// ShutdownServer シャットダウン
func ShutdownServer(ctx context.Context, client ServerAPI, zone string, id types.ID, force bool) error {
	return shutdown(ctx, &serverHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	}, force)
}

/************************************************
 * LoadBalancer
 ***********************************************/

// BootLoadBalancer 起動
func BootLoadBalancer(ctx context.Context, client LoadBalancerAPI, zone string, id types.ID) error {
	return boot(ctx, &loadBalancerHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	})
}

// ShutdownLoadBalancer シャットダウン
func ShutdownLoadBalancer(ctx context.Context, client LoadBalancerAPI, zone string, id types.ID, force bool) error {
	return shutdown(ctx, &loadBalancerHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	}, force)
}

/************************************************
 * Database
 ***********************************************/

// BootDatabase 起動
func BootDatabase(ctx context.Context, client DatabaseAPI, zone string, id types.ID) error {
	return boot(ctx, &databaseHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	})
}

// ShutdownDatabase シャットダウン
func ShutdownDatabase(ctx context.Context, client DatabaseAPI, zone string, id types.ID, force bool) error {
	return shutdown(ctx, &databaseHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	}, force)
}

/************************************************
 * VPCRouter
 ***********************************************/

// BootVPCRouter 起動
func BootVPCRouter(ctx context.Context, client VPCRouterAPI, zone string, id types.ID) error {
	return boot(ctx, &vpcRouterHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	})
}

// ShutdownVPCRouter シャットダウン
func ShutdownVPCRouter(ctx context.Context, client VPCRouterAPI, zone string, id types.ID, force bool) error {
	return shutdown(ctx, &vpcRouterHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	}, force)
}

/************************************************
 * NFS
 ***********************************************/

// BootNFS 起動
func BootNFS(ctx context.Context, client NFSAPI, zone string, id types.ID) error {
	return boot(ctx, &nfsHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	})
}

// ShutdownNFS シャットダウン
func ShutdownNFS(ctx context.Context, client NFSAPI, zone string, id types.ID, force bool) error {
	return shutdown(ctx, &nfsHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	}, force)
}

/************************************************
 * MobileGateway
 ***********************************************/

// BootMobileGateway 起動
func BootMobileGateway(ctx context.Context, client MobileGatewayAPI, zone string, id types.ID) error {
	return boot(ctx, &mobileGatewayHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	})
}

// ShutdownMobileGateway シャットダウン
//
// HACK: forceオプションは現在指定不能になっているが、互換性維持のためにここの引数は残しておく
//
//	forceは指定しても利用されない
func ShutdownMobileGateway(ctx context.Context, client MobileGatewayAPI, zone string, id types.ID, force bool) error {
	return shutdown(ctx, &mobileGatewayHandler{
		ctx:    ctx,
		client: client,
		zone:   zone,
		id:     id,
	}, false)
}

type handler interface {
	boot() error
	shutdown(force bool) error
	read() (interface{}, error)
}

var mu sync.Mutex

func initDefaults() {
	mu.Lock()
	defer mu.Unlock()

	if BootRetrySpan == 0 {
		BootRetrySpan = defaults.DefaultPowerHelperBootRetrySpan
	}
	if ShutdownRetrySpan == 0 {
		ShutdownRetrySpan = defaults.DefaultPowerHelperShutdownRetrySpan
	}
	if InitialRequestTimeout == 0 {
		InitialRequestTimeout = defaults.DefaultPowerHelperInitialRequestTimeout
	}
	if InitialRequestRetrySpan == 0 {
		InitialRequestRetrySpan = defaults.DefaultPowerHelperInitialRequestRetrySpan
	}
}

func boot(ctx context.Context, h handler) error {
	initDefaults()

	// 初回リクエスト、409+still_creatingの場合は一定期間リトライする
	if err := powerRequestWithRetry(ctx, h.boot); err != nil {
		return err
	}

	retryTimer := time.NewTicker(BootRetrySpan)
	defer retryTimer.Stop()

	inProcess := false

	waiter := iaas.WaiterForUp(h.read)
	compCh, progressCh, errCh := waiter.WaitForStateAsync(ctx)

	var state interface{}

	for {
		select {
		case <-ctx.Done():
			return errors.New("canceled")
		case <-compCh:
			return nil
		case s := <-progressCh:
			state = s
		case <-retryTimer.C:
			if inProcess {
				continue
			}
			if state != nil && state.(accessor.InstanceStatus).GetInstanceStatus().IsDown() {
				if err := h.boot(); err != nil {
					if err, ok := err.(iaas.APIError); ok {
						// 初回リクエスト以降で409を受け取った場合はAPI側で受け入れ済とみなしこれ以上リトライしない
						if err.ResponseCode() == http.StatusConflict {
							inProcess = true
							continue
						}
					}
					return err
				}
			}
		case err := <-errCh:
			return err
		}
	}
}

func shutdown(ctx context.Context, h handler, force bool) error {
	initDefaults()

	// 初回リクエスト、409+still_creatingの場合は一定期間リトライする
	if err := powerRequestWithRetry(ctx, func() error { return h.shutdown(force) }); err != nil {
		return err
	}

	retryTimer := time.NewTicker(ShutdownRetrySpan)
	defer retryTimer.Stop()

	inProcess := false

	waiter := iaas.WaiterForDown(h.read)
	compCh, progressCh, errCh := waiter.WaitForStateAsync(ctx)

	var state interface{}

	for {
		select {
		case <-compCh:
			return nil
		case s := <-progressCh:
			state = s
		case <-retryTimer.C:
			if inProcess {
				continue
			}
			if state != nil && state.(accessor.InstanceStatus).GetInstanceStatus().IsUp() {
				if err := h.shutdown(force); err != nil {
					if err, ok := err.(iaas.APIError); ok {
						// 初回リクエスト以降で409を受け取った場合はAPI側で受け入れ済とみなしこれ以上リトライしない
						if err.ResponseCode() == http.StatusConflict {
							inProcess = true
							continue
						}
					}
					return err
				}
			}
		case err := <-errCh:
			return err
		}
	}
}

func powerRequestWithRetry(ctx context.Context, fn func() error) error {
	ctx, cancel := context.WithTimeout(ctx, InitialRequestTimeout)
	defer cancel()

	retryTimer := time.NewTicker(InitialRequestRetrySpan)
	defer retryTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			if err != nil {
				return fmt.Errorf("powerRequestWithRetry: timed out: %s", err)
			}
			return nil
		case <-retryTimer.C:
			err := fn()
			if err != nil {
				if iaas.IsStillCreatingError(err) {
					continue
				}
				return err
			}
			return nil
		}
	}
}
