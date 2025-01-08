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
	"github.com/stretchr/testify/assert"
)

func TestZoneOp_Find(t *testing.T) {
	t.Parallel()

	client := iaas.NewZoneOp(singletonAPICaller())

	zoneFindResult, err := client.Find(context.Background(), &iaas.FindCondition{Count: 1})
	assert.NoError(t, err)
	assert.Len(t, zoneFindResult.Zones, 1)
}
