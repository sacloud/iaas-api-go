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
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/sacloud/iaas-api-go/internal/define"
	"github.com/sacloud/iaas-api-go/internal/dsl"
)

// typeNameToClass は DSL の TypeName をさくらクラウド API の "Class" フィールド値にマッピングする。
// 共有エンドポイントグループ（commonserviceitem / appliance）の union 型生成に使用する。
var typeNameToClass = map[string]string{
	// commonserviceitem グループ
	"DNS":                           "dns",
	"GSLB":                          "gslb",
	"AutoBackup":                    "autobackup",
	"AutoScale":                     "autoscale",
	"CertificateAuthority":          "certificateauthority",
	"ContainerRegistry":             "containerregistry",
	"EnhancedDB":                    "enhanceddb",
	"ESME":                          "esme",
	"LocalRouter":                   "localrouter",
	"ProxyLB":                       "proxylb",
	"SIM":                           "sim",
	"SimpleMonitor":                 "simplemon",
	"SimpleNotificationGroup":       "saknoticegroup",
	"SimpleNotificationDestination": "saknoticedestination",
	// appliance グループ
	"Database":      "database",
	"LoadBalancer":  "loadbalancer",
	"MobileGateway": "mobilegateway",
	"NFS":           "nfs",
	"VPCRouter":     "vpcrouter",
}

// pathNameToGroupName は共有エンドポイントの pathName を PascalCase のグループ名にマッピングする。
var pathNameToGroupName = map[string]string{
	"commonserviceitem": "CommonServiceItem",
	"appliance":         "Appliance",
}

// opParam は TypeSpec オペレーションの単一パラメータを表す。
type opParam struct {
	Decorator string // "@path"、"" など
	Name      string
	TSType    string
	Optional  bool // true のとき TypeSpec で `name?: type` と出力する
}

// opEntry は TypeSpec interface に出力する単一オペレーションの情報を保持する。
type opEntry struct {
	MethodNameLower string
	PathFormat      string
	HttpMethodLower string
	Params          []opParam
	ReturnType      string // TypeSpec の戻り値型名。envelope なしの場合は "void"
}

// fatField は fat model の1フィールドを表す。
type fatField struct {
	Name     string
	TSType   string
	Optional bool // true のとき `name?: type` と出力する
}

// fatModelDef は複数リソースのリクエスト型を統合した fat model を保持する。
// union → anyOf の代わりに、全バリアントのフィールドを1つのモデルに統合する。
type fatModelDef struct {
	Name     string
	AddClass bool // true のとき Class: string フィールドを先頭に出力する（create 系）
	Fields   []fatField
}

// opKey は HTTP メソッド + 解決済みパスで一意にオペレーションを識別する。
type opKey struct {
	method string
	path   string
}

// resolveOpPath は DSL のパスフォーマットを TypeSpec 互換のパス文字列に変換する。
func resolveOpPath(op *dsl.Operation, api *dsl.Resource) string {
	pf := op.GetPathFormat()
	pf = strings.ReplaceAll(pf, "{{.rootURL}}", "")
	pf = strings.ReplaceAll(pf, "{{.pathSuffix}}", api.GetPathSuffix())
	pf = strings.ReplaceAll(pf, "{{.pathName}}", api.GetPathName())
	pf = replaceGoTemplatePlaceholders(pf)
	if !strings.HasPrefix(pf, "/") {
		pf = "/" + pf
	}
	// DSL の childResourceName に先頭スラッシュが含まれると "//" が生じるため除去する
	// 例: simple_monitor.go の "/activity/responsetimesec" → "//activity/..." となるケース
	for strings.Contains(pf, "//") {
		pf = strings.ReplaceAll(pf, "//", "/")
	}
	return pf
}

// buildOpParams はオペレーションの全パラメータリストを構築する。
// @path パラメータを先に出力し、その後にボディパラメータを続ける。
func buildOpParams(op *dsl.Operation, resolvedPath string) []opParam {
	pathParams := extractPathParams(resolvedPath)
	pathParamSet := map[string]bool{}
	for _, p := range pathParams {
		pathParamSet[p] = true
	}

	var params []opParam
	for _, p := range pathParams {
		params = append(params, opParam{Decorator: "@path", Name: p, TSType: "string"})
	}
	for _, arg := range op.Arguments {
		if pathParamSet[arg.PathFormatName()] {
			continue
		}
		params = append(params, opParam{
			Name:   arg.ArgName(),
			TSType: goArgTypeToTS(arg.TypeName()),
		})
	}
	return params
}

// bodyArg は @path でないオペレーション引数を表す。
type bodyArg struct {
	name   string
	tsType string
}

// bodyArgs はオペレーションから @path でない引数を抽出して返す。
func bodyArgs(op *dsl.Operation, resolvedPath string) []bodyArg {
	pathParamSet := map[string]bool{}
	for _, p := range extractPathParams(resolvedPath) {
		pathParamSet[p] = true
	}
	var args []bodyArg
	for _, arg := range op.Arguments {
		if pathParamSet[arg.PathFormatName()] {
			continue
		}
		args = append(args, bodyArg{arg.ArgName(), goArgTypeToTS(arg.TypeName())})
	}
	return args
}

// primaryOpForKey は同一 opKey を持つ複数オペレーションから代表を選ぶ。
// 名前が短いもの（例: "Update" > "UpdateSettings"）を優先する。
func primaryOpForKey(candidates []*dsl.Operation) *dsl.Operation {
	best := candidates[0]
	for _, op := range candidates[1:] {
		if len(op.Name) < len(best.Name) {
			best = op
		}
	}
	return best
}

// typeSuffixAfterPrefix はリソースの TypeName プレフィックスを除いた型名サフィックスを返す。
// 例: resource="DNS", type="DNSCreateRequest" → "CreateRequest"
func typeSuffixAfterPrefix(resourceTypeName, tsType string) string {
	if strings.HasPrefix(tsType, resourceTypeName) {
		return tsType[len(resourceTypeName):]
	}
	return tsType
}

// responseTypeForOp はオペレーションの TypeSpec 戻り値型名を返す。
// レスポンス envelope がある場合はその PascalCase 名、なければ "void"。
func responseTypeForOp(op *dsl.Operation) string {
	if op.HasResponseEnvelope() {
		return upperFirst(op.ResponseEnvelopeStructName())
	}
	return "void"
}

// commonTypeSuffix は各リソースのバリアント型名からプレフィックスを除いた共通サフィックスを返す。
// サフィックスが全バリアントで一致しない場合は "" を返す。
func commonTypeSuffix(resources []*dsl.Resource, argVariants map[*dsl.Resource]string) string {
	var suffixes []string
	for _, res := range resources {
		t, ok := argVariants[res]
		if !ok || t == "" {
			continue // このオペレーションを持たないリソースはスキップ
		}
		suffixes = append(suffixes, typeSuffixAfterPrefix(res.TypeName(), t))
	}
	if len(suffixes) == 0 {
		return ""
	}
	first := suffixes[0]
	for _, s := range suffixes[1:] {
		if s != first {
			return ""
		}
	}
	return first
}

// generateOps はリソース別ディレクトリ構造で ops.tsp を生成する。
// 単一リソース → resources/{name}/ops.tsp
// 共有グループ → resources/{group_snake}/ops.tsp
func generateOps() {
	// API を pathName でグループ化
	pathNameAPIs := map[string][]*dsl.Resource{}
	var pathNameOrder []string
	seen := map[string]bool{}
	for _, api := range define.APIs {
		pn := api.GetPathName()
		if !seen[pn] {
			seen[pn] = true
			pathNameOrder = append(pathNameOrder, pn)
		}
		pathNameAPIs[pn] = append(pathNameAPIs[pn], api)
	}

	// 各グループを処理
	for _, pn := range pathNameOrder {
		apis := pathNameAPIs[pn]

		if len(apis) == 1 {
			// 単一リソース: resources/{name}/ops.tsp に出力
			api := apis[0]
			outFile := filepath.Join(resourcesDir, api.FileSafeName(), "ops.tsp")
			generateIndividualFile(api, outFile)
		} else {
			// 共有エンドポイントグループ: resources/{group_snake}/ops.tsp に出力
			groupName, ok := pathNameToGroupName[pn]
			if !ok {
				groupName = upperFirst(pn) // フォールバック
			}
			outFile := filepath.Join(resourcesDir, toSnake(groupName), "ops.tsp")
			generateSharedGroupFile(groupName, pn, apis, outFile)
		}
	}
}

// generateIndividualFile は単一リソース API の TypeSpec ファイルを生成する。
// 同一 method+path のオペレーション群は @sharedRoute を使わず 1つに統合する。
// ボディ引数は全オペレーションに共通なら required、一部にのみ存在する場合は optional にする。
func generateIndividualFile(api *dsl.Resource, outputPath string) {
	// opKey ごとにオペレーションをグループ化（元の順序を保持）
	type opGroup struct {
		key opKey
		ops []*dsl.Operation
	}
	var groups []opGroup
	groupIdx := map[opKey]int{}

	for _, op := range api.Operations {
		path := resolveOpPath(op, api)
		k := opKey{strings.ToLower(op.Method), path}
		if idx, ok := groupIdx[k]; ok {
			groups[idx].ops = append(groups[idx].ops, op)
		} else {
			groupIdx[k] = len(groups)
			groups = append(groups, opGroup{key: k, ops: []*dsl.Operation{op}})
		}
	}

	var ops []opEntry
	for _, g := range groups {
		path := g.key.path
		primary := primaryOpForKey(g.ops)
		pathParams := extractPathParams(path)
		pathParamSet := map[string]bool{}
		for _, p := range pathParams {
			pathParamSet[p] = true
		}

		// @path パラメータ
		var params []opParam
		for _, p := range pathParams {
			params = append(params, opParam{Decorator: "@path", Name: p, TSType: "string"})
		}

		if len(g.ops) == 1 {
			// 重複なし: エンベロープがあれば @body でラップして使用
			if primary.HasRequestEnvelope() {
				// リクエストエンベロープを使用（例: IconCreateRequestEnvelope → {"Icon": {...}}）
				params = append(params, opParam{
					Decorator: "@body",
					Name:      "body",
					TSType:    upperFirst(primary.RequestEnvelopeStructName()),
				})
			} else {
				for _, arg := range primary.Arguments {
					if pathParamSet[arg.PathFormatName()] {
						continue
					}
					params = append(params, opParam{
						Name:   arg.ArgName(),
						TSType: goArgTypeToTS(arg.TypeName()),
					})
				}
			}
		} else {
			// 重複あり: 全オペレーションのボディ引数をマージ
			// 全オペレーションに存在する引数は required、一部のみは optional
			params = append(params, mergeBodyArgs(g.ops, pathParamSet)...)
		}

		ops = append(ops, opEntry{
			MethodNameLower: lowerFirst(primary.Name),
			PathFormat:      path,
			HttpMethodLower: g.key.method,
			Params:          params,
			ReturnType:      responseTypeForOp(primary),
		})
	}

	type fileParam struct {
		TypeName   string
		Operations []opEntry
	}
	writeFile(individualOpTmpl, fileParam{TypeName: api.TypeName(), Operations: ops}, outputPath, template.FuncMap{"trimSpace": strings.TrimSpace})
}

// mergeBodyArgs は同一 method+path を持つ複数オペレーションのボディ引数をマージする。
// 全オペレーションに存在する引数は required、一部のみに存在する引数は optional にする。
func mergeBodyArgs(ops []*dsl.Operation, pathParamSet map[string]bool) []opParam {
	// 引数名の出現順序と出現回数を追跡
	type argInfo struct {
		name   string
		tsType string
		count  int // 何オペレーションに存在するか
	}
	seen := map[string]*argInfo{}
	var order []string

	for _, op := range ops {
		for _, arg := range op.Arguments {
			if pathParamSet[arg.PathFormatName()] {
				continue
			}
			name := arg.ArgName()
			ts := goArgTypeToTS(arg.TypeName())
			if info, ok := seen[name]; ok {
				info.count++
				// 型が異なる場合は最初に見た型を採用
				_ = ts
			} else {
				seen[name] = &argInfo{name: name, tsType: ts, count: 1}
				order = append(order, name)
			}
		}
	}

	total := len(ops)
	var params []opParam
	for _, name := range order {
		info := seen[name]
		params = append(params, opParam{
			Name:     info.name,
			TSType:   info.tsType,
			Optional: info.count < total, // 全 op に存在しない場合は optional
		})
	}
	return params
}

// generateSharedGroupFile は同一 pathName を共有するリソースグループの TypeSpec ファイルを生成する。
// 共有オペレーションには union 型を生成し、リソース固有オペレーションは個別 interface に出力する。
func generateSharedGroupFile(groupName, pathName string, resources []*dsl.Resource, outputPath string) {
	// --- 各リソースの opKey → []*dsl.Operation マップを構築 ---
	type resourceOpMap struct {
		resource *dsl.Resource
		byKey    map[opKey][]*dsl.Operation
	}
	resourceOps := make([]resourceOpMap, len(resources))
	for i, res := range resources {
		m := map[opKey][]*dsl.Operation{}
		for _, op := range res.Operations {
			path := resolveOpPath(op, res)
			k := opKey{strings.ToLower(op.Method), path}
			m[k] = append(m[k], op)
		}
		resourceOps[i] = resourceOpMap{res, m}
	}

	// --- 各 opKey が何リソースで使われているか数える ---
	keyResourceCount := map[opKey]int{}
	for _, rm := range resourceOps {
		for k := range rm.byKey {
			keyResourceCount[k]++
		}
	}

	// 2リソース以上で共有されている opKey を特定
	isShared := map[opKey]bool{}
	for k, c := range keyResourceCount {
		if c > 1 {
			isShared[k] = true
		}
	}

	// --- 最初のリソースのオペレーション順序で共有キーを並べる ---
	var orderedSharedKeys []opKey
	{
		seenK := map[opKey]bool{}
		for _, op := range resources[0].Operations {
			path := resolveOpPath(op, resources[0])
			k := opKey{strings.ToLower(op.Method), path}
			if isShared[k] && !seenK[k] {
				orderedSharedKeys = append(orderedSharedKeys, k)
				seenK[k] = true
			}
		}
		// 最初のリソースにない共有キーを末尾に追加
		for k := range isShared {
			if !seenK[k] {
				orderedSharedKeys = append(orderedSharedKeys, k)
				seenK[k] = true
			}
		}
	}

	// --- 共有オペレーションと fat model を構築 ---
	// union → anyOf は ogen が "complex anyOf" として未対応のため、
	// 全バリアントのフィールドを統合した fat model を生成する。
	var fatModels []fatModelDef
	var sharedOps []opEntry

	for _, k := range orderedSharedKeys {
		// このキーを持つ最初のリソースから代表オペレーションを取得
		var repRes *dsl.Resource
		var repOps []*dsl.Operation
		for _, rm := range resourceOps {
			if ops, ok := rm.byKey[k]; ok {
				repRes = rm.resource
				repOps = ops
				break
			}
		}
		repOp := primaryOpForKey(repOps)
		resolvedPath := resolveOpPath(repOp, repRes)

		// @path パラメータを取得（全リソース共通）
		pathParams := extractPathParams(resolvedPath)

		// 各リソースのボディ引数を収集（代表オペレーションを使用）
		resBodyArgs := map[*dsl.Resource][]bodyArg{}
		for _, rm := range resourceOps {
			ops, ok := rm.byKey[k]
			if !ok {
				continue
			}
			resBodyArgs[rm.resource] = bodyArgs(primaryOpForKey(ops), resolvedPath)
		}

		// パラメータリストを構築: @path を先に出力
		var params []opParam
		for _, p := range pathParams {
			params = append(params, opParam{Decorator: "@path", Name: p, TSType: "string"})
		}

		// ボディ引数: 全リソースで型が同じなら直接使用、異なれば fat model を生成
		firstArgs := resBodyArgs[resources[0]]
		for argIdx, firstArg := range firstArgs {
			// このオペレーションを持つリソースのみバリアント型を収集
			argVariants := map[*dsl.Resource]string{}
			allSame := true
			for _, rm := range resourceOps {
				args, ok := resBodyArgs[rm.resource]
				if !ok || argIdx >= len(args) {
					continue // このオペレーションを持たないリソースはスキップ
				}
				argVariants[rm.resource] = args[argIdx].tsType
				if args[argIdx].tsType != firstArg.tsType {
					allSame = false
				}
			}

			if allSame {
				// 全バリアントが同じ型 → そのまま使用
				params = append(params, opParam{Name: firstArg.name, TSType: firstArg.tsType})
			} else {
				// 型がリソースによって異なる → fat model を生成
				// fat model 名はリソース TypeName プレフィックスを除いた共通サフィックスから決定
				suffix := commonTypeSuffix(resources, argVariants)
				var fatModelName string
				if suffix != "" {
					fatModelName = groupName + suffix
				} else {
					fatModelName = groupName + upperFirst(lowerFirst(repOp.Name)) + upperFirst(firstArg.name)
				}

				// 同名の fat model がまだなければ追加
				alreadyAdded := false
				for _, fm := range fatModels {
					if fm.Name == fatModelName {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					// create (POST) の場合のみ Class フィールドを追加する
					addClass := strings.ToUpper(repOp.Method) == "POST"
					totalVariants := 0
					for _, rm := range resourceOps {
						if _, ok := rm.byKey[k]; ok {
							totalVariants++
						}
					}
					fatModels = append(fatModels, buildFatModelDefs(fatModelName, argVariants, resources, totalVariants, addClass)...)
				}
				params = append(params, opParam{Name: firstArg.name, TSType: fatModelName})
			}
		}

		sharedOps = append(sharedOps, opEntry{
			MethodNameLower: lowerFirst(repOp.Name),
			PathFormat:      resolvedPath,
			HttpMethodLower: k.method,
			Params:          params,
			ReturnType:      responseTypeForOp(repOp),
		})
	}

	// --- リソース固有オペレーション（ユニークパス）を収集 ---
	type resourceInterface struct {
		TypeName   string
		Operations []opEntry
	}
	var resourceInterfaces []resourceInterface

	for _, rm := range resourceOps {
		// ユニークオペレーションを元の順序で出力（opKey の重複は最初の1つのみ）
		seenUniqueKey := map[opKey]bool{}
		var ops []opEntry
		for _, op := range rm.resource.Operations {
			path := resolveOpPath(op, rm.resource)
			k := opKey{strings.ToLower(op.Method), path}
			if isShared[k] {
				continue // 共有オペレーションはスキップ
			}
			if seenUniqueKey[k] {
				continue // リソース内重複は最初の1つのみ出力
			}
			seenUniqueKey[k] = true
			ops = append(ops, opEntry{
				MethodNameLower: lowerFirst(op.Name),
				PathFormat:      path,
				HttpMethodLower: k.method,
				Params:          buildOpParams(op, path),
				ReturnType:      responseTypeForOp(op),
			})
		}

		if len(ops) > 0 {
			resourceInterfaces = append(resourceInterfaces, resourceInterface{
				TypeName:   rm.resource.TypeName(),
				Operations: ops,
			})
		}
	}

	// TypeSpec は interface 内の op 名の重複を禁止するため、衝突を解消する
	nameCount := map[string]int{}
	for i := range sharedOps {
		name := sharedOps[i].MethodNameLower
		if nameCount[name] > 0 {
			// パスに固有のパラメータがあればそれを suffix として使用
			extra := extraPathParamSuffix(sharedOps[i].PathFormat)
			if extra != "" {
				sharedOps[i].MethodNameLower = name + "By" + upperFirst(extra)
			} else {
				sharedOps[i].MethodNameLower = fmt.Sprintf("%s%d", name, nameCount[name]+1)
			}
		}
		nameCount[name]++
	}

	// 安定した出力順序のためにソート
	sort.Slice(resourceInterfaces, func(i, j int) bool {
		return resourceInterfaces[i].TypeName < resourceInterfaces[j].TypeName
	})

	type fileParam struct {
		GroupName          string
		FatModels          []fatModelDef
		SharedOps          []opEntry
		ResourceInterfaces []resourceInterface
	}
	writeFile(sharedGroupOpTmpl, fileParam{
		GroupName:          groupName,
		FatModels:          fatModels,
		SharedOps:          sharedOps,
		ResourceInterfaces: resourceInterfaces,
	}, outputPath, template.FuncMap{"trimSpace": strings.TrimSpace})
}

// --- パス・テンプレートユーティリティ ---

var pathParamRe = regexp.MustCompile(`\{(\w+)\}`)

// extractPathParams はパス文字列から {paramName} を抽出して返す。
func extractPathParams(path string) []string {
	matches := pathParamRe.FindAllStringSubmatch(path, -1)
	seen := map[string]bool{}
	var params []string
	for _, m := range matches {
		if !seen[m[1]] {
			seen[m[1]] = true
			params = append(params, m[1])
		}
	}
	return params
}

// {{if...}}...{{end}} ブロックを内容を保持しつつ除去する正規表現
var goTmplCondRe = regexp.MustCompile(`\{\{if[^}]*\}\}(.*?)\{\{end\}\}`)

// {{.xxx}} を {xxx} に変換する正規表現
var goTmplRe = regexp.MustCompile(`\{\{\.(\w+)\}\}`)

// replaceGoTemplatePlaceholders は Go テンプレートの記法を TypeSpec のパスパラメータ記法に変換する。
func replaceGoTemplatePlaceholders(s string) string {
	s = goTmplCondRe.ReplaceAllString(s, "$1")
	return goTmplRe.ReplaceAllString(s, "{$1}")
}

// goArgTypeToTS は Go の型文字列を TypeSpec の型名に変換する。
func goArgTypeToTS(goType string) string {
	if strings.HasPrefix(goType, "map[") {
		return "Record<unknown>"
	}
	if strings.HasPrefix(goType, "[]*") {
		return goArgTypeToTS(goType[3:]) + "[]"
	}
	if strings.HasPrefix(goType, "[]") {
		return goArgTypeToTS(goType[2:]) + "[]"
	}
	goType = strings.TrimPrefix(goType, "*")
	if idx := strings.LastIndex(goType, "."); idx >= 0 {
		goType = goType[idx+1:]
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
		return goType
	}
}

// extraPathParamSuffix はパスから zone/id 以外の最初のパスパラメータ名を返す。
// op 名の重複解消用サフィックスとして使用する。
func extraPathParamSuffix(path string) string {
	standard := map[string]bool{"zone": true, "id": true}
	for _, p := range extractPathParams(path) {
		if !standard[p] {
			return p
		}
	}
	return ""
}

// --- テンプレート ---

// individualOpTmpl は単一リソースの TypeSpec interface テンプレート。
const individualOpTmpl = `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-typespec'; DO NOT EDIT

// TypeSpec operation definition for {{ .TypeName }}

import "@typespec/http";

using TypeSpec.Http;

namespace Sacloud.IaaS;

@tag("{{ .TypeName }}")
interface {{ .TypeName }}Op {
{{ range .Operations }}
  @{{ .HttpMethodLower }}
  @route("{{ .PathFormat }}")
  op {{ .MethodNameLower }}(
    {{ range .Params }}{{ if .Decorator }}{{ .Decorator }} {{ end }}{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }},
    {{ end }}
  ): {{ if and (eq .HttpMethodLower "post") (ne .ReturnType "void") }}{@statusCode _: 201; ...{{ .ReturnType }}}{{ else if eq .HttpMethodLower "delete" }}{@statusCode _: 200; is_ok: boolean}{{ else }}{{ .ReturnType }}{{ end }} | ApiError;
{{ end }}
}
`

// sharedGroupOpTmpl は共有エンドポイントグループの TypeSpec テンプレート。
// union 型 + グループ共有 interface + リソース固有 interface を出力する。
const sharedGroupOpTmpl = `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-typespec'; DO NOT EDIT

// TypeSpec operations for {{ .GroupName }} shared endpoint group

import "@typespec/http";

using TypeSpec.Http;

namespace Sacloud.IaaS;
{{ range .FatModels }}
model {{ .Name }} {
  {{- if .AddClass }}
  Class: string;
  {{- end }}
  {{- range .Fields }}
  {{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }};
  {{- end }}
}
{{ end }}
@tag("{{ .GroupName }}")
interface {{ .GroupName }}Op {
{{ range .SharedOps }}
  @{{ .HttpMethodLower }}
  @route("{{ .PathFormat }}")
  op {{ .MethodNameLower }}(
    {{ range .Params }}{{ if .Decorator }}{{ .Decorator }} {{ end }}{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }},
    {{ end }}
  ): {{ if and (eq .HttpMethodLower "post") (ne .ReturnType "void") }}{@statusCode _: 201; ...{{ .ReturnType }}}{{ else if eq .HttpMethodLower "delete" }}{@statusCode _: 200; is_ok: boolean}{{ else }}{{ .ReturnType }}{{ end }} | ApiError;
{{ end }}
}
{{ range .ResourceInterfaces }}
@tag("{{ .TypeName }}")
interface {{ .TypeName }}Op {
{{ range .Operations }}
  @{{ .HttpMethodLower }}
  @route("{{ .PathFormat }}")
  op {{ .MethodNameLower }}({{ range .Params }}{{ if .Decorator }}{{ .Decorator }} {{ end }}{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }}, {{ end }}): {{ if and (eq .HttpMethodLower "post") (ne .ReturnType "void") }}{@statusCode _: 201; ...{{ .ReturnType }}}{{ else if eq .HttpMethodLower "delete" }}{@statusCode _: 200; is_ok: boolean}{{ else }}{{ .ReturnType }}{{ end }} | ApiError;
{{ end }}
}
{{ end }}`
