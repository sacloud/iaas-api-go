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
	// define.APIs.Models() から naked 型名 → モデル名のマップを自動構築
	for _, model := range define.APIs.Models() {
		if model.NakedType == nil {
			continue
		}
		goType := model.NakedType.GoTypeSourceCode()
		goType = strings.TrimPrefix(goType, "*")
		if idx := strings.LastIndex(goType, "."); idx >= 0 {
			goType = goType[idx+1:]
		}
		if _, exists := m[goType]; !exists {
			m[goType] = model.Name
		}
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
type resolvedPayload struct {
	Name   string
	TSType string
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

// envelopeInfo は1オペレーション分のエンベロープ情報を保持する。
type envelopeInfo struct {
	HasRequestEnvelope           bool
	HasResponseEnvelope          bool
	RequestEnvelopeStructName    string
	ResponseEnvelopeStructName   string
	RequestEnvelopeTypeSpecName  string
	ResponseEnvelopeTypeSpecName string
	IsRequestSingular            bool
	IsRequestPlural              bool
	IsResponseSingular           bool
	IsResponsePlural             bool
	RequestPayloads              []resolvedPayload
	ResponsePayloads             []*dsl.EnvelopePayloadDesc
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
	{{ .Name }}: {{ .TSType }}{{ if $isPlural }}[]{{ end }};
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
	@doc("Total count of target resources")
	Total: int32;

	@doc("Current page number")
	From: int32;

	@doc("Count of current page")
	Count: int32;

{{- range .ResponsePayloads }}
	/**
	 * {{ .Name }}
	 */
	{{ .Name }}: {{ goTypeToTypeSpec .TypeName }}{{ if $isPlural }}[]{{ end }};
{{- end }}
{{- else }}
	@doc("is_ok - Operation result indicator")
	is_ok: boolean;

	@doc("success - API result status")
	Success?: boolean;

{{- range .ResponsePayloads }}
	/**
	 * {{ .Name }}
	 */
	{{ .Name }}: {{ goTypeToTypeSpec .TypeName }};
{{- end }}
{{- end }}
}
{{- end }}
{{ end }}`

func generateEnvelopes() {
	for _, api := range define.APIs {
		var envelopes []envelopeInfo

		for _, op := range api.Operations {
			hasRequestEnvelope := op.HasRequestEnvelope()
			hasResponseEnvelope := op.HasResponseEnvelope()

			if !hasRequestEnvelope && !hasResponseEnvelope {
				continue
			}

			reqStructName := op.RequestEnvelopeStructName()
			respStructName := op.ResponseEnvelopeStructName()

			// Sort/Include/Exclude は定義しない（AGENTS.md: 複雑性が高すぎる）
			skipFields := map[string]bool{"Sort": true, "Include": true, "Exclude": true}
			var reqPayloads []resolvedPayload
			for _, p := range op.RequestPayloads() {
				if skipFields[p.Name] {
					continue
				}
				reqPayloads = append(reqPayloads, resolvedPayload{
					Name:   p.Name,
					TSType: resolveRequestPayloadTSType(op, p),
				})
			}

			envelopes = append(envelopes, envelopeInfo{
				HasRequestEnvelope:           hasRequestEnvelope,
				HasResponseEnvelope:          hasResponseEnvelope,
				RequestEnvelopeStructName:    reqStructName,
				ResponseEnvelopeStructName:   respStructName,
				RequestEnvelopeTypeSpecName:  upperFirst(reqStructName),
				ResponseEnvelopeTypeSpecName: upperFirst(respStructName),
				IsRequestSingular:            op.IsRequestSingular(),
				IsRequestPlural:              op.IsRequestPlural(),
				IsResponseSingular:           op.IsResponseSingular(),
				IsResponsePlural:             op.IsResponsePlural(),
				RequestPayloads:              reqPayloads,
				ResponsePayloads:             op.ResponsePayloads(),
			})
		}

		if len(envelopes) == 0 {
			// エンベロープがなければスキップ
			continue
		}

		// nakedInlineModels に対応する naked 型が使われている場合、ExtraModels に収集する
		// レスポンスペイロードのみ対象（リクエストペイロードは引数モデルから解決済み）
		usedInlineModels := map[string]bool{}
		var extraModels []fatModelDef
		for _, env := range envelopes {
			for _, p := range env.ResponsePayloads {
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
