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

package fake

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/size"
)

var initOnce sync.Once

func ds() Store {
	InitDataStore()
	return DataStore
}

// DataStore fakeドライバが利用するデータストア
var DataStore Store

// InitDataStore データストアの初期化
func InitDataStore() {
	initOnce.Do(func() {
		if DataStore == nil {
			DataStore = NewInMemoryStore()
		}
		if err := DataStore.Init(); err != nil {
			panic(err)
		}

		// pool(id,ip,subnet,etc...)
		p := initValuePool(DataStore)

		if DataStore.NeedInitData() {
			if err := initData(DataStore, p); err != nil {
				panic(err)
			}
		}
	})
}

func initData(s Store, p *valuePool) error {
	initGeneratedArchives(s, p)
	initBills(s, p)
	initCoupons(s, p)
	initGeneratedCDROMs(s, p)
	initNotes(s, p)
	initSwitch(s, p)
	initZones(s, p)
	initRegions(s, p)
	initPrivateHostPlan(s, p)
	initDiskPlan(s, p)
	initLicensePlan(s, p)
	initInternetPlan(s, p)
	initServerPlan(s, p)
	initServiceClass(s, p)
	return nil
}

func initGeneratedArchives(s Store, p *valuePool) {
	for zone, archives := range initArchives { // generated by internal/tools/gen-api-fake-public-archive
		for _, archive := range archives {
			s.Put(ResourceArchive, zone, archive.ID, archive)
		}
	}
}

func initBills(s Store, p *valuePool) {
	bills := []*iaas.Bill{
		{
			ID:             p.generateID(),
			Amount:         1080,
			Date:           time.Now(),
			MemberID:       "dummy00000",
			Paid:           false,
			PayLimit:       time.Now().AddDate(0, 1, 0),
			PaymentClassID: 999,
		},
	}
	for _, bill := range bills {
		s.Put(ResourceBill, iaas.APIDefaultZone, bill.ID, bill)
		initBillDetails(s, p, bill.ID)
	}
}

func initBillDetails(s Store, p *valuePool, billID types.ID) {
	details := []*iaas.BillDetail{
		{
			ID:               p.generateID(),
			Amount:           108,
			Description:      "description",
			ServiceClassID:   999,
			ServiceClassPath: "fake/cloud/dummy",
			Usage:            100,
			FormattedUsage:   "1d, 2h",
			ServiceUsagePath: "fake/cloud/usetime",
			Zone:             "tk1a",
			ContractEndAt:    time.Now(),
		},
	}
	s.Put(ResourceBill+"Details", iaas.APIDefaultZone, billID, &details)
}

func initGeneratedCDROMs(s Store, p *valuePool) {
	for zone, cdroms := range initCDROM { // generated by internal/tools/gen-api-fake-data
		for _, cdrom := range cdroms {
			s.Put(ResourceCDROM, zone, cdrom.ID, cdrom)
		}
	}
}

func initCoupons(s Store, p *valuePool) {
	coupons := []*iaas.Coupon{
		{
			ID:             p.generateID(),
			MemberID:       "dummy00000",
			ContractID:     p.generateID(),
			ServiceClassID: 999,
			Discount:       20000,
			AppliedAt:      time.Now().AddDate(0, -1, 0),
			UntilAt:        time.Now().AddDate(0, 1, 0),
		},
	}
	for _, c := range coupons {
		s.Put(ResourceCoupon, iaas.APIDefaultZone, c.ID, c)
	}
}

func initNotes(s Store, p *valuePool) {
	notes := []*iaas.Note{
		{
			ID:    1,
			Name:  "sys-nfs",
			Class: "json",
			Content: `
{
	"plans":{
		"HDD":[
			{"size": 100,"availability":"available","planId":101},
			{"size": 500,"availability":"available","planId":102},
			{"size": 1024,"availability":"available","planId":103},
			{"size": 2048,"availability":"available","planId":104},
			{"size": 4096,"availability":"available","planId":105},
			{"size": 8192,"availability":"available","planId":106},
			{"size": 12288,"availability":"available","planId":107}
		],
		"SSD":[
			{"size": 20,"availability":"available","planId":200},
			{"size": 100,"availability":"available","planId":201},
			{"size": 500,"availability":"available","planId":202},
			{"size": 1024,"availability":"available","planId":203},
			{"size": 2048,"availability":"available","planId":204},
			{"size": 4096,"availability":"available","planId":205}
		]
	}
}`,
		},
		{
			ID:    2,
			Name:  "sys-database",
			Class: "json",
			Scope: "shared",
			Content: `
{
  "Products": [],
  "Backup": {
    "LockLimit": 8,
    "RotateLimit": 8
  },
  "AppliancePlans": [
    {
      "Class": "database",
      "Model": "Proxy",
      "CPU": 4,
      "MemoryMB": 4096,
      "DiskSizes": [
        {
          "SizeMB": 102400,
          "DisplaySize": 90,
          "PlanID": 50721,
          "ServiceClass": "cloud/appliance/database/4core4gb-100gb-proxy"
        },
        {
          "SizeMB": 256000,
          "DisplaySize": 240,
          "PlanID": 50722,
          "ServiceClass": "cloud/appliance/database/4core4gb-250gb-proxy"
        },
        {
          "SizeMB": 512000,
          "DisplaySize": 500,
          "PlanID": 50723,
          "ServiceClass": "cloud/appliance/database/4core4gb-500gb-proxy"
        },
        {
          "SizeMB": 1048576,
          "DisplaySize": 1000,
          "PlanID": 50724,
          "ServiceClass": "cloud/appliance/database/4core4gb-1tb-proxy"
        }
      ]
    },
    {
      "Class": "database",
      "Model": "Proxy",
      "CPU": 4,
      "MemoryMB": 8192,
      "DiskSizes": [
        {
          "SizeMB": 102400,
          "DisplaySize": 90,
          "PlanID": 50725,
          "ServiceClass": "cloud/appliance/database/4core8gb-100gb-proxy"
        },
        {
          "SizeMB": 256000,
          "DisplaySize": 240,
          "PlanID": 50726,
          "ServiceClass": "cloud/appliance/database/4core8gb-250gb-proxy"
        },
        {
          "SizeMB": 512000,
          "DisplaySize": 500,
          "PlanID": 50727,
          "ServiceClass": "cloud/appliance/database/4core8gb-500gb-proxy"
        },
        {
          "SizeMB": 1048576,
          "DisplaySize": 1000,
          "PlanID": 50728,
          "ServiceClass": "cloud/appliance/database/4core8gb-1tb-proxy"
        }
      ]
    }
  ]
}
`,
		},
	}
	for _, note := range notes {
		s.Put(ResourceNote, iaas.APIDefaultZone, note.ID, note)
	}
}

func initSwitch(s Store, p *valuePool) {
	sharedNetMask := net.CIDRMask(p.SharedNetMaskLen, 32)
	sharedNet := p.CurrentSharedIP.Mask(sharedNetMask)

	sharedSegmentSwitch = &iaas.Switch{
		ID:             p.generateID(),
		Name:           "スイッチ",
		Scope:          types.Scopes.Shared,
		Description:    "共有セグメント用スイッチ",
		NetworkMaskLen: p.SharedNetMaskLen,
		DefaultRoute:   p.SharedDefaultGateway.String(),
		Subnets: []*iaas.SwitchSubnet{
			{
				ID:             1,
				DefaultRoute:   p.SharedDefaultGateway.String(),
				NextHop:        p.SharedDefaultGateway.String(),
				StaticRoute:    p.SharedDefaultGateway.String(),
				NetworkAddress: sharedNet.String(),
				NetworkMaskLen: p.SharedNetMaskLen,
			},
		},
	}
	for _, zone := range zones {
		s.Put(ResourceSwitch, zone, sharedSegmentSwitch.ID, sharedSegmentSwitch)
	}
}

func initZones(s Store, p *valuePool) {
	// zones
	zones := []*iaas.Zone{
		{
			ID:           21001,
			Name:         "tk1a",
			Description:  "東京第1ゾーン",
			DisplayOrder: 1,
			Region: &iaas.Region{
				ID:          210,
				Name:        "東京",
				Description: "東京",
				NameServers: []string{
					"210.188.224.10",
					"210.188.224.11",
				},
			},
		},
		{
			ID:           21002,
			Name:         "tk1b",
			Description:  "東京第2ゾーン",
			DisplayOrder: 2,
			Region: &iaas.Region{
				ID:          210,
				Name:        "東京",
				Description: "東京",
				NameServers: []string{
					"210.188.224.10",
					"210.188.224.11",
				},
			},
		},
		{
			ID:           31001,
			Name:         "is1a",
			Description:  "石狩第1ゾーン",
			DisplayOrder: 3,
			Region: &iaas.Region{
				ID:          310,
				Name:        "石狩",
				Description: "石狩",
				NameServers: []string{
					"133.242.0.3",
					"133.242.0.4",
				},
			},
		},
		{
			ID:           31002,
			Name:         "is1b",
			Description:  "石狩第2ゾーン",
			DisplayOrder: 4,
			Region: &iaas.Region{
				ID:          310,
				Name:        "石狩",
				Description: "石狩",
				NameServers: []string{
					"133.242.0.3",
					"133.242.0.4",
				},
			},
		},
		{
			ID:           29001,
			Name:         "tk1v",
			Description:  "Sandbox",
			DisplayOrder: 5,
			IsDummy:      true,
			Region: &iaas.Region{
				ID:          210,
				Name:        "東京",
				Description: "東京",
				NameServers: []string{
					"210.188.224.10",
					"210.188.224.11",
				},
			},
		},
	}

	for _, zone := range zones {
		s.Put(ResourceZone, iaas.APIDefaultZone, zone.ID, zone)
	}
}

func initRegions(s Store, p *valuePool) {
	regions := []*iaas.Region{
		{
			ID:          210,
			Name:        "東京",
			Description: "東京",
			NameServers: []string{
				"210.188.224.10",
				"210.188.224.11",
			},
		},
		{
			ID:          290,
			Name:        "Sandbox",
			Description: "Sandbox",
			NameServers: []string{
				"133.242.0.3",
				"133.242.0.4",
			},
		},
		{
			ID:          310,
			Name:        "石狩",
			Description: "石狩",
			NameServers: []string{
				"133.242.0.3",
				"133.242.0.4",
			},
		},
	}

	for _, region := range regions {
		s.Put(ResourceRegion, iaas.APIDefaultZone, region.ID, region)
	}
}

func initPrivateHostPlan(s Store, p *valuePool) {
	plans := []*iaas.PrivateHostPlan{
		{
			ID:           112900526366,
			Name:         "200Core 224GB 標準",
			Class:        "dynamic",
			CPU:          200,
			MemoryMB:     229376,
			Availability: types.Availabilities.Available,
		},
		{
			ID:           113102196877,
			Name:         "200Core 224GB Windows",
			Class:        "ms_windows",
			CPU:          200,
			MemoryMB:     229376,
			Availability: types.Availabilities.Available,
		},
	}
	for _, zone := range zones {
		for _, plan := range plans {
			s.Put(ResourcePrivateHostPlan, zone, plan.ID, plan)
		}
	}
}

func initServerPlan(s Store, p *valuePool) {
	plans := []*iaas.ServerPlan{
		{
			ID:           p.generateID(),
			Name:         "プラン/1Core-1GB",
			CPU:          1,
			MemoryMB:     1 * size.MiB,
			GPU:          0,
			CPUModel:     "uncategorized",
			Commitment:   types.Commitments.Standard,
			Generation:   100,
			Availability: types.Availabilities.Available,
		},
		{
			ID:           p.generateID(),
			Name:         "プラン/2Core-4GB",
			CPU:          2,
			MemoryMB:     4 * size.MiB,
			GPU:          0,
			CPUModel:     "uncategorized",
			Commitment:   types.Commitments.Standard,
			Generation:   100,
			Availability: types.Availabilities.Available,
		},
		{
			ID:           p.generateID(),
			Name:         "GPUプラン/4Core-56GB-1GPU",
			CPU:          4,
			MemoryMB:     56 * 1024 * size.MiB,
			GPU:          1,
			CPUModel:     "uncategorized",
			Commitment:   types.Commitments.Standard,
			Generation:   200,
			Availability: types.Availabilities.Available,
		},
		{
			ID:           p.generateID(),
			Name:         "コア専有プラン/32Core-120GB",
			CPU:          32,
			MemoryMB:     120 * 1024 * size.MiB,
			GPU:          0,
			CPUModel:     "amd_epyc_7713p",
			Commitment:   types.Commitments.DedicatedCPU,
			Generation:   200,
			Availability: types.Availabilities.Available,
		},
		// TODO add more plans
	}

	for _, zone := range zones {
		for _, plan := range plans {
			s.Put(ResourceServerPlan, zone, plan.ID, plan)
		}
	}
}

func initInternetPlan(s Store, p *valuePool) {
	bandWidthList := []int{100, 250, 500, 1000, 1500, 2000, 2500, 3000, 5000}

	var plans []*iaas.InternetPlan

	for _, bw := range bandWidthList {
		plans = append(plans, &iaas.InternetPlan{
			ID:            types.ID(bw),
			BandWidthMbps: bw,
			Name:          fmt.Sprintf("%dMbps共有", bw),
			Availability:  types.Availabilities.Available,
		})
	}

	for _, zone := range zones {
		for _, plan := range plans {
			s.Put(ResourceInternetPlan, zone, plan.ID, plan)
		}
	}
}

func initDiskPlan(s Store, p *valuePool) {
	plans := []*iaas.DiskPlan{
		{
			ID:           2,
			Name:         "HDDプラン",
			Availability: types.Availabilities.Available,
			StorageClass: "iscsi1204",
			Size: []*iaas.DiskPlanSizeInfo{
				{
					Availability:  types.Availabilities.Available,
					DisplaySize:   20,
					DisplaySuffix: "GB",
					SizeMB:        20 * size.GiB,
				},
				{
					Availability:  types.Availabilities.Available,
					DisplaySize:   40,
					DisplaySuffix: "GB",
					SizeMB:        40 * size.GiB,
				},
			},
		},
		{
			ID:           4,
			Name:         "SSDプラン",
			Availability: types.Availabilities.Available,
			StorageClass: "iscsi1204",
			Size: []*iaas.DiskPlanSizeInfo{
				{
					Availability:  types.Availabilities.Available,
					DisplaySize:   20,
					DisplaySuffix: "GB",
					SizeMB:        20 * size.GiB,
				},
				{
					Availability:  types.Availabilities.Available,
					DisplaySize:   40,
					DisplaySuffix: "GB",
					SizeMB:        40 * size.GiB,
				},
			},
		},
		// TODO add more size-info
	}

	for _, zone := range zones {
		for _, plan := range plans {
			s.Put(ResourceDiskPlan, zone, plan.ID, plan)
		}
	}
}

func initLicensePlan(s Store, p *valuePool) {
	plans := []*iaas.LicenseInfo{
		{
			ID:         types.ID(10001),
			Name:       "Windows RDS SAL",
			TermsOfUse: "1ライセンスにつき、1人のユーザが利用できます。",
		},
	}

	for _, zone := range zones {
		for _, plan := range plans {
			s.Put(ResourceLicenseInfo, zone, plan.ID, plan)
		}
	}
}

func initServiceClass(s Store, p *valuePool) {
	classes := []*iaas.ServiceClass{
		{
			ID:               types.ID(50050),
			ServiceClassName: "plan/1",
			ServiceClassPath: "cloud/plan/1",
			DisplayName:      "プラン1(ディスクなし)",
			IsPublic:         true,
			Price: &iaas.Price{
				Base:    0,
				Daily:   108,
				Hourly:  10,
				Monthly: 2139,
			},
		},
		{
			ID:               types.ID(50051),
			ServiceClassName: "plan/2",
			ServiceClassPath: "cloud/plan/2",
			DisplayName:      "プラン2(ディスクなし)",
			IsPublic:         true,
			Price: &iaas.Price{
				Base:    0,
				Daily:   172,
				Hourly:  17,
				Monthly: 3425,
			},
		},
	}

	for _, zone := range zones {
		for _, class := range classes {
			class.Price.Zone = zone
			s.Put(ResourceServiceClass, zone, class.ID, class)
		}
	}
}
