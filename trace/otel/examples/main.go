// Copyright 2022-2023 The sacloud/iaas-api-go Authors
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
	"context"
	"log"
	"net/http"
	"os"
	"time"

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/api"
	"github.com/sacloud/iaas-api-go/ostype"
	sacloudotel "github.com/sacloud/iaas-api-go/trace/otel"
	"github.com/sacloud/iaas-api-go/types"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// ref: https://github.com/open-telemetry/opentelemetry-go/blob/v1.2.0/example/jaeger/main.go

// Example ローカルのJaegerを利用する例
func main() {
	ctx := context.Background()

	tp, err := tracerProvider("http://localhost:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}

	otel.SetTracerProvider(tp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	// サンプルAPIリクエスト
	op(ctx)

	// Jaeger UI( http://localhost:16686/search など)を開くとトレースが確認できるはず
}

func tracerProvider(url string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in an Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("libsacloud"),
			attribute.String("version", iaas.Version),
		)),
	)
	return tp, nil
}

func op(ctx context.Context) {
	httpClient := &http.Client{}
	sacloudotel.Initialize()
	httpClient.Transport = otelhttp.NewTransport(http.DefaultTransport)

	caller := api.NewCallerWithOptions(&api.CallerOptions{
		Options: &client.Options{
			AccessToken:       os.Getenv("SAKURACLOUD_ACCESS_TOKEN"),
			AccessTokenSecret: os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET"),
			HttpClient:        httpClient,
		},
	})
	archiveOp := iaas.NewArchiveOp(caller)

	// normal operation
	archiveOp.Find(ctx, "is1a", &iaas.FindCondition{ // nolint
		Count:  1,
		From:   0,
		Filter: ostype.ArchiveCriteria[ostype.Ubuntu],
	})

	// invalid operation(not foundエラーになるはず)
	archiveOp.Read(ctx, "is1a", types.ID(1)) // nolint
}
