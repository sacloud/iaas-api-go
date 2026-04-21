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
	"bytes"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"text/template"
	"unicode"
)

// repoRoot は util.go の位置から3階層上のリポジトリルートの絶対パス。
var repoRoot string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("failed to get current file path")
	}
	// internal/tools/gen-typespec/util.go → ../../.. でリポジトリルートを解決
	repoRoot = filepath.Join(filepath.Dir(filename), "../../..")
}

// absPath は repoRoot からの相対パスを絶対パスに変換する。
func absPath(relPath string) string {
	if filepath.IsAbs(relPath) {
		return relPath
	}
	return filepath.Join(repoRoot, relPath)
}

// ensureDir は repoRoot からの相対パスのディレクトリを作成する。
func ensureDir(relPath string) {
	dir := absPath(relPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create directory %s: %v", dir, err)
	}
}

// writeFile はテンプレートを param でレンダリングして outputPath（repoRoot 相対）に書き出す。
func writeFile(tmplStr string, param interface{}, outputPath string, funcMap template.FuncMap) {
	outputPath = absPath(outputPath)
	log.Printf("Writing to: %s", outputPath)

	tmpl := template.New("t")
	if funcMap != nil {
		tmpl = tmpl.Funcs(funcMap)
	}
	template.Must(tmpl.Parse(tmplStr))

	buf := bytes.NewBufferString("")
	if err := tmpl.Execute(buf, param); err != nil {
		log.Fatalf("writing output: %s", err)
	}

	// 出力先ディレクトリを作成
	if _, err := os.Stat(filepath.Dir(outputPath)); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0750); err != nil {
			log.Fatal(err)
		}
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

// lowerFirst は PascalCase 文字列を camelCase に変換する。
// 先頭に連続する大文字がある場合（例: IMEILock）は頭字語全体を小文字にする（imeiLock）。
func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)

	// 先頭の連続する大文字の数を数える
	count := 0
	for count < len(runes) && unicode.IsUpper(runes[count]) {
		count++
	}
	if count == 0 {
		return s
	}

	// 複数の大文字が続いた後に小文字が来る場合、最後の大文字は次の単語の先頭なので残す
	// 例: IMEILock → I,M,E,I を小文字 + Lock → imeiLock
	toLower := count
	if count > 1 && count < len(runes) {
		toLower = count - 1
	}

	result := make([]rune, len(runes))
	copy(result, runes)
	for i := 0; i < toLower; i++ {
		result[i] = unicode.ToLower(result[i])
	}
	return string(result)
}

// upperFirst は文字列の先頭文字を大文字にする。
func upperFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// toSnake は PascalCase 文字列を snake_case に変換する。
func toSnake(s string) string {
	var out []rune
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) && i > 0 {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(r))
	}
	return string(out)
}
