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
	"context"
	"testing"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/testutil"
	"github.com/stretchr/testify/assert"
)

func TestServerPlanOp_Find(t *testing.T) {
	t.Parallel()

	client := iaas.NewServerPlanOp(singletonAPICaller())

	searched, err := client.Find(context.Background(), iaas.APIDefaultZone, &iaas.FindCondition{Count: 1})
	assert.NoError(t, err)

	err = testutil.DoAsserts(
		testutil.AssertLenFunc(t, searched.ServerPlans, 1, "ServerPlans"),
		testutil.AssertNotEmptyFunc(t, searched.ServerPlans[0].ID, "ServerPlans.ID"),
		testutil.AssertNotEmptyFunc(t, searched.ServerPlans[0].Name, "ServerPlans.Name"),
		testutil.AssertNotEmptyFunc(t, searched.ServerPlans[0].CPU, "ServerPlans.CPU"),
		testutil.AssertNotEmptyFunc(t, searched.ServerPlans[0].CPUModel, "ServerPlans.CPUModel"),
		testutil.AssertNotEmptyFunc(t, searched.ServerPlans[0].Commitment, "ServerPlans.Commitment"),
		testutil.AssertNotEmptyFunc(t, searched.ServerPlans[0].Generation, "ServerPlans.Generation"),
		testutil.AssertNotEmptyFunc(t, searched.ServerPlans[0].Availability, "ServerPlans.Availability"),
	)
	assert.NoError(t, err)
}
