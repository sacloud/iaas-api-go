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

// Package query は v2 リソースの検索 / 参照チェック系ヘルパーを提供する。
//
// v1 の github.com/sacloud/iaas-api-go/helper/query の v2 相当。
// v1 のような iaas.FindCondition / search.Filter には依存せず、v2 の
// typed FindRequest (client.ArchiveFindRequest 等) を組み立てて呼び出す。
package query

import (
	"context"

	"github.com/sacloud/iaas-api-go/v2/client"
)

// ArchiveFinder は Archive の検索 I/F。
// iaas.ArchiveAPI の List はこれを満たす。
type ArchiveFinder interface {
	List(ctx context.Context, req *client.ArchiveFindRequest) (*client.ArchiveFindResponseEnvelope, error)
}

// ArchiveReader は Archive の Read I/F。
type ArchiveReader interface {
	Read(ctx context.Context, id int64) (*client.ArchiveReadResponseEnvelope, error)
}

// DiskReader は Disk の Read I/F。
type DiskReader interface {
	Read(ctx context.Context, id int64) (*client.DiskReadResponseEnvelope, error)
}

// ServerPlanFinder は ServerPlan の検索 I/F。
type ServerPlanFinder interface {
	List(ctx context.Context, req *client.ServerPlanFindRequest) (*client.ServerPlanFindResponseEnvelope, error)
}

// ServerReader は Server の Read I/F。
type ServerReader interface {
	Read(ctx context.Context, id int64) (*client.ServerReadResponseEnvelope, error)
}

// ServerFinder は Server の検索 I/F。
type ServerFinder interface {
	List(ctx context.Context, req *client.ServerFindRequest) (*client.ServerFindResponseEnvelope, error)
}

// SwitchFinder は Switch の検索 I/F。
type SwitchFinder interface {
	List(ctx context.Context, req *client.SwitchFindRequest) (*client.SwitchFindResponseEnvelope, error)
}

// SwitchReader は Switch の Read + GetServers I/F。
type SwitchReader interface {
	Read(ctx context.Context, id int64) (*client.SwitchReadResponseEnvelope, error)
	GetServers(ctx context.Context, id int64) (*client.SwitchGetServersResponseEnvelope, error)
}

// ApplianceLister は Appliance の一覧 I/F (iaas.ApplianceAPI.List)。
type ApplianceLister interface {
	List(ctx context.Context, q client.OptString) (*client.DatabaseFindResponseEnvelope, error)
}

// MobileGatewaySIMLister は MobileGateway ごとに SIM を列挙する I/F。
type MobileGatewaySIMLister interface {
	ListSIM(ctx context.Context, id int64) (*client.MobileGatewayListSIMResponseEnvelope, error)
}

// NoteFinder は Note (StartupScript) の検索 I/F。
type NoteFinder interface {
	List(ctx context.Context, req *client.NoteFindRequest) (*client.NoteFindResponseEnvelope, error)
}
