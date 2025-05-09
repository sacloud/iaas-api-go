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

package define

import (
	"github.com/sacloud/iaas-api-go/internal/define/names"
	"github.com/sacloud/iaas-api-go/internal/define/ops"
	"github.com/sacloud/iaas-api-go/internal/dsl"
	"github.com/sacloud/iaas-api-go/internal/dsl/meta"
	"github.com/sacloud/iaas-api-go/naked"
)

const (
	bridgeAPIName     = "Bridge"
	bridgeAPIPathName = "bridge"
)

var bridgeAPI = &dsl.Resource{
	Name:       bridgeAPIName,
	PathName:   bridgeAPIPathName,
	PathSuffix: dsl.CloudAPISuffix,
	Operations: dsl.Operations{
		// find
		ops.Find(bridgeAPIName, bridgeNakedType, findParameter, bridgeView),

		// create
		ops.Create(bridgeAPIName, bridgeNakedType, bridgeCreateParam, bridgeView),

		// read
		ops.Read(bridgeAPIName, bridgeNakedType, bridgeView),

		// update
		ops.Update(bridgeAPIName, bridgeNakedType, bridgeUpdateParam, bridgeView),

		// delete
		ops.Delete(bridgeAPIName),
	},
}

var (
	bridgeNakedType = meta.Static(naked.Bridge{})

	bridgeView = &dsl.Model{
		Name:      bridgeAPIName,
		NakedType: bridgeNakedType,
		Fields: []*dsl.FieldDesc{
			fields.ID(),
			fields.Name(),
			fields.Description(),
			fields.CreatedAt(),
			fields.Region(),
			fields.BridgeInfo(),
			fields.SwitchInZone(),
		},
	}

	bridgeCreateParam = &dsl.Model{
		Name:      names.CreateParameterName(bridgeAPIName),
		NakedType: bridgeNakedType,
		Fields: []*dsl.FieldDesc{
			fields.Name(),
			fields.Description(),
		},
	}

	bridgeUpdateParam = &dsl.Model{
		Name:      names.UpdateParameterName(bridgeAPIName),
		NakedType: bridgeNakedType,
		Fields: []*dsl.FieldDesc{
			fields.Name(),
			fields.Description(),
		},
	}
)
