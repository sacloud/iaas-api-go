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

// Package ostype is define OS type of SakuraCloud public archive
package ostype

//go:generate stringer -type=ArchiveOSType

// ArchiveOSType パブリックアーカイブOS種別
type ArchiveOSType int

const (
	// Custom OS種別:カスタム
	Custom ArchiveOSType = iota

	// AlmaLinux OS種別: Alma Linux
	AlmaLinux
	// AlmaLinux10 OS種別: Alma Linux10
	AlmaLinux10
	// AlmaLinux9 OS種別: Alma Linux9
	AlmaLinux9

	// RockyLinux OS種別: Rocky Linux
	RockyLinux
	// RockyLinux10 OS種別: Rocky Linux10
	RockyLinux10
	// RockyLinux9 OS種別: Rocky Linux9
	RockyLinux9

	// MiracleLinux OS種別: MIRACLE LINUX
	MiracleLinux
	// MiracleLinux8 OS種別: MIRACLE LINUX8
	MiracleLinux8
	// MiracleLinux9 OS種別: MIRACLE LINUX9
	MiracleLinux9

	// Ubuntu OS種別:Ubuntu
	Ubuntu
	// Ubuntu2404 OS種別:Ubuntu
	Ubuntu2404
	// Ubuntu2204 OS種別:Ubuntu(Jammy Jellyfish)
	Ubuntu2204

	// Debian OS種別:Debian
	Debian
	// Debian11 OS種別:Debian11
	Debian11
	// Debian12 OS種別:Debian12
	Debian12

	// Kusanagi OS種別:Kusanagi(CentOS)
	Kusanagi
)

// ArchiveOSTypes アーカイブ種別のリスト
var ArchiveOSTypes = []ArchiveOSType{
	AlmaLinux,
	AlmaLinux10,
	AlmaLinux9,
	RockyLinux,
	RockyLinux10,
	RockyLinux9,
	MiracleLinux,
	MiracleLinux8,
	MiracleLinux9,
	Ubuntu,
	Ubuntu2404,
	Ubuntu2204,
	Debian,
	Debian11,
	Debian12,
	Kusanagi,
}

// OSTypeShortNames OSTypeとして利用できる文字列のリスト
var OSTypeShortNames = []string{
	"almalinux", "almalinux10", "almalinux9",
	"rockylinux", "rockylinux10", "rockylinux9",
	"miracle", "miraclelinux", "miracle8", "miraclelinux8", "miracle9", "miraclelinux9",
	"ubuntu", "ubuntu2404", "ubuntu2204",
	"debian", "debian11", "debian12",
	"kusanagi",
}

// IsSupportDiskEdit ディスクの修正機能をフルサポートしているか(Windowsは一部サポートのためfalseを返す)
func (o ArchiveOSType) IsSupportDiskEdit() bool {
	switch o {
	case
		AlmaLinux, AlmaLinux10, AlmaLinux9,
		RockyLinux, RockyLinux10, RockyLinux9,
		MiracleLinux, MiracleLinux8, MiracleLinux9,
		Ubuntu, Ubuntu2404, Ubuntu2204,
		Debian, Debian11, Debian12,
		Kusanagi:
		return true
	default:
		return false
	}
}

// StrToOSType 文字列からArchiveOSTypesへの変換
func StrToOSType(osType string) ArchiveOSType {
	switch osType {
	case "almalinux":
		return AlmaLinux
	case "almalinux10":
		return AlmaLinux10
	case "almalinux9":
		return AlmaLinux9
	case "rockylinux":
		return RockyLinux
	case "rockylinux10":
		return RockyLinux10
	case "rockylinux9":
		return RockyLinux9
	case "miracle", "miraclelinux":
		return MiracleLinux
	case "miracle8", "miraclelinux8":
		return MiracleLinux8
	case "miracle9", "miraclelinux9":
		return MiracleLinux9
	case "ubuntu":
		return Ubuntu
	case "ubuntu2404":
		return Ubuntu2404
	case "ubuntu2204":
		return Ubuntu2204
	case "debian":
		return Debian
	case "debian11":
		return Debian11
	case "debian12":
		return Debian12
	case "kusanagi":
		return Kusanagi
	default:
		return Custom
	}
}
