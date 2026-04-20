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

package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/sacloud/iaas-api-go/internal/define"
	"github.com/sacloud/iaas-api-go/internal/dsl"
	"github.com/sacloud/iaas-api-go/internal/dsl/meta"
)

const resourcesDir = "spec/typespec/resources/"

// varToEnumType: "InterfaceDrivers" → "EInterfaceDriver"
var varToEnumType = map[string]string{}

// enumMemberName: "EInterfaceDriver" + "VirtIO" → "virtio"
var enumMemberName = map[string]map[string]string{}

// allModelsByName は全リソースの DSL モデルを名前でアクセスできるようにしたマップ。
// generateSharedGroupFile の fat model 構築（buildFatModel）で参照する。
var allModelsByName map[string]*dsl.Model

func init() {
	buildEnumMaps()
	buildAllModels()
}

// buildAllModels は全 API の全モデルを名前でアクセスできるよう allModelsByName を初期化する。
func buildAllModels() {
	allModelsByName = map[string]*dsl.Model{}
	for _, api := range define.APIs {
		for _, m := range resourceModelsForAPI(api) {
			if _, exists := allModelsByName[m.Name]; !exists {
				allModelsByName[m.Name] = m
			}
		}
	}
}

// buildEnumMaps は types パッケージを解析して enum 変数名→型名、メンバー名→TypeSpec名のマップを構築する
func buildEnumMaps() {
	typesDir := filepath.Join(repoRoot, "types")

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, typesDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		log.Fatalf("failed to parse types dir: %v", err)
	}
	pkg := pkgs["types"]
	if pkg == nil {
		return
	}

	// enum ベース型を収集
	enumBaseTypes := map[string]string{}
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts := spec.(*ast.TypeSpec)
				ident, ok := ts.Type.(*ast.Ident)
				if !ok {
					continue
				}
				if ident.Name == "string" || ident.Name == "int" {
					enumBaseTypes[ts.Name.Name] = ident.Name
				}
			}
		}
	}

	// var の struct リテラルから varName→typeName とメンバー名マップを構築
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
					continue
				}
				// &struct{...} パターン（UnaryExpr）も対応する
				val := vs.Values[0]
				if ue, ok := val.(*ast.UnaryExpr); ok {
					val = ue.X
				}
				cl, ok := val.(*ast.CompositeLit)
				if !ok {
					continue
				}
				st, ok := cl.Type.(*ast.StructType)
				if !ok {
					continue
				}

				// struct フィールドの型から対象の enum 型を特定
				var targetType string
				for _, field := range st.Fields.List {
					ident, ok := field.Type.(*ast.Ident)
					if !ok {
						continue
					}
					if _, ok := enumBaseTypes[ident.Name]; ok {
						targetType = ident.Name
						break
					}
				}
				if targetType == "" {
					continue
				}

				varName := vs.Names[0].Name
				varToEnumType[varName] = targetType

				members := map[string]string{}
				for _, elt := range cl.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					goName := kv.Key.(*ast.Ident).Name
					members[goName] = strings.ToLower(goName)
				}
				enumMemberName[targetType] = members
			}
		}
	}
}

// convertDefaultValue は Go の DefaultValue 文字列を TypeSpec のデフォルト値表現に変換する。
// 変換できない場合は空文字列を返す。
func convertDefaultValue(defaultValue string) string {
	// 関数呼び出しや time パッケージは TypeSpec では表現不可
	if strings.Contains(defaultValue, "(") || strings.Contains(defaultValue, "time.") {
		return ""
	}
	// types.VarName.MemberName 形式（enum 型）のデフォルト値は出力しない。
	// TypeSpec で enum フィールドにデフォルト値を付けると OpenAPI で allOf + default の組み合わせになり、
	// ogen が "complex defaults" として未対応扱いにするため。
	if strings.HasPrefix(defaultValue, "types.") {
		return ""
	}
	return defaultValue
}

// resolveEnumDefault は Go の enum デフォルト値を TypeSpec 表現に変換して返す。
// enum 型でない場合は空文字列を返す。
// convertDefaultValue が enum を省略する際のコメント出力用。
func resolveEnumDefault(defaultValue string) string {
	if !strings.HasPrefix(defaultValue, "types.") {
		return ""
	}
	parts := strings.SplitN(strings.TrimPrefix(defaultValue, "types."), ".", 2)
	if len(parts) != 2 {
		return ""
	}
	varName, memberName := parts[0], parts[1]
	if typeName, ok := varToEnumType[varName]; ok {
		if members, ok := enumMemberName[typeName]; ok {
			if tsMember, ok := members[memberName]; ok {
				return typeName + "." + tsMember
			}
		}
	}
	return ""
}

// goTypeAliasMap は Go の型エイリアス・スカラー型を TypeSpec の型名にマッピングする。
// var struct{} パターンを持たない types パッケージの型や、
// 配列エイリアス型（DNSRecords など）をここで扱う。
var goTypeAliasMap = map[string]string{
	// Go 配列エイリアス → TypeSpec 配列型
	"DNSRecords":                     "DNSRecord[]",
	"GSLBServers":                    "GSLBServer[]",
	"LoadBalancerServers":            "LoadBalancerServer[]",
	"LoadBalancerVirtualIPAddresses": "LoadBalancerVirtualIPAddress[]",
	// types パッケージのスカラー型（enum var 定義なし → string へ変換）
	"ArchiveShareKey":            "string",
	"ExternalPermission":         "string",
	"WebUI":                      "string",
	"EnhancedDBType":             "string",
	"EnhancedDBRegion":           "string",
	"PacketFilterNetwork":        "string",
	"PacketFilterPort":           "string",
	"VPCFirewallNetwork":         "string",
	"VPCFirewallPort":            "string",
	"EProxyLBFixedStatusCode":    "int32", // HTTP ステータスコード（数値）
	"EProxyLBRedirectStatusCode": "int32", // HTTP ステータスコード（数値）
	"StringNumber":               "int32", // JSON では文字列だが数値として扱う
	// 検索パラメータ型（TypeSpec モデルなし）
	"SortKeys": "unknown",
	"Filter":   "Record<unknown>",
}

// modelFieldTypeToTS は Go の型名を TypeSpec の型名に変換する（再帰対応）。
// テンプレートには "goTypeToTypeSpec" キーで渡す。
func modelFieldTypeToTS(goType string) string {
	// map 型
	if strings.HasPrefix(goType, "map[") {
		return "Record<unknown>"
	}
	// interface{}
	if goType == "interface{}" {
		return "unknown"
	}

	// スライス型（[]*T と []T）は内側の型を再帰的に変換して [] を後置する
	if strings.HasPrefix(goType, "[]*") {
		return modelFieldTypeToTS(goType[3:]) + "[]"
	}
	if strings.HasPrefix(goType, "[]") {
		return modelFieldTypeToTS(goType[2:]) + "[]"
	}

	// ポインタ型を除去
	goType = strings.TrimPrefix(goType, "*")

	// パッケージ名を処理（time.Time / types.ID / types.Tags / types.EXxx など）
	if idx := strings.LastIndex(goType, "."); idx != -1 {
		pkg := goType[:idx]
		typeName := goType[idx+1:]
		switch {
		case pkg == "time" && typeName == "Time":
			return "utcDateTime"
		case pkg == "types" && typeName == "ID":
			return "int64"
		case pkg == "types" && typeName == "Tags":
			return "string[]" // Tags は []string
		case pkg == "types" && typeName == "StringFlag":
			return "string" // StringFlag は "True"/"False" を表す真偽値のラッパー型（v2ではstringで表現）
		default:
			goType = typeName // パッケージ名を除去して TypeSpec 型名として使用
		}
	}

	switch goType {
	case "string":
		return "string"
	case "int", "int32":
		return "int32"
	case "int64":
		return "int64"
	case "float32":
		return "float32"
	case "float64":
		return "float64"
	case "bool":
		return "boolean"
	default:
		// エイリアスマップに一致するものを優先して使用
		if ts, ok := goTypeAliasMap[goType]; ok {
			return ts
		}
		return goType // モデル型はそのまま TypeSpec の型名として使用
	}
}

// modelFieldExclusions は特定モデルで TypeSpec 出力をスキップするフィールド名のセット。
// DSL 定義には存在するが OpenAPI ドキュメントに含めないフィールドを指定する。
var modelFieldExclusions = map[string]map[string]bool{
	// Sort/Filter は型が未定義（unknown/Record<unknown>）になるためドキュメントから除外する
	// Include/Exclude は今後非推奨・廃止予定のためドキュメントから除外する
	"FindCondition": {"Sort": true, "Filter": true, "Include": true, "Exclude": true},
	// naked.Interface の client-side 仮想フィールド。API レスポンスには含まれず、v1 の
	// UnmarshalJSON が Switch 情報から算出している。API 定義を記述する v2 TypeSpec には載せない。
	// 表示用にこの値が必要な downstream（usacloud 等）は上位層で算出する必要がある。
	"InterfaceView":          {"UpstreamType": true},
	"VPCRouterInterface":     {"UpstreamType": true, "Index": true},
	"MobileGatewayInterface": {"UpstreamType": true, "Index": true},
	// naked.Server には BundleInfo フィールドが存在せず（Disk/Archive にのみ定義）、実 API レスポンス
	// にも含まれない。downstream からの参照も無いため除外する。DSL 側の `fields.BundleInfo()` 登録
	// は過去の名残と思われる。
	"Server": {"BundleInfo": true},
	// SourceInfo は他ゾーンから転送されたアーカイブにのみ含まれる。ArchiveUnderZone.ID は
	// `X-Sakura-Bigint-As-Int: 1` ヘッダに反して文字列で返ってくる仕様で decode が困難。
	// usacloud / terraform いずれも参照していないため除外する。
	"Archive": {"SourceInfo": true},
	// Appliance の Remark.Switch.ID / Remark.Zone.ID は `X-Sakura-Bigint-As-Int` ヘッダに反して文字列で返る
	// （上位の Appliance.Switch.ID / Zone.ID は int のため齟齬がある）。downstream は上位のフィールドを
	// 使うので除外する。共有エンドポイントの response は Database を代表型として使うため DatabaseRemark
	// のみで実害が出るが、将来 per-resource envelope を emit した場合に備えて個別 Remark も同様に除外しておく。
	"DatabaseRemark":      {"Switch": true, "Zone": true},
	"LoadBalancerRemark":  {"Switch": true, "Zone": true},
	"NFSRemark":           {"Switch": true, "Zone": true},
	"MobileGatewayRemark": {"Switch": true, "Zone": true},
	"VPCRouterRemark":     {"Switch": true, "Zone": true},
	// VPCRouter の Interfaces レスポンスは 8 スロットで未使用枠に null が混在する
	// (`[{...}, null, null, ...]`) ため ogen の strict decode と相性が悪い。共有 Appliance 応答の
	// 代表型 Database からは Interfaces を丸ごと除外する。downstream (terraform / usacloud) は
	// Appliance.Interfaces は直接使わないため実害なし。
	"Database": {"Interfaces": true},
}

// tsModelField はテンプレートに渡す TypeSpec フィールド情報。
// mapconv で Foo.ID にマッピングされるフィールドは Foo?: { ID: int64 } に変換済み。
type tsModelField struct {
	Name       string
	TSType     string
	Optional   bool
	TSDefault  string // TypeSpec デフォルト値（空なら省略）
	EnumDefault string // enum デフォルトのコメント用
	OtherDefault string // その他デフォルト値のコメント用
}

// nakedFieldIsNullable は naked 型の指定フィールドが null になりえるかを返す。
// json タグに omitempty が含まれる場合、または フィールドがポインタ型の場合に true を返す。
func nakedFieldIsNullable(nakedRT reflect.Type, fieldName string) bool {
	if nakedRT == nil {
		return false
	}
	// スライス・ポインタ型は要素型まで辿る
	for nakedRT.Kind() == reflect.Slice || nakedRT.Kind() == reflect.Ptr {
		nakedRT = nakedRT.Elem()
	}
	if nakedRT.Kind() != reflect.Struct {
		return false
	}
	sf, ok := nakedRT.FieldByName(fieldName)
	if !ok {
		return false
	}
	if sf.Type.Kind() == reflect.Ptr {
		return true
	}
	tag := sf.Tag.Get("json")
	for _, part := range strings.Split(tag, ",") {
		if part == "omitempty" {
			return true
		}
	}
	return false
}

// resolvedModel はテンプレート出力用に解決済みのモデル定義（本体 or 合成サブモデル）。
type resolvedModel struct {
	Name   string
	Fields []tsModelField
}

// fieldNullabilityOverrides は v1 naked 型の宣言と実 API の挙動が食い違うフィールドに対して、
// 明示的に nullable 扱いに切り替えるホワイトリスト。
// キー = TypeSpec モデル名（合成サブモデル名または DSL モデル名）、値 = nullable にするフィールド名のセット。
//
// 典型例: naked.UserSubnet.DefaultRoute は `string` 非ポインタだが、実 API は
// `{"DefaultRoute": null}` を返すため、v2 TypeSpec の該当フィールドは optional にする必要がある。
// 運用: 実 API で null 由来の decode 失敗に遭遇したら、該当の (モデル名, フィールド名) を
// このマップに追記し、同時に verify-typespec 側に期待チェックを追加する。
var fieldNullabilityOverrides = map[string]map[string]bool{
	// Server response の Interfaces[].Switch.UserSubnet は DefaultRoute が null で返ることがある
	"InterfaceViewSwitchUserSubnet": {
		"DefaultRoute":   true,
		"NetworkMaskLen": true,
	},
	"VPCRouterInterfaceSwitchUserSubnet": {
		"DefaultRoute":   true,
		"NetworkMaskLen": true,
	},
	// Switch 本体の UserSubnet も同様
	"SwitchUserSubnet": {
		"DefaultRoute":   true,
		"NetworkMaskLen": true,
	},
	// Internet レスポンス下にネストされる Switch の情報は {ID, (場合により IPv6Nets)} のみ返る。
	// Description/Tags は API 実レスポンスに含まれないため nullable にする。
	"SwitchInfo": {
		"Description": true,
		"Tags":        true,
	},
	// Bridge.BridgeInfo は別ゾーンに接続された Bridge のみが値を持つ（単一ゾーン時は API が返さない）。
	// required のままだと Create 直後の response decode が `invalid: BridgeInfo (field required)` で失敗する。
	"Bridge": {
		"BridgeInfo": true,
	},
	// Database 型は共有 Appliance endpoint の代表レスポンス型として使われるため、
	// Database 固有の InterfaceSettings / IPAddresses は他 Appliance（NFS/LB/VPCR 等）のレスポンスには
	// 含まれない。required のままだと NFS などを decode するときに失敗するので optional 化する。
	"Database": {
		"InterfaceSettings": true,
		"IPAddresses":       true,
		"Disk":              true,
	},
	// DatabaseSettingCommon の WebUI / SourceNetwork は実 API レスポンスで省略されることがある
	// （ユーザが指定しない場合、API は WebUI を返さない）。required のままだと decode が失敗する。
	"DatabaseSettingCommon": {
		"WebUI":         true,
		"SourceNetwork": true,
	},
}

// fieldNode は DSL モデルのフィールドを mapconv 経路でツリー化したノード。
// ルートが DSL モデル自体に対応し、子ノードがそのモデルのフィールド（または mapconv によってネスト
// される中間セグメント）となる。葉ノードは `leafField` に元 DSL フィールドを持つ。
type fieldNode struct {
	name       string
	children   []*fieldNode
	childIdx   map[string]*fieldNode
	leafField  *dsl.FieldDesc // 葉ノードでのみ設定
	leafIsArr  bool           // セグメント側に [] があった／フィールド型自体が配列の場合 true
}

func (n *fieldNode) getOrCreate(name string, isArr bool) *fieldNode {
	if n.childIdx == nil {
		n.childIdx = map[string]*fieldNode{}
	}
	if c, ok := n.childIdx[name]; ok {
		if isArr {
			c.leafIsArr = true
		}
		return c
	}
	c := &fieldNode{name: name, leafIsArr: isArr}
	n.childIdx[name] = c
	n.children = append(n.children, c)
	return c
}

// insertField は segs（mapconv パス）に沿って f を tree に挿入する。
func insertField(root *fieldNode, segs []string, f *dsl.FieldDesc) {
	node := root
	for i, seg := range segs {
		isArr := strings.HasPrefix(seg, "[]")
		if isArr {
			seg = seg[2:]
		}
		child := node.getOrCreate(seg, isArr)
		if i == len(segs)-1 {
			child.leafField = f
		}
		node = child
	}
}

// isResourceRefShortcut は child が {ID} のみを持つ葉構造（つまり `Foo.ID` 単独）か判定する。
// true なら ResourceRef | null でショートカットできる。
func isResourceRefShortcut(child *fieldNode) bool {
	if len(child.children) != 1 {
		return false
	}
	gc := child.children[0]
	return gc.name == "ID" && gc.leafField != nil && len(gc.children) == 0
}

// resolveModel は DSL モデルを TypeSpec モデル群に変換する。
// mapconv を tree に展開し、任意階層のネストに対応する:
//   - `Foo.ID` 単独 → `Foo?: ResourceRef | null`
//   - `Foo.Bar` / `Foo.Bar.Baz` などそれ以外 → 合成サブモデル `{Parent}{Foo}` を生成し `Foo?: ... | null`
//   - 合成名が既存 DSL モデルと衝突する場合は既存モデルを再利用し、合成を省略する
//   - 深さ 3 以上でも再帰的に同じルールでサブモデルを生成する
//
// 配列セグメント（`[]X`）を含む mapconv パスは tree ネスト化せず、当該フィールドを平坦にそのまま
// 出力する。fat_model/既存挙動との互換のため。
func resolveModel(m *dsl.Model) []resolvedModel {
	exclusions := modelFieldExclusions[m.Name]

	var nakedRT reflect.Type
	if m.HasNakedType() {
		if st, ok := m.NakedType.(*meta.StaticType); ok {
			nakedRT = st.ReflectType
		}
	}

	root := &fieldNode{name: m.Name}
	for _, f := range m.Fields {
		if exclusions[f.Name] {
			continue
		}
		var segs []string
		if f.Tags != nil {
			segs, _, _ = parseMapconvPath(f.Tags.MapConv)
		}
		if len(segs) == 0 {
			segs = []string{f.Name}
		}
		// 配列セグメントを含むパスは tree nesting を適用しない（既存挙動維持）
		hasArrayInPath := false
		for _, s := range segs {
			if strings.HasPrefix(s, "[]") {
				hasArrayInPath = true
				break
			}
		}
		if hasArrayInPath {
			insertField(root, []string{f.Name}, f)
			continue
		}
		insertField(root, segs, f)
	}

	return emitFromFieldTree(m.Name, root, nakedRT)
}

// emitFromFieldTree は tree を根から辿って TypeSpec モデルを生成する。
// - modelName: この呼び出しが emit する model の名前
// - node:      その model のフィールド集合（= node.children）
// - nakedRT:   この model に対応する naked struct の reflect.Type（null チェック用、無ければ nil）
func emitFromFieldTree(modelName string, node *fieldNode, nakedRT reflect.Type) []resolvedModel {
	var mainFields []tsModelField
	var subs []resolvedModel

	// 合成サブモデル（例: `DatabaseRemark`）側でフィールドをスキップしたいケース向けに、
	// ルートの `resolveModel` だけでなく再帰呼び出しでも modelFieldExclusions を参照する。
	subExclusions := modelFieldExclusions[modelName]

	for _, child := range node.children {
		if subExclusions[child.name] {
			continue
		}
		nullable := nakedFieldIsNullable(nakedRT, child.name)
		if overrides, ok := fieldNullabilityOverrides[modelName]; ok && overrides[child.name] {
			nullable = true
		}

		// 葉ノード
		if child.leafField != nil && len(child.children) == 0 {
			tsType := modelFieldTypeToTS(child.leafField.TypeName())
			if child.leafIsArr && !strings.HasSuffix(tsType, "[]") {
				tsType = tsType + "[]"
			}
			if nullable {
				tsType = tsType + " | null"
			}
			tsDefault := convertDefaultValue(child.leafField.DefaultValue)
			mainFields = append(mainFields, tsModelField{
				Name:      child.name,
				TSType:    tsType,
				Optional:  nullable,
				TSDefault: tsDefault,
				EnumDefault: func() string {
					if tsDefault == "" {
						return resolveEnumDefault(child.leafField.DefaultValue)
					}
					return ""
				}(),
				OtherDefault: func() string {
					if tsDefault == "" && resolveEnumDefault(child.leafField.DefaultValue) == "" && child.leafField.DefaultValue != "" {
						return child.leafField.DefaultValue
					}
					return ""
				}(),
			})
			continue
		}

		// 中間ノード
		// `.ID` のみを持つ → ResourceRef | null に短縮
		subName := modelName + child.name
		_, reuseExisting := allModelsByName[subName]

		if isResourceRefShortcut(child) && !reuseExisting {
			tsType := "ResourceRef"
			if child.leafIsArr {
				tsType += "[]"
			}
			tsType += " | null"
			mainFields = append(mainFields, tsModelField{
				Name:     child.name,
				TSType:   tsType,
				Optional: true,
			})
			continue
		}

		// サブモデル参照（既存再利用 or 新規合成）
		var subRT reflect.Type
		if nakedRT != nil {
			if sf, ok := nakedRT.FieldByName(child.name); ok {
				t := sf.Type
				for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice {
					t = t.Elem()
				}
				if t.Kind() == reflect.Struct {
					subRT = t
				}
			}
		}

		if !reuseExisting {
			subs = append(subs, emitFromFieldTree(subName, child, subRT)...)
		}

		tsType := subName
		if child.leafIsArr {
			tsType += "[]"
		}
		if nullable {
			tsType += " | null"
		}
		mainFields = append(mainFields, tsModelField{
			Name:     child.name,
			TSType:   tsType,
			Optional: nullable,
		})
	}

	return append([]resolvedModel{{Name: modelName, Fields: mainFields}}, subs...)
}

// resourceModels は1リソース分の全モデルを1ファイルに出力するためのテンプレートパラメータ。
type resourceModels struct {
	Models []resolvedModel
}

const modelsTmpl = `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-typespec'; DO NOT EDIT

import "@typespec/http";

namespace Sacloud.IaaS;
{{ range .Models }}
model {{ .Name }} {
	{{- range .Fields }}
	{{- if .TSDefault }}
	{{.Name}}{{ if .Optional }}?{{ end }}: {{.TSType}} = {{.TSDefault}};
	{{- else if .EnumDefault }}
	// Default: {{.EnumDefault}} (ogen の complex defaults 未対応のため省略)
	{{.Name}}{{ if .Optional }}?{{ end }}: {{.TSType}};
	{{- else if .OtherDefault }}
	// Default value: {{.OtherDefault}}
	{{.Name}}{{ if .Optional }}?{{ end }}: {{.TSType}};
	{{- else }}
	{{.Name}}{{ if .Optional }}?{{ end }}: {{.TSType}};
	{{- end }}
	{{- end }}
}
{{ end }}`

// resourceModelsForAPI は api のオペレーションから全モデルを重複なく収集する。
func resourceModelsForAPI(api *dsl.Resource) []*dsl.Model {
	ms := dsl.Models{}
	for _, op := range api.Operations {
		ms = append(ms, op.Models()...)
	}
	return ms.UniqByName()
}

// computeRequestModelMerges は同一 method+path の op 群で同名 payload に異なる request model が
// 割り当てられているケースを検出し、primary の model 名に variant の DSL model を合流させる計画を返す。
//
// 背景: v1 DSL では Go メソッドの使いやすさを優先して 1 エンドポイントを複数メソッドに分割している
// （例: POST /archive の Create（SourceDisk/SourceArchive 指定）と CreateBlank（SizeMB 指定））が、
// API 定義としてはどちらも同じ 1 エンドポイントで、wire の envelope も同一構造。
// 公式マニュアルが定義するフィールドはすべて受けられるよう、primary の model にすべての variant の
// フィールドを union する。variant 側の request model は v2 では emit しない。
//
// 戻り値:
//   - merges:   primary model 名 → 合流対象の DSL model リスト（primary 自身を先頭に含む）
//   - skipSet:  models.tsp で emit を省略する variant model 名のセット
func computeRequestModelMerges(api *dsl.Resource) (merges map[string][]*dsl.Model, skipSet map[string]bool) {
	merges = map[string][]*dsl.Model{}
	skipSet = map[string]bool{}

	type opGroup struct {
		key opKey
		ops []*dsl.Operation
	}
	var groups []opGroup
	groupIdx := map[opKey]int{}
	for _, op := range api.Operations {
		if opIsExcluded(api, op) {
			continue
		}
		k := opKey{strings.ToLower(op.Method), resolveOpPath(op, api)}
		if idx, ok := groupIdx[k]; ok {
			groups[idx].ops = append(groups[idx].ops, op)
		} else {
			groupIdx[k] = len(groups)
			groups = append(groups, opGroup{key: k, ops: []*dsl.Operation{op}})
		}
	}

	for _, g := range groups {
		if len(g.ops) < 2 {
			continue
		}
		primary := primaryOpForKey(g.ops)

		type payloadEntry struct {
			models []*dsl.Model
			seen   map[string]bool
		}
		payloadMap := map[string]*payloadEntry{}

		collect := func(op *dsl.Operation) {
			for _, p := range op.RequestPayloads() {
				for _, arg := range op.Arguments {
					model, ok := arg.Type.(*dsl.Model)
					if !ok {
						continue
					}
					destField := strings.SplitN(arg.MapConvTag, ",", 2)[0]
					if destField != p.Name {
						continue
					}
					entry := payloadMap[p.Name]
					if entry == nil {
						entry = &payloadEntry{seen: map[string]bool{}}
						payloadMap[p.Name] = entry
					}
					if !entry.seen[model.Name] {
						entry.seen[model.Name] = true
						entry.models = append(entry.models, model)
					}
					break
				}
			}
		}
		collect(primary)
		for _, op := range g.ops {
			if op != primary {
				collect(op)
			}
		}

		for _, entry := range payloadMap {
			if len(entry.models) < 2 {
				continue
			}
			primaryModel := entry.models[0]
			merges[primaryModel.Name] = entry.models
			for _, m := range entry.models[1:] {
				skipSet[m.Name] = true
			}
		}
	}
	return merges, skipSet
}

// mergedDSLFields は複数 DSL model の Fields を union する。
// 同一フィールド名（mapconv root が同一の場合も含む）は最初の model の定義を採用する。
func mergedDSLFields(models []*dsl.Model) []*dsl.FieldDesc {
	seen := map[string]bool{}
	var result []*dsl.FieldDesc
	for _, m := range models {
		for _, f := range m.Fields {
			key := f.Name
			if f.Tags != nil {
				mc := strings.SplitN(f.Tags.MapConv, ",", 2)[0]
				if mc != "" {
					if segs := strings.Split(mc, "."); len(segs) > 0 && !strings.HasPrefix(segs[0], "[]") {
						key = segs[0]
					}
				}
			}
			if seen[key] {
				continue
			}
			seen[key] = true
			result = append(result, f)
		}
	}
	return result
}

// filteredModelsForAPI は resourceModelsForAPI から除外対象 op 経由のみで参照されるモデルを除いたもの。
// excludedOps に指定された op は仕様として生成対象外のため、そのリクエストモデルも emit しない。
func filteredModelsForAPI(api *dsl.Resource) []*dsl.Model {
	ms := dsl.Models{}
	for _, op := range api.Operations {
		if opIsExcluded(api, op) {
			continue
		}
		ms = append(ms, op.Models()...)
	}
	return ms.UniqByName()
}

func generateModels() {
	// 全リソースにまたがる出力済みモデル名を追跡し、重複定義を防ぐ。
	// 同じモデル（DSL 本体 or 合成サブモデル）が複数リソースで現れる場合、最初に現れたリソースのファイルにのみ出力する。
	outputtedModels := map[string]bool{}

	for _, api := range define.APIs {
		allModels := filteredModelsForAPI(api)
		merges, skipSet := computeRequestModelMerges(api)

		var newModels []resolvedModel
		for _, m := range allModels {
			if skipSet[m.Name] {
				// 同一パスの別 op で primary に合流させるため個別には emit しない
				continue
			}
			if outputtedModels[m.Name] {
				continue
			}

			// primary model（合流対象あり）の場合、variant の Fields を union した一時 DSL model を作って処理
			if variants, isMergePrimary := merges[m.Name]; isMergePrimary {
				m = &dsl.Model{
					Name:      m.Name,
					NakedType: m.NakedType,
					IsArray:   m.IsArray,
					Fields:    mergedDSLFields(variants),
				}
			}

			for _, rm := range resolveModel(m) {
				if outputtedModels[rm.Name] {
					continue
				}
				outputtedModels[rm.Name] = true
				newModels = append(newModels, rm)
			}
		}

		if len(newModels) == 0 {
			continue
		}

		outFile := filepath.Join(resourcesDir, api.FileSafeName(), "models.tsp")
		writeFile(modelsTmpl, resourceModels{Models: newModels}, outFile, nil)
	}
}
