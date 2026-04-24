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
	"time"
)

const (
	// DefaultTimeoutDuration は参照チェック待機のデフォルトタイムアウト。
	DefaultTimeoutDuration = 1 * time.Hour
	// DefaultTick は参照チェックの poll 間隔のデフォルト。
	DefaultTick = 5 * time.Second
)

// CheckReferencedOption は IsXxxReferenced の poll オプション。
type CheckReferencedOption struct {
	Timeout time.Duration
	Tick    time.Duration
}

// DefaultCheckReferencedOption は参照チェックのデフォルトオプション。
var DefaultCheckReferencedOption = CheckReferencedOption{
	Timeout: DefaultTimeoutDuration,
	Tick:    DefaultTick,
}

func (c *CheckReferencedOption) init() {
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeoutDuration
	}
	if c.Tick <= 0 {
		c.Tick = DefaultTick
	}
}

// waitWhileReferenced は f が true を返している間 poll する。
// f が false (参照なし) or error を返したら終了する。
func waitWhileReferenced(ctx context.Context, option CheckReferencedOption, f func() (bool, error)) error {
	option.init()

	ctx, cancel := context.WithTimeout(ctx, option.Timeout)
	defer cancel()

	if found, err := f(); !found || err != nil {
		return err
	}
	t := time.NewTicker(option.Tick)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if found, err := f(); !found || err != nil {
				return err
			}
		}
	}
}
