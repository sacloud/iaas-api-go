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

// Package ostype is define OS type of SakuraCloud public archive
package ostype

//go:generate stringer -type=ArchiveOSType

// ArchiveOSType パブリックアーカイブOS種別
type ArchiveOSType int

const (
	// Custom OS種別:カスタム
	Custom ArchiveOSType = iota

	// CentOS OS種別:CentOS
	CentOS
	// CentOS8Stream OS種別:CentOS8Stream
	CentOS8Stream
	// CentOS7 OS種別:CentOS7
	CentOS7

	// AlmaLinux OS種別: Alma Linux
	AlmaLinux
	// RockyLinux OS種別: Rocky Linux
	RockyLinux
	// MiracleLinux OS種別: MIRACLE LINUX
	MiracleLinux

	// Ubuntu OS種別:Ubuntu
	Ubuntu
	// Ubuntu2004 OS種別:Ubuntu(Focal Fossa)
	Ubuntu2004
	// Ubuntu1804 OS種別:Ubuntu(Bionic)
	Ubuntu1804

	// Debian OS種別:Debian
	Debian
	// Debian10 OS種別:Debian10
	Debian10
	// Debian11 OS種別:Debian11
	Debian11

	// RancherOS OS種別:RancherOS
	RancherOS
	// K3OS OS種別: k3OS
	K3OS

	// Kusanagi OS種別:Kusanagi(CentOS)
	Kusanagi
)

// ArchiveOSTypes アーカイブ種別のリスト
var ArchiveOSTypes = []ArchiveOSType{
	CentOS,
	CentOS8Stream,
	CentOS7,
	AlmaLinux,
	RockyLinux,
	MiracleLinux,
	Ubuntu,
	Ubuntu2004,
	Ubuntu1804,
	Debian,
	Debian10,
	Debian11,
	RancherOS,
	K3OS,
	Kusanagi,
}

// OSTypeShortNames OSTypeとして利用できる文字列のリスト
var OSTypeShortNames = []string{
	"centos", "centos8stream", "centos7",
	"almalinux", "rockylinux", "miracle", "miraclelinux",
	"ubuntu", "ubuntu2004", "ubuntu1804",
	"debian", "debian10", "debian11",
	"rancheros", "k3os", "kusanagi",
}

// IsSupportDiskEdit ディスクの修正機能をフルサポートしているか(Windowsは一部サポートのためfalseを返す)
func (o ArchiveOSType) IsSupportDiskEdit() bool {
	switch o {
	case CentOS, CentOS8Stream, CentOS7,
		AlmaLinux, RockyLinux, MiracleLinux,
		Ubuntu, Ubuntu2004, Ubuntu1804,
		Debian, Debian10, Debian11,
		RancherOS, K3OS, Kusanagi:
		return true
	default:
		return false
	}
}

// StrToOSType 文字列からArchiveOSTypesへの変換
func StrToOSType(osType string) ArchiveOSType {
	switch osType {
	case "centos":
		return CentOS
	case "centos8stream":
		return CentOS8Stream
	case "centos7":
		return CentOS7
	case "almalinux":
		return AlmaLinux
	case "rockylinux":
		return RockyLinux
	case "miracle", "miraclelinux":
		return MiracleLinux
	case "ubuntu":
		return Ubuntu
	case "ubuntu2004":
		return Ubuntu2004
	case "ubuntu1804":
		return Ubuntu1804
	case "debian":
		return Debian
	case "debian10":
		return Debian10
	case "debian11":
		return Debian11
	case "rancheros":
		return RancherOS
	case "k3os":
		return K3OS
	case "kusanagi":
		return Kusanagi
	default:
		return Custom
	}
}
