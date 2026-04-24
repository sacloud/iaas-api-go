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

package query

import (
	"context"
	"fmt"
	"strconv"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
)

// ReferenceFinder は参照チェックに必要な各種 Finder をまとめたもの。
// 呼び出し側は iaas.NewServerOp(c) / iaas.NewSwitchOp(c) 等を詰めるか、
// NewReferenceFinder を使うと便利。SIM 参照チェックを使う場合は Appliance /
// MobileGateway 両方を設定すること。
type ReferenceFinder struct {
	Server        ServerFinder
	Switch        SwitchReader
	Appliance     ApplianceLister
	MobileGateway MobileGatewaySIMLister
}

// NewReferenceFinder は v2 client.Client から ReferenceFinder を生成する便利関数。
func NewReferenceFinder(c *client.Client) ReferenceFinder {
	return ReferenceFinder{
		Server:        iaas.NewServerOp(c),
		Switch:        iaas.NewSwitchOp(c),
		Appliance:     iaas.NewApplianceOp(c),
		MobileGateway: iaas.NewMobileGatewayOp(c),
	}
}

// IsPrivateHostReferenced は専有ホストが Server で参照されている場合 true。
func IsPrivateHostReferenced(ctx context.Context, r ReferenceFinder, privateHostID int64) (bool, error) {
	servers, err := listAllServers(ctx, r.Server)
	if err != nil {
		return false, err
	}
	for _, s := range servers {
		if s.PrivateHost.IsSet() && s.PrivateHost.Value.ID == privateHostID {
			return true, nil
		}
	}
	return false, nil
}

// WaitWhilePrivateHostIsReferenced は PrivateHost が参照されている間 poll する。
func WaitWhilePrivateHostIsReferenced(ctx context.Context, r ReferenceFinder, privateHostID int64, option CheckReferencedOption) error {
	return waitWhileReferenced(ctx, option, func() (bool, error) {
		return IsPrivateHostReferenced(ctx, r, privateHostID)
	})
}

// IsPacketFilterReferenced は PacketFilter が Server の Interface で参照されている場合 true。
func IsPacketFilterReferenced(ctx context.Context, r ReferenceFinder, packetFilterID int64) (bool, error) {
	servers, err := listAllServers(ctx, r.Server)
	if err != nil {
		return false, err
	}
	for _, s := range servers {
		for _, iface := range s.Interfaces {
			if iface.PacketFilter.IsSet() && iface.PacketFilter.Value.ID.Value == packetFilterID {
				return true, nil
			}
		}
	}
	return false, nil
}

// WaitWhilePacketFilterIsReferenced は PacketFilter が参照されている間 poll する。
func WaitWhilePacketFilterIsReferenced(ctx context.Context, r ReferenceFinder, packetFilterID int64, option CheckReferencedOption) error {
	return waitWhileReferenced(ctx, option, func() (bool, error) {
		return IsPacketFilterReferenced(ctx, r, packetFilterID)
	})
}

// IsCDROMReferenced は CD-ROM (ISO) が Server に挿入されている場合 true。
func IsCDROMReferenced(ctx context.Context, r ReferenceFinder, cdromID int64) (bool, error) {
	servers, err := listAllServers(ctx, r.Server)
	if err != nil {
		return false, err
	}
	for _, s := range servers {
		if s.Instance.IsSet() && s.Instance.Value.CDROM.IsSet() && s.Instance.Value.CDROM.Value.ID == cdromID {
			return true, nil
		}
	}
	return false, nil
}

// WaitWhileCDROMIsReferenced は CD-ROM が参照されている間 poll する。
func WaitWhileCDROMIsReferenced(ctx context.Context, r ReferenceFinder, cdromID int64, option CheckReferencedOption) error {
	return waitWhileReferenced(ctx, option, func() (bool, error) {
		return IsCDROMReferenced(ctx, r, cdromID)
	})
}

// IsDiskReferenced は Disk が Server に接続されている場合 true。
func IsDiskReferenced(ctx context.Context, r ReferenceFinder, diskID int64) (bool, error) {
	servers, err := listAllServers(ctx, r.Server)
	if err != nil {
		return false, err
	}
	for _, s := range servers {
		for _, d := range s.Disks {
			if d.ID.Value == diskID {
				return true, nil
			}
		}
	}
	return false, nil
}

// WaitWhileDiskIsReferenced は Disk が参照されている間 poll する。
func WaitWhileDiskIsReferenced(ctx context.Context, r ReferenceFinder, diskID int64, option CheckReferencedOption) error {
	return waitWhileReferenced(ctx, option, func() (bool, error) {
		return IsDiskReferenced(ctx, r, diskID)
	})
}

// IsSwitchReferenced は Switch が Server で参照されている場合 true。
// 注: v1 では HybridConnectionID の検査と Appliance (DB/LB/NFS/MGW/VPCRouter) の
// Interface 走査も含んでいたが、v2 spec ではこれらのフィールドが未公開のため省略している。
// Switch.GetServers で接続済み Server のみを検査する。
// Appliance 系の参照検査が必要になった時点で v2 spec への追加と合わせて拡張する。
func IsSwitchReferenced(ctx context.Context, r ReferenceFinder, switchID int64) (bool, error) {
	if r.Switch == nil {
		return false, nil
	}
	resp, err := r.Switch.GetServers(ctx, switchID)
	if err != nil {
		if iaas.IsNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("query: GetServers failed: %w", err)
	}
	return resp != nil && len(resp.Servers) > 0, nil
}

// WaitWhileSwitchIsReferenced は Switch が参照されている間 poll する。
func WaitWhileSwitchIsReferenced(ctx context.Context, r ReferenceFinder, switchID int64, option CheckReferencedOption) error {
	return waitWhileReferenced(ctx, option, func() (bool, error) {
		return IsSwitchReferenced(ctx, r, switchID)
	})
}

// IsBridgeReferenced は Bridge が Switch.Bridge で参照されている場合 true。
func IsBridgeReferenced(ctx context.Context, switchFinder SwitchFinder, bridgeID int64) (bool, error) {
	switches, err := listAllSwitches(ctx, switchFinder)
	if err != nil {
		return false, err
	}
	for _, sw := range switches {
		if sw.Bridge.IsSet() && sw.Bridge.Value.ID == bridgeID {
			return true, nil
		}
	}
	return false, nil
}

// WaitWhileBridgeIsReferenced は Bridge が参照されている間 poll する。
func WaitWhileBridgeIsReferenced(ctx context.Context, switchFinder SwitchFinder, bridgeID int64, option CheckReferencedOption) error {
	return waitWhileReferenced(ctx, option, func() (bool, error) {
		return IsBridgeReferenced(ctx, switchFinder, bridgeID)
	})
}

// IsSIMReferenced は SIM が MobileGateway に登録されている場合 true。
// SIM.ResourceID は文字列 (int64 の 10 進表記) なので strconv で比較する。
// Appliance + MobileGateway 両方の Finder が必要。
func IsSIMReferenced(ctx context.Context, r ReferenceFinder, simID int64) (bool, error) {
	if r.Appliance == nil || r.MobileGateway == nil {
		return false, nil
	}
	resp, err := r.Appliance.List(ctx, client.OptString{})
	if err != nil {
		return false, fmt.Errorf("query: Appliance.List failed: %w", err)
	}
	target := strconv.FormatInt(simID, 10)
	for _, app := range resp.Appliances {
		sims, err := r.MobileGateway.ListSIM(ctx, app.ID.Value)
		if err != nil {
			if iaas.IsNotFoundError(err) {
				continue
			}
			return false, fmt.Errorf("query: ListSIM failed: %w", err)
		}
		for _, sim := range sims.SIM {
			if sim.ResourceID.Value == target {
				return true, nil
			}
		}
	}
	return false, nil
}

// WaitWhileSIMIsReferenced は SIM が参照されている間 poll する。
func WaitWhileSIMIsReferenced(ctx context.Context, r ReferenceFinder, simID int64, option CheckReferencedOption) error {
	return waitWhileReferenced(ctx, option, func() (bool, error) {
		return IsSIMReferenced(ctx, r, simID)
	})
}

// ---------- internal helpers ----------

func listAllServers(ctx context.Context, f ServerFinder) ([]client.Server, error) {
	if f == nil {
		return nil, nil
	}
	resp, err := f.List(ctx, &client.ServerFindRequest{Count: 10000})
	if err != nil {
		return nil, fmt.Errorf("query: Server.List failed: %w", err)
	}
	return resp.Servers, nil
}

func listAllSwitches(ctx context.Context, f SwitchFinder) ([]client.Switch, error) {
	if f == nil {
		return nil, nil
	}
	resp, err := f.List(ctx, &client.SwitchFindRequest{Count: 10000})
	if err != nil {
		return nil, fmt.Errorf("query: Switch.List failed: %w", err)
	}
	return resp.Switches, nil
}
