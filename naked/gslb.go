// Copyright 2022 The sacloud/iaas-api-go Authors
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

// GSLB GSLB
type GSLB struct {
	ID           types.ID            `json:",omitempty" yaml:"id,omitempty" structs:",omitempty"`
	Name         string              `json:",omitempty" yaml:"name,omitempty" structs:",omitempty"`
	Description  string              `yaml:"description"`
	Tags         types.Tags          `yaml:"tags"`
	Icon         *Icon               `json:",omitempty" yaml:"icon,omitempty" structs:",omitempty"`
	CreatedAt    *time.Time          `json:",omitempty" yaml:"created_at,omitempty" structs:",omitempty"`
	ModifiedAt   *time.Time          `json:",omitempty" yaml:"modified_at,omitempty" structs:",omitempty"`
	Availability types.EAvailability `json:",omitempty" yaml:"availability,omitempty" structs:",omitempty"`
	ServiceClass string              `json:",omitempty" yaml:"service_class,omitempty" structs:",omitempty"`
	Provider     *Provider           `json:",omitempty" yaml:"provider,omitempty" structs:",omitempty"`
	Settings     *GSLBSettings       `json:",omitempty" yaml:"settings,omitempty" structs:",omitempty"`
	SettingsHash string              `json:",omitempty" yaml:"settings_hash,omitempty" structs:",omitempty"`
	Status       *GSLBStatus         `json:",omitempty" yaml:"status,omitempty" structs:",omitempty"`
}

// GSLBSettingsUpdate GSLB
type GSLBSettingsUpdate struct {
	Settings     *GSLBSettings `json:",omitempty" yaml:"settings,omitempty" structs:",omitempty"`
	SettingsHash string        `json:",omitempty" yaml:"settings_hash,omitempty" structs:",omitempty"`
}

// GSLBSettings GSLB?????????
type GSLBSettings struct {
	GSLB *GSLBSetting `json:",omitempty" yaml:"gslb,omitempty" structs:",omitempty"`
}

// GSLBSetting GSLB?????????
type GSLBSetting struct {
	DelayLoop   int              `json:",omitempty" yaml:"delay_loop,omitempty" structs:",omitempty"`
	HealthCheck *GSLBHealthCheck `json:",omitempty" yaml:"health_check,omitempty" structs:",omitempty"`
	Weighted    types.StringFlag `yaml:"weighted"`
	Servers     []*GSLBServer    `yaml:"servers"`
	SorryServer string           `json:",omitempty" yaml:",omitempty" structs:",omitempty"` // ????????????????????????
}

// GSLBHealthCheck ?????????????????????
type GSLBHealthCheck struct {
	Protocol types.Protocol     `json:",omitempty" yaml:"protocol,omitempty" structs:""` // ???????????????
	Host     string             `json:",omitempty" yaml:"host,omitempty" structs:""`     // ???????????????
	Path     string             `json:",omitempty" yaml:"path,omitempty" structs:""`     // HTTP/HTTPS?????????????????????????????????
	Status   types.StringNumber `json:",omitempty" yaml:"status,omitempty" structs:""`   // ????????????????????????????????????
	Port     types.StringNumber `json:",omitempty" yaml:"port,omitempty" structs:""`     // ???????????????
}

// GSLBServer GSLB?????????????????????
type GSLBServer struct {
	IPAddress string             `json:",omitempty" yaml:"ip_address,omitempty" structs:",omitempty"` // IP????????????
	Enabled   types.StringFlag   `yaml:"enabled" `                                                    // ??????/??????
	Weight    types.StringNumber `json:",omitempty" yaml:"weight,omitempty" structs:",omitempty"`     // ????????????
}

// GSLBStatus GSLB???????????????
type GSLBStatus struct {
	FQDN string `json:",omitempty" yaml:"fqdn,omitempty" structs:",omitempty"`
}
