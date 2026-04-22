// Copyright 2022-2026 The sacloud/iaas-api-go Authors
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

// gen-v2-op は v2/<bucket>_op_gen.go を生成する。
//
// 仕組み:
//  1. v2/client/ を AST で走査
//  2. Invoker インターフェースを見つけ、各 <Bucket>Op<Action> メソッドを列挙
//  3. 対応する <Bucket>Op<Action>Params 構造体のフィールドを抽出
//  4. バケットごとに *_op_gen.go を emit（インターフェース + 実装 + コンストラクタ）
//
// ラッパー層の使用感は simple-notification-api-go に揃える:
//   - メソッド名は Action のみ（<Bucket>Op プレフィックスを剥がす）
//   - Params のフィールドはメソッド引数にフラット化
//   - `Q OptString` + 対応する `<Bucket>FindRequest` が存在する場合は、
//     typed request (`*client.<Bucket>FindRequest`) を受け取り内部で ToOptString
//   - ogen が返す `*<Op>OK` タイプ（IsOk のみ）は破棄し、戻り値は error のみ
//   - それ以外のレスポンスは ogen 生成型をそのまま露出
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("gen-v2-op: ")
}

// operation は Invoker インターフェースの 1 メソッドを表す。
type operation struct {
	Name        string // e.g. "NoteOpCreate"
	Bucket      string // e.g. "Note"
	Action      string // e.g. "Create"
	HasRequest  bool
	RequestType string // e.g. "*NoteCreateRequestEnvelope"
	ParamsType  string // e.g. "NoteOpCreateParams"
	ReturnType  string // e.g. "*NoteCreateResponseEnvelope"
}

type paramsField struct {
	Name string // e.g. "Zone", "ID", "Q"
	Type string // e.g. "string", "OptString"
}

func main() {
	clientDir := absPath("v2/client")
	outDir := absPath("v2")

	ops, paramsFields, findRequests := parseClientPackage(clientDir)
	if len(ops) == 0 {
		log.Fatalf("no operations found in %s", clientDir)
	}

	// Group by bucket
	byBucket := map[string][]operation{}
	for _, op := range ops {
		byBucket[op.Bucket] = append(byBucket[op.Bucket], op)
	}

	var buckets []string
	for b := range byBucket {
		buckets = append(buckets, b)
	}
	sort.Strings(buckets)

	total := 0
	generatedFiles := map[string]bool{}
	for _, b := range buckets {
		bucketOps := byBucket[b]
		sort.Slice(bucketOps, func(i, j int) bool { return bucketOps[i].Action < bucketOps[j].Action })

		src, err := generateBucketFile(b, bucketOps, paramsFields, findRequests)
		if err != nil {
			log.Fatalf("generate %s: %v", b, err)
		}

		outPath := filepath.Join(outDir, snakeCase(b)+"_gen.go")
		if err := os.WriteFile(outPath, src, 0o644); err != nil {
			log.Fatalf("write %s: %v", outPath, err)
		}
		generatedFiles[filepath.Base(outPath)] = true
		total += len(bucketOps)
		log.Printf("generated: %s (%d ops)", filepath.Base(outPath), len(bucketOps))
	}

	// 既存の <bucket>_gen.go のうち今回生成対象外になったものを削除する。
	// excludedOps による op 除去で bucket が空になった場合、stale なラッパーファイルが残って
	// 未定義シンボル参照のビルド失敗を起こすため。
	entries, err := os.ReadDir(outDir)
	if err == nil {
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, "_gen.go") {
				continue
			}
			if generatedFiles[name] {
				continue
			}
			p := filepath.Join(outDir, name)
			if err := os.Remove(p); err != nil {
				log.Printf("warning: failed to remove stale %s: %v", name, err)
			} else {
				log.Printf("removed stale: %s", name)
			}
		}
	}

	log.Printf("done: %d buckets, %d operations", len(buckets), total)
}

// parseClientPackage は v2/client/ を AST で走査し、
// Invoker メソッド / Params 構造体フィールド / <Bucket>FindRequest 型を抽出する。
func parseClientPackage(dir string) ([]operation, map[string][]paramsField, map[string]bool) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		// テストファイル (_test.go) は除外
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		log.Fatalf("parse dir %s: %v", dir, err)
	}
	clientAst, ok := pkgs["client"]
	if !ok {
		log.Fatalf("package client not found in %s", dir)
	}

	var ops []operation
	paramsFields := map[string][]paramsField{}
	findRequests := map[string]bool{}

	for _, f := range clientAst.Files {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if iface, ok := ts.Type.(*ast.InterfaceType); ok && ts.Name.Name == "Invoker" {
					ops = extractInvokerOps(iface, fset)
				}
				if strct, ok := ts.Type.(*ast.StructType); ok {
					if strings.HasSuffix(ts.Name.Name, "Params") {
						paramsFields[ts.Name.Name] = extractStructFields(strct, fset)
					}
					if strings.HasSuffix(ts.Name.Name, "FindRequest") {
						findRequests[ts.Name.Name] = true
					}
				}
			}
		}
	}
	return ops, paramsFields, findRequests
}

// extractInvokerOps は Invoker の各メソッド宣言から operation を作る。
func extractInvokerOps(iface *ast.InterfaceType, fset *token.FileSet) []operation {
	var ops []operation
	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue // embedded
		}
		name := method.Names[0].Name
		fn, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		bucket, action := splitOpName(name)
		if bucket == "" || action == "" {
			continue // not <Bucket>Op<Action>
		}
		op := operation{Name: name, Bucket: bucket, Action: action}

		// Params:
		//   (ctx, params ParamsType)
		//   (ctx, request *ReqType, params ParamsType)
		if fn.Params == nil {
			continue
		}
		for i, p := range fn.Params.List {
			if i == 0 {
				continue // ctx
			}
			typeStr := typeString(p.Type, fset)
			if strings.HasSuffix(typeStr, "Params") {
				op.ParamsType = typeStr
			} else {
				op.HasRequest = true
				op.RequestType = typeStr
			}
		}
		// ParamsType が無い op は zone / id / query を一切持たない（body のみ）形。
		// ogen がそうした op では Params 構造体自体を生成しないため、ラッパー側でも
		// params を組み立てずに request だけを渡す。

		// Return: (*T, error)
		if fn.Results == nil || len(fn.Results.List) == 0 {
			continue
		}
		op.ReturnType = typeString(fn.Results.List[0].Type, fset)

		ops = append(ops, op)
	}
	return ops
}

// splitOpName は "<Bucket>Op<Action>" を分割する。
// バケット側が空 / アクション側が空のケースは ("", "") を返す。
func splitOpName(name string) (bucket, action string) {
	// 最後に出現する "Op[A-Z]" を境界とする
	for i := 0; i+2 < len(name); i++ {
		if name[i] == 'O' && name[i+1] == 'p' && isUpperASCII(name[i+2]) {
			// bucket 側はこれ以降 "Op[A-Z]" を持たない最も手前で切る
			return name[:i], name[i+2:]
		}
	}
	return "", ""
}

func isUpperASCII(b byte) bool { return b >= 'A' && b <= 'Z' }

func extractStructFields(strct *ast.StructType, fset *token.FileSet) []paramsField {
	var out []paramsField
	for _, f := range strct.Fields.List {
		for _, n := range f.Names {
			out = append(out, paramsField{Name: n.Name, Type: typeString(f.Type, fset)})
		}
	}
	return out
}

func typeString(e ast.Expr, fset *token.FileSet) string {
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, e)
	return buf.String()
}

// ---- code emission ----

type bucketFile struct {
	Bucket string
	Ops    []opView
}

type opView struct {
	Bucket       string
	StructName   string    // 非公開版 (例: "noteOp")
	MethodName   string    // ラッパーのメソッド名（Find→List 等のリネーム後）
	Action       string    // エラーメッセージ用（MethodName と同じ）
	OpName       string    // ogen 側の元メソッド名 (例: "NoteOpFind")
	MethodArgs   []argView // ctx 以降の引数。ctx 自体は含めない
	ParamsType   string    // "client.<ParamsType>"。NoParams の場合は空
	NoParams     bool      // ogen 側が Params 構造体を生成しない op（body のみ or 引数なし）
	ParamAssigns []assignView
	HasOptAssign bool       // Q+FindRequest の nil check が必要
	OptAssign    assignView // Q の代入（if req != nil { ... } でガード）
	HasRequest   bool
	RequestName  string // "request"
	RequestType  string // "*client.<ReqType>"
	ReturnsError bool   // true: error のみ返す
	ReturnType   string // "*client.<RespType>"
}

type argView struct {
	Name string
	Type string
}

type assignView struct {
	Field string // "Zone"
	Value string // "zone", "req.ToOptString()" etc.
}

func generateBucketFile(bucket string, ops []operation, paramsFields map[string][]paramsField, findRequests map[string]bool) ([]byte, error) {
	bf := bucketFile{Bucket: bucket}
	for _, op := range ops {
		v, err := buildOpView(op, paramsFields, findRequests)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op.Name, err)
		}
		bf.Ops = append(bf.Ops, v)
	}

	var buf bytes.Buffer
	if err := fileTmpl.Execute(&buf, bf); err != nil {
		return nil, err
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		// デバッグ用に整形前の出力を dump
		fmt.Fprintln(os.Stderr, buf.String())
		return nil, fmt.Errorf("gofmt: %w", err)
	}
	return src, nil
}

// actionRenames は ogen のオペレーション名（<Bucket>Op<Action>）から得られる
// Action を、ラッパー上でのメソッド名にマップする。sacloud-sdk-go/api/AGENTS.md の
// 命名規約（List / Read / Create / Update / Delete）に寄せるための最小リネーム。
var actionRenames = map[string]string{
	"Find": "List",
}

func buildOpView(op operation, paramsFields map[string][]paramsField, findRequests map[string]bool) (opView, error) {
	methodName := op.Action
	if renamed, ok := actionRenames[op.Action]; ok {
		methodName = renamed
	}
	v := opView{
		Bucket:     op.Bucket,
		StructName: lowerFirst(op.Bucket) + "Op",
		MethodName: methodName,
		Action:     methodName,
		OpName:     op.Name,
	}
	if op.ParamsType == "" {
		v.NoParams = true
	} else {
		v.ParamsType = "client." + op.ParamsType
	}

	fields := paramsFields[op.ParamsType]
	findReqType := op.Bucket + "FindRequest"
	hasTypedFind := findRequests[findReqType]

	for _, fld := range fields {
		if fld.Name == "Q" && fld.Type == "OptString" && hasTypedFind {
			v.MethodArgs = append(v.MethodArgs, argView{Name: "req", Type: "*client." + findReqType})
			v.HasOptAssign = true
			v.OptAssign = assignView{Field: "Q", Value: "req.ToOptString()"}
			continue
		}
		// 通常フィールド: そのままメソッド引数に
		argName := lowerFirst(fld.Name)
		argType := qualifyType(fld.Type)
		v.MethodArgs = append(v.MethodArgs, argView{Name: argName, Type: argType})
		v.ParamAssigns = append(v.ParamAssigns, assignView{Field: fld.Name, Value: argName})
	}

	if op.HasRequest {
		v.HasRequest = true
		v.RequestName = "request"
		v.RequestType = qualifyType(op.RequestType)
		v.MethodArgs = append(v.MethodArgs, argView{Name: v.RequestName, Type: v.RequestType})
	}

	// Return 型の決定: *<...>OK なら error のみ
	ret := op.ReturnType
	if strings.HasPrefix(ret, "*") && strings.HasSuffix(ret, "OK") {
		v.ReturnsError = true
	} else {
		v.ReturnsError = false
		v.ReturnType = qualifyType(ret)
	}

	return v, nil
}

// qualifyType は ogen パッケージ内部の型参照に client. プレフィックスを付ける。
// 例: "*NoteCreateRequestEnvelope" → "*client.NoteCreateRequestEnvelope"
//     "OptString" → "client.OptString"
//     "string" / "int" / 組み込み型はそのまま
func qualifyType(t string) string {
	// ポインタ
	if strings.HasPrefix(t, "*") {
		return "*" + qualifyType(t[1:])
	}
	// スライス
	if strings.HasPrefix(t, "[]") {
		return "[]" + qualifyType(t[2:])
	}
	// 組み込み型 / 既にパッケージ修飾済み
	if isBuiltinGoType(t) || strings.Contains(t, ".") {
		return t
	}
	// ogen 生成型（UpperCamel）はすべて client パッケージ
	if t != "" && isUpperASCII(t[0]) {
		return "client." + t
	}
	return t
}

func isBuiltinGoType(t string) bool {
	switch t {
	case "bool", "byte", "complex64", "complex128", "error", "float32", "float64",
		"int", "int8", "int16", "int32", "int64", "rune", "string",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "any":
		return true
	}
	return false
}

// lowerFirst は "ID"→"id", "Zone"→"zone", "URL"→"url" のように先頭を小文字化。
// 全て大文字の場合は全て小文字、そうでなければ先頭 1 文字のみ。
func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	if strings.ToUpper(s) == s {
		return strings.ToLower(s)
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// snakeCase は UpperCamel を snake_case に変換する。
// 連続する大文字はアクロニムとして単一の単語扱い。
//   "Note" → "note"
//   "AuthStatus" → "auth_status"
//   "SSHKey" → "ssh_key"
//   "CDROM" → "cdrom"
//   "IPv6Addr" → "ipv6_addr"（"IPv" 系の特殊ケース）
//   "VPCRouter" → "vpc_router"
//
// "IPv6"/"IPv4" はアクロニムと小文字 1 文字 ("v") が混在する変則で、素の規則では
// "i_pv6_..." のように誤分割される。ここでは事前正規化で吸収する。
func snakeCase(s string) string {
	s = normalizeAcronyms(s)
	runes := []rune(s)
	var out strings.Builder
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			var next rune
			if i+1 < len(runes) {
				next = runes[i+1]
			}
			// 直前が小文字 / 直後が小文字の場合にアンダースコア
			if unicode.IsLower(prev) || (next != 0 && unicode.IsLower(next)) {
				out.WriteByte('_')
			}
		}
		out.WriteRune(unicode.ToLower(r))
	}
	return out.String()
}

// normalizeAcronyms は snake_case アルゴリズムが苦手な語を事前に変形する。
// 現時点では "IPv6" / "IPv4" だけ（このコードベースに出現するもの）。
func normalizeAcronyms(s string) string {
	s = strings.ReplaceAll(s, "IPv6", "Ipv6")
	s = strings.ReplaceAll(s, "IPv4", "Ipv4")
	return s
}

// ---- file template ----

var fileTmpl = template.Must(template.New("file").Funcs(template.FuncMap{
	"lowerFirst": lowerFirst,
}).Parse(fileTemplate))

const fileTemplate = `// Copyright 2022-2026 The sacloud/iaas-api-go Authors
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

// Code generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-v2-op'; DO NOT EDIT.

package iaas

import (
	"context"

	"github.com/sacloud/iaas-api-go/v2/client"
)

{{$structName := printf "%sOp" (lowerFirst .Bucket) -}}
// {{.Bucket}}API は {{.Bucket}} リソースに対する操作インターフェース。
type {{.Bucket}}API interface {
{{- range .Ops}}
	{{.MethodName}}(ctx context.Context{{range .MethodArgs}}, {{.Name}} {{.Type}}{{end}}) ({{if not .ReturnsError}}{{.ReturnType}}, {{end}}error)
{{- end}}
}

var _ {{.Bucket}}API = (*{{$structName}})(nil)

type {{$structName}} struct {
	client *client.Client
}

// New{{.Bucket}}Op は {{.Bucket}}API 実装を返す。
func New{{.Bucket}}Op(c *client.Client) {{.Bucket}}API {
	return &{{$structName}}{client: c}
}
{{range .Ops}}
func (op *{{$structName}}) {{.MethodName}}(ctx context.Context{{range .MethodArgs}}, {{.Name}} {{.Type}}{{end}}) ({{if not .ReturnsError}}{{.ReturnType}}, {{end}}error) {
	{{- if not .NoParams}}
	params := {{.ParamsType}}{ {{- range $i, $a := .ParamAssigns}}{{if $i}}, {{end}}{{$a.Field}}: {{$a.Value}}{{end -}} }
	{{- end}}
	{{- if .HasOptAssign}}
	if req != nil {
		params.{{.OptAssign.Field}} = {{.OptAssign.Value}}
	}
	{{- end}}
	{{if .ReturnsError}}_, err{{else}}resp, err{{end}} := op.client.{{.OpName}}(ctx{{if .HasRequest}}, {{.RequestName}}{{end}}{{if not .NoParams}}, params{{end}})
	if err != nil {
		return {{if not .ReturnsError}}nil, {{end}}wrapOpErr("{{.Bucket}}.{{.MethodName}}", err)
	}
	return {{if not .ReturnsError}}resp, {{end}}nil
}
{{end}}`

// absPath は repo ルートからの相対パスを絶対パスに変換する。
// gen-find-request と同じ前提で、実行時のカレントは通常 iaas-api-go ルート。
func absPath(rel string) string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getwd: %v", err)
	}
	if strings.Contains(wd, "/internal/tools/") {
		idx := strings.Index(wd, "/internal/tools/")
		wd = wd[:idx]
	}
	return filepath.Join(wd, rel)
}
