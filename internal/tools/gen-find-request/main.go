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

// gen-find-request は v2/client/find_request_gen.go を生成する。
//
// 概要: Find 系エンドポイントの q= クエリ用にリソース別の typed request / filter
// struct を提供する。ユーザーは `XxxFindRequest{Count, From, Filter: XxxFindFilter{...}}`
// を構築し、`.ToOptString()` で `client.BridgeOpFindParams.Q` に渡す。
//
// ページングは全リソース一律 Count/From を持つ。フィルタは以下のいずれかを持つ:
//   - Name          (partial match, 部分一致。スペース区切りで AND)
//   - Tags          (exact match AND, 配列)
//   - Scope         (scoped リソース用)
//   - Class         (Appliance 個別リソース用)
//   - Provider.Class (CommonServiceItem 系, JSON タグ "Provider.Class")
//
// Sort / Include / Exclude は意図的に定義しない（AGENTS.md 参照）。
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("gen-find-request: ")
}

// filterFields はリソースがサポートするフィルタフィールドを表す。
type filterFields struct {
	Name          bool
	Tags          bool
	Scope         bool
	Class         bool // Appliance 個別リソース用
	ProviderClass bool // CommonServiceItem 系
}

// manifest はリソース名（TypeName）→ フィルタサポートフィールド。
// ここに列挙されたリソースのみ FindRequest/FindFilter が生成される。
//
// ※ 追加時の指針:
//   - Name: ほぼ全リソースが対応。プランや Region/Zone 等は部分一致で検索できる
//   - Tags: Name+Tags セットで検索可能なユーザー作成リソースのみ
//   - Scope: Archive/CDROM/Disk/Icon/Note/Switch 等の shared/user スコープ分岐があるもの
//   - Class: Appliance 個別リソース (DB/LB/MGW/NFS/VPC) の絞り込み用
//   - ProviderClass: CommonServiceItem 系 (DNS/GSLB/ProxyLB/SIM/etc)
var manifest = map[string]filterFields{
	// ユーザー作成リソース (Name+Tags)
	"Archive":              {Name: true, Tags: true, Scope: true},
	"Bridge":               {Name: true},
	"CDROM":                {Name: true, Tags: true, Scope: true},
	"Disk":                 {Name: true, Tags: true, Scope: true},
	"Icon":                 {Name: true, Tags: true, Scope: true},
	"Interface":            {Name: true},
	"Internet":             {Name: true, Tags: true},
	"License":              {Name: true},
	"Note":                 {Name: true, Tags: true, Scope: true},
	"PacketFilter":         {Name: true},
	"PrivateHost":          {Name: true, Tags: true},
	"Server":               {Name: true, Tags: true},
	"SSHKey":               {Name: true},
	"Subnet":               {Name: true},
	"Switch":               {Name: true, Tags: true, Scope: true},

	// プラン系 (Name のみ。Class を持つ PrivateHostPlan は別扱い)
	"DiskPlan":             {Name: true},
	"InternetPlan":         {Name: true},
	"LicenseInfo":          {Name: true},
	"PrivateHostPlan":      {Name: true, Class: true},
	"ServerPlan":           {Name: true},
	"ServiceClass":         {Name: true},

	// Appliance 個別 (Name+Tags+Class)
	"Database":             {Name: true, Tags: true, Class: true},
	"LoadBalancer":         {Name: true, Tags: true, Class: true},
	"MobileGateway":        {Name: true, Tags: true, Class: true},
	"NFS":                  {Name: true, Tags: true, Class: true},
	"VPCRouter":            {Name: true, Tags: true, Class: true},

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

type resourceGen struct {
	TypeName string
	Fields   filterFields
}

// HasFilter は Filter struct に載せるフィールドが存在するかどうか。
func (r resourceGen) HasFilter() bool {
	return r.Fields.Name || r.Fields.Tags || r.Fields.Scope || r.Fields.Class || r.Fields.ProviderClass
}

const fileHeader = `// Copyright 2022-2025 The sacloud/iaas-api-go Authors
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

// Code generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-find-request'; DO NOT EDIT.

package client

import "encoding/json"

`

const tmpl = `
// {{.TypeName}}FindFilter は {{.TypeName}} の Find クエリで指定可能なフィルタ条件。
// 省略されたフィールド (zero value) は q= に含まれない。
type {{.TypeName}}FindFilter struct {
{{- if .Fields.Name}}
	// Name は部分一致検索（スペース区切りで AND 結合）。
	Name string ` + "`json:\"Name,omitempty\"`" + `
{{- end}}
{{- if .Fields.Tags}}
	// Tags はタグ配列の完全一致 AND 検索。
	Tags []string ` + "`json:\"Tags,omitempty\"`" + `
{{- end}}
{{- if .Fields.Scope}}
	// Scope は "shared" / "user" のいずれか。部分一致。
	Scope string ` + "`json:\"Scope,omitempty\"`" + `
{{- end}}
{{- if .Fields.Class}}
	// Class は Appliance のサブクラス (例: "database"/"loadbalancer")。部分一致。
	Class string ` + "`json:\"Class,omitempty\"`" + `
{{- end}}
{{- if .Fields.ProviderClass}}
	// ProviderClass は CommonServiceItem の Provider.Class。部分一致。
	ProviderClass string ` + "`json:\"Provider.Class,omitempty\"`" + `
{{- end}}
}

// {{.TypeName}}FindRequest は {{.TypeName}} Find エンドポイント用の q= クエリ値。
// ToOptString() で OptString に変換して Params.Q に渡す。
type {{.TypeName}}FindRequest struct {
	// Count は取得件数の上限 (0 はサーバーデフォルト)。
	Count int32 ` + "`json:\"Count,omitempty\"`" + `
	// From はページング開始オフセット。
	From int32 ` + "`json:\"From,omitempty\"`" + `
{{- if .HasFilter}}
	// Filter はフィルタ条件。未指定の場合は q 全体から省略される。
	Filter {{.TypeName}}FindFilter ` + "`json:\"Filter,omitempty\"`" + `
{{- end}}
}

// Marshal は JSON エンコード後の文字列を返す。エラーは起こらない想定。
func (r *{{.TypeName}}FindRequest) Marshal() string {
	b, err := json.Marshal(r)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// ToOptString は Marshal 結果を OptString にラップする。
// q クエリパラメータ (Params.Q) への代入用。
func (r *{{.TypeName}}FindRequest) ToOptString() OptString {
	return NewOptString(r.Marshal())
}
`

const emptyFilterNote = `// メモ: Filter struct にフィールドが無いリソースは {{.TypeName}}FindFilter を生成しない。
// Count / From のみの {{.TypeName}}FindRequest を提供する。
`

func main() {
	resources := make([]resourceGen, 0, len(manifest))
	for name, fields := range manifest {
		resources = append(resources, resourceGen{TypeName: name, Fields: fields})
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].TypeName < resources[j].TypeName })

	t := template.Must(template.New("file").Parse(tmpl))
	var buf bytes.Buffer
	buf.WriteString(fileHeader)
	for _, r := range resources {
		if err := t.Execute(&buf, r); err != nil {
			log.Fatalf("execute template for %s: %v", r.TypeName, err)
		}
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		// デバッグ用に整形前の出力を dump
		fmt.Fprintln(os.Stderr, buf.String())
		log.Fatalf("gofmt failed: %v", err)
	}

	outPath := absPath("v2/client/find_request_gen.go")
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", filepath.Dir(outPath), err)
	}
	if err := os.WriteFile(outPath, src, 0o644); err != nil {
		log.Fatalf("write %s: %v", outPath, err)
	}
	log.Printf("generated: %s (%d resources)", outPath, len(resources))
}

// absPath は repo ルートからの相対パスを絶対パスに変換する。
func absPath(rel string) string {
	// 上位の iaas-api-go ルートからの相対パスで指定されることを前提とする
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}
	// 実行時のカレントディレクトリは通常 iaas-api-go のルート
	// （gen-typespec と同じ前提。go run ./internal/tools/gen-find-request で実行）
	if strings.Contains(wd, "/internal/tools/") {
		// ツールディレクトリから実行された場合、ルートを推定
		idx := strings.Index(wd, "/internal/tools/")
		wd = wd[:idx]
	}
	return filepath.Join(wd, rel)
}
