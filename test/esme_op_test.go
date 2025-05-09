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
	"os"
	"testing"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
)

func TestESMEOpCRUD(t *testing.T) {
	testutil.PreCheckEnvsFunc("SAKURACLOUD_ESME_DESTINATION")(t)

	destination := os.Getenv("SAKURACLOUD_ESME_DESTINATION")

	testutil.RunCRUD(t, &testutil.CRUDTestCase{
		Parallel: true,

		SetupAPICallerFunc: singletonAPICaller,

		Create: &testutil.CRUDTestFunc{
			Func: testESMECreate,
			CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
				ExpectValue:  createESMEExpected,
				IgnoreFields: ignoreESMEFields,
			}),
		},

		Read: &testutil.CRUDTestFunc{
			Func: testESMERead,
			CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
				ExpectValue:  createESMEExpected,
				IgnoreFields: ignoreESMEFields,
			}),
		},

		Updates: []*testutil.CRUDTestFunc{
			{
				Func: testESMEUpdate,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateESMEExpected,
					IgnoreFields: ignoreESMEFields,
				}),
			},
			{
				Func: testESMEUpdateToMin,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateESMEToMinExpected,
					IgnoreFields: ignoreESMEFields,
				}),
			},
			// send SMS
			{
				Func: func(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
					client := iaas.NewESMEOp(caller)
					_, err := client.SendMessageWithGeneratedOTP(ctx, ctx.ID, &iaas.ESMESendMessageWithGeneratedOTPRequest{
						Destination: destination,
						Sender:      "libsacloud-test",
						DomainName:  "www.example.com",
					})
					return nil, err
				},
			},
			{
				Func: func(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
					client := iaas.NewESMEOp(caller)
					logs, err := client.Logs(ctx, ctx.ID)
					if err != nil {
						return nil, err
					}
					return nil, testutil.DoAsserts(
						testutil.AssertLenFunc(t, logs, 1, "Logs"),
					)
				},
			},
			{
				Func: func(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
					client := iaas.NewESMEOp(caller)
					_, err := client.SendMessageWithInputtedOTP(ctx, ctx.ID, &iaas.ESMESendMessageWithInputtedOTPRequest{
						Destination: destination,
						Sender:      "libsacloud-test",
						DomainName:  "www.example.com",
						OTP:         "397397",
					})
					return nil, err
				},
			},
			{
				Func: func(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
					client := iaas.NewESMEOp(caller)
					logs, err := client.Logs(ctx, ctx.ID)
					if err != nil {
						return nil, err
					}
					return nil, testutil.DoAsserts(
						testutil.AssertLenFunc(t, logs, 2, "Logs"),
					)
				},
			},
		},

		Delete: &testutil.CRUDTestDeleteFunc{
			Func: testESMEDelete,
		},
	})
}

var (
	ignoreESMEFields = []string{
		"ID",
		"Class",
		"Settings",
		"SettingsHash",
		"CreatedAt",
		"ModifiedAt",
	}
	createESMEParam = &iaas.ESMECreateRequest{
		Name:        testutil.ResourceName("esme"),
		Description: "desc",
		Tags:        []string{"tag1", "tag2"},
	}
	createESMEExpected = &iaas.ESME{
		Name:         createESMEParam.Name,
		Description:  createESMEParam.Description,
		Tags:         createESMEParam.Tags,
		Availability: types.Availabilities.Available,
	}
	updateESMEParam = &iaas.ESMEUpdateRequest{
		Name:        testutil.ResourceName("esme-upd"),
		Description: "desc-upd",
		Tags:        []string{"tag1-upd", "tag2-upd"},
		IconID:      testIconID,
	}
	updateESMEExpected = &iaas.ESME{
		Name:         updateESMEParam.Name,
		Description:  updateESMEParam.Description,
		Tags:         updateESMEParam.Tags,
		Availability: types.Availabilities.Available,
		IconID:       testIconID,
	}

	updateESMEToMinParam = &iaas.ESMEUpdateRequest{
		Name: testutil.ResourceName("esme-to-min"),
	}
	updateESMEToMinExpected = &iaas.ESME{
		Name:         updateESMEToMinParam.Name,
		Availability: types.Availabilities.Available,
	}
)

func testESMECreate(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewESMEOp(caller)
	return client.Create(ctx, createESMEParam)
}

func testESMERead(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewESMEOp(caller)
	return client.Read(ctx, ctx.ID)
}

func testESMEUpdate(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewESMEOp(caller)
	return client.Update(ctx, ctx.ID, updateESMEParam)
}

func testESMEUpdateToMin(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewESMEOp(caller)
	return client.Update(ctx, ctx.ID, updateESMEToMinParam)
}

func testESMEDelete(ctx *testutil.CRUDTestContext, caller iaas.APICaller) error {
	client := iaas.NewESMEOp(caller)
	return client.Delete(ctx, ctx.ID)
}
