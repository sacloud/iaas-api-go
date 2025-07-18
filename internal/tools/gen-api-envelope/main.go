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
	"log"
	"path/filepath"

	"github.com/sacloud/iaas-api-go/internal/define"
	"github.com/sacloud/iaas-api-go/internal/tools"
)

const destination = "zz_envelopes.go"

func init() {
	log.SetFlags(0)
	log.SetPrefix("gen-api-envelope: ")
}

func main() {
	outputPath := filepath.Join(tools.ProjectRootPath(), destination)

	tools.WriteFileWithTemplate(&tools.TemplateConfig{
		OutputPath: outputPath,
		Template:   tmpl,
		Parameter:  define.APIs,
	})

	log.Printf("generated: %s\n", outputPath)
}

const tmpl = `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-api-envelope'; DO NOT EDIT

package iaas

import (
{{- range .ImportStatements "github.com/sacloud/iaas-api-go/types" "github.com/sacloud/iaas-api-go/naked" "github.com/sacloud/iaas-api-go/search" }}
	{{ . }}
{{- end }}
)

{{- range . }}
{{- range .Operations -}}

{{ if .HasRequestEnvelope }}
// {{ .RequestEnvelopeStructName }} is envelop of API request
type {{ .RequestEnvelopeStructName }} struct {
{{ if .IsRequestSingular }}
	{{- range .RequestPayloads}}
	{{.Name}} {{.TypeName}} {{.TagString}}
	{{- end }}
{{- else if .IsRequestPlural -}}
	{{- range .RequestPayloads}}
	{{.Name}} []{{.TypeName}} {{.TagString}}
	{{- end }}
{{ end }}
}
{{ end }}

{{ if .HasResponseEnvelope }}
// {{ .ResponseEnvelopeStructName }} is envelop of API response
type {{ .ResponseEnvelopeStructName }} struct {
{{- if .IsResponsePlural -}}
	Total       int        ` + "`" + `json:",omitempty"` + "`" + ` // トータル件数
	From        int        ` + "`" + `json:",omitempty"` + "`" + ` // ページング開始ページ
	Count       int        ` + "`" + `json:",omitempty"` + "`" + ` // 件数
{{ else }}
	IsOk    bool  ` + "`" + `json:"is_ok,omitempty"` + "`" + ` // is_ok項目
	Success types.APIResult  ` + "`" + `json:",omitempty"` + "`" + `      // success項目
{{ end }}
{{ if .IsResponseSingular }}
	{{- range .ResponsePayloads}}
	{{.Name}} {{.TypeName}} {{.TagString}}
	{{- end }}
{{- else if .IsResponsePlural -}}
	{{- range .ResponsePayloads}}
	{{.Name}} []{{.TypeName}} {{.TagString}}
	{{- end }}
{{ end }}
}
{{ end }}

{{- end -}}
{{- end -}}
`
