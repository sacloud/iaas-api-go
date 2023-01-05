// Copyright 2022-2023 The sacloud/iaas-api-go Authors
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
	"github.com/sacloud/iaas-api-go/internal/define/ops"
	"github.com/sacloud/iaas-api-go/internal/dsl"
	"github.com/sacloud/iaas-api-go/internal/dsl/meta"
	"github.com/sacloud/iaas-api-go/naked"
)

const (
	diskPlanAPIName     = "DiskPlan"
	diskPlanAPIPathName = "product/disk"
)

var diskPlanAPI = &dsl.Resource{
	Name:       diskPlanAPIName,
	PathName:   diskPlanAPIPathName,
	PathSuffix: dsl.CloudAPISuffix,
	Operations: dsl.Operations{
		ops.Find(diskPlanAPIName, diskPlanNakedType, findParameter, diskPlanView),
		ops.Read(diskPlanAPIName, diskPlanNakedType, diskPlanView),
	},
}

var (
	diskPlanNakedType = meta.Static(naked.DiskPlan{})
	diskPlanView      = &dsl.Model{
		Name:      diskPlanAPIName,
		NakedType: diskPlanNakedType,
		Fields: []*dsl.FieldDesc{
			fields.ID(),
			fields.Name(),
			fields.Def("StorageClass", meta.TypeString),
			fields.Availability(),
			{
				Name: "Size",
				Type: &dsl.Model{
					Name:      "DiskPlanSizeInfo",
					NakedType: meta.Static(naked.DiskPlanSizeInfo{}),
					IsArray:   true,
					Fields: []*dsl.FieldDesc{
						fields.Availability(),
						fields.Def("DisplaySize", meta.TypeInt),
						fields.Def("DisplaySuffix", meta.TypeString),
						fields.SizeMB(),
					},
				},
				Tags: &dsl.FieldTags{
					MapConv: "[]Size,recursive",
				},
			},
		},
	}
)
