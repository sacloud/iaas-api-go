# iaas-api-go

さくらのクラウドの IaaS API クライアント。

現在、v2 を開発中。

## v1 と v2 の違い

v1 では Go の DSL 定義からクライアントを生成していた。

しかし、他の API Client と同様に openapi 定義から生成されるようにしていきたい。
また、openapi 定義自体を、ドキュメントサイトで公開していくようにしていきたい。

そのために、v2 では以下のフローへ移行を行う：

```
OpenAPI 定義
  → Go クライアント生成 (ogen)
```

そのための前段階として、Go DSL から TypeSpec ファイルを生成するところから初めている。
typespec 定義が十分なクオリティに達した段階で、このディレクトリの外に openapi 定義の管理は移行する想定。

```
Go DSL 定義
  → TypeSpec 生成 (gen-typespec)
  → OpenAPI 生成 (tsp compile)
  → Go クライアント生成 (ogen)
```

`cd spec && pnpm run build` で全工程が実行される。

## バージョン運用方針（v1 / v2）

当面は **同一 main branch で v1 と v2 を併存**させて開発する。

- v1: 従来の Go DSL 直生成系（既存利用者向けの互換維持）
- v2: TypeSpec/OpenAPI/ogen 経由の新系統（非互換変更の受け皿）
- 生成物の出力先、ビルドスクリプト、CI ジョブは v1/v2 で分離して管理する

将来、保守負荷やリリース運用の都合で必要になれば、v1 maintenance ブランチを切って分離する。
ただし、Go の import path / module path は途中で揺らさず、メジャーバージョンごとの規約を維持する。

[Go Modules: v2 and Beyond](https://go.dev/blog/v2-go-modules) を参考にすること。

## 生成フロー

Go の DSL 定義から TypeSpec を生成し、OpenAPI 経由で Go クライアントを生成する。

```
Go DSL 定義
  → TypeSpec 生成 (gen-typespec)
  → OpenAPI 生成 (tsp compile)
  → Go クライアント生成 (ogen)
```

`cd spec && pnpm run build` で全工程が実行される。

```
pnpm run build
  = generate        # DSL → TypeSpec ファイル生成
  + format          # tsp format
  + lint            # TypeSpec → OpenAPI (tsp-output/@typespec/openapi3/openapi.yaml)
  + generate:client # OpenAPI → Go クライアント (iaas/client/)
```

## TypeSpec のファイル構成

### 手動作成ファイル（spec/src/）

* `spec/src/main.tsp` - TypeSpec のエントリーポイント
  * `resources.tsp` と `types.tsp` をインポートする

### 自動生成ファイル（spec/typespec/）

```
spec/typespec/
  resources/
    {resource_name}/      # 例: archive/, dns/, server/ など
      models.tsp          # モデル定義
      ops.tsp             # オペレーション定義（単一リソース）
      envelopes.tsp       # リクエスト/レスポンスエンベロープ
      results.tsp         # レスポンス結果型
    appliance/            # 共有グループ（Database, LoadBalancer, MobileGateway, NFS, VPCRouter）
      ops.tsp
    common_service_item/  # 共有グループ（AutoBackup, AutoScale, DNS, GSLB など）
      ops.tsp
  resources.tsp           # 全リソースファイルの import をまとめた生成ファイル
  types.tsp               # enum 定義（gen-typespec が types/ パッケージを AST 解析して生成）
  main.tsp                # spec/src/main.tsp のコピー（copy:main で配置）
```

### ジェネレータ（internal/tools/gen-typespec/）

| ファイル | 役割 |
|---|---|
| `main.go` | エントリポイント。各生成関数を順番に呼び出す |
| `util.go` | 共通ユーティリティ（repoRoot, absPath, writeFile, lowerFirst 等） |
| `models.go` | `resources/{name}/models.tsp` を生成 |
| `ops.go` | `resources/{name}/ops.tsp` を生成（共有グループは `resources/{group}/ops.tsp`） |
| `fat_model.go` | 共有グループの fat model をツリー構造で構築する |
| `envelopes.go` | `resources/{name}/envelopes.tsp` を生成 |
| `results.go` | `resources/{name}/results.tsp` を生成 |
| `types.go` | `types/` パッケージを AST 解析して `spec/typespec/types.tsp`（enum 定義）を生成 |

## 設計上の決定と背景

### 共有エンドポイントグループの fat model

**問題**: `appliance` と `commonserviceitem` の 2 エンドポイントは複数リソース（Database/LoadBalancer 等）が同一パスを共有している。当初は TypeSpec の union 型でバリアントを表現したが、OpenAPI の `anyOf` に変換され ogen が "complex anyOf" として未実装扱いにしていた。

**対策**: union をやめ、全バリアントのフィールドを 1 つの model に統合した **fat model** を生成する（`ops.go` の `generateSharedGroupFile`、`fat_model.go`）。

- DSL の mapconv タグ（`"Remark.Plan.ID"`、`"Remark.[]Servers.IPAddress"` 等）を解析してネスト構造を復元する（`fat_model.go`）
- 全バリアントに存在するフィールドは required、一部のみのフィールドは optional
- create 系（POST）の fat model には `Class: string` を必須フィールドとして先頭に追加（API 仕様で必須）
- update 系（PUT）の fat model には `Class` を含めない
- 複数バリアントで同一パスに異なる型が存在する場合は `unknown` にフォールバック

この変更により `spec/ogen.yml` の `ignore_not_implemented` は完全に除去できた。

### enum デフォルト値の省略

**問題**: enum 型（`types.` プレフィックス）のデフォルト値を TypeSpec で出力すると、OpenAPI で `$ref` + `default` の組み合わせになり ogen が "complex defaults" として未実装扱いにしていた。

**対策**: `models.go` の `convertDefaultValue` で enum 型のデフォルト値は出力しない。代わりに `// Default: EInterfaceDriver.virtio (ogen の complex defaults 未対応のため省略)` というコメントのみ出力する。

### convenient errors

**問題**: ogen の convenient errors（エラーハンドリングの簡略化）を使うには全オペレーションにエラーレスポンスが必要。

**対策**: `spec/src/main.tsp` に `@error model ApiError` を定義し、全オペレーションの戻り値を `ReturnType | ApiError` にしている（`ops.go` のテンプレート）。ogen が `default` レスポンスを認識して convenient errors を生成する。

## ogen "Type is not defined, using any" 警告について

`pnpm run generate:client` 実行時に残る警告はすべて意図的。ogen がその型を `jx.Raw`（any）で生成するが対処不要。

### モニターデータ系（13件）

`DiskMonitorResponseEnvelopeData`、`InterfaceMonitorResponseEnvelopeData` 等すべての `*MonitorResponseEnvelopeData`。

**理由**: モニター API の `Data` フィールドはキーがタイムスタンプ文字列・値がメトリクス名をキーにした動的構造（`naked.MonitorValues` が `UnmarshalJSON` でカスタムデシリアライズ）のため、静的な TypeSpec モデルに変換できない。`envelopes.go` の `nakedTypeToTSName` で `"MonitorValues": "unknown"` と明示している。

### fat model の型競合（3件）

- `ApplianceCreateRequestSettings` / `ApplianceUpdateRequestSettings`:
  Database（`Settings` 配下に中間ノード）vs VPCRouter（`Settings: VPCRouterSetting` の葉ノード）
- `ApplianceCreateRequestRemarkSwitch`:
  Database（`Remark.Switch.ID` スカラー）vs VPCRouter（`Remark.Switch: ApplianceConnectedSwitch` オブジェクト）
- `CommonServiceItemCreateRequestStatusRegion`:
  ProxyLB（`Status.Region: EProxyLBRegion`）vs EnhancedDB（`Status.Region: EnhancedDBRegion`）

いずれも `fat_model.go` のツリーマージで同パスに異なる型が存在するため `unknown` になる正常動作。

### 動的 map 型（2件）

- `DatabaseParameterSettingsItem`: `DatabaseParameter.Settings` が `map[string]interface{}` → `Record<unknown>`
- `DiskEditNoteVariablesItem`: `DiskEditNote.Variables` が `map[string]interface{}` → `Record<unknown>`

`Record<unknown>` の値型が未定義のため警告が出るが、DSL 定義通り動的マップであり `unknown` が正しい。


## openapi 定義規約

 * 全ての API 呼び出しには `X-Sakura-Bigint-As-Int: 1` をつける。
 * GET の sort/include/exclude は定義しない。ページングとフィルタは定義する。
   * ページングはサーバー側で行った方が良いため。
   * include/exclude は複雑性が高すぎる。
 * oneOf は使用しない
 * 将来増えうる値は enum ではなく string にする
 * リソース単位で tag をつける。

