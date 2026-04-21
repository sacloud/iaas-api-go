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
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sacloud/iaas-api-go/internal/define"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("gen-typespec: ")
}

func main() {
	// 生成前に resources/ 配下の古い .tsp ファイルをすべて削除する。
	// これにより、前回の実行で生成されたが今回は不要なファイルが残るのを防ぐ。
	cleanResourcesDir()
	generateTypes()
	generateModels()
	generateOps()
	generateEnvelopes()
	generateResults()
	generateResourcesTsp()
	generateMainTsp()
	// 旧構造ディレクトリ・ファイルを削除する
	removeObsoleteFiles()
	// fieldmanifest により除外されたフィールド一覧をレポート出力する
	writeExcludedFieldsReport()
}

// removeObsoleteFiles は旧ディレクトリ構造の残存ファイル・ディレクトリを削除する。
func removeObsoleteFiles() {
	// 削除対象の .tsp ファイル
	obsoleteTspFiles := []string{
		"spec/typespec/envelopes.tsp",
		"spec/typespec/models.tsp",
		"spec/typespec/ops.tsp",
		"spec/typespec/results.tsp",
	}
	for _, f := range obsoleteTspFiles {
		p := absPath(f)
		if _, err := os.Stat(p); err == nil {
			if err := os.Remove(p); err != nil {
				log.Printf("warning: failed to remove %s: %v", p, err)
			} else {
				log.Printf("Removed obsolete file: %s", p)
			}
		}
	}

	// 削除対象のディレクトリ（再帰的に削除）
	obsoleteDirs := []string{
		"spec/typespec/envelopes",
		"spec/typespec/models",
		"spec/typespec/ops",
		"spec/typespec/results",
	}
	for _, d := range obsoleteDirs {
		p := absPath(d)
		if _, err := os.Stat(p); err == nil {
			if err := os.RemoveAll(p); err != nil {
				log.Printf("warning: failed to remove dir %s: %v", p, err)
			} else {
				log.Printf("Removed obsolete directory: %s", p)
			}
		}
	}
}

// cleanResourcesDir は spec/typespec/resources/ 配下の全 .tsp ファイルを再帰的に削除する。
func cleanResourcesDir() {
	resourcesDirAbs := absPath(resourcesDir)
	// ディレクトリが存在しない場合はスキップ
	if _, err := os.Stat(resourcesDirAbs); os.IsNotExist(err) {
		return
	}
	entries, err := os.ReadDir(resourcesDirAbs)
	if err != nil {
		log.Fatalf("failed to read resources dir: %v", err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subDir := filepath.Join(resourcesDirAbs, e.Name())
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			log.Fatalf("failed to read sub dir %s: %v", subDir, err)
		}
		for _, se := range subEntries {
			if !se.IsDir() && strings.HasSuffix(se.Name(), ".tsp") {
				if err := os.Remove(filepath.Join(subDir, se.Name())); err != nil {
					log.Fatalf("failed to remove %s: %v", se.Name(), err)
				}
			}
		}
		// サブディレクトリが空になった場合は削除
		remainingEntries, err := os.ReadDir(subDir)
		if err == nil && len(remainingEntries) == 0 {
			if err := os.Remove(subDir); err != nil {
				log.Fatalf("failed to remove dir %s: %v", subDir, err)
			}
		}
	}
}

// generateResourcesTsp は spec/typespec/resources.tsp を生成する。
// TypeSpec が前方参照を解決できるよう、ファイルタイプ別にまとめて出力する:
//  1. 全リソースの models.tsp（モデル定義）
//  2. 全リソースの ops.tsp（オペレーション定義、モデルを参照する）
//  3. 全リソースの envelopes.tsp（エンベロープ定義）
//  4. 全リソースの results.tsp（result定義）
//
// 個別リソース（auto_backup など）は FileSafeName() のディレクトリを、
// 共有グループ（common_service_item, appliance）はグループ名のディレクトリを使う。
func generateResourcesTsp() {
	resourcesDirAbs := absPath(resourcesDir)

	// 個別リソースのディレクトリ名リスト（出現順）
	// 共有グループに属するリソースも FileSafeName() でディレクトリを持つ（models/envelopes/results）
	var individualDirs []string
	// 共有グループのディレクトリ名リスト（ops のみ）
	var sharedGroupDirs []string
	seenIndividual := map[string]bool{}
	seenGroup := map[string]bool{}

	for _, api := range define.APIs {
		pn := api.GetPathName()
		if isSharedGroup(pn) {
			// 共有グループ ops ディレクトリを登録
			groupName := pathNameToGroupName[pn]
			groupDir := toSnake(groupName)
			if !seenGroup[groupDir] {
				seenGroup[groupDir] = true
				sharedGroupDirs = append(sharedGroupDirs, groupDir)
			}
			// 個別リソースのディレクトリも登録（models/envelopes/results 用）
			if !seenIndividual[api.FileSafeName()] {
				seenIndividual[api.FileSafeName()] = true
				individualDirs = append(individualDirs, api.FileSafeName())
			}
		} else {
			// 単一リソース
			if !seenIndividual[api.FileSafeName()] {
				seenIndividual[api.FileSafeName()] = true
				individualDirs = append(individualDirs, api.FileSafeName())
			}
		}
	}

	// 各ディレクトリのファイルセットを収集するヘルパー
	getFileSet := func(dirName string) map[string]bool {
		subDir := filepath.Join(resourcesDirAbs, dirName)
		entries, err := os.ReadDir(subDir)
		if err != nil {
			return nil
		}
		fileSet := map[string]bool{}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".tsp") {
				fileSet[e.Name()] = true
			}
		}
		return fileSet
	}

	var lines []string
	lines = append(lines, "// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-typespec'; DO NOT EDIT", "")

	// 1. 全リソースの models.tsp を先に出力（共有グループが参照するため）
	for _, dirName := range individualDirs {
		fileSet := getFileSet(dirName)
		if fileSet["models.tsp"] {
			lines = append(lines, fmt.Sprintf("import \"./resources/%s/models.tsp\";", dirName))
		}
	}

	// 2. 全 ops.tsp を出力
	//    個別リソースの ops → 共有グループの ops の順
	for _, dirName := range individualDirs {
		pn := apiPathNameForDir(dirName)
		if isSharedGroup(pn) {
			// 共有グループに属するリソースの ops は共有グループ側で出力するためスキップ
			continue
		}
		fileSet := getFileSet(dirName)
		if fileSet["ops.tsp"] {
			lines = append(lines, fmt.Sprintf("import \"./resources/%s/ops.tsp\";", dirName))
		}
	}
	for _, groupDir := range sharedGroupDirs {
		fileSet := getFileSet(groupDir)
		if fileSet["ops.tsp"] {
			lines = append(lines, fmt.Sprintf("import \"./resources/%s/ops.tsp\";", groupDir))
		}
	}

	// 3. 全 envelopes.tsp を出力
	for _, dirName := range individualDirs {
		fileSet := getFileSet(dirName)
		if fileSet["envelopes.tsp"] {
			lines = append(lines, fmt.Sprintf("import \"./resources/%s/envelopes.tsp\";", dirName))
		}
	}

	// 4. 全 results.tsp を出力
	for _, dirName := range individualDirs {
		fileSet := getFileSet(dirName)
		if fileSet["results.tsp"] {
			lines = append(lines, fmt.Sprintf("import \"./resources/%s/results.tsp\";", dirName))
		}
	}

	lines = append(lines, "")

	content := strings.Join(lines, "\n")
	writeFile(content, nil, "spec/typespec/resources.tsp", nil)
	log.Printf("generated: spec/typespec/resources.tsp\n")
}

// apiPathNameForDir は FileSafeName（ディレクトリ名）から pathName を返す。
// 共有グループ判定に使用する。
func apiPathNameForDir(dirName string) string {
	for _, api := range define.APIs {
		if api.FileSafeName() == dirName {
			return api.GetPathName()
		}
	}
	return dirName
}

// isSharedGroup は pathName が共有エンドポイントグループかどうかを返す。
func isSharedGroup(pathName string) bool {
	_, ok := pathNameToGroupName[pathName]
	return ok
}

// generateMainTsp は spec/typespec/main.tsp を生成する。
func generateMainTsp() {
	// Markdown のインラインコード用バッククォートは Go の raw string を終端してしまうので、
	// プレースホルダ @BT@ に置き換えて組み立てる。
	content := `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-typespec'; DO NOT EDIT
// Copyright 2022-2025 The sacloud/iaas-api-go Authors

import "@typespec/http";
import "@typespec/openapi3";

import "./resources.tsp";
import "./types.tsp";

@service(#{
  title: "Sakura Cloud IaaS API",
})
@doc("""
# さくらのクラウド IaaS API

## 必須リクエストヘッダ

この API を呼び出す際は、すべてのリクエストに以下のヘッダーを付与してください。

- @BT@X-Sakura-Bigint-As-Int: 1@BT@

このヘッダーにより、bigint 型の値が JSON 文字列ではなく整数として返却されます。

## Find 系エンドポイントのクエリ形式

本定義書では Find 系エンドポイントの検索条件を @BT@?q={JSON}@BT@ 形式のクエリパラメータとして記述している。
しかし **現行サーバー実装は @BT@q=@BT@ キーを受け付けず、
@BT@?{JSON}@BT@（クエリ文字列の先頭に @BT@?@BT@ を置き、続けて生の JSON オブジェクトを書く形。@BT@q=@BT@ というキー名は無い）形式を要求する**。

### 表現の違い

本定義書（および生成された OpenAPI）上の表現:

@BT@@BT@@BT@
GET /is1a/api/cloud/1.1/bridge?q=%7B%22Count%22%3A3%7D
@BT@@BT@@BT@

実サーバーが受理する表現:

@BT@@BT@@BT@
GET /is1a/api/cloud/1.1/bridge?{"Count":3}
@BT@@BT@@BT@

### クライアント実装が取るべき対応

以下のいずれかで辻褄合わせを行う必要がある。

- HTTP 送信直前に URL の @BT@q=<urlencoded-json>@BT@ を @BT@<rawjson>@BT@ に書き換える
  （参考実装: @BT@v2/client/find_transport.go@BT@ の @BT@findQueryRewriteTransport@BT@）
- あるいは独自に検索条件を組み立てて @BT@?{JSON}@BT@ を直接生成する

将来サーバー側が本定義どおり @BT@?q={JSON}@BT@ を受理するようになれば、書き換え層は不要になる。
""")
@useAuth(BasicAuth)
namespace Sacloud.IaaS;

using TypeSpec.Http;

// カスタム型定義
scalar ID extends string;
scalar Tags;
scalar StringNumber extends string;
scalar StringFlag extends string;

// API エラーレスポンス
@error
model ApiError {
  is_fatal?: boolean;
  serial?: string;
  status?: string;
  error_code?: string;
  error_msg?: string;
}

// リソース参照（ID のみ保持）
// create/update リクエストで他リソースを参照する際に使用する
model ResourceRef {
  ID: int64;
}
`

	content = strings.ReplaceAll(content, "@BT@", "`")
	writeFile(content, nil, "spec/typespec/main.tsp", nil)
	log.Printf("generated: spec/typespec/main.tsp\n")
}
