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
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sacloud/iaas-api-go/internal/define"
	"github.com/sacloud/iaas-api-go/internal/dsl"
)


// nakedTypeToTSName は naked 型名 → TypeSpec 型名のマップ。
// init() で define.APIs.Models() から動的に構築し、手動エントリで補完する。
var nakedTypeToTSName map[string]string

func init() {
	nakedTypeToTSName = buildNakedToTSNameMap()
}

// buildNakedToTSNameMap は define.APIs から naked 型名 → TypeSpec モデル名のマップを構築する。
// naked 型名と TypeSpec モデル名が異なるケースを自動的に解決する。
func buildNakedToTSNameMap() map[string]string {
	// 手動マッピング（動的検出できないもの）
	// naked 型名と TypeSpec モデル名が異なるケース、またはモデルなしの型を網羅する
	m := map[string]string{
		// naked 型名 → TypeSpec モデル名（name が異なるもの）
		"OpeningFTPServer":                            "FTPServer",
		"AutoBackupSettingsUpdate":                    "AutoBackupUpdateSettingsRequest",
		"AutoScaleSettingsUpdate":                     "AutoScaleUpdateSettingsRequest",
		"AutoScaleRunningStatus":                      "AutoScaleStatus",
		"CertificateAuthorityAddClientParameter":      "CertificateAuthorityAddClientParam",
		"CertificateAuthorityClientDetail":            "CertificateAuthorityClient",
		"CertificateAuthorityAddServerParameter":      "CertificateAuthorityAddServerParam",
		"CertificateAuthorityServerDetail":            "CertificateAuthorityServer",
		"ContainerRegistrySettingsUpdate":             "ContainerRegistryUpdateSettingsRequest",
		"DatabaseSettingsUpdate":                      "DatabaseUpdateSettingsRequest",
		"DatabaseStatusResponse":                      "DatabaseStatus",
		"DatabaseParameterSetting":                    "DatabaseParameter",
		"DiskEdit":                                    "DiskEditRequest",
		"DNSSettingsUpdate":                           "DNSUpdateSettingsRequest",
		"EnhancedDBPasswordSettings":                  "EnhancedDBSetPasswordRequest",
		"EnhancedDBConfigSettings":                    "EnhancedDBSetConfigRequest",
		"GSLBSettingsUpdate":                          "GSLBUpdateSettingsRequest",
		"LoadBalancerSettingsUpdate":                  "LoadBalancerUpdateSettingsRequest",
		"LocalRouterSettingsUpdate":                   "LocalRouterUpdateSettingsRequest",
		"MobileGatewaySettingsUpdate":                 "MobileGatewayUpdateSettingsRequest",
		"ProxyLBSettingsUpdate":                       "ProxyLBUpdateSettingsRequest",
		"ProxyLBPlanChange":                           "ProxyLBChangePlanRequest",
		"SimpleMonitorSettingsUpdate":                 "SimpleMonitorUpdateSettingsRequest",
		"SimpleMonitorHealthCheckStatus":              "SimpleMonitorHealthStatus",
		"SimpleNotificationGroupSettingsUpdate":       "SimpleNotificationGroupUpdateSettingsRequest",
		"SimpleNotificationDestinationSettingsUpdate": "SimpleNotificationDestinationUpdateSettingsRequest",
		"SimpleNotificationDestinationRunningStatus":  "SimpleNotificationDestinationStatus",
		"TrafficMonitoringConfig":                     "MobileGatewayTrafficControl",
		"TrafficStatus":                               "MobileGatewayTrafficStatus",
		"VPCRouterSettingsUpdate":                     "VPCRouterUpdateSettingsRequest",
		"VPCRouterPingResult":                         "VPCRouterPingResults",
		"ESMESendSMSRequest":                          "ESMESendMessageWithInputtedOTPRequest",
		"ESMESendSMSResponse":                         "ESMESendMessageResult",
		// TypeSpec モデルなし → unknown
		"MonitorValues":            "unknown",
		"SortKeys":                 "unknown",
		"Filter":                   "Record<unknown>",
		"KMSKey":                   "unknown",
		"DedicatedStorageContract": "unknown",
		// MobileGatewaySIMGroup は nakedInlineModels でインライン定義するため除外
	}
	// define.APIs.Models() から naked 型名 → モデル名のマップを自動構築。
	// 同じ naked 型を共有するモデルが複数ある場合（例: naked.Switch を使う Switch と BridgeInfo）、
	// モデル名が Go 型名と一致するものを優先する（自己参照モデルをエンベロープの自然な型名にするため）。
	type mapEntry struct {
		name      string
		nameMatch bool
	}
	candidates := map[string]mapEntry{}
	for _, model := range define.APIs.Models() {
		if model.NakedType == nil {
			continue
		}
		goType := model.NakedType.GoTypeSourceCode()
		goType = strings.TrimPrefix(goType, "*")
		if idx := strings.LastIndex(goType, "."); idx >= 0 {
			goType = goType[idx+1:]
		}
		if _, manual := m[goType]; manual {
			// 手動マッピングは上書きしない
			continue
		}
		nameMatch := model.Name == goType
		if existing, ok := candidates[goType]; ok {
			if existing.nameMatch || !nameMatch {
				// 既に name-match 候補が採用済み、または新しい候補も name-match でないなら維持
				continue
			}
		}
		candidates[goType] = mapEntry{name: model.Name, nameMatch: nameMatch}
	}
	for goType, c := range candidates {
		m[goType] = c.name
	}
	return m
}

// envelopePayloadTypeToTS は Go型名→TypeSpec型名変換関数（再帰対応）。
// テンプレートには "goTypeToTypeSpec" キーで渡す。
func envelopePayloadTypeToTS(goType string) string {
	// map 型
	if strings.HasPrefix(goType, "map[") {
		return "Record<unknown>"
	}

	// スライス型（[]*T と []T）は内側の型を再帰的に変換して [] を後置する
	if strings.HasPrefix(goType, "[]*") {
		return envelopePayloadTypeToTS(goType[3:]) + "[]"
	}
	if strings.HasPrefix(goType, "[]") {
		return envelopePayloadTypeToTS(goType[2:]) + "[]"
	}

	// ポインタ型を削除
	goType = strings.TrimPrefix(goType, "*")

	// パッケージ名を処理（time.Time は utcDateTime に変換）
	if idx := strings.LastIndex(goType, "."); idx != -1 {
		pkg := goType[:idx]
		typeName := goType[idx+1:]
		if pkg == "time" && typeName == "Time" {
			return "utcDateTime"
		}
		goType = typeName
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
		// naked 型名と TypeSpec モデル名が異なる場合はマッピングを使用
		if ts, ok := nakedTypeToTSName[goType]; ok {
			return ts
		}
		return goType // モデル型はそのまま TypeSpec の型名として使用
	}
}

// nakedInlineModels は naked 型に対応する TypeSpec モデル定義を envelopes.tsp にインライン出力するための定義。
// nakedTypeToTSName で型名が自動解決できない場合に、envelopes.tsp 内にモデルを直接出力する。
// 現時点では使用エントリなし（将来の追加用に仕組みのみ保持）。
var nakedInlineModels = map[string]fatModelDef{}

// stripNakedTypeName は Go の完全修飾型名からパッケージ名を除いた裸の型名を返す。
// 例: "*naked.MobileGatewaySIMGroup" → "MobileGatewaySIMGroup"
func stripNakedTypeName(goType string) string {
	goType = strings.TrimPrefix(goType, "*")
	if strings.HasPrefix(goType, "[]*") {
		return stripNakedTypeName(goType[3:])
	}
	if strings.HasPrefix(goType, "[]") {
		return stripNakedTypeName(goType[2:])
	}
	if idx := strings.LastIndex(goType, "."); idx != -1 {
		return goType[idx+1:]
	}
	return goType
}

// resolvedPayload はエンベロープ内の1フィールド分の情報（TypeSpec型名解決済み）。
// 同一 method+path の op 群を合成したエンベロープでは、一部 op のみに存在する payload は Optional になる。
type resolvedPayload struct {
	Name     string
	TSType   string
	Optional bool
}

// resolveRequestPayloadTSType はオペレーションの引数からリクエストペイロードの TypeSpec 型名を解決する。
// MappableArgument でマッピングされた Model 引数がある場合はそのモデル名を使い、
// なければ naked 型名から envelopePayloadTypeToTS で変換したものを返す。
func resolveRequestPayloadTSType(op *dsl.Operation, payload *dsl.EnvelopePayloadDesc) string {
	for _, arg := range op.Arguments {
		if model, ok := arg.Type.(*dsl.Model); ok {
			// MapConvTag は "{destField},recursive" の形式
			destField := strings.SplitN(arg.MapConvTag, ",", 2)[0]
			if destField == payload.Name {
				return model.Name
			}
		}
	}
	return envelopePayloadTypeToTS(payload.TypeName())
}

// envelopeInfo は1エンベロープ分の情報を保持する（単一 op または合成 op 群）。
type envelopeInfo struct {
	HasRequestEnvelope           bool
	HasResponseEnvelope           bool
	RequestEnvelopeStructName    string
	ResponseEnvelopeStructName   string
	RequestEnvelopeTypeSpecName  string
	ResponseEnvelopeTypeSpecName string
	IsRequestSingular            bool
	IsRequestPlural              bool
	IsResponseSingular           bool
	IsResponsePlural             bool
	RequestPayloads              []resolvedPayload
	ResponsePayloads             []resolvedPayload
}

// resourceEnvelopes は1リソース分の全エンベロープをまとめたテンプレートパラメータ。
type resourceEnvelopes struct {
	Envelopes   []envelopeInfo
	ExtraModels []fatModelDef // nakedInlineModels から収集したインラインモデル定義
}

// envelopesTmpl は1リソース分の全エンベロープを1ファイルに出力するテンプレート。
// range .Envelopes の中では . が envelopeInfo になるため、$.IsRequestPlural ではなく
// .IsRequestPlural を使う。RequestPayloads の range では with で外側 envelope を保持する。
const envelopesTmpl = `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-typespec'; DO NOT EDIT

import "@typespec/http";

namespace Sacloud.IaaS;
{{ range .ExtraModels }}
model {{ .Name }} {
{{- range .Fields }}
  {{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }};
{{- end }}
}
{{ end }}
{{ range .Envelopes }}
{{- if .HasRequestEnvelope }}
/**
 * Request envelope for {{ .RequestEnvelopeStructName }}
 */
{{- $isPlural := .IsRequestPlural }}
model {{ .RequestEnvelopeTypeSpecName }} {
{{- range .RequestPayloads }}
	/**
	 * {{ .Name }}
	 */
	{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }}{{ if $isPlural }}[]{{ end }};
{{- end }}
}
{{- end }}
{{- if .HasResponseEnvelope }}
/**
 * Response envelope for {{ .ResponseEnvelopeStructName }}
 */
{{- $isPlural := .IsResponsePlural }}
model {{ .ResponseEnvelopeTypeSpecName }} {
{{- if .IsResponsePlural }}
	@doc("対象リソースの総件数")
	Total: int32;

	@doc("現在のページ番号")
	From: int32;

	@doc("現在のページの件数")
	Count: int32;

{{- range .ResponsePayloads }}
	/**
	 * {{ .Name }}
	 */
	{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }}{{ if $isPlural }}[]{{ end }};
{{- end }}
{{- else }}
	@doc("オペレーションが成功したかどうかを示すフラグ。成功判定にはこのフィールドを用いること。")
	is_ok: boolean;
{{- range .ResponsePayloads }}
	/**
	 * {{ .Name }}
	 */
	{{ .Name }}{{ if .Optional }}?{{ end }}: {{ .TSType }};
{{- end }}
{{- end }}
}
{{- end }}
{{ end }}`

// buildMergedEnvelopeInfos は api の全オペレーションを (method, path) でグループ化し、
// グループ単位でマージされた envelopeInfo のリストを返す。
// 複数 op が同一 method+path を共有する場合（例: Disk の Create/CreateWithConfig/...）、
// 各 op の request/response payload を union でマージする:
//   - 全 op に存在する payload は required
//   - 一部 op のみに存在する payload は optional
//
// エンベロープ名は primaryOpForKey（最短名の op）のものを採用する。
// 出力順は最初に現れたグループの順を維持する。
// ops.go からも同じ仕組みで「このグループの統合エンベロープ名」を参照するため、同じ関数を使う。
func buildMergedEnvelopeInfos(api *dsl.Resource) ([]envelopeInfo, map[opKey]string) {
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

	// Sort/Include/Exclude は定義しない（AGENTS.md: 複雑性が高すぎる）
	skipFields := map[string]bool{"Sort": true, "Include": true, "Exclude": true}

	var envelopes []envelopeInfo
	envelopeNameByKey := map[opKey]string{}
	seenEnvelope := map[string]bool{}

	for _, g := range groups {
		primary := primaryOpForKey(g.ops)
		hasReq, hasResp := false, false
		for _, op := range g.ops {
			if op.HasRequestEnvelope() {
				hasReq = true
			}
			if op.HasResponseEnvelope() {
				hasResp = true
			}
		}
		if !hasReq && !hasResp {
			continue
		}

		reqName := primary.RequestEnvelopeStructName()
		respName := primary.ResponseEnvelopeStructName()
		if hasReq {
			envelopeNameByKey[g.key] = upperFirst(reqName)
		}

		// primary op を先頭に並べて payload TS type の解決で primary の argument model が優先されるようにする。
		// （api.Operations 順では primary 以外が先に visit される可能性があり、envelope 名と payload 型名が
		//  別の op に由来する不整合を招くため）
		visitOrder := []*dsl.Operation{primary}
		for _, op := range g.ops {
			if op != primary {
				visitOrder = append(visitOrder, op)
			}
		}

		// リクエスト payload を union でマージ
		total := len(g.ops)
		reqIndex := map[string]*resolvedPayload{}
		reqCount := map[string]int{}
		var reqOrder []string
		for _, op := range visitOrder {
			for _, p := range op.RequestPayloads() {
				if skipFields[p.Name] {
					continue
				}
				if _, exists := reqIndex[p.Name]; !exists {
					reqIndex[p.Name] = &resolvedPayload{
						Name:   p.Name,
						TSType: resolveRequestPayloadTSType(op, p),
					}
					reqOrder = append(reqOrder, p.Name)
				}
				reqCount[p.Name]++
			}
		}
		var reqPayloads []resolvedPayload
		for _, name := range reqOrder {
			p := *reqIndex[name]
			p.Optional = reqCount[name] < total
			reqPayloads = append(reqPayloads, p)
		}

		// レスポンス payload を union でマージ
		respIndex := map[string]*resolvedPayload{}
		respCount := map[string]int{}
		var respOrder []string
		for _, op := range visitOrder {
			for _, p := range op.ResponsePayloads() {
				if _, exists := respIndex[p.Name]; !exists {
					respIndex[p.Name] = &resolvedPayload{
						Name:   p.Name,
						TSType: envelopePayloadTypeToTS(p.TypeName()),
					}
					respOrder = append(respOrder, p.Name)
				}
				respCount[p.Name]++
			}
		}
		var respPayloads []resolvedPayload
		for _, name := range respOrder {
			p := *respIndex[name]
			p.Optional = respCount[name] < total
			respPayloads = append(respPayloads, p)
		}

		reqTS := upperFirst(reqName)
		respTS := upperFirst(respName)
		// 同名エンベロープが既に登録されていればスキップ（複数グループで primary が同じ名前になる稀ケース）
		if (hasReq && seenEnvelope[reqTS]) || (hasResp && seenEnvelope[respTS]) {
			continue
		}
		if hasReq {
			seenEnvelope[reqTS] = true
		}
		if hasResp {
			seenEnvelope[respTS] = true
		}

		envelopes = append(envelopes, envelopeInfo{
			HasRequestEnvelope:           hasReq,
			HasResponseEnvelope:          hasResp,
			RequestEnvelopeStructName:    reqName,
			ResponseEnvelopeStructName:   respName,
			RequestEnvelopeTypeSpecName:  reqTS,
			ResponseEnvelopeTypeSpecName: respTS,
			IsRequestSingular:            primary.IsRequestSingular(),
			IsRequestPlural:              primary.IsRequestPlural(),
			IsResponseSingular:           primary.IsResponseSingular(),
			IsResponsePlural:             primary.IsResponsePlural(),
			RequestPayloads:              reqPayloads,
			ResponsePayloads:             respPayloads,
		})
	}

	return envelopes, envelopeNameByKey
}

func generateEnvelopes() {
	for _, api := range define.APIs {
		envelopes, _ := buildMergedEnvelopeInfos(api)

		if len(envelopes) == 0 {
			// エンベロープがなければスキップ
			continue
		}

		// nakedInlineModels に対応する naked 型が使われている場合、ExtraModels に収集する
		// レスポンスペイロードのみ対象（リクエストペイロードは引数モデルから解決済み）
		// envelopeInfo は TypeSpec 型名しか持たないため、naked 型名の取得には元の DSL op を直接辿る。
		usedInlineModels := map[string]bool{}
		var extraModels []fatModelDef
		for _, op := range api.Operations {
			for _, p := range op.ResponsePayloads() {
				name := stripNakedTypeName(p.TypeName())
				if def, ok := nakedInlineModels[name]; ok && !usedInlineModels[def.Name] {
					usedInlineModels[def.Name] = true
					extraModels = append(extraModels, def)
				}
			}
		}

		outFile := filepath.Join(resourcesDir, api.FileSafeName(), "envelopes.tsp")
		writeFile(envelopesTmpl, resourceEnvelopes{Envelopes: envelopes, ExtraModels: extraModels}, outFile, template.FuncMap{
			"lowerFirst":       lowerFirst,
			"goTypeToTypeSpec": envelopePayloadTypeToTS,
		})
	}
}
