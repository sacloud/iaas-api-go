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
	"text/template"
)

// descriptionFuncs はモデル/エンベロープテンプレートで @doc(...) の値をエスケープするための FuncMap。
// TypeSpec の `"..."` 文字列リテラルに埋め込む際、バックスラッシュとダブルクォートをエスケープする。
// 改行は単純置換で `\n` にする (説明文は基本的に単行想定)。
var descriptionFuncs = template.FuncMap{
	"docEscape": func(s string) string {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `\"`)
		s = strings.ReplaceAll(s, "\n", `\n`)
		return s
	},
}

// fieldDescriptions は TypeSpec モデル・エンベロープのフィールドに付与する日本語説明文を、
// (モデル名, フィールド名) 単位で保持するレジストリ。
// さくらのクラウド公式マニュアル (https://manual.sakura.ad.jp/cloud-api/1.1/) を一次情報源とする。
//
// 登録はリソース別 descriptions_<resource>.go の init() 関数から行う。
// 未登録のフィールドは commonFieldDescriptions (モデル名にかかわらず適用される fallback) を参照し、
// それでも見つからない場合は生成物に description を emit しない（従来のフィールド名 placeholder のまま）。
var fieldDescriptions = map[string]map[string]string{}

// registerFieldDescriptions はリソース別の説明 map を fieldDescriptions にマージする。
// 同一 (model, field) で異なる説明が登録された場合は panic して早期に検知する。
func registerFieldDescriptions(src map[string]map[string]string) {
	for model, fields := range src {
		dst, ok := fieldDescriptions[model]
		if !ok {
			dst = map[string]string{}
			fieldDescriptions[model] = dst
		}
		for f, desc := range fields {
			if existing, dup := dst[f]; dup && existing != desc {
				panic("duplicate field description: " + model + "." + f)
			}
			dst[f] = desc
		}
	}
}

// lookupFieldDescription は (modelName, fieldName) に対応する説明を返す。
// モデル固有の登録が優先され、見つからなければ commonFieldDescriptions を参照する。
// 何も登録されていなければ空文字列を返す。
func lookupFieldDescription(modelName, fieldName string) string {
	if m, ok := fieldDescriptions[modelName]; ok {
		if d, ok := m[fieldName]; ok {
			return d
		}
	}
	if d, ok := commonFieldDescriptions[fieldName]; ok {
		return d
	}
	return ""
}
