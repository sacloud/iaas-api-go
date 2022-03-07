// Copyright 2022 The sacloud/iaas-api-go Authors
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

const destination = "zz_result.go"

func init() {
	log.SetFlags(0)
	log.SetPrefix("gen-api-result: ")
}

func main() {
	outputPath := destination
	tools.WriteFileWithTemplate(&tools.TemplateConfig{
		OutputPath: filepath.Join(tools.ProjectRootPath(), outputPath),
		Template:   tmpl,
		Parameter:  define.APIs,
	})
	log.Printf("generated: %s\n", outputPath)
}

const tmpl = `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-api-result'; DO NOT EDIT

package iaas

{{- range . }}
{{- range .Operations -}}

{{ if .HasResults }}
// {{ .ResultTypeName }} represents the Result of API 
type {{ .ResultTypeName }} struct {
{{- if .IsResponsePlural -}}
	Total       int        ` + "`" + `json:",omitempty"` + "`" + ` // Total count of target resources
	From        int        ` + "`" + `json:",omitempty"` + "`" + ` // Current page number
	Count       int        ` + "`" + `json:",omitempty"` + "`" + ` // Count of current page
{{ else }}
	IsOk    bool  ` + "`" + `json:",omitempty"` + "`" + ` // is_ok
{{ end }}
{{ if .IsResponseSingular }}
	{{- range .Results}}
	{{.DestField}} {{.GoTypeSourceCode}} {{.TagString}}
	{{- end }}
{{- else if .IsResponsePlural -}}
	{{- range .Results}}
	{{.DestField}} []{{.GoTypeSourceCode}} {{.TagString}}
	{{- end }}
{{ end }}
}

{{ if .IsResponsePlural }}{{ if eq (len .Results) 1 }}
// Values returns find results
func (r *{{ .ResultTypeName }}) Values() []interface{} {
	var results []interface{}
	for _ , v := range r.{{ (index .Results 0).DestField }} {
		results = append(results, v)
	}
	return results
}
{{ end }}{{ end }}

{{ end }}

{{- end -}}
{{- end -}}
`
