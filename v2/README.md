# iaas-api-go v2 (work in progress)

> [!IMPORTANT]
> 現在、v2 は開発中です。予告なくインターフェースが変更されることがあります。v1 とは互換性がありません。

さくらのクラウド IaaS API Go言語向け APIライブラリ v2

マニュアル: https://manual.sakura.ad.jp/cloud/


## 概要
iaas-api-go v2 はさくらのクラウド IaaS API を Go 言語から利用するための API ライブラリです。
v1 では Go の DSL 定義からクライアントを生成していましたが、v2 では TypeSpec / OpenAPI 定義から
ogen で自動生成する方式に切り替えています。

使用感は [sacloud/simple-notification-api-go](https://github.com/sacloud/simple-notification-api-go) に揃えており、
[saclient-go](https://github.com/sacloud/saclient-go)（認証・プロファイル解決・ヘッダ付与・リトライ等の共通土台）
の上にリソースごとの Op インターフェースが載る薄いラッパー構成です。

## 利用イメージ

### スクリプト (Note) の作成と検索
```go
package main

import (
	"context"
	"fmt"
	"os"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/saclient-go"
)

func main() {
	// デフォルトで usacloud 互換プロファイル or 環境変数
	// (SAKURA_ACCESS_TOKEN{_SECRET} / SAKURA_ZONE 等) が利用される
	var sc saclient.Client
	if err := sc.SetEnviron(os.Environ()); err != nil {
		panic(err)
	}
	ctx := context.Background()

	c, err := iaas.NewClient(&sc)
	if err != nil {
		panic(err)
	}

	noteOp := iaas.NewNoteOp(c)
	zone := "tk1v"

	// Note を作成
	createResp, err := noteOp.Create(ctx, zone, &client.NoteCreateRequestEnvelope{
		Note: client.NoteCreateRequest{
			Name:    client.NewOptNilString("my-note"),
			Class:   client.NewOptNilString("shell"),
			Content: client.NewOptNilString("#!/bin/bash\necho hello"),
			Tags:    []string{"sample"},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created Note ID: %d\n", createResp.Note.ID.Value)

	// Note を検索
	findResp, err := noteOp.List(ctx, zone, &client.NoteFindRequest{
		Count: 50,
		Filter: client.NoteFindFilter{
			Name: "my-note",
		},
	})
	if err != nil {
		panic(err)
	}
	for _, n := range findResp.Notes {
		fmt.Printf("%d: %s\n", n.ID.Value, n.Name.Value)
	}
}
```

### サーバーの検索と電源投入
```go
package main

import (
	"context"
	"fmt"
	"os"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/saclient-go"
)

func main() {
	var sc saclient.Client
	if err := sc.SetEnviron(os.Environ()); err != nil {
		panic(err)
	}
	ctx := context.Background()

	c, err := iaas.NewClient(&sc)
	if err != nil {
		panic(err)
	}

	serverOp := iaas.NewServerOp(c)
	zone := "tk1v"

	// Name でサーバーを検索
	findResp, err := serverOp.List(ctx, zone, &client.ServerFindRequest{
		Count: 10,
		Filter: client.ServerFindFilter{
			Name: "my-server",
		},
	})
	if err != nil {
		panic(err)
	}

	// マッチしたサーバーを順に起動
	for _, s := range findResp.Servers {
		id := fmt.Sprintf("%d", s.ID.Value)
		if err := serverOp.Boot(ctx, zone, id, &client.ServerBootRequestEnvelope{}); err != nil {
			panic(err)
		}
		fmt.Printf("Booted: %d (%s)\n", s.ID.Value, s.Name.Value)
	}
}
```

⚠️ v1.0 に達するまでは互換性のない形で変更される可能性がありますのでご注意ください。


## v2 での主な変更点
- 生成パイプラインを Go DSL 直生成から **TypeSpec → OpenAPI → ogen** に移行
- ラッパー層（`v2/*_op_gen.go`）は `internal/tools/gen-v2-op/` から自動生成
- 認証・リトライ・ヘッダ付与等の土台は `saclient-go` に一元化
- 詳細は [`AGENTS.md`](../AGENTS.md) を参照


## License

`iaas-api-go` Copyright (C) 2022- The sacloud/iaas-api-go authors.
This project is published under [Apache 2.0 License](../LICENSE).
