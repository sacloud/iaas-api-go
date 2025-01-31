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

package test

import (
	"testing"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/power"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
)

func TestLoadBalancerOp_CRUD(t *testing.T) {
	testutil.RunCRUD(t, &testutil.CRUDTestCase{
		Parallel: true,

		SetupAPICallerFunc: singletonAPICaller,
		Setup: setupSwitchFunc("lb",
			createLoadBalancerParam,
			createLoadBalancerExpected,
			updateLoadBalancerExpected,
			updateLoadBalancerSettingsExpected,
			updateLoadBalancerToMin1Expected,
			updateLoadBalancerToMin2Expected,
		),
		Create: &testutil.CRUDTestFunc{
			Func: testLoadBalancerCreate,
			CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
				ExpectValue:  createLoadBalancerExpected,
				IgnoreFields: ignoreLoadBalancerFields,
			}),
		},

		Read: &testutil.CRUDTestFunc{
			Func: testLoadBalancerRead,
			CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
				ExpectValue:  createLoadBalancerExpected,
				IgnoreFields: ignoreLoadBalancerFields,
			}),
		},

		Updates: []*testutil.CRUDTestFunc{
			{
				Func: testLoadBalancerUpdate,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateLoadBalancerExpected,
					IgnoreFields: ignoreLoadBalancerFields,
				}),
			},
			{
				Func: testLoadBalancerUpdateSettings,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateLoadBalancerSettingsExpected,
					IgnoreFields: ignoreLoadBalancerFields,
				}),
			},
			{
				Func: testLoadBalancerUpdateToMin1,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateLoadBalancerToMin1Expected,
					IgnoreFields: ignoreLoadBalancerFields,
				}),
			},
			{
				Func: testLoadBalancerUpdateToMin2,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateLoadBalancerToMin2Expected,
					IgnoreFields: ignoreLoadBalancerFields,
				}),
			},
		},

		Shutdown: func(ctx *testutil.CRUDTestContext, caller iaas.APICaller) error {
			client := iaas.NewLoadBalancerOp(caller)
			return power.ShutdownLoadBalancer(ctx, client, testZone, ctx.ID, true)
		},

		Delete: &testutil.CRUDTestDeleteFunc{
			Func: testLoadBalancerDelete,
		},

		Cleanup: cleanupSwitchFunc("lb"),
	})
}

var (
	ignoreLoadBalancerFields = []string{
		"ID",
		"Class",
		"Availability",
		"InstanceStatus",
		"InstanceHostName",
		"InstanceHostInfoURL",
		"InstanceStatusChangedAt",
		"Interfaces",
		"Switch",
		"ZoneID",
		"CreatedAt",
		"ModifiedAt",
		"SettingsHash",
	}

	createLoadBalancerParam = &iaas.LoadBalancerCreateRequest{
		PlanID:         types.LoadBalancerPlans.HighSpec,
		VRID:           100,
		IPAddresses:    []string{"192.168.0.11", "192.168.0.12"},
		NetworkMaskLen: 24,
		DefaultRoute:   "192.168.0.1",
		Name:           testutil.ResourceName("lb"),
		Description:    "desc",
		Tags:           []string{"tag1", "tag2"},
		VirtualIPAddresses: []*iaas.LoadBalancerVirtualIPAddress{
			{
				VirtualIPAddress: "192.168.0.101",
				Port:             types.StringNumber(80),
				DelayLoop:        types.StringNumber(10),
				SorryServer:      "192.168.0.2",
				Description:      "vip1 desc",
				Servers: []*iaas.LoadBalancerServer{
					{
						IPAddress: "192.168.0.201",
						Port:      80,
						Enabled:   true,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:       types.LoadBalancerHealthCheckProtocols.HTTP,
							Path:           "/index.html",
							ResponseCode:   200,
							Retry:          2,
							ConnectTimeout: 6,
						},
					},
					{
						IPAddress: "192.168.0.202",
						Port:      80,
						Enabled:   true,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
							Path:         "/index.html",
							ResponseCode: 200,
						},
					},
				},
			},
			{
				VirtualIPAddress: "192.168.0.102",
				Port:             80,
				DelayLoop:        10,
				SorryServer:      "192.168.0.2",
				Description:      "vip2 desc",
				Servers: []*iaas.LoadBalancerServer{
					{
						IPAddress: "192.168.0.203",
						Port:      80,
						Enabled:   true,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
							Path:         "/index.html",
							ResponseCode: 200,
						},
					},
					{
						IPAddress: "192.168.0.204",
						Port:      80,
						Enabled:   true,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTP,
							Path:         "/index.html",
							ResponseCode: 200,
						},
					},
				},
			},
		},
	}
	createLoadBalancerExpected = &iaas.LoadBalancer{
		Name:               createLoadBalancerParam.Name,
		Description:        createLoadBalancerParam.Description,
		Tags:               createLoadBalancerParam.Tags,
		Availability:       types.Availabilities.Available,
		InstanceStatus:     types.ServerInstanceStatuses.Up,
		PlanID:             createLoadBalancerParam.PlanID,
		DefaultRoute:       createLoadBalancerParam.DefaultRoute,
		NetworkMaskLen:     createLoadBalancerParam.NetworkMaskLen,
		IPAddresses:        createLoadBalancerParam.IPAddresses,
		VRID:               createLoadBalancerParam.VRID,
		VirtualIPAddresses: createLoadBalancerParam.VirtualIPAddresses,
	}
	updateLoadBalancerParam = &iaas.LoadBalancerUpdateRequest{
		Name:        testutil.ResourceName("lb-upd"),
		Tags:        []string{"tag1-upd", "tag2-upd"},
		Description: "desc-upd",
		IconID:      testIconID,
		VirtualIPAddresses: []*iaas.LoadBalancerVirtualIPAddress{
			{
				VirtualIPAddress: "192.168.0.111",
				Port:             81,
				DelayLoop:        11,
				SorryServer:      "192.168.0.3",
				Description:      "vip1 desc-upd",
				Servers: []*iaas.LoadBalancerServer{
					{
						IPAddress: "192.168.0.211",
						Port:      81,
						Enabled:   false,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTPS,
							Path:         "/index-upd.html",
							ResponseCode: 201,
						},
					},
					{
						IPAddress: "192.168.0.212",
						Port:      81,
						Enabled:   false,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:       types.LoadBalancerHealthCheckProtocols.HTTPS,
							Path:           "/index-upd.html",
							ResponseCode:   201,
							Retry:          3,
							ConnectTimeout: 12,
						},
					},
				},
			},
			{
				VirtualIPAddress: "192.168.0.112",
				Port:             81,
				DelayLoop:        11,
				SorryServer:      "192.168.0.3",
				Description:      "vip2 desc-upd",
				Servers: []*iaas.LoadBalancerServer{
					{
						IPAddress: "192.168.0.213",
						Port:      81,
						Enabled:   false,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTPS,
							Path:         "/index-upd.html",
							ResponseCode: 201,
						},
					},
					{
						IPAddress: "192.168.0.214",
						Port:      81,
						Enabled:   false,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTPS,
							Path:         "/index-upd.html",
							ResponseCode: 201,
						},
					},
				},
			},
		},
	}
	updateLoadBalancerExpected = &iaas.LoadBalancer{
		Name:               updateLoadBalancerParam.Name,
		Description:        updateLoadBalancerParam.Description,
		Tags:               updateLoadBalancerParam.Tags,
		IconID:             testIconID,
		Availability:       types.Availabilities.Available,
		PlanID:             createLoadBalancerParam.PlanID,
		InstanceStatus:     types.ServerInstanceStatuses.Up,
		DefaultRoute:       createLoadBalancerParam.DefaultRoute,
		NetworkMaskLen:     createLoadBalancerParam.NetworkMaskLen,
		IPAddresses:        createLoadBalancerParam.IPAddresses,
		VRID:               createLoadBalancerParam.VRID,
		VirtualIPAddresses: updateLoadBalancerParam.VirtualIPAddresses,
	}
	updateLoadBalancerSettingsParam = &iaas.LoadBalancerUpdateSettingsRequest{
		VirtualIPAddresses: []*iaas.LoadBalancerVirtualIPAddress{
			{
				VirtualIPAddress: "192.168.0.121",
				Port:             82,
				DelayLoop:        11,
				SorryServer:      "192.168.0.4",
				Description:      "vip1 desc-upd",
				Servers: []*iaas.LoadBalancerServer{
					{
						IPAddress: "192.168.0.221",
						Port:      82,
						Enabled:   false,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTPS,
							Path:         "/index-upd.html",
							ResponseCode: 201,
						},
					},
					{
						IPAddress: "192.168.0.222",
						Port:      82,
						Enabled:   false,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTPS,
							Path:         "/index-upd.html",
							ResponseCode: 201,
						},
					},
				},
			},
			{
				VirtualIPAddress: "192.168.0.122",
				Port:             82,
				DelayLoop:        11,
				SorryServer:      "192.168.0.4",
				Description:      "vip2 desc-upd",
				Servers: []*iaas.LoadBalancerServer{
					{
						IPAddress: "192.168.0.223",
						Port:      82,
						Enabled:   false,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTPS,
							Path:         "/index-upd.html",
							ResponseCode: 201,
						},
					},
					{
						IPAddress: "192.168.0.224",
						Port:      82,
						Enabled:   false,
						HealthCheck: &iaas.LoadBalancerServerHealthCheck{
							Protocol:     types.LoadBalancerHealthCheckProtocols.HTTPS,
							Path:         "/index-upd.html",
							ResponseCode: 201,
						},
					},
				},
			},
		},
	}
	updateLoadBalancerSettingsExpected = &iaas.LoadBalancer{
		Name:               updateLoadBalancerParam.Name,
		Description:        updateLoadBalancerParam.Description,
		Tags:               updateLoadBalancerParam.Tags,
		IconID:             testIconID,
		Availability:       types.Availabilities.Available,
		PlanID:             createLoadBalancerParam.PlanID,
		InstanceStatus:     types.ServerInstanceStatuses.Up,
		DefaultRoute:       createLoadBalancerParam.DefaultRoute,
		NetworkMaskLen:     createLoadBalancerParam.NetworkMaskLen,
		IPAddresses:        createLoadBalancerParam.IPAddresses,
		VRID:               createLoadBalancerParam.VRID,
		VirtualIPAddresses: updateLoadBalancerSettingsParam.VirtualIPAddresses,
	}
	updateLoadBalancerToMin1Param = &iaas.LoadBalancerUpdateRequest{
		Name: testutil.ResourceName("lb-to-min1"),
		VirtualIPAddresses: []*iaas.LoadBalancerVirtualIPAddress{
			{
				VirtualIPAddress: "192.168.0.111",
				Port:             80,
				Servers:          iaas.LoadBalancerServers{},
			},
		},
	}
	updateLoadBalancerToMin1Expected = &iaas.LoadBalancer{
		Name:           updateLoadBalancerToMin1Param.Name,
		Availability:   types.Availabilities.Available,
		PlanID:         createLoadBalancerParam.PlanID,
		InstanceStatus: types.ServerInstanceStatuses.Up,
		DefaultRoute:   createLoadBalancerParam.DefaultRoute,
		NetworkMaskLen: createLoadBalancerParam.NetworkMaskLen,
		IPAddresses:    createLoadBalancerParam.IPAddresses,
		VRID:           createLoadBalancerParam.VRID,
		VirtualIPAddresses: []*iaas.LoadBalancerVirtualIPAddress{
			{
				VirtualIPAddress: "192.168.0.111",
				Port:             80,
				DelayLoop:        10, // default value
				Servers:          []*iaas.LoadBalancerServer{},
			},
		},
	}
	updateLoadBalancerToMin2Param = &iaas.LoadBalancerUpdateRequest{
		Name:               testutil.ResourceName("lb-to-min2"),
		VirtualIPAddresses: iaas.LoadBalancerVirtualIPAddresses{},
	}
	updateLoadBalancerToMin2Expected = &iaas.LoadBalancer{
		Name:               updateLoadBalancerToMin2Param.Name,
		Availability:       types.Availabilities.Available,
		PlanID:             createLoadBalancerParam.PlanID,
		InstanceStatus:     types.ServerInstanceStatuses.Up,
		DefaultRoute:       createLoadBalancerParam.DefaultRoute,
		NetworkMaskLen:     createLoadBalancerParam.NetworkMaskLen,
		IPAddresses:        createLoadBalancerParam.IPAddresses,
		VRID:               createLoadBalancerParam.VRID,
		VirtualIPAddresses: iaas.LoadBalancerVirtualIPAddresses{},
	}
)

func testLoadBalancerCreate(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewLoadBalancerOp(caller)
	return client.Create(ctx, testZone, createLoadBalancerParam)
}

func testLoadBalancerRead(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewLoadBalancerOp(caller)
	return client.Read(ctx, testZone, ctx.ID)
}

func testLoadBalancerUpdate(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewLoadBalancerOp(caller)
	return client.Update(ctx, testZone, ctx.ID, updateLoadBalancerParam)
}

func testLoadBalancerUpdateSettings(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewLoadBalancerOp(caller)
	return client.UpdateSettings(ctx, testZone, ctx.ID, updateLoadBalancerSettingsParam)
}

func testLoadBalancerUpdateToMin1(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewLoadBalancerOp(caller)
	return client.Update(ctx, testZone, ctx.ID, updateLoadBalancerToMin1Param)
}

func testLoadBalancerUpdateToMin2(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewLoadBalancerOp(caller)
	return client.Update(ctx, testZone, ctx.ID, updateLoadBalancerToMin2Param)
}

func testLoadBalancerDelete(ctx *testutil.CRUDTestContext, caller iaas.APICaller) error {
	client := iaas.NewLoadBalancerOp(caller)
	return client.Delete(ctx, testZone, ctx.ID)
}
