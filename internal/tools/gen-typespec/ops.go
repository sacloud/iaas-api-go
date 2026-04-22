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
	"github.com/sacloud/iaas-api-go/internal/tools/findmanifest"
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
	Optional  bool   // true のとき TypeSpec で `name?: type` と出力する
	Doc       string // non-empty のとき @doc("""...""") をパラメータに付与する（例: Find の q パラメータ）
	Example   string // non-empty のとき @example("...") をパラメータに付与する。Doc と独立
}

// opTemplateFuncs はパラメータの @doc 出力などで使う共通 FuncMap。
//
// indent はテンプレート内で複数行文字列を TypeSpec のトリプル引用符内に埋めるために使う。
// TypeSpec の `"""..."""` は閉じる `"""` の行頭インデントを基準に各行のインデントを剥がす仕様。
// 各行に n スペースを前置することで、閉じ `"""` と揃えて正しく dedent させる。
var opTemplateFuncs = template.FuncMap{
	"trimSpace": strings.TrimSpace,
	"indent": func(n int, s string) string {
		pad := strings.Repeat(" ", n)
		return pad + strings.ReplaceAll(s, "\n", "\n"+pad)
	},
}

// buildFindQDoc は Find 系エンドポイントの q パラメータに付ける @doc 内容（Markdown）を返す。
// typeName は manifest のキー（個別リソース TypeName もしくは "Appliance" / "CommonServiceItem"）。
// manifest にエントリが無い／Filter を持たないリソースでは Filter 節を省略する。
//
// 例示文は @example 側で行うのでここには書かない。ワイヤー形式（`?q=` → `?{JSON}`）や
// Sort / Include / Exclude 非サポートは本 API 全体の @doc（main.tsp）で説明済みなので触れない。
func buildFindQDoc(typeName string) string {
	f, ok := findmanifest.Manifest[typeName]
	if !ok {
		f = findmanifest.GroupManifest[typeName]
	}

	var b strings.Builder
	b.WriteString("Find 検索条件をシリアライズした JSON 文字列を渡す。\n\n")
	b.WriteString("**指定可能なトップレベルフィールド:**\n\n")
	b.WriteString("- `Count` (int): 取得件数の上限\n")
	b.WriteString("- `From` (int): 開始オフセット")

	if f.HasAny() {
		b.WriteString("\n- `Filter` (object): このエンドポイントで指定可能なフィルタキーは以下\n")
		if f.Name {
			b.WriteString("    - `Name` (string): 部分一致。スペース区切りで AND 結合\n")
		}
		if f.Tags {
			b.WriteString("    - `Tags` (string[]): タグ完全一致の AND 結合\n")
		}
		if f.Scope {
			b.WriteString("    - `Scope` (string): `\"shared\"` または `\"user\"` の部分一致\n")
		}
		if f.Class {
			b.WriteString("    - `Class` (string): Appliance のサブクラス（例 `database` / `loadbalancer`）の部分一致\n")
		}
		if f.ProviderClass {
			b.WriteString("    - `Provider.Class` (string): CommonServiceItem のプロバイダ種別の部分一致\n")
		}
		return strings.TrimRight(b.String(), "\n")
	}
	return b.String()
}

// buildFindQExample は q パラメータの @example に渡す JSON 文字列を返す。
// manifest から代表的なフィルタ 1 つを選んで例示する。
//
// 注意: @typespec/openapi3 v1.11 時点では、parameter に付けた @example は
// OpenAPI YAML の `example` フィールドに変換されない（compile はエラー無く通る）。
// それでも TypeSpec ファイル自体が公開ドキュメントとなる運用のため、ソース側に
// 例示を残す価値はある。将来 emitter が対応した時点で自動的に反映される。
func buildFindQExample(typeName string) string {
	f, ok := findmanifest.Manifest[typeName]
	if !ok {
		f = findmanifest.GroupManifest[typeName]
	}
	switch {
	case f.Name:
		return `{"Count":10,"From":0,"Filter":{"Name":"foo"}}`
	case f.Tags:
		return `{"Count":10,"From":0,"Filter":{"Tags":["foo"]}}`
	case f.Scope:
		return `{"Count":10,"From":0,"Filter":{"Scope":"user"}}`
	case f.Class:
		return `{"Count":10,"From":0,"Filter":{"Class":"database"}}`
	case f.ProviderClass:
		return `{"Count":10,"From":0,"Filter":{"Provider.Class":"dns"}}`
	default:
		return `{"Count":10,"From":0}`
	}
}

// opEntry は TypeSpec interface に出力する単一オペレーションの情報を保持する。
type opEntry struct {
	MethodNameLower string
	PathFormat      string
	HttpMethodLower string
	Params          []opParam
	ReturnType      string // TypeSpec の戻り値型名。envelope なしの場合は "void"
	SuccessStatus   int    // POST のときの成功 status code（200/201/202）。POST 以外では未使用
	Summary         string // @summary("...") に出力する日本語文字列。buildSummary() で合成
}

// postSuccessStatus は POST エンドポイントが返す HTTP ステータスコードを解決する。
// 実 API を観測して判明したものを個別に登録する。未登録の POST は 201 Created を想定。
// 実 API が異なるコードを返すようになったら、統合テストで decode が失敗して発覚するので
// そのときにこの表を更新する。
//
// キー = 解決済みの TypeSpec ルート文字列（`@route("...")` の値と同じ）。
func postSuccessStatus(resolvedPath string) int {
	if code, ok := postStatusCodeOverrides[resolvedPath]; ok {
		return code
	}
	return 201
}

// postStatusCodeOverrides は POST ステータスコードの個別上書きマップ。
// 観測済みエンドポイントのみ登録し、未登録は 201 にフォールバックする。
// キーは resolveOpPath 後の TypeSpec ルート（zone / pathSuffix を含まない形）。
var postStatusCodeOverrides = map[string]int{
	// 202 Accepted（非同期受付完了）を返すエンドポイント
	"/internet":  202,
	"/appliance": 202,

	// 200 OK を返す sub-action 系 POST（既存リソースに対する操作で「新規リソース作成」では無い）
	"/internet/{id}/ipv6net": 200,
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
// zone と pathSuffix (api/cloud/1.1) は OpenAPI の servers: 側に移動したので、
// TypeSpec の @route からは落とす。
func resolveOpPath(op *dsl.Operation, api *dsl.Resource) string {
	pf := op.GetPathFormat()
	pf = strings.ReplaceAll(pf, "{{.rootURL}}", "")
	pf = strings.ReplaceAll(pf, "{{.zone}}", "")
	pf = strings.ReplaceAll(pf, "{{.pathSuffix}}", "")
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

// pathParamDocs は @path パラメータ名に対して付与する @doc 文言。
// 未登録のキーは @doc 無しで出力される。
// zone は @server のサーバー変数として扱うため、パスパラメータには出現しない。
var pathParamDocs = map[string]string{
	"id": "対象リソースの ID。数値を 10 進文字列で指定する。",
	"accountID":       "契約（アカウント）の ID。",
	"bridgeID":        "ブリッジ ID。",
	"clientID":        "CA クライアント証明書の ID（`cli_xxxx` 形式）。",
	"destZoneID":      "転送先ゾーンの ID。",
	"destination":     "ping の宛先 IP アドレスまたはホスト名。",
	"index":           "インターフェースのインデックス（0 始まり）。",
	"ipAddress":       "対象の IPv4 アドレス。",
	"MemberCode":      "会員コード。",
	"month":           "対象月（1〜12）。",
	"nicIndex":        "NIC のインデックス（0 始まり）。",
	"packetFilterID":  "パケットフィルタ ID。",
	"serverID":        "サーバ ID。",
	"simID":           "SIM ID。",
	"sourceArchiveID": "コピー元アーカイブ ID。",
	"subnetID":        "サブネット ID。",
	"switchID":        "スイッチ ID。",
	"username":        "コンテナレジストリのユーザ名。",
	"year":            "対象年（西暦 4 桁）。",
}

// newPathParam は @path パラメータを 1 つ組み立てる。pathParamDocs に該当キーがあれば @doc を付与する。
func newPathParam(name string) opParam {
	p := opParam{Decorator: "@path", Name: name, TSType: "string"}
	if doc, ok := pathParamDocs[name]; ok {
		p.Doc = doc
	}
	return p
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
		params = append(params, newPathParam(p))
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

// fatModelExists は同名の fatModelDef が既に存在するかチェックする。
func fatModelExists(defs []fatModelDef, name string) bool {
	for _, d := range defs {
		if d.Name == name {
			return true
		}
	}
	return false
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
// エンベロープを持つ op はすべて統合エンベロープ（buildMergedEnvelopeInfos で生成）を @body として参照する。
func generateIndividualFile(api *dsl.Resource, outputPath string) {
	// opKey ごとにオペレーションをグループ化（元の順序を保持）
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
		path := resolveOpPath(op, api)
		k := opKey{strings.ToLower(op.Method), path}
		if idx, ok := groupIdx[k]; ok {
			groups[idx].ops = append(groups[idx].ops, op)
		} else {
			groupIdx[k] = len(groups)
			groups = append(groups, opGroup{key: k, ops: []*dsl.Operation{op}})
		}
	}

	// opKey -> 統合エンベロープ名（envelopes.go の generateEnvelopes と同じ基準）
	_, envelopeNameByKey := buildMergedEnvelopeInfos(api)

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
			params = append(params, newPathParam(p))
		}

		if envelopeName, ok := envelopeNameByKey[g.key]; ok && envelopeName != "" {
			// Find 系 (GET + XxxFindRequestEnvelope) は @query q?: string に差し替える。
			// サーバーは将来 `?q={json}` を受け付ける予定（未実装）。OpenAPI はその未来形で記述し、
			// ワイヤーレベルでは findQueryRewriteTransport が `q=` を剥がして `?{json}` に変換する。
			if strings.ToLower(g.key.method) == "get" && strings.HasSuffix(envelopeName, "FindRequestEnvelope") {
				params = append(params, opParam{
					Decorator: "@query",
					Name:      "q",
					TSType:    "string",
					Optional:  true,
					Doc:       buildFindQDoc(api.TypeName()),
					Example:   buildFindQExample(api.TypeName()),
				})
			} else {
				// グループのいずれかにリクエストエンベロープがある → 統合エンベロープを @body で使用
				// （単一 op の場合は従来通り、そのエンベロープ名。複数 op の場合は primary のエンベロープ名に
				// 全バリアントの payload が union でマージ済み）
				params = append(params, opParam{
					Decorator: "@body",
					Name:      "body",
					TSType:    envelopeName,
				})
			}
		} else {
			// エンベロープなし: primary op の引数をボディ引数として出力
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

		ops = append(ops, opEntry{
			MethodNameLower: lowerFirst(primary.Name),
			PathFormat:      path,
			HttpMethodLower: g.key.method,
			Params:          params,
			ReturnType:      responseTypeForOp(primary),
			SuccessStatus:   postSuccessStatus(path),
			Summary:         buildSummary(api.TypeName(), lowerFirst(primary.Name)),
		})
	}

	type fileParam struct {
		TypeName   string
		Operations []opEntry
	}
	writeFile(individualOpTmpl, fileParam{TypeName: api.TypeName(), Operations: ops}, outputPath, opTemplateFuncs)
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

		// 共有グループの body は DSL の Arg 名ではなく実 API の JSON 構造に合わせる必要がある。
		// DSL 上は以下の 2 パターン:
		// - MappableArgument: MapConvTag = "<PayloadName>,recursive" → body を `{ <PayloadName>: T }` で包む (Create/Update)
		// - PassthroughModelArgument: MapConvTag = ",squash"       → body は flat (Shutdown の Force, Find の Count/From/Filter 等)
		// この判定は primary op の最初の非 path 引数の MapConvTag から行う。
		shouldWrap := false
		wrapKey := ""
		if len(resources) > 0 {
			if firstOps, ok := resourceOps[0].byKey[k]; ok {
				primary := primaryOpForKey(firstOps)
				pathSet := map[string]bool{}
				for _, p := range extractPathParams(resolvedPath) {
					pathSet[p] = true
				}
				for _, arg := range primary.Arguments {
					if pathSet[arg.PathFormatName()] {
						continue
					}
					// "<dest>,recursive" パターンで wrap する
					parts := strings.SplitN(arg.MapConvTag, ",", 2)
					if len(parts) == 2 && parts[0] != "" && strings.Contains(parts[1], "recursive") {
						shouldWrap = true
						wrapKey = parts[0]
					}
					break // 最初の 1 つだけ判定
				}
			}
		}

		// パラメータリストを構築: @path を先に出力
		var params []opParam
		for _, p := range pathParams {
			params = append(params, newPathParam(p))
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
				// wrap が必要なら envelope モデル名を TSType に、不要なら型そのものを使う
				tsType := firstArg.tsType
				// Find 系共有グループ (GET + FindCondition) は @query q?: string に差し替え。
				// 詳細は generateIndividualFile 側の同種分岐コメント参照。
				if strings.ToLower(k.method) == "get" && tsType == "FindCondition" {
					params = append(params, opParam{
						Decorator: "@query",
						Name:      "q",
						TSType:    "string",
						Optional:  true,
						Doc:       buildFindQDoc(groupName),
						Example:   buildFindQExample(groupName),
					})
					continue
				}
				if shouldWrap && len(firstArgs) == 1 {
					envName := groupName + upperFirst(lowerFirst(repOp.Name)) + "RequestEnvelope"
					if !fatModelExists(fatModels, envName) {
						fatModels = append(fatModels, fatModelDef{
							Name:   envName,
							Fields: []fatField{{Name: wrapKey, TSType: firstArg.tsType}},
						})
					}
					tsType = envName
				}
				params = append(params, opParam{Decorator: "@body", Name: "body", TSType: tsType})
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
				// fat model も wrap が必要なら envelope モデルでラップする
				tsType := fatModelName
				if shouldWrap && len(firstArgs) == 1 {
					envName := groupName + upperFirst(lowerFirst(repOp.Name)) + "RequestEnvelope"
					if !fatModelExists(fatModels, envName) {
						fatModels = append(fatModels, fatModelDef{
							Name:   envName,
							Fields: []fatField{{Name: wrapKey, TSType: fatModelName}},
						})
					}
					tsType = envName
				}
				params = append(params, opParam{Decorator: "@body", Name: "body", TSType: tsType})
			}
		}

		sharedOps = append(sharedOps, opEntry{
			MethodNameLower: lowerFirst(repOp.Name),
			PathFormat:      resolvedPath,
			HttpMethodLower: k.method,
			Params:          params,
			ReturnType:      responseTypeForOp(repOp),
			SuccessStatus:   postSuccessStatus(resolvedPath),
			Summary:         buildSummary(groupName, lowerFirst(repOp.Name)),
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
			if opIsExcluded(rm.resource, op) {
				continue // excludedOps で明示的に除外された op はスキップ
			}
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
				SuccessStatus:   postSuccessStatus(path),
				Summary:         buildSummary(rm.resource.TypeName(), lowerFirst(op.Name)),
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
	}, outputPath, opTemplateFuncs)
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

// extraPathParamSuffix はパスから id 以外の最初のパスパラメータ名を返す。
// op 名の重複解消用サフィックスとして使用する。
// zone は @server の変数扱いで @route には出現しないので考慮不要。
func extraPathParamSuffix(path string) string {
	standard := map[string]bool{"id": true}
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
  @summary("{{ .Summary }}")
  @{{ .HttpMethodLower }}
  @route("{{ .PathFormat }}")
  op {{ .MethodNameLower }}(
    {{ range .Params }}{{ if .Doc }}@doc("""
{{ indent 4 .Doc }}
    """)
    {{ end }}{{ if .Example }}@example("""
    {{ .Example }}
    """)
    {{ end }}{{ if .Decorator }}{{ .Decorator }} {{ end }}{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }},
    {{ end }}
  ): {{ if and (eq .HttpMethodLower "post") (ne .ReturnType "void") }}{@statusCode _: {{ .SuccessStatus }}; ...{{ .ReturnType }}}{{ else if eq .HttpMethodLower "delete" }}{@statusCode _: 200; is_ok: boolean}{{ else if and (eq .ReturnType "void") (ne .HttpMethodLower "get") }}{@statusCode _: 200; is_ok: boolean}{{ else }}{{ .ReturnType }}{{ end }} | ApiError;
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
  @summary("{{ .Summary }}")
  @{{ .HttpMethodLower }}
  @route("{{ .PathFormat }}")
  op {{ .MethodNameLower }}(
    {{ range .Params }}{{ if .Doc }}@doc("""
{{ indent 4 .Doc }}
    """)
    {{ end }}{{ if .Example }}@example("""
    {{ .Example }}
    """)
    {{ end }}{{ if .Decorator }}{{ .Decorator }} {{ end }}{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }},
    {{ end }}
  ): {{ if and (eq .HttpMethodLower "post") (ne .ReturnType "void") }}{@statusCode _: {{ .SuccessStatus }}; ...{{ .ReturnType }}}{{ else if eq .HttpMethodLower "delete" }}{@statusCode _: 200; is_ok: boolean}{{ else if and (eq .ReturnType "void") (ne .HttpMethodLower "get") }}{@statusCode _: 200; is_ok: boolean}{{ else }}{{ .ReturnType }}{{ end }} | ApiError;
{{ end }}
}
{{ range .ResourceInterfaces }}
@tag("{{ .TypeName }}")
interface {{ .TypeName }}Op {
{{ range .Operations }}
  @summary("{{ .Summary }}")
  @{{ .HttpMethodLower }}
  @route("{{ .PathFormat }}")
  op {{ .MethodNameLower }}({{ range .Params }}{{ if .Doc }}@doc("""
{{ indent 4 .Doc }}
    """) {{ end }}{{ if .Example }}@example("""
    {{ .Example }}
    """) {{ end }}{{ if .Decorator }}{{ .Decorator }} {{ end }}{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }}, {{ end }}): {{ if and (eq .HttpMethodLower "post") (ne .ReturnType "void") }}{@statusCode _: {{ .SuccessStatus }}; ...{{ .ReturnType }}}{{ else if eq .HttpMethodLower "delete" }}{@statusCode _: 200; is_ok: boolean}{{ else if and (eq .ReturnType "void") (ne .HttpMethodLower "get") }}{@statusCode _: 200; is_ok: boolean}{{ else }}{{ .ReturnType }}{{ end }} | ApiError;
{{ end }}
}
{{ end }}`
