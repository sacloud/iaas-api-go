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

// Code generated by "stringer -type=ArchiveOSType"; DO NOT EDIT.

package ostype

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Custom-0]
	_ = x[CentOS-1]
	_ = x[CentOS7-2]
	_ = x[AlmaLinux-3]
	_ = x[AlmaLinux9-4]
	_ = x[AlmaLinux8-5]
	_ = x[RockyLinux-6]
	_ = x[RockyLinux9-7]
	_ = x[RockyLinux8-8]
	_ = x[MiracleLinux-9]
	_ = x[MiracleLinux8-10]
	_ = x[MiracleLinux9-11]
	_ = x[Ubuntu-12]
	_ = x[Ubuntu2204-13]
	_ = x[Ubuntu2004-14]
	_ = x[Ubuntu1804-15]
	_ = x[Debian-16]
	_ = x[Debian10-17]
	_ = x[Debian11-18]
	_ = x[Kusanagi-19]
}

const _ArchiveOSType_name = "CustomCentOSCentOS7AlmaLinuxAlmaLinux9AlmaLinux8RockyLinuxRockyLinux9RockyLinux8MiracleLinuxMiracleLinux8MiracleLinux9UbuntuUbuntu2204Ubuntu2004Ubuntu1804DebianDebian10Debian11Kusanagi"

var _ArchiveOSType_index = [...]uint8{0, 6, 12, 19, 28, 38, 48, 58, 69, 80, 92, 105, 118, 124, 134, 144, 154, 160, 168, 176, 184}

func (i ArchiveOSType) String() string {
	if i < 0 || i >= ArchiveOSType(len(_ArchiveOSType_index)-1) {
		return "ArchiveOSType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ArchiveOSType_name[_ArchiveOSType_index[i]:_ArchiveOSType_index[i+1]]
}
