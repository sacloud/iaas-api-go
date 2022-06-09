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
	"github.com/sacloud/iaas-api-go/internal/dsl"
	"github.com/sacloud/iaas-api-go/internal/tools"
)

const destination = "trace/zz_api_tracer.go"

func init() {
	log.SetFlags(0)
	log.SetPrefix("gen-api-tracer: ")
}

func main() {
	dsl.IsOutOfSacloudPackage = true

	tools.WriteFileWithTemplate(&tools.TemplateConfig{
		OutputPath: filepath.Join(tools.ProjectRootPath(), destination),
		Template:   tmpl,
		Parameter:  define.APIs,
	})
	log.Printf("generated: %s\n", filepath.Join(tools.ProjectRootPath(), destination))
}

const tmpl = `// generated by 'github.com/sacloud/iaas-api-go/internal/tools/gen-api-tracer'; DO NOT EDIT

package trace

import (
{{- range .ImportStatements "context" "encoding/json" "log" "sync"}}
	{{ . }}
{{- end }}
)

var initOnce sync.Once

// AddClientFactoryHooks add client factory hooks
func AddClientFactoryHooks() {
	initOnce.Do(func(){
		addClientFactoryHooks()
	})
}

func addClientFactoryHooks() {
{{ range . -}} 
	iaas.AddClientFacotyHookFunc("{{.TypeName}}", func(in interface{}) interface{} {
		return New{{.TypeName}}Tracer(in.(iaas.{{.TypeName}}API))
	})
{{ end -}}
}

{{ range . }} {{$typeName := .TypeName}} {{ $resource := . }}
/************************************************* 
* {{ $typeName }}Tracer
*************************************************/

// {{ $typeName }}Tracer is for trace {{ $typeName }}Op operations
type {{ $typeName }}Tracer struct {
	Internal iaas.{{$typeName}}API
}

// New{{ $typeName}}Tracer creates new {{ $typeName}}Tracer instance
func New{{ $typeName}}Tracer(in iaas.{{$typeName}}API) iaas.{{$typeName}}API {
	return &{{ $typeName}}Tracer {
		Internal: in,
	}
}

{{ range .Operations }}{{$returnErrStatement := .ReturnErrorStatement}}{{ $operationName := .MethodName }}
// {{ .MethodName }} is API call with trace log
func (t *{{ $typeName }}Tracer) {{ .MethodName }}(ctx context.Context{{if not $resource.IsGlobal}}, zone string{{end}}{{ range .Arguments }}, {{ .ArgName }} {{ .TypeName }}{{ end }}) {{.ResultsStatement}} {
	log.Println("[TRACE] {{ $typeName }}API.{{ .MethodName }} start")
	targetArguments := struct {
{{ if not $resource.IsGlobal }}
		Argzone string
{{ end -}}
{{ range .Arguments -}}
		Arg{{.ArgName}} {{.TypeName}} ` + "`json:\"{{.ArgName}}\"`" + `
{{ end -}}
	} {
{{if not $resource.IsGlobal -}}
		Argzone: zone,
{{ end -}}
{{ range .Arguments -}}
		Arg{{.ArgName}}: {{.ArgName}},
{{ end -}}
	}
	if d, err := json.Marshal(targetArguments); err == nil {
		log.Printf("[TRACE] \targs: %s\n", string(d))
	}

	defer func() {
		log.Println("[TRACE] {{ $typeName }}API.{{ .MethodName }} end")
	}()

	{{range .ResultsTypeInfo}}{{.VarName}}, {{end}}err := t.Internal.{{ .MethodName }}(ctx{{if not $resource.IsGlobal}}, zone{{end}}{{ range .Arguments }}, {{ .ArgName }}{{ end }})
	targetResults := struct {
{{ range .ResultsTypeInfo -}}
		{{.FieldName}} {{.Type.GoTypeSourceCode}} 
{{ end -}}
		Error error
	} {
{{ range .ResultsTypeInfo -}}
		{{.FieldName}}: {{.VarName}},
{{ end -}}
		Error: err,
	}
	if d, err := json.Marshal(targetResults); err == nil {
		log.Printf("[TRACE] \tresults: %s\n", string(d))
	}

	return {{range .ResultsTypeInfo}}{{.VarName}}, {{end}}err
}
{{- end -}}

{{ end }}
`
