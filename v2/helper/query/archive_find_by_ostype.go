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

	"github.com/sacloud/iaas-api-go/v2/client"
)

// ArchiveOSType はさくらのクラウドのパブリックアーカイブ OS 種別。
// v1 ostype.ArchiveOSType の v2 相当 (v1 依存しない形で複製)。
type ArchiveOSType int

const (
	// Custom は OS 種別: カスタム。
	Custom ArchiveOSType = iota
	// AlmaLinux は Alma Linux (最新安定版)。
	AlmaLinux
	AlmaLinux10
	AlmaLinux9
	AlmaLinux8
	RockyLinux
	RockyLinux10
	RockyLinux9
	RockyLinux8
	MiracleLinux
	MiracleLinux9
	MiracleLinux8
	Ubuntu
	Ubuntu2404
	Ubuntu2204
	Debian
	Debian12
	Debian11
	Kusanagi
)

// archiveCriteria は OS 種別ごとに ArchiveFindFilter の Tags / Scope を返す。
// Tag は AND 検索、Scope は "shared" 固定。
var archiveCriteria = map[ArchiveOSType]client.ArchiveFindFilter{
	AlmaLinux:     {Tags: []string{"current-stable", "distro-alma"}, Scope: "shared"},
	AlmaLinux10:   {Tags: []string{"alma-10-latest"}, Scope: "shared"},
	AlmaLinux9:    {Tags: []string{"alma-9-latest"}, Scope: "shared"},
	AlmaLinux8:    {Tags: []string{"alma-8-latest"}, Scope: "shared"},
	RockyLinux:    {Tags: []string{"current-stable", "distro-rocky"}, Scope: "shared"},
	RockyLinux10:  {Tags: []string{"rocky-10-latest"}, Scope: "shared"},
	RockyLinux9:   {Tags: []string{"rocky-9-latest"}, Scope: "shared"},
	RockyLinux8:   {Tags: []string{"rocky-8-latest"}, Scope: "shared"},
	MiracleLinux:  {Tags: []string{"current-stable", "distro-miracle"}, Scope: "shared"},
	MiracleLinux9: {Tags: []string{"miracle-9-latest"}, Scope: "shared"},
	MiracleLinux8: {Tags: []string{"miracle-8-latest"}, Scope: "shared"},
	Ubuntu:        {Tags: []string{"current-stable", "distro-ubuntu"}, Scope: "shared"},
	Ubuntu2404:    {Tags: []string{"ubuntu-24.04-latest"}, Scope: "shared"},
	Ubuntu2204:    {Tags: []string{"ubuntu-22.04-latest"}, Scope: "shared"},
	Debian:        {Tags: []string{"current-stable", "distro-debian"}, Scope: "shared"},
	Debian12:      {Tags: []string{"debian-12-latest"}, Scope: "shared"},
	Debian11:      {Tags: []string{"debian-11-latest"}, Scope: "shared"},
	Kusanagi:      {Tags: []string{"current-stable", "pkg-kusanagi"}, Scope: "shared"},
}

// FindArchiveByOSType は OS 種別ごとの最新安定版アーカイブを取得する。
func FindArchiveByOSType(ctx context.Context, finder ArchiveFinder, os ArchiveOSType) (*client.Archive, error) {
	filter, ok := archiveCriteria[os]
	if !ok {
		return nil, fmt.Errorf("unsupported ArchiveOSType: %d", os)
	}
	req := &client.ArchiveFindRequest{
		Count:  1,
		Filter: filter,
	}
	resp, err := finder.List(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Archives) == 0 {
		return nil, fmt.Errorf("archive not found for ArchiveOSType: %d", os)
	}
	a := resp.Archives[0]
	return &a, nil
}
