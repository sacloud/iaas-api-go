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
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/defaults"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/stretchr/testify/require"
)

func TestPowerHandler(t *testing.T) {
	t.Parallel()

	defaultInterval := defaults.DefaultStatePollingInterval

	defaults.DefaultStatePollingInterval = 10 * time.Millisecond
	BootRetrySpan = time.Millisecond
	ShutdownRetrySpan = time.Millisecond
	defer func() {
		defaults.DefaultStatePollingInterval = defaultInterval
		BootRetrySpan = 0
		ShutdownRetrySpan = 0
	}()

	ctx := context.Background()
	t.Run("boot", func(t *testing.T) {
		handler := &dummyPowerHandler{
			ignoreBootCount: 3,
			instanceStatus:  types.ServerInstanceStatuses.Down,
		}
		err := boot(ctx, handler)
		require.NoError(t, err)
		require.Equal(t, handler.ignoreBootCount+1, handler.bootCount)
	})
	t.Run("shutdown", func(t *testing.T) {
		handler := &dummyPowerHandler{
			ignoreShutdownCount: 3,
			instanceStatus:      types.ServerInstanceStatuses.Up,
		}
		err := shutdown(ctx, handler, true)
		require.NoError(t, err)
		require.Equal(t, handler.ignoreShutdownCount+1, handler.shutdownCount)
	})
}

type dummyPowerHandler struct {
	bootCount           int
	shutdownCount       int
	ignoreBootCount     int
	ignoreShutdownCount int
	instanceStatus      types.EServerInstanceStatus

	mu sync.Mutex
}

func (d *dummyPowerHandler) boot() error {
	d.bootCount++
	if d.bootCount > d.ignoreBootCount {
		go d.toggleInstanceStatus()
		return iaas.NewAPIError("DUMMY", nil, http.StatusConflict, nil)
	}
	return nil
}
func (d *dummyPowerHandler) shutdown(force bool) error {
	d.shutdownCount++
	if d.shutdownCount > d.ignoreShutdownCount {
		go d.toggleInstanceStatus()
		return iaas.NewAPIError("DUMMY", nil, http.StatusConflict, nil)
	}
	return nil
}

func (d *dummyPowerHandler) read() (interface{}, error) {
	return d, nil
}

func (d *dummyPowerHandler) toggleInstanceStatus() {
	time.Sleep(100 * time.Millisecond)

	d.mu.Lock()
	defer d.mu.Unlock()

	switch d.instanceStatus {
	case types.ServerInstanceStatuses.Up:
		d.instanceStatus = types.ServerInstanceStatuses.Down
	case types.ServerInstanceStatuses.Down:
		d.instanceStatus = types.ServerInstanceStatuses.Up
	}
}

// GetInstanceStatus .
func (d *dummyPowerHandler) GetInstanceStatus() types.EServerInstanceStatus {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.instanceStatus
}

// SetInstanceStatus .
func (d *dummyPowerHandler) SetInstanceStatus(v types.EServerInstanceStatus) {
	d.instanceStatus = v
}

func TestPower_powerRequestWithRetry(t *testing.T) {
	InitialRequestRetrySpan = 1 * time.Millisecond
	InitialRequestTimeout = 100 * time.Millisecond

	// 最初のシャットダウンが受け入れられる(エラーにならない)まで409-still_creating時にリトライする
	// エラーなしの場合は即時return nilする
	t.Run("retry when received 409 and still_creating response", func(t *testing.T) {
		retried := 0
		maxRetry := 3
		err := powerRequestWithRetry(context.Background(), func() error {
			if retried < maxRetry {
				retried++
				return iaas.NewAPIError("GET", nil, http.StatusConflict, &iaas.APIErrorResponse{
					IsFatal:      true,
					Serial:       "xxx",
					Status:       "409 Conflict",
					ErrorCode:    "still_creating",
					ErrorMessage: "xxx",
				})
			}
			return nil
		})

		if err != nil {
			t.Fatalf("got unexpected error: %s", err)
		}
		if retried != maxRetry {
			t.Fatalf("powerRequest was not retried: expected: %d, actual: %d", maxRetry, retried)
		}
	})
	// 409時のリトライにはタイムアウトを設定する
	t.Run("retry when received 409 and still_creating should be timed out", func(t *testing.T) {
		err := powerRequestWithRetry(context.Background(), func() error {
			return iaas.NewAPIError("GET", nil, http.StatusConflict, &iaas.APIErrorResponse{
				IsFatal:      true,
				Serial:       "xxx",
				Status:       "409 Conflict",
				ErrorCode:    "still_creating",
				ErrorMessage: "xxx",
			})
		})

		require.EqualError(t, err, "powerRequestWithRetry: timed out: context deadline exceeded")
	})
	// その他のエラーは即時returnする
	t.Run("force return error when received unexpected error", func(t *testing.T) {
		expected := iaas.NewAPIError("GET", nil, http.StatusNotFound, &iaas.APIErrorResponse{
			IsFatal:      true,
			Serial:       "xxx",
			Status:       "404 NotFound",
			ErrorCode:    "not_found",
			ErrorMessage: "xxx",
		})
		err := powerRequestWithRetry(context.Background(), func() error { return expected })

		require.EqualValues(t, expected, err)
	})
}
