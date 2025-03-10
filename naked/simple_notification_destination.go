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

package naked

import (
	"time"

	"github.com/sacloud/iaas-api-go/types"
)

// SimpleNotificationDestination シンプル通知
type SimpleNotificationDestination struct {
	ID           types.ID                               `json:",omitempty" yaml:"id,omitempty" structs:",omitempty"`
	Name         string                                 `json:",omitempty" yaml:"name,omitempty" structs:",omitempty"`
	Description  string                                 `yaml:"description"`
	Tags         types.Tags                             `yaml:"tags"`
	Icon         *Icon                                  `json:",omitempty" yaml:"icon,omitempty" structs:",omitempty"`
	CreatedAt    *time.Time                             `json:",omitempty" yaml:"created_at,omitempty" structs:",omitempty"`
	ModifiedAt   *time.Time                             `json:",omitempty" yaml:"modified_at,omitempty" structs:",omitempty"`
	Availability types.EAvailability                    `json:",omitempty" yaml:"availability,omitempty" structs:",omitempty"`
	Provider     *Provider                              `json:",omitempty" yaml:"provider,omitempty" structs:",omitempty"`
	Settings     *SimpleNotificationDestinationSettings `json:",omitempty" yaml:"settings,omitempty" structs:",omitempty"`
	SettingsHash string                                 `json:",omitempty" yaml:"settings_hash,omitempty" structs:",omitempty"`
	Status       *SimpleNotificationDestinationStatus   `json:",omitempty" yaml:"status" structs:",omitempty"`
	ServiceClass string                                 `json:",omitempty" yaml:"service_class,omitempty" structs:",omitempty"`
}

// SimpleNotificationDestinationSettingsUpdate シンプル通知更新パラメータ
type SimpleNotificationDestinationSettingsUpdate struct {
	Settings     *SimpleNotificationDestinationSettings `json:",omitempty" yaml:"settings,omitempty" structs:",omitempty"`
	SettingsHash string                                 `json:",omitempty" yaml:"settings_hash,omitempty" structs:",omitempty"`
}

// SimpleNotificationDestinationSettings セッティング
type SimpleNotificationDestinationSettings struct {
	Type     types.ESimpleNotificationDestinationTypes `json:",omitempty" yaml:",omitempty"`
	Value    string                                    `json:",omitempty" yaml:",omitempty"`
	Disabled bool
}

// SimpleNotificationDestinationStatus ステータス
type SimpleNotificationDestinationStatus struct {
	Disabled     bool
	ErrorMessage string `json:",omitempty" yaml:",omitempty"`
	IsValid      bool
}

// SimpleNotificationDestinationRunningStatus /statusの戻り値
type SimpleNotificationDestinationRunningStatus struct {
	IsValid    bool
	ModifiedAt *time.Time `json:",omitempty" yaml:"modified_at,omitempty" structs:",omitempty"`
}
