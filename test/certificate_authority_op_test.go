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

package test

import (
	"testing"
	"time"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
)

func TestCertificateAuthorityOp_CRUD(t *testing.T) {
	testutil.RunCRUD(t, &testutil.CRUDTestCase{
		Parallel: true,

		SetupAPICallerFunc: singletonAPICaller,

		Create: &testutil.CRUDTestFunc{
			Func: testCertificateAuthorityCreate,
			CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
				ExpectValue:  createCertificateAuthorityExpected,
				IgnoreFields: ignoreCertificateAuthorityFields,
			}),
		},

		Read: &testutil.CRUDTestFunc{
			Func: testCertificateAuthorityRead,
			CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
				ExpectValue:  createCertificateAuthorityExpected,
				IgnoreFields: ignoreCertificateAuthorityFields,
			}),
		},

		Updates: []*testutil.CRUDTestFunc{
			{
				Func: testCertificateAuthorityUpdate,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateCertificateAuthorityExpected,
					IgnoreFields: ignoreCertificateAuthorityFields,
				}),
			},
		},

		Delete: &testutil.CRUDTestDeleteFunc{
			Func: testCertificateAuthorityDelete,
		},
	})
}

var (
	ignoreCertificateAuthorityFields = []string{
		"ID",
		"Class",
		"SettingsHash",
		"CreatedAt",
		"ModifiedAt",
		"Subject", // fakeドライバーで算出していないため
		"NotAfter",
	}
	createCertificateAuthorityParam = &iaas.CertificateAuthorityCreateRequest{
		Name:             testutil.ResourceName("certificate-authority"),
		Description:      "desc",
		Tags:             []string{"tag1", "tag2"},
		Country:          "JP",
		Organization:     "usacloud",
		OrganizationUnit: []string{"u1", "u2"},
		CommonName:       "u2.u1.uiaas.jp",
		NotAfter:         time.Now().Add(10 * 24 * time.Hour),
	}
	createCertificateAuthorityExpected = &iaas.CertificateAuthority{
		Name:             createCertificateAuthorityParam.Name,
		Description:      createCertificateAuthorityParam.Description,
		Tags:             createCertificateAuthorityParam.Tags,
		Availability:     types.Availabilities.Available,
		Country:          createCertificateAuthorityParam.Country,
		Organization:     createCertificateAuthorityParam.Organization,
		OrganizationUnit: createCertificateAuthorityParam.OrganizationUnit,
		CommonName:       createCertificateAuthorityParam.CommonName,
	}
	updateCertificateAuthorityParam = &iaas.CertificateAuthorityUpdateRequest{
		Name:        createCertificateAuthorityParam.Name,
		Description: "desc-upd",
		Tags:        []string{"tag1-upd", "tag2-upd"},
		IconID:      testIconID,
	}
	updateCertificateAuthorityExpected = &iaas.CertificateAuthority{
		Name:             createCertificateAuthorityParam.Name,
		Description:      updateCertificateAuthorityParam.Description,
		IconID:           testIconID,
		Tags:             updateCertificateAuthorityParam.Tags,
		Availability:     types.Availabilities.Available,
		Country:          createCertificateAuthorityParam.Country,
		Organization:     createCertificateAuthorityParam.Organization,
		OrganizationUnit: createCertificateAuthorityParam.OrganizationUnit,
		CommonName:       createCertificateAuthorityParam.CommonName,
	}
)

func testCertificateAuthorityCreate(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewCertificateAuthorityOp(caller)
	return client.Create(ctx, createCertificateAuthorityParam)
}

func testCertificateAuthorityRead(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewCertificateAuthorityOp(caller)
	return client.Read(ctx, ctx.ID)
}

func testCertificateAuthorityUpdate(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewCertificateAuthorityOp(caller)
	return client.Update(ctx, ctx.ID, updateCertificateAuthorityParam)
}

func testCertificateAuthorityDelete(ctx *testutil.CRUDTestContext, caller iaas.APICaller) error {
	client := iaas.NewCertificateAuthorityOp(caller)
	return client.Delete(ctx, ctx.ID)
}
