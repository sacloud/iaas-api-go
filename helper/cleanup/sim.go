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

package cleanup

import (
	"context"
	"fmt"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/types"
)

// DeleteSIM SIMの無効化&削除
func DeleteSIM(ctx context.Context, client iaas.SIMAPI, id types.ID) error {
	sim, err := query.FindSIMByID(ctx, client, id)
	if err != nil {
		return err
	}
	if sim.Info.Activated {
		if err := client.Deactivate(ctx, id); err != nil {
			return err
		}
	}
	return client.Delete(ctx, id)
}

// DeleteSIMWithReferencedCheck 他リソースからの参照を確認した上でリソースの削除を行う
func DeleteSIMWithReferencedCheck(ctx context.Context, caller iaas.APICaller, zones []string, id types.ID, option query.CheckReferencedOption) error {
	if err := query.WaitWhileSIMIsReferenced(ctx, caller, zones, id, option); err != nil {
		return fmt.Errorf("sim[%s] is still being used by other resources: %s", id, err)
	}
	return DeleteSIM(ctx, iaas.NewSIMOp(caller), id)
}
