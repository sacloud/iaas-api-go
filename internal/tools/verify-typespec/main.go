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

// verify-typespec は generator が特別対応した既知の TypeSpec モデル・envelope について
// 期待するフィールドが含まれ続けているかを post-generation で検証する。
// 退行（リファクタ中に merge ロジックが外れる等）を検出するのが目的。
//
// 実行: `go run ./internal/tools/verify-typespec` （`spec/package.json` の verify ステップから呼ばれる想定）
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var repoRoot string

func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("failed to get current file path")
	}
	// internal/tools/verify-typespec/main.go → ../../.. でリポジトリルートを解決
	repoRoot = filepath.Join(filepath.Dir(filename), "../../..")
}

// check は 1 件の検証ケース。
type check struct {
	// label は失敗時メッセージで使う人間向けラベル。
	label string
	// tsp は repoRoot 相対の対象ファイル。
	tsp string
	// 以下の何れかを使う。
	// - model != "" && fieldsIncluded: 指定 model ブロック中に fieldsIncluded の各フィールドが定義されていること
	// - model != "" && notPresent: 指定 model が tsp に存在しないこと（merge により吸収されたはず）
	// - payloadTypeInModel != "": 指定 model ブロック中の payload フィールド (payloadTypeInModel) が payloadType で参照されていること
	model              string
	fieldsIncluded     []string
	fieldsAbsent       []string // 指定フィールドが model 内に存在してはいけない
	notPresent         bool
	payloadFieldName   string // 例: "Switch"
	payloadType        string // 例: "Switch"（"BridgeInfo" を期待してしまう退行を防ぐ）
	fieldOptionalName  string // 指定フィールドが optional / required のどちらで宣言されているかを検証するときに使う
	fieldOptionalWant  bool   // true = optional であること (`X?:` or `X?:`)、false = required (`X:`)
	rawContains        string // tsp ファイル全体にこの文字列が含まれるかを検証（model ブロックに収まらないトップレベル宣言の検査用）
}

var checks = []check{
	// Archive: POST /archive で Create（SourceDisk/SourceArchive）と CreateBlank（SizeMB）を 1 定義に統合
	{
		label:          "Archive POST /archive envelope body model merges Create + CreateBlank",
		tsp:            "spec/typespec/resources/archive/models.tsp",
		model:          "ArchiveCreateRequest",
		fieldsIncluded: []string{"SourceDisk", "SourceArchive", "SizeMB", "Name", "Description", "Tags", "Icon"},
	},
	{
		label:      "ArchiveCreateBlankRequest is absorbed into ArchiveCreateRequest (must not be emitted)",
		tsp:        "spec/typespec/resources/archive/models.tsp",
		model:      "ArchiveCreateBlankRequest",
		notPresent: true,
	},

	// Archive: POST /archive/:sid/to/zone/:did で Transfer と CreateFromShared を統合
	{
		label:          "Archive transfer endpoint merges Transfer + CreateFromShared",
		tsp:            "spec/typespec/resources/archive/models.tsp",
		model:          "ArchiveTransferRequest",
		fieldsIncluded: []string{"SizeMB", "SourceSharedKey", "Name", "Description", "Tags", "Icon"},
	},
	{
		label:      "ArchiveCreateRequestFromShared is absorbed (must not be emitted)",
		tsp:        "spec/typespec/resources/archive/models.tsp",
		model:      "ArchiveCreateRequestFromShared",
		notPresent: true,
	},

	// Archive: PUT /archive/:id/ftp で Share と OpenFTP を統合（wire payload 名は違うので envelope に両方入る）
	{
		label:          "Archive PUT /archive/:id/ftp envelope merges Share + OpenFTP",
		tsp:            "spec/typespec/resources/archive/envelopes.tsp",
		model:          "ArchiveShareRequestEnvelope",
		fieldsIncluded: []string{"Shared", "ChangePassword"},
	},

	// Disk: POST /disk で 4 variant の optional payload を envelope に union
	{
		label:          "Disk POST /disk envelope merges Create + 3 variants",
		tsp:            "spec/typespec/resources/disk/envelopes.tsp",
		model:          "DiskCreateRequestEnvelope",
		fieldsIncluded: []string{"Disk", "DistantFrom", "KMSKey", "Config", "BootAtAvailable", "TargetDedicatedStorageContract"},
	},
	// Disk: 本体 payload Disk は required、側次 payload（KMSKey/DistantFrom）は optional
	{
		label:             "Disk envelope: Disk payload required",
		tsp:               "spec/typespec/resources/disk/envelopes.tsp",
		model:             "DiskCreateRequestEnvelope",
		fieldOptionalName: "Disk",
		fieldOptionalWant: false,
	},
	{
		label:             "Disk envelope: KMSKey payload optional (not unconditionally sent by v1)",
		tsp:               "spec/typespec/resources/disk/envelopes.tsp",
		model:             "DiskCreateRequestEnvelope",
		fieldOptionalName: "KMSKey",
		fieldOptionalWant: true,
	},
	{
		label:             "Disk envelope: DistantFrom payload optional",
		tsp:               "spec/typespec/resources/disk/envelopes.tsp",
		model:             "DiskCreateRequestEnvelope",
		fieldOptionalName: "DistantFrom",
		fieldOptionalWant: true,
	},

	// Server: DELETE /server/:id で Delete + DeleteWithDisks を統合（WithDisk が optional 追加される）
	{
		label:          "Server DELETE /server/:id envelope merges Delete + DeleteWithDisks",
		tsp:            "spec/typespec/resources/server/envelopes.tsp",
		model:          "ServerDeleteRequestEnvelope",
		fieldsIncluded: []string{"WithDisk"},
	},

	// Server: PUT /server/:id/power で Boot + BootWithVariables を統合
	{
		label:          "Server PUT /server/:id/power envelope merges Boot + BootWithVariables",
		tsp:            "spec/typespec/resources/server/envelopes.tsp",
		model:          "ServerBootRequestEnvelope",
		fieldsIncluded: []string{"UserBootVariables"},
	},

	// Switch: POST /switch response envelope は Switch 型を参照するべき（以前 naked.Switch を共有する BridgeInfo に誤解決されるバグがあった）
	{
		label:            "SwitchCreateResponseEnvelope.Switch references Switch model (regression guard)",
		tsp:              "spec/typespec/resources/switch/envelopes.tsp",
		model:            "SwitchCreateResponseEnvelope",
		payloadFieldName: "Switch",
		payloadType:      "Switch",
	},
	{
		label:            "SwitchReadResponseEnvelope.Switch references Switch model",
		tsp:              "spec/typespec/resources/switch/envelopes.tsp",
		model:            "SwitchReadResponseEnvelope",
		payloadFieldName: "Switch",
		payloadType:      "Switch",
	},

	// 深さ3 mapconv のネスト展開 (Archive SourceArchiveInfo: ArchiveUnderZone.{Account,Zone,ID}.* )
	{
		label:          "SourceArchiveInfo has nested ArchiveUnderZone (depth-3 mapconv flattening regression guard)",
		tsp:            "spec/typespec/resources/archive/models.tsp",
		model:          "SourceArchiveInfo",
		fieldsIncluded: []string{"ArchiveUnderZone"},
	},
	{
		label:        "Flattened AccountID/ZoneID/ZoneName must NOT appear at SourceArchiveInfo top-level",
		tsp:          "spec/typespec/resources/archive/models.tsp",
		model:        "SourceArchiveInfo",
		fieldsAbsent: []string{"AccountID", "ZoneID", "ZoneName"},
	},
	{
		label:          "SourceArchiveInfoArchiveUnderZone carries Account and Zone sub-refs",
		tsp:            "spec/typespec/resources/archive/models.tsp",
		model:          "SourceArchiveInfoArchiveUnderZone",
		fieldsIncluded: []string{"ID", "Account", "Zone"},
	},

	// 深さ3 mapconv (Server Instance.Host.Name / Instance.CDROM.ID / Instance.HostInfoURL など)
	{
		label:          "Server Instance has nested Host submodel (Instance.Host.Name / Instance.Host.InfoURL)",
		tsp:            "spec/typespec/resources/server/models.tsp",
		model:          "ServerInstance",
		fieldsIncluded: []string{"Host", "CDROM"},
	},
	{
		label:          "ServerInstanceHost has Name and InfoURL",
		tsp:            "spec/typespec/resources/server/models.tsp",
		model:          "ServerInstanceHost",
		fieldsIncluded: []string{"Name", "InfoURL"},
	},

	// int ベースの enum は scalar extends int32 であるべき（string にすると JSON 数値の decode が失敗する）
	{
		label:       "EPlanGeneration is declared as int32 scalar",
		tsp:         "spec/typespec/types.tsp",
		rawContains: "scalar EPlanGeneration extends int32;",
	},
	{
		label:       "ENFSSize is declared as int32 scalar",
		tsp:         "spec/typespec/types.tsp",
		rawContains: "scalar ENFSSize extends int32;",
	},
	{
		label:       "EProxyLBPlan is declared as int32 scalar",
		tsp:         "spec/typespec/types.tsp",
		rawContains: "scalar EProxyLBPlan extends int32;",
	},

	// fieldNullabilityOverrides で nullable 化したフィールドの確認（v1 naked 型と実 API の挙動差異の救済）
	{
		label:             "InterfaceViewSwitchUserSubnet.DefaultRoute is optional (API returns null)",
		tsp:               "spec/typespec/resources/database/models.tsp",
		model:             "InterfaceViewSwitchUserSubnet",
		fieldOptionalName: "DefaultRoute",
		fieldOptionalWant: true,
	},
	{
		label:             "InterfaceViewSwitchUserSubnet.NetworkMaskLen is optional",
		tsp:               "spec/typespec/resources/database/models.tsp",
		model:             "InterfaceViewSwitchUserSubnet",
		fieldOptionalName: "NetworkMaskLen",
		fieldOptionalWant: true,
	},
	{
		label:             "SwitchUserSubnet.DefaultRoute is optional",
		tsp:               "spec/typespec/resources/switch/models.tsp",
		model:             "SwitchUserSubnet",
		fieldOptionalName: "DefaultRoute",
		fieldOptionalWant: true,
	},
	{
		label:             "SwitchUserSubnet.NetworkMaskLen is optional",
		tsp:               "spec/typespec/resources/switch/models.tsp",
		model:             "SwitchUserSubnet",
		fieldOptionalName: "NetworkMaskLen",
		fieldOptionalWant: true,
	},
}

func main() {
	failed := 0
	for _, c := range checks {
		if err := run(c); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL: %s\n  %v\n", c.label, err)
			failed++
		}
	}
	if failed > 0 {
		fmt.Fprintf(os.Stderr, "\nverify-typespec: %d check(s) failed\n", failed)
		os.Exit(1)
	}
	fmt.Printf("verify-typespec: all %d check(s) passed\n", len(checks))
}

func run(c check) error {
	path := filepath.Join(repoRoot, c.tsp)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", c.tsp, err)
	}
	content := string(data)

	if c.rawContains != "" {
		if !strings.Contains(content, c.rawContains) {
			return fmt.Errorf("%s does not contain expected text %q", c.tsp, c.rawContains)
		}
		return nil
	}

	if c.notPresent {
		if modelExists(content, c.model) {
			return fmt.Errorf("model %q should have been absorbed into its primary (via merge), but it is still emitted in %s", c.model, c.tsp)
		}
		return nil
	}

	block, ok := modelBlock(content, c.model)
	if !ok {
		return fmt.Errorf("model %q not found in %s", c.model, c.tsp)
	}

	if c.payloadFieldName != "" {
		re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(c.payloadFieldName) + `\??:\s*([A-Za-z0-9_]+)`)
		m := re.FindStringSubmatch(block)
		if m == nil {
			return fmt.Errorf("model %q has no payload field %q", c.model, c.payloadFieldName)
		}
		if m[1] != c.payloadType {
			return fmt.Errorf("model %q field %q has type %q, expected %q", c.model, c.payloadFieldName, m[1], c.payloadType)
		}
	}

	if c.fieldOptionalName != "" {
		got, ok := fieldOptional(block, c.fieldOptionalName)
		if !ok {
			return fmt.Errorf("model %q has no field %q", c.model, c.fieldOptionalName)
		}
		if got != c.fieldOptionalWant {
			want := "required"
			if c.fieldOptionalWant {
				want = "optional"
			}
			actual := "required"
			if got {
				actual = "optional"
			}
			return fmt.Errorf("model %q field %q is %s, expected %s", c.model, c.fieldOptionalName, actual, want)
		}
	}

	for _, f := range c.fieldsIncluded {
		if !fieldInBlock(block, f) {
			return fmt.Errorf("model %q missing field %q in %s", c.model, f, c.tsp)
		}
	}
	for _, f := range c.fieldsAbsent {
		if fieldInBlock(block, f) {
			return fmt.Errorf("model %q unexpectedly contains field %q (should be nested under a sub-model) in %s", c.model, f, c.tsp)
		}
	}
	return nil
}

// fieldOptional は model block から指定フィールドを探し、optional (`X?:`) なら true、required (`X:`) なら false を返す。
// フィールドが見つからなければ ok=false。
func fieldOptional(block, fieldName string) (bool, bool) {
	re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(fieldName) + `(\??):\s`)
	m := re.FindStringSubmatch(block)
	if m == nil {
		return false, false
	}
	return m[1] == "?", true
}

// modelExists は tsp 本文に `model <Name> {` の宣言があるかを返す。
func modelExists(content, name string) bool {
	re := regexp.MustCompile(`(?m)^model\s+` + regexp.QuoteMeta(name) + `\s*\{`)
	return re.MatchString(content)
}

// modelBlock は `model <Name> { ... }` の中身（中括弧の間の文字列）を返す。
func modelBlock(content, name string) (string, bool) {
	re := regexp.MustCompile(`(?m)^model\s+` + regexp.QuoteMeta(name) + `\s*\{`)
	loc := re.FindStringIndex(content)
	if loc == nil {
		return "", false
	}
	// 宣言行の `{` の位置から対応する `}` までを取り出す（TypeSpec はこのレベルでネスト block を含まない想定）
	start := loc[1]
	depth := 1
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[start:i], true
			}
		}
	}
	return "", false
}

// fieldInBlock は model block 中に行頭（インデント含む）からフィールド宣言 `Name:` or `Name?:` があるかを返す。
func fieldInBlock(block, fieldName string) bool {
	re := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(fieldName) + `\??:\s`)
	return re.MatchString(block)
}

