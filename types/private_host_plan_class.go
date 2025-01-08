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

package types

import "strings"

const (
	// PrivateHostClassDynamic 標準
	PrivateHostClassDynamic = "dynamic"
	// PrivateHostClassWindows Windows
	PrivateHostClassWindows = "ms_windows"
)

// PrivateHostClasses PrivateHost.Classに指定できる有効な文字列
var PrivateHostClasses = []string{PrivateHostClassDynamic, PrivateHostClassWindows}

// PrivateHostClassString PrivateHost.Classに指定できる有効な文字列(スペース区切り)
var PrivateHostClassString = strings.Join(PrivateHostClasses, " ")
