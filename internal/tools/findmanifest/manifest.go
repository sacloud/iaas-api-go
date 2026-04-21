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

// Package findmanifest holds the filter-field manifest for Find 系エンドポイントの
// q クエリパラメータ。gen-find-request（クライアント struct を生成）と
// gen-typespec（q パラメータの @doc / @example を生成）が共通に参照する。
//
// Sort / Include / Exclude は意図的に定義しない（AGENTS.md「Find クエリ設計」参照）。
package findmanifest

// FilterFields は 1 リソースがサポートするフィルタフィールド。
// 全ての Find エンドポイントは Count / From を持つため、ここでは Filter 配下のフィールドだけを扱う。
type FilterFields struct {
	Name          bool
	Tags          bool
	Scope         bool
	Class         bool // Appliance 個別リソース用
	ProviderClass bool // CommonServiceItem 系
}

// HasAny は Filter struct に載せるフィールドが一つでもあるかどうかを返す。
func (f FilterFields) HasAny() bool {
	return f.Name || f.Tags || f.Scope || f.Class || f.ProviderClass
}

// Manifest はリソース名（TypeName）→ サポートフィルタ。
// 追加時の指針:
//   - Name: ほぼ全リソースが対応。プランや Region/Zone 等は部分一致で検索できる
//   - Tags: Name+Tags セットで検索可能なユーザー作成リソースのみ
//   - Scope: Archive/CDROM/Disk/Icon/Note/Switch 等の shared/user スコープ分岐があるもの
//   - Class: Appliance 個別リソース (DB/LB/MGW/NFS/VPC) の絞り込み用
//   - ProviderClass: CommonServiceItem 系 (DNS/GSLB/ProxyLB/SIM/etc)
var Manifest = map[string]FilterFields{
	// ユーザー作成リソース (Name+Tags)
	"Archive":      {Name: true, Tags: true, Scope: true},
	"Bridge":       {Name: true},
	"CDROM":        {Name: true, Tags: true, Scope: true},
	"Disk":         {Name: true, Tags: true, Scope: true},
	"Icon":         {Name: true, Tags: true, Scope: true},
	"Interface":    {Name: true},
	"Internet":     {Name: true, Tags: true},
	"License":      {Name: true},
	"Note":         {Name: true, Tags: true, Scope: true},
	"PacketFilter": {Name: true},
	"PrivateHost":  {Name: true, Tags: true},
	"Server":       {Name: true, Tags: true},
	"SSHKey":       {Name: true},
	"Subnet":       {Name: true},
	"Switch":       {Name: true, Tags: true, Scope: true},

	// プラン系 (Name のみ。Class を持つ PrivateHostPlan は別扱い)
	"DiskPlan":        {Name: true},
	"InternetPlan":    {Name: true},
	"LicenseInfo":     {Name: true},
	"PrivateHostPlan": {Name: true, Class: true},
	"ServerPlan":      {Name: true},
	"ServiceClass":    {Name: true},

	// Appliance 個別 (Name+Tags+Class)
	"Database":      {Name: true, Tags: true, Class: true},
	"LoadBalancer":  {Name: true, Tags: true, Class: true},
	"MobileGateway": {Name: true, Tags: true, Class: true},
	"NFS":           {Name: true, Tags: true, Class: true},
	"VPCRouter":     {Name: true, Tags: true, Class: true},

	// CommonServiceItem 系 (Name+Tags+ProviderClass)
	"AutoBackup":                    {Name: true, Tags: true, ProviderClass: true},
	"AutoScale":                     {Name: true, Tags: true, ProviderClass: true},
	"CertificateAuthority":          {Name: true, Tags: true, ProviderClass: true},
	"ContainerRegistry":             {Name: true, Tags: true, ProviderClass: true},
	"DNS":                           {Name: true, Tags: true, ProviderClass: true},
	"EnhancedDB":                    {Name: true, Tags: true, ProviderClass: true},
	"ESME":                          {Name: true, Tags: true, ProviderClass: true},
	"GSLB":                          {Name: true, Tags: true, ProviderClass: true},
	"LocalRouter":                   {Name: true, Tags: true, ProviderClass: true},
	"ProxyLB":                       {Name: true, Tags: true, ProviderClass: true},
	"SIM":                           {Name: true, Tags: true, ProviderClass: true},
	"SimpleMonitor":                 {Name: true, Tags: true, ProviderClass: true},
	"SimpleNotificationDestination": {Name: true, Tags: true, ProviderClass: true},
	"SimpleNotificationGroup":       {Name: true, Tags: true, ProviderClass: true},

	// Facility (IDで問い合わせが普通だが Name 部分一致もサポート)
	"Region": {Name: true},
	"Zone":   {Name: true},
}

// GroupManifest は共有エンドポイントグループ（appliance / commonserviceitem）の
// Find に対するフィルタフィールド。個別 FindRequest/FindFilter は生成しないので
// Manifest には入れず、gen-typespec の q @doc 生成でのみ参照する。
//
// 各グループの Find レスポンスは全バリアントを含む。Filter フィールドは全バリアントが
// 共通で持つフィルタ（Name, Tags）+ そのグループ固有の絞り込みキー（Class または
// Provider.Class）で構成される。
var GroupManifest = map[string]FilterFields{
	"Appliance":         {Name: true, Tags: true, Class: true},
	"CommonServiceItem": {Name: true, Tags: true, ProviderClass: true},
}
