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

package examples

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// tspIdentRe は TypeSpec のプレーン識別子として安全に使えるキー名のパターン。
// マッチしない場合はバッククォート識別子（`my.key`）としてクォートする。
var tspIdentRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// encodeKey は TypeSpec object literal のプロパティ名を整形する。
//   - 識別子として有効な文字列はそのまま（例: ID / Name / is_ok）
//   - それ以外はバッククォートで囲む（例: ` + "`" + `Provider.Class` + "`" + `）
func encodeKey(k string) string {
	if tspIdentRe.MatchString(k) {
		return k
	}
	// バッククォート中のバッククォートは TypeSpec ではエスケープ不可なので、
	// 含まれていたら明示的にエラーを促す代わりにここでは置換しておく。
	return "`" + strings.ReplaceAll(k, "`", "_") + "`"
}

// ToTSPLiteral は JSON 文字列を TypeSpec の value literal（#{...} / #[...]）に変換する。
// @opExample(#{returnType: <literal>}) に流し込むために使う。
//
// ポリシー:
//   - 大きな整数を float64 に落とさないよう json.Decoder.UseNumber を使う。
//   - オブジェクトキーは安全側で常にダブルクォート文字列として出力する。
//   - インデントは indent で指定されたスペース数を深さごとに積む。
func ToTSPLiteral(jsonStr string, indent int) (string, error) {
	dec := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
	dec.UseNumber()
	var v interface{}
	if err := dec.Decode(&v); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	var sb strings.Builder
	if err := encode(&sb, v, indent, 0); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func encode(sb *strings.Builder, v interface{}, indent, depth int) error {
	switch val := v.(type) {
	case nil:
		sb.WriteString("null")
	case bool:
		if val {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case json.Number:
		// 整数・浮動のどちらも JSON に現れた通りのテキストで出力する。
		// TypeSpec は任意精度の数値リテラルを受け付けるのでそのまま渡してよい。
		sb.WriteString(val.String())
	case string:
		sb.WriteString(strconv.Quote(val))
	case []interface{}:
		return encodeArray(sb, val, indent, depth)
	case map[string]interface{}:
		return encodeObject(sb, val, indent, depth)
	default:
		return fmt.Errorf("unsupported JSON value type %T", v)
	}
	return nil
}

func encodeArray(sb *strings.Builder, arr []interface{}, indent, depth int) error {
	if len(arr) == 0 {
		sb.WriteString("#[]")
		return nil
	}
	sb.WriteString("#[")
	for i, el := range arr {
		if indent > 0 {
			sb.WriteString("\n")
			sb.WriteString(pad(indent, depth+1))
		}
		if err := encode(sb, el, indent, depth+1); err != nil {
			return err
		}
		if i < len(arr)-1 {
			sb.WriteString(",")
			if indent == 0 {
				sb.WriteString(" ")
			}
		}
	}
	if indent > 0 {
		sb.WriteString("\n")
		sb.WriteString(pad(indent, depth))
	}
	sb.WriteString("]")
	return nil
}

func encodeObject(sb *strings.Builder, obj map[string]interface{}, indent, depth int) error {
	if len(obj) == 0 {
		sb.WriteString("#{}")
		return nil
	}
	// キー順を安定させる（YAML/OpenAPI diff を最小にするため）。
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb.WriteString("#{")
	for i, k := range keys {
		if indent > 0 {
			sb.WriteString("\n")
			sb.WriteString(pad(indent, depth+1))
		}
		sb.WriteString(encodeKey(k))
		sb.WriteString(": ")
		if err := encode(sb, obj[k], indent, depth+1); err != nil {
			return err
		}
		if i < len(keys)-1 {
			sb.WriteString(",")
			if indent == 0 {
				sb.WriteString(" ")
			}
		}
	}
	if indent > 0 {
		sb.WriteString("\n")
		sb.WriteString(pad(indent, depth))
	}
	sb.WriteString("}")
	return nil
}

func pad(indent, depth int) string {
	return strings.Repeat(" ", indent*depth)
}
