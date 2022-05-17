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

package ostype

import (
	"github.com/sacloud/iaas-api-go/search"
	"github.com/sacloud/iaas-api-go/search/keys"
)

// ArchiveCriteria OSTypeごとのアーカイブ検索条件
var ArchiveCriteria = map[ArchiveOSType]search.Filter{
	CentOS: {
		search.Key(keys.Tags): search.TagsAndEqual("distro-centos"),
	},
	CentOS8Stream: {
		search.Key(keys.Tags): search.TagsAndEqual("distro-ver-8-stream", "distro-centos"),
	},
	CentOS7: {
		search.Key(keys.Tags): search.TagsAndEqual("centos-7-latest"),
	},
	AlmaLinux: {
		search.Key(keys.Tags): search.TagsAndEqual("current-stable", "distro-alma"),
	},
	RockyLinux: {
		search.Key(keys.Tags): search.TagsAndEqual("current-stable", "distro-rocky"),
	},
	MiracleLinux: {
		search.Key(keys.Tags): search.TagsAndEqual("current-stable", "distro-miracle"),
	},
	Ubuntu: {
		search.Key(keys.Tags): search.TagsAndEqual("current-stable", "distro-ubuntu"),
	},
	Ubuntu2004: {
		search.Key(keys.Tags): search.TagsAndEqual("ubuntu-20.04-latest"),
	},
	Ubuntu1804: {
		search.Key(keys.Tags): search.TagsAndEqual("ubuntu-18.04-latest"),
	},
	Debian: {
		search.Key(keys.Tags): search.TagsAndEqual("current-stable", "distro-debian"),
	},
	Debian10: {
		search.Key(keys.Tags): search.TagsAndEqual("debian-10-latest"),
	},
	Debian11: {
		search.Key(keys.Tags): search.TagsAndEqual("debian-11-latest"),
	},
	RancherOS: {
		search.Key(keys.Tags): search.TagsAndEqual("current-stable", "distro-rancheros"),
	},
	K3OS: {
		search.Key(keys.Tags): search.TagsAndEqual("current-stable", "distro-k3os"),
	},
	Kusanagi: {
		search.Key(keys.Tags): search.TagsAndEqual("current-stable", "pkg-kusanagi"),
	},
}
