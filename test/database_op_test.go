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
	"context"
	"testing"

	"github.com/sacloud/iaas-api-go/helper/query"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/power"
	"github.com/sacloud/iaas-api-go/helper/wait"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
)

func TestDatabaseOpCRUD(t *testing.T) {
	testutil.RunCRUD(t, &testutil.CRUDTestCase{
		Parallel: true,

		SetupAPICallerFunc: singletonAPICaller,
		Setup: setupSwitchFunc("db",
			createDatabaseParam,
			createDatabaseExpected,
			updateDatabaseSettingsExpected,
			updateDatabaseExpected,
			updateDatabaseToFullExpected,
			updateDatabaseToMinExpected,
		),
		Create: &testutil.CRUDTestFunc{
			Func: testDatabaseCreate,
			CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
				ExpectValue:  createDatabaseExpected,
				IgnoreFields: ignoreDatabaseFields,
			}),
		},
		Read: &testutil.CRUDTestFunc{
			Func: testDatabaseRead,
			CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
				ExpectValue:  createDatabaseExpected,
				IgnoreFields: ignoreDatabaseFields,
			}),
		},
		Updates: []*testutil.CRUDTestFunc{
			{
				Func: testDatabaseUpdateSettings,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateDatabaseSettingsExpected,
					IgnoreFields: ignoreDatabaseFields,
				}),
			},
			{
				Func: testDatabaseUpdate,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateDatabaseExpected,
					IgnoreFields: ignoreDatabaseFields,
				}),
			},
			{
				Func: testDatabaseUpdate,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateDatabaseExpected,
					IgnoreFields: ignoreDatabaseFields,
				}),
			},
			{
				Func: testDatabaseUpdateToFull,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateDatabaseToFullExpected,
					IgnoreFields: ignoreDatabaseFields,
				}),
			},
			{
				Func: testDatabaseUpdateToMin,
				CheckFunc: testutil.AssertEqualWithExpected(&testutil.CRUDTestExpect{
					ExpectValue:  updateDatabaseToMinExpected,
					IgnoreFields: ignoreDatabaseFields,
				}),
			},
			// parameter settings
			{
				Func: func(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
					dbOp := iaas.NewDatabaseOp(caller)
					err := dbOp.SetParameter(ctx, testZone, ctx.ID, map[string]interface{}{
						"MariaDB/server.cnf/mysqld/max_connections": 50,
					})
					if err != nil {
						return nil, err
					}
					return dbOp.GetParameter(ctx, testZone, ctx.ID)
				},
				CheckFunc: func(t testutil.TestT, ctx *testutil.CRUDTestContext, v interface{}) error {
					param := v.(*iaas.DatabaseParameter)
					return testutil.DoAsserts(
						testutil.AssertLenFunc(t, param.Settings, 1, "Settings"),
						testutil.AssertEqualFunc(t, float64(50), param.Settings["MariaDB/server.cnf/mysqld/max_connections"], "Settings.Value"),
					)
				},
				SkipExtractID: true,
			},
			// reset parameter
			{
				Func: func(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
					dbOp := iaas.NewDatabaseOp(caller)
					err := dbOp.SetParameter(ctx, testZone, ctx.ID, map[string]interface{}{
						"MariaDB/server.cnf/mysqld/max_connections": nil,
					})
					if err != nil {
						return nil, err
					}
					return dbOp.GetParameter(ctx, testZone, ctx.ID)
				},
				CheckFunc: func(t testutil.TestT, ctx *testutil.CRUDTestContext, v interface{}) error {
					param := v.(*iaas.DatabaseParameter)
					return testutil.DoAsserts(
						testutil.AssertLenFunc(t, param.Settings, 0, "Settings"),
					)
				},
				SkipExtractID: true,
			},
		},
		Shutdown: func(ctx *testutil.CRUDTestContext, caller iaas.APICaller) error {
			client := iaas.NewDatabaseOp(caller)
			return power.ShutdownDatabase(ctx, client, testZone, ctx.ID, true)
		},

		Delete: &testutil.CRUDTestDeleteFunc{
			Func: testDatabaseDelete,
		},

		Cleanup: cleanupSwitchFunc("db"),
	})
}

var (
	ignoreDatabaseFields = []string{
		"ID",
		"Class",
		"Tags", // Create(POST)時は指定したタグが返ってくる。その後利用可能になったらデータベースの種類に応じて@MariaDBxxxのようなタグが付与される
		"Availability",
		"InstanceStatus",
		"InstanceHostName",
		"InstanceHostInfoURL",
		"InstanceStatusChangedAt",
		"Interfaces",
		"ZoneID",
		"CreatedAt",
		"ModifiedAt",
		"SettingsHash",
	}

	createDatabaseParam = &iaas.DatabaseCreateRequest{
		PlanID:         types.DatabasePlans.DB10GB,
		IPAddresses:    []string{"192.168.0.11"},
		NetworkMaskLen: 24,
		DefaultRoute:   "192.168.0.1",
		Name:           testutil.ResourceName("db"),
		Description:    "desc",
		Tags:           []string{"tag1", "tag2"},

		Conf: &iaas.DatabaseRemarkDBConfCommon{
			DatabaseName:     types.RDBMSVersions[types.RDBMSTypesMariaDB].Name,
			DatabaseVersion:  types.RDBMSVersions[types.RDBMSTypesMariaDB].Version,
			DatabaseRevision: "10.4.12",
			DefaultUser:      "exa.mple",
			UserPassword:     "LibsacloudExamplePassword01",
		},
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     5432,
			DefaultUser:     "exa.mple",
			UserPassword:    "LibsacloudExamplePassword01",
			ReplicaUser:     "replica",
			ReplicaPassword: "replica-user-password",
		},
		ReplicationSetting: &iaas.DatabaseReplicationSetting{
			Model: types.DatabaseReplicationModels.MasterSlave,
		},
	}
	createDatabaseExpected = &iaas.Database{
		Name:               createDatabaseParam.Name,
		Description:        createDatabaseParam.Description,
		Availability:       types.Availabilities.Available,
		PlanID:             createDatabaseParam.PlanID,
		DefaultRoute:       createDatabaseParam.DefaultRoute,
		NetworkMaskLen:     createDatabaseParam.NetworkMaskLen,
		IPAddresses:        createDatabaseParam.IPAddresses,
		Conf:               createDatabaseParam.Conf,
		CommonSetting:      createDatabaseParam.CommonSetting,
		ReplicationSetting: createDatabaseParam.ReplicationSetting,
	}
	updateDatabaseSettingsParam = &iaas.DatabaseUpdateSettingsRequest{
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     54322,
			DefaultUser:     "exa.mple.upd",
			UserPassword:    "LibsacloudExamplePassword01up1",
			ReplicaUser:     "replica-upd",
			ReplicaPassword: "replica-user-password-upd",
		},
		ReplicationSetting: createDatabaseParam.ReplicationSetting,
	}
	updateDatabaseSettingsExpected = &iaas.Database{
		Name:           createDatabaseParam.Name,
		Description:    createDatabaseParam.Description,
		Availability:   types.Availabilities.Available,
		PlanID:         createDatabaseParam.PlanID,
		InstanceStatus: types.ServerInstanceStatuses.Up,
		DefaultRoute:   createDatabaseParam.DefaultRoute,
		NetworkMaskLen: createDatabaseParam.NetworkMaskLen,
		IPAddresses:    createDatabaseParam.IPAddresses,
		Conf:           createDatabaseParam.Conf,
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     54322,
			DefaultUser:     "exa.mple.upd",
			UserPassword:    "LibsacloudExamplePassword01up1",
			ReplicaUser:     "replica-upd",
			ReplicaPassword: "replica-user-password-upd",
		},
		ReplicationSetting: createDatabaseParam.ReplicationSetting,
	}
	updateDatabaseParam = &iaas.DatabaseUpdateRequest{
		Name:        testutil.ResourceName("db-upd"),
		Tags:        []string{"tag1-upd", "tag2-upd"},
		Description: "desc-upd",
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     5432,
			DefaultUser:     "exa.mple",
			UserPassword:    "LibsacloudExamplePassword02",
			ReplicaUser:     "replica",
			ReplicaPassword: "replica-user-password",
		},
		ReplicationSetting: &iaas.DatabaseReplicationSetting{
			Model: types.DatabaseReplicationModels.MasterSlave,
		},
	}
	updateDatabaseExpected = &iaas.Database{
		Name:           updateDatabaseParam.Name,
		Description:    updateDatabaseParam.Description,
		Availability:   types.Availabilities.Available,
		PlanID:         createDatabaseParam.PlanID,
		InstanceStatus: types.ServerInstanceStatuses.Up,
		DefaultRoute:   createDatabaseParam.DefaultRoute,
		NetworkMaskLen: createDatabaseParam.NetworkMaskLen,
		IPAddresses:    createDatabaseParam.IPAddresses,
		Conf:           createDatabaseParam.Conf,
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     5432,
			DefaultUser:     "exa.mple",
			UserPassword:    "LibsacloudExamplePassword02",
			ReplicaUser:     "replica",
			ReplicaPassword: "replica-user-password",
		},
		ReplicationSetting: createDatabaseParam.ReplicationSetting,
	}
	updateDatabaseToFullParam = &iaas.DatabaseUpdateRequest{
		Name:        testutil.ResourceName("db-to-full"),
		Tags:        []string{"tag1-upd", "tag2-upd"},
		Description: "desc-upd",
		BackupSetting: &iaas.DatabaseSettingBackup{
			Rotate: 3,
			Time:   "00:00",
			DayOfWeek: []types.EBackupSpanWeekday{
				types.BackupSpanWeekdays.Sunday,
				types.BackupSpanWeekdays.Monday,
			},
		},
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     54321,
			DefaultUser:     "exa.mple",
			UserPassword:    "LibsacloudExamplePassword03",
			SourceNetwork:   []string{"192.168.11.0/24", "192.168.12.0/24"},
			ReplicaUser:     "replica",
			ReplicaPassword: "replica-user-password",
		},
		ReplicationSetting: &iaas.DatabaseReplicationSetting{
			Model: types.DatabaseReplicationModels.MasterSlave,
		},
		IconID: testIconID,
	}
	updateDatabaseToFullExpected = &iaas.Database{
		Name:           updateDatabaseToFullParam.Name,
		Description:    updateDatabaseToFullParam.Description,
		Availability:   types.Availabilities.Available,
		PlanID:         createDatabaseParam.PlanID,
		InstanceStatus: types.ServerInstanceStatuses.Up,
		DefaultRoute:   createDatabaseParam.DefaultRoute,
		NetworkMaskLen: createDatabaseParam.NetworkMaskLen,
		IPAddresses:    createDatabaseParam.IPAddresses,
		Conf:           createDatabaseParam.Conf,
		CommonSetting: &iaas.DatabaseSettingCommon{
			ServicePort:     54321,
			DefaultUser:     "exa.mple",
			UserPassword:    "LibsacloudExamplePassword03",
			SourceNetwork:   []string{"192.168.11.0/24", "192.168.12.0/24"},
			ReplicaUser:     "replica",
			ReplicaPassword: "replica-user-password",
		},
		BackupSetting:      updateDatabaseToFullParam.BackupSetting,
		ReplicationSetting: createDatabaseParam.ReplicationSetting,
		IconID:             updateDatabaseToFullParam.IconID,
	}
	updateDatabaseToMinParam = &iaas.DatabaseUpdateRequest{
		Name: testutil.ResourceName("db-to-min"),
		CommonSetting: &iaas.DatabaseSettingCommon{
			DefaultUser:     "exa.mple",
			UserPassword:    "LibsacloudExamplePassword04",
			ReplicaUser:     "replica",
			ReplicaPassword: "replica-user-password",
		},
		ReplicationSetting: &iaas.DatabaseReplicationSetting{
			Model: types.DatabaseReplicationModels.MasterSlave,
		},
	}
	updateDatabaseToMinExpected = &iaas.Database{
		Name:           updateDatabaseToMinParam.Name,
		Availability:   types.Availabilities.Available,
		PlanID:         createDatabaseParam.PlanID,
		InstanceStatus: types.ServerInstanceStatuses.Up,
		DefaultRoute:   createDatabaseParam.DefaultRoute,
		NetworkMaskLen: createDatabaseParam.NetworkMaskLen,
		IPAddresses:    createDatabaseParam.IPAddresses,
		Conf:           createDatabaseParam.Conf,
		CommonSetting: &iaas.DatabaseSettingCommon{
			DefaultUser:     "exa.mple",
			UserPassword:    "LibsacloudExamplePassword04",
			ReplicaUser:     "replica",
			ReplicaPassword: "replica-user-password",
		},
		ReplicationSetting: createDatabaseParam.ReplicationSetting,
	}
)

func testDatabaseCreate(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewDatabaseOp(caller)
	db, err := client.Create(ctx, testZone, createDatabaseParam)
	if err != nil {
		return nil, err
	}
	return wait.UntilDatabaseIsUp(ctx, client, testZone, db.ID)
}

func testDatabaseRead(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewDatabaseOp(caller)
	return client.Read(ctx, testZone, ctx.ID)
}

func testDatabaseUpdateSettings(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewDatabaseOp(caller)
	return client.UpdateSettings(ctx, testZone, ctx.ID, updateDatabaseSettingsParam)
}

func testDatabaseUpdate(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewDatabaseOp(caller)
	return client.Update(ctx, testZone, ctx.ID, updateDatabaseParam)
}

func testDatabaseUpdateToFull(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewDatabaseOp(caller)
	return client.Update(ctx, testZone, ctx.ID, updateDatabaseToFullParam)
}

func testDatabaseUpdateToMin(ctx *testutil.CRUDTestContext, caller iaas.APICaller) (interface{}, error) {
	client := iaas.NewDatabaseOp(caller)
	return client.Update(ctx, testZone, ctx.ID, updateDatabaseToMinParam)
}

func testDatabaseDelete(ctx *testutil.CRUDTestContext, caller iaas.APICaller) error {
	client := iaas.NewDatabaseOp(caller)
	return client.Delete(ctx, testZone, ctx.ID)
}

// TestCreateProxyDatabase 冗長化オプションが有効なデータベースアプライアンスの作成テスト
func TestCreateProxyDatabase(t *testing.T) {
	if !testutil.IsAccTest() {
		t.Skip()
	}

	ctx := context.Background()
	caller := testutil.SingletonAPICaller()
	name := testutil.ResourceName("proxy-database")

	sw, err := iaas.NewSwitchOp(caller).Create(ctx, testutil.TestZone(), &iaas.SwitchCreateRequest{
		Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}

	planID, _, err := query.GetProxyDatabasePlan(ctx, iaas.NewNoteOp(caller), 4, 4, 90)
	if err != nil {
		t.Fatal(err)
	}

	dbOp := iaas.NewDatabaseOp(caller)
	db, err := dbOp.Create(ctx, testutil.TestZone(), &iaas.DatabaseCreateRequest{
		PlanID:   planID,
		SwitchID: sw.ID,
		IPAddresses: []string{
			"192.168.22.111",
			"192.168.22.112",
		},
		NetworkMaskLen: 24,
		DefaultRoute:   "192.168.22.1",
		Conf: &iaas.DatabaseRemarkDBConfCommon{
			DatabaseName: types.RDBMSTypesPostgreSQL.String(),
			// DatabaseVersion: "10.5", // debug
			DefaultUser:  "sacloud",
			UserPassword: "TestUserPassword01",
		},
		CommonSetting: &iaas.DatabaseSettingCommon{
			DefaultUser:  "sacloud",
			UserPassword: testutil.WithRandomPrefix("password"),
		},
		InterfaceSettings: []*iaas.DatabaseSettingsInterface{
			{
				VirtualIPAddress: "192.168.22.11",
				Index:            1,
			},
		},
		Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := wait.UntilDatabaseIsUp(ctx, dbOp, testutil.TestZone(), db.ID); err != nil {
		t.Fatal(err)
	}
}
