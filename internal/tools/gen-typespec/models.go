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
	"strings"
	"text/template"

	"github.com/sacloud/iaas-api-go/internal/define"
	"github.com/sacloud/iaas-api-go/internal/dsl"
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
	"FindCondition": {"Sort": true, "Filter": true},
}

// resourceModels は1リソース分の全モデルを1ファイルに出力するためのテンプレートパラメータ。
type resourceModels struct {
	Models []*dsl.Model
}

const modelsTmpl = `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-typespec'; DO NOT EDIT

import "@typespec/http";

namespace Sacloud.IaaS;
{{ range .Models }}
model {{ .Name }} {
	{{- range filteredFields .Name .Fields }}
	{{- $tsDefault := convertDefaultValue .DefaultValue }}
	{{- $enumDefault := resolveEnumDefault .DefaultValue }}
	{{if $tsDefault }}{{.Name}}: {{goTypeToTypeSpec .TypeName}} = {{$tsDefault}};{{else}}{{if $enumDefault }}// Default: {{$enumDefault}} (ogen の complex defaults 未対応のため省略)
	{{else if .DefaultValue }}// Default value: {{.DefaultValue}}
	{{end}}{{.Name}}: {{goTypeToTypeSpec .TypeName}};{{end}}{{if .HasTag }}
	// Go tags: ` + "`" + `{{.TagString}}` + "`" + `
	{{end}}
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

func generateModels() {
	// 全リソースにまたがる出力済みモデル名を追跡し、重複定義を防ぐ。
	// 同じモデルが複数リソースで参照されている場合、最初に現れたリソースのファイルにのみ出力する。
	outputtedModels := map[string]bool{}

	for _, api := range define.APIs {
		allModels := resourceModelsForAPI(api)

		// 未出力のモデルのみ抽出
		var newModels []*dsl.Model
		for _, m := range allModels {
			if !outputtedModels[m.Name] {
				newModels = append(newModels, m)
				outputtedModels[m.Name] = true
			}
		}

		if len(newModels) == 0 {
			// このリソースで新たに定義するモデルがなければスキップ
			continue
		}

		outFile := filepath.Join(resourcesDir, api.FileSafeName(), "models.tsp")
		writeFile(modelsTmpl, resourceModels{Models: newModels}, outFile, template.FuncMap{
			"goTypeToTypeSpec":    modelFieldTypeToTS,
			"convertDefaultValue": convertDefaultValue,
			"resolveEnumDefault":  resolveEnumDefault,
			"filteredFields": func(modelName string, fields []*dsl.FieldDesc) []*dsl.FieldDesc {
				exclusions := modelFieldExclusions[modelName]
				if len(exclusions) == 0 {
					return fields
				}
				var result []*dsl.FieldDesc
				for _, f := range fields {
					if !exclusions[f.Name] {
						result = append(result, f)
					}
				}
				return result
			},
		})
	}
}
