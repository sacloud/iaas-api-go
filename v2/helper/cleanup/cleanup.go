// Copyright 2022-2026 The sacloud/iaas-api-go Authors
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

// Package cleanup は v2 リソースの削除前処理を含む削除ヘルパーを提供する。
//
// v1 の github.com/sacloud/iaas-api-go/helper/cleanup の v2 相当。
// query / power に依存し、参照チェックおよび必要に応じたシャットダウンを
// 行ってから delete を呼ぶ。
package cleanup

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/iaas-api-go/v2/helper/power"
	"github.com/sacloud/iaas-api-go/v2/helper/query"
)

// ---------- Server ----------

// ServerCleanupAPI は DeleteServer が必要とする最小 interface。
type ServerCleanupAPI interface {
	power.ServerPowerAPI
	Delete(ctx context.Context, id int64, request *client.ServerDeleteRequestEnvelope) error
}

// DeleteServer はサーバを削除する。起動中なら強制シャットダウン、withDisks=true なら
// 接続されたディスクもまとめて削除する。
func DeleteServer(ctx context.Context, op ServerCleanupAPI, id int64, withDisks bool) error {
	resp, err := op.Read(ctx, id)
	if err != nil {
		return err
	}

	if resp.Server.Instance.IsSet() && string(resp.Server.Instance.Value.Status.Value) == "up" {
		if err := power.ShutdownServer(ctx, op, id, true); err != nil {
			return fmt.Errorf("cleanup: shutdown server[%d] failed: %w", id, err)
		}
	}

	req := &client.ServerDeleteRequestEnvelope{}
	if withDisks {
		for _, d := range resp.Server.Disks {
			req.WithDisk = append(req.WithDisk, client.ID(strconv.FormatInt(d.ID.Value, 10)))
		}
	}
	if err := op.Delete(ctx, id, req); err != nil {
		return fmt.Errorf("cleanup: delete server[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- Disk ----------

// DiskCleanupAPI は DeleteDisk が必要とする最小 interface。
type DiskCleanupAPI interface {
	Delete(ctx context.Context, id int64) error
}

// DeleteDisk は参照されていないことを確認してから Disk を削除する。
func DeleteDisk(ctx context.Context, op DiskCleanupAPI, r query.ReferenceFinder, id int64, option query.CheckReferencedOption) error {
	if err := query.WaitWhileDiskIsReferenced(ctx, r, id, option); err != nil {
		return fmt.Errorf("cleanup: disk[%d] is still being used: %w", id, err)
	}
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete disk[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- CDROM ----------

// CDROMCleanupAPI は DeleteCDROM が必要とする最小 interface。
type CDROMCleanupAPI interface {
	Delete(ctx context.Context, id int64) error
}

// DeleteCDROM は参照されていないことを確認してから CD-ROM を削除する。
func DeleteCDROM(ctx context.Context, op CDROMCleanupAPI, r query.ReferenceFinder, id int64, option query.CheckReferencedOption) error {
	if err := query.WaitWhileCDROMIsReferenced(ctx, r, id, option); err != nil {
		return fmt.Errorf("cleanup: cdrom[%d] is still being used: %w", id, err)
	}
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete cdrom[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- Switch ----------

// SwitchCleanupAPI は DeleteSwitch が必要とする最小 interface。
type SwitchCleanupAPI interface {
	Delete(ctx context.Context, id int64) error
}

// DeleteSwitch は参照されていないことを確認してから Switch を削除する。
func DeleteSwitch(ctx context.Context, op SwitchCleanupAPI, r query.ReferenceFinder, id int64, option query.CheckReferencedOption) error {
	if err := query.WaitWhileSwitchIsReferenced(ctx, r, id, option); err != nil {
		return fmt.Errorf("cleanup: switch[%d] is still being used: %w", id, err)
	}
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete switch[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- Bridge ----------

// BridgeCleanupAPI は DeleteBridge が必要とする最小 interface。
type BridgeCleanupAPI interface {
	Delete(ctx context.Context, id int64) error
}

// DeleteBridge は参照されていないことを確認してから Bridge を削除する。
func DeleteBridge(ctx context.Context, op BridgeCleanupAPI, switchFinder query.SwitchFinder, id int64, option query.CheckReferencedOption) error {
	// Bridge は Switch.Bridge での参照をチェック
	if err := query.WaitWhileBridgeIsReferenced(ctx, switchFinder, id, option); err != nil {
		return fmt.Errorf("cleanup: bridge[%d] is still being used: %w", id, err)
	}
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete bridge[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- PacketFilter ----------

// PacketFilterCleanupAPI は DeletePacketFilter が必要とする最小 interface。
type PacketFilterCleanupAPI interface {
	Delete(ctx context.Context, id int64) error
}

// DeletePacketFilter は参照されていないことを確認してから PacketFilter を削除する。
func DeletePacketFilter(ctx context.Context, op PacketFilterCleanupAPI, r query.ReferenceFinder, id int64, option query.CheckReferencedOption) error {
	if err := query.WaitWhilePacketFilterIsReferenced(ctx, r, id, option); err != nil {
		return fmt.Errorf("cleanup: packet filter[%d] is still being used: %w", id, err)
	}
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete packet filter[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- PrivateHost ----------

// PrivateHostCleanupAPI は DeletePrivateHost が必要とする最小 interface。
type PrivateHostCleanupAPI interface {
	Delete(ctx context.Context, id int64) error
}

// DeletePrivateHost は参照されていないことを確認してから PrivateHost を削除する。
func DeletePrivateHost(ctx context.Context, op PrivateHostCleanupAPI, r query.ReferenceFinder, id int64, option query.CheckReferencedOption) error {
	if err := query.WaitWhilePrivateHostIsReferenced(ctx, r, id, option); err != nil {
		return fmt.Errorf("cleanup: private host[%d] is still being used: %w", id, err)
	}
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete private host[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- SIM ----------

// SIMCleanupAPI は DeleteSIM が必要とする最小 interface。
// v2 では SIM は CommonServiceItem の一種として管理されるため、実際には
// CommonServiceItem の Delete を呼ぶことになる。
type SIMCleanupAPI interface {
	Delete(ctx context.Context, id int64) error
}

// DeleteSIM は参照されていないことを確認してから SIM を削除する。
func DeleteSIM(ctx context.Context, op SIMCleanupAPI, r query.ReferenceFinder, id int64, option query.CheckReferencedOption) error {
	if err := query.WaitWhileSIMIsReferenced(ctx, r, id, option); err != nil {
		return fmt.Errorf("cleanup: sim[%d] is still being used: %w", id, err)
	}
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete sim[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- Internet ----------

// InternetCleanupAPI は DeleteInternet が必要とする最小 interface。
type InternetCleanupAPI interface {
	Delete(ctx context.Context, id int64) error
}

// DeleteInternet は Internet (ルータ+スイッチ) を削除する。
// ルータ+スイッチ配下の Switch は Internet の削除時に自動的に連動削除される。
func DeleteInternet(ctx context.Context, op InternetCleanupAPI, id int64) error {
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete internet[%d] failed: %w", id, err)
	}
	return nil
}

// ---------- MobileGateway ----------

// MobileGatewayCleanupAPI は DeleteMobileGateway が必要とする最小 interface。
// 起動中の MobileGateway はシャットダウンしてから削除する。
type MobileGatewayCleanupAPI interface {
	power.AppliancePowerAPI
	Delete(ctx context.Context, id int64) error
}

// DeleteMobileGateway は Appliance として MobileGateway を削除する。
// 起動中ならシャットダウンしてから削除する。
func DeleteMobileGateway(ctx context.Context, op MobileGatewayCleanupAPI, id int64) error {
	resp, err := op.Read(ctx, id)
	if err != nil {
		return err
	}
	if resp.Appliance.Instance.IsSet() && string(resp.Appliance.Instance.Value.Status.Value) == "up" {
		if err := power.ShutdownAppliance(ctx, op, id, true); err != nil {
			return fmt.Errorf("cleanup: shutdown mobile gateway[%d] failed: %w", id, err)
		}
	}
	if err := op.Delete(ctx, id); err != nil {
		return fmt.Errorf("cleanup: delete mobile gateway[%d] failed: %w", id, err)
	}
	return nil
}
