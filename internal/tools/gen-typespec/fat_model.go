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
	"strings"

	"github.com/sacloud/iaas-api-go/internal/dsl"
)

// jsonNode は mapconv タグから構築した JSON 構造のツリーノード。
// 中間ノード: children に子ノードを持つ（TypeSpec の中間モデルに対応）
// 葉ノード: children は空で leafType を持つ（TypeSpec のスカラー/モデル型に対応）
type jsonNode struct {
	children    map[string]*jsonNode
	childOrder  []string // 安定した出力順序
	leafType    string   // 葉ノードの TypeSpec 型（e.g., "int64", "DatabaseSettingCommon"）
	isArray     bool     // このノードが配列（[]Seg 記法またはフィールド自体が配列）
	count       int      // このノードが何バリアントに存在するか
	hasConflict bool     // 型競合あり → unknown を使う
}

func newJsonNode() *jsonNode {
	return &jsonNode{children: map[string]*jsonNode{}}
}

// parseMapconvPath は mapconv タグ文字列をパスセグメントと修飾子に分解する。
// 空パス（mapconv なし or ",omitempty"）の場合は nil を返す。
// path 内の各セグメントは "[]" プレフィックスを含む場合がある（e.g., "[]Servers"）。
func parseMapconvPath(mapconv string) (segs []string, recursive bool, omitempty bool) {
	if mapconv == "" {
		return nil, false, false
	}
	parts := strings.Split(mapconv, ",")
	pathStr := parts[0]
	for _, mod := range parts[1:] {
		switch strings.TrimSpace(mod) {
		case "recursive":
			recursive = true
		case "omitempty":
			omitempty = true
		}
	}
	// 代替パス "A/B" の最初を使用（書き込み時のパス）
	if idx := strings.Index(pathStr, "/"); idx >= 0 {
		pathStr = pathStr[:idx]
	}
	if pathStr == "" {
		return nil, recursive, omitempty
	}
	return strings.Split(pathStr, "."), recursive, omitempty
}

// setLeaf は指定ノードに葉型を設定する。
// 中間ノード（children あり）に葉を設定しようとした場合は競合扱い。
func setLeaf(node *jsonNode, tsType string, isArray bool) {
	if len(node.children) > 0 {
		// 既に中間ノードとして使われている → 競合
		if !node.hasConflict {
			node.hasConflict = true
		}
		return
	}
	if node.leafType != "" && node.leafType != tsType {
		node.hasConflict = true
		node.leafType = ""
		return
	}
	if !node.hasConflict {
		node.leafType = tsType
		if isArray {
			node.isArray = true
		}
	}
}

// getOrCreate は指定キーの子ノードを取得または生成して返す。
func getOrCreate(node *jsonNode, key string, isArray bool) *jsonNode {
	// 既に葉として使われているなら競合
	if node.leafType != "" && !node.hasConflict {
		node.hasConflict = true
		node.leafType = ""
	}

	child, exists := node.children[key]
	if !exists {
		child = newJsonNode()
		node.children[key] = child
		node.childOrder = append(node.childOrder, key)
	}
	if isArray {
		child.isArray = true
	}
	return child
}

// addFieldToTree はフィールドを mapconv パスに沿ってツリーに追加する。
func addFieldToTree(root *jsonNode, segs []string, tsType string, isLeafArray bool) {
	if len(segs) == 0 {
		return
	}

	node := root
	for i, seg := range segs {
		segIsArray := strings.HasPrefix(seg, "[]")
		if segIsArray {
			seg = seg[2:]
		}

		if i == len(segs)-1 {
			// 最後のセグメント → 葉ノード
			child := getOrCreate(node, seg, segIsArray || isLeafArray)
			child.count++
			setLeaf(child, tsType, false) // isArray は既に node.isArray に設定済み
		} else {
			// 中間セグメント → ノードを辿る
			child := getOrCreate(node, seg, segIsArray)
			child.count++
			node = child
		}
	}
}

// buildFatModelTree は複数バリアントのモデルから統合ツリーを構築する。
// argVariants: リソース → TypeSpec 型名のマッピング
// resources: リソースの順序（安定した出力のため）
func buildFatModelTree(argVariants map[*dsl.Resource]string, resources []*dsl.Resource) *jsonNode {
	root := newJsonNode()

	for _, res := range resources {
		tsTypeName, ok := argVariants[res]
		if !ok || tsTypeName == "" {
			continue
		}
		m, ok := allModelsByName[tsTypeName]
		if !ok {
			continue
		}

		for _, f := range m.Fields {
			mapconv := ""
			if f.Tags != nil {
				mapconv = f.Tags.MapConv
			}
			segs, _, _ := parseMapconvPath(mapconv)

			// フィールドの TypeSpec 型を取得
			fTSType := modelFieldTypeToTS(f.TypeName())

			// パス内に [] セグメントがある場合、フィールドの配列性はパスで表現される
			hasArrayInPath := false
			for _, s := range segs {
				if strings.HasPrefix(s, "[]") {
					hasArrayInPath = true
					break
				}
			}

			// 配列型を要素型に分解
			isLeafArray := false
			if strings.HasSuffix(fTSType, "[]") {
				fTSType = fTSType[:len(fTSType)-2]
				if !hasArrayInPath {
					// パスに [] がなければフィールド自体が配列
					isLeafArray = true
				}
				// パスに [] がある場合は [] セグメントで配列を表現するため isLeafArray = false のまま
			}

			// mapconv が空の場合はフィールド名をパスとして使う
			if len(segs) == 0 {
				segs = []string{f.Name}
			}

			addFieldToTree(root, segs, fTSType, isLeafArray)
		}
	}

	return root
}

// generateFatModelsFromTree はツリーから TypeSpec モデル群を生成する。
// fatModelAlwaysOptionalTop は共有グループ fat model のトップレベルで常に optional 扱いにする
// フィールドのセット。v1 naked 型ではポインタで omitempty されているため、全 variant に現れても
// 実 API は省略可能なもの。
// Fat model のデフォルト（全 variant にあれば required）だと、例えば Icon: {ID: 0} を常に送る
// 羽目になり、Sakura Cloud 側が「不正な Icon.ID=0」と解釈して 503 を返すケースがある。
var fatModelAlwaysOptionalTop = map[string]bool{
	"Icon":         true,
	"Plan":         true,
	"Disk":         true,
	"Settings":     true,
	"SettingsHash": true,
}

// 戻り値: メインモデルを先頭に、中間モデルを後続に含むリスト
func generateFatModelsFromTree(name string, node *jsonNode, totalVariants int, addClass bool) []fatModelDef {
	return generateFatModelsFromTreeWithDepth(name, node, totalVariants, addClass, 0)
}

// generateFatModelsFromTreeWithDepth は depth=0 の top-level フィールドのみ
// `fatModelAlwaysOptionalTop` を適用して optional にする。
func generateFatModelsFromTreeWithDepth(name string, node *jsonNode, totalVariants int, addClass bool, depth int) []fatModelDef {
	mainModel := fatModelDef{
		Name:     name,
		AddClass: addClass,
	}
	var extraModels []fatModelDef

	for _, key := range node.childOrder {
		child := node.children[key]
		optional := child.count < totalVariants
		if depth == 0 && fatModelAlwaysOptionalTop[key] {
			optional = true
		}

		if child.hasConflict {
			// 型競合 → unknown
			tsType := "unknown"
			if child.isArray {
				tsType = "unknown[]"
			}
			mainModel.Fields = append(mainModel.Fields, fatField{
				Name: key, TSType: tsType, Optional: optional,
			})
			continue
		}

		if len(child.children) == 0 {
			// 葉ノード
			tsType := child.leafType
			if tsType == "" {
				tsType = "unknown"
			}
			if child.isArray {
				tsType += "[]"
			}
			mainModel.Fields = append(mainModel.Fields, fatField{
				Name: key, TSType: tsType, Optional: optional,
			})
		} else {
			// 中間ノード → 再帰的に処理して中間モデルを生成
			childModelName := name + upperFirst(key)
			childModels := generateFatModelsFromTreeWithDepth(childModelName, child, child.count, false, depth+1)
			extraModels = append(extraModels, childModels...)

			tsType := childModelName
			if child.isArray {
				tsType += "[]"
			}
			mainModel.Fields = append(mainModel.Fields, fatField{
				Name: key, TSType: tsType, Optional: optional,
			})
		}
	}

	return append([]fatModelDef{mainModel}, extraModels...)
}

// buildFatModelDefs は共有グループの fat model 定義群を生成して返す。
// addClass が true の場合、メインモデルに Class: string フィールドを追加する（create 系）。
func buildFatModelDefs(name string, argVariants map[*dsl.Resource]string, resources []*dsl.Resource, totalVariants int, addClass bool) []fatModelDef {
	tree := buildFatModelTree(argVariants, resources)
	return generateFatModelsFromTree(name, tree, totalVariants, addClass)
}
