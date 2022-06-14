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

package query

import (
	"context"
	"testing"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/stretchr/testify/require"
)

func TestGetDatabasePlan(t *testing.T) {
	if !testutil.IsAccTest() {
		t.Skip()
	}
	caller := testutil.SingletonAPICaller()
	finder := iaas.NewNoteOp(caller)

	// 4CPU/4GBメモリ/90GBディスク
	plan, serviceClass, err := GetProxyDatabasePlan(context.Background(), finder, 4, 4, 90)
	require.NoError(t, err)
	require.NotEqual(t, plan, types.ID(0))
	require.Equal(t, "cloud/appliance/database/4core4gb-100gb-proxy", serviceClass)
}
