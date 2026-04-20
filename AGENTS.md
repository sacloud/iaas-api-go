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
  + verify          # 生成物の退行検査（internal/tools/verify-typespec）
  + format          # tsp format
  + lint            # TypeSpec → OpenAPI (tsp-output/@typespec/openapi3/openapi.yaml)
  + generate:client # OpenAPI → Go クライアント (iaas/client/)
```

`verify` は「特別対応した既知のモデル / envelope が退行していないか」を `spec/typespec/` の生成ファイルに対してチェックする。現状のチェック対象:
- 同一パス相乗り op の envelope merge（Archive / Disk / Server / 他）で期待フィールドが残っているか
- merge により吸収される variant model（`ArchiveCreateBlankRequest` など）が再び emit されていないか
- Switch のレスポンス envelope が `BridgeInfo` に誤解決されないこと（過去のバグの再発防止）

チェックを追加・更新する場合は `internal/tools/verify-typespec/main.go` の `checks` スライスに追記する。

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
| `excluded_ops.go` | 生成対象から除外する DSL op のマップ（通常は空。例外的な除外用の最終手段） |

### 生成物検証（`internal/tools/verify-typespec/`）

`main.go` — 生成された `spec/typespec/resources/*/*.tsp` を読んで、特別対応した model / envelope が期待フィールドを保っているかをチェックする。ビルドパイプラインの `verify` ステップから呼ばれる。

## 設計上の決定と背景

### 共有エンドポイントグループの fat model

**問題**: `appliance` と `commonserviceitem` の 2 エンドポイントは複数リソース（Database/LoadBalancer 等）が同一パスを共有している。当初は TypeSpec の union 型でバリアントを表現したが、OpenAPI の `anyOf` に変換され ogen が "complex anyOf" として未実装扱いにしていた。

**対策**: union をやめ、全バリアントのフィールドを 1 つの model に統合した **fat model** を生成する（`ops.go` の `generateSharedGroupFile`、`fat_model.go`）。

- DSL の mapconv タグ（`"Remark.Plan.ID"`、`"Remark.[]Servers.IPAddress"` 等）を解析してネスト構造を復元する（`fat_model.go`）
- 全バリアントに存在するフィールドは required、一部のみのフィールドは optional
- create 系（POST）の fat model には `Class: string` を必須フィールドとして先頭に追加（API 仕様で必須）
- update 系（PUT）の fat model には `Class` を含めない
- 複数バリアントで同一パスに異なる型が存在する場合は `unknown` にフォールバック
- 全バリアントに存在しても実 API では optional な top-level フィールド（`Icon` / `Plan` / `Disk` / `Settings` / `SettingsHash` など、v1 naked 型でポインタ + omitempty のもの）は `fatModelAlwaysOptionalTop` で強制的に optional にする。そうしないと `Icon: {ID: 0}` を常に送信してしまい API に弾かれる（503 等）。

この変更により `spec/ogen.yml` の `ignore_not_implemented` は完全に除去できた。

### 共有エンドポイントのリクエスト body ラッピング

**問題**: 共有グループ ops（`ApplianceOp` / `CommonServiceItemOp`）は DSL 上 `param: ApplianceCreateRequest` のような引数を持つが、TypeSpec に `@body` decorator 無しで出力すると `{"param": ApplianceCreateRequest}` という wrapper で送信されてしまう。実 API は `{"Appliance": ApplianceCreateRequest}` を期待するため不整合が生じる。

**対策**（`ops.go` の `generateSharedGroupFile`）: primary op の非 path 引数の `MapConvTag` を見て、

- `"<payload>,recursive"` パターン（= `MappableArgument`、Create/Update）→ 専用 envelope モデル（`ApplianceCreateRequestEnvelope { Appliance: ApplianceCreateRequest }` など）を生成して `@body body: EnvelopeName` で渡す。body は `{"Appliance": {...}}` となる。
- `,squash` パターン（= `PassthroughModelArgument`、Find の `FindCondition`、Shutdown の `ShutdownOption` 等）→ wrap せず `@body body: <ArgType>` で直接渡す。body は該当モデルの構造そのまま（`{"Count": 0, "From": 0, "Filter": {...}}` 等）。

### 共有グループ response の代表型（Database）と field optional 化

**問題**: 共有 `ApplianceOp.create` / `read` / `find` の response envelope は、fat model 的に 1 つの代表型を使う必要がある。現在は Database（アルファベット最初）を採用している（`DatabaseCreateResponseEnvelope { Appliance: Database }` 等）。つまり NFS / LoadBalancer / MobileGateway / VPCRouter の Create/Read/Find レスポンスも全て `Database` 型に decode される。

**対策**: Database 固有のフィールド（`InterfaceSettings` / `IPAddresses` / `Disk`）は他 Appliance のレスポンスに含まれないため、`models.go` の `fieldNullabilityOverrides["Database"]` で optional にしている。また `DatabaseSettingCommon.WebUI` / `SourceNetwork` も API が省略することがあるため同様に optional 化。

**Remark.Switch / Remark.Zone の除外**: Appliance レスポンス中 `Remark.Switch.ID` / `Remark.Zone.ID` は `X-Sakura-Bigint-As-Int: 1` ヘッダに反して文字列で返ってくる（top-level の `Appliance.Switch.ID` は int で返るのに齟齬がある）。downstream は top-level 側を使うので、`DatabaseRemark` 等から Switch/Zone を `modelFieldExclusions` で除外している。

### enum デフォルト値の省略

**問題**: enum 型（`types.` プレフィックス）のデフォルト値を TypeSpec で出力すると、OpenAPI で `$ref` + `default` の組み合わせになり ogen が "complex defaults" として未実装扱いにしていた。

**対策**: `models.go` の `convertDefaultValue` で enum 型のデフォルト値は出力しない。代わりに `// Default: EInterfaceDriver.virtio (ogen の complex defaults 未対応のため省略)` というコメントのみ出力する。

### 同一 method+path を共有する複数 op の統合

**背景**: v1 DSL は Go メソッドの使いやすさを優先して「1 つの API エンドポイント」を複数の Go メソッドに分割しているケースがある（例: POST `/archive` の Create と CreateBlank、POST `/disk` の Create / CreateWithConfig / CreateOnDedicatedStorage / CreateOnDedicatedStorageWithConfig、POST `/archive/:sid/to/zone/:did` の Transfer と CreateFromShared、PUT `/archive/:id/ftp` の Share と OpenFTP、PUT `/server/:id/power` の Boot と BootWithVariables、DELETE `/server/:id` の Delete と DeleteWithDisks）。
しかしこれは v1 の overengineering であり、**実 API は同じエンドポイント 1 個**。v2 TypeSpec は実 API を記述する目的なので、これらは 1 オペレーションに統合する。

**真実のソース**: [さくらのクラウドマニュアル](https://manual.sakura.ad.jp/cloud-api/1.1/) が記述するフィールドが API 定義に含まれるべきもの。v1 DSL はあくまで「使われているフィールドの集合」の参考情報。

**対策**:
- **オペレーション名**: `primaryOpForKey`（名前が最短の op）を採用（例: Create/CreateBlank → `create`、Transfer/CreateFromShared → `transfer`、Boot/BootWithVariables → `boot`）。
- **エンベロープ payload の union**（`envelopes.go` の `buildMergedEnvelopeInfos`）: 各 op の request/response payload を payload 名で union。全 op に存在する payload は required、一部のみなら optional。envelope 生成時に primary op を先に visit して payload TS 型名が primary の argument model に解決されるようにしている（`visitOrder`）。
- **request model の union**（`models.go` の `computeRequestModelMerges` / `mergedDSLFields`）: 同 payload 名に異なる DSL request model が割り当てられているケース（例: Archive Create の `ArchiveCreateRequest` と CreateBlank の `ArchiveCreateBlankRequest`）は、primary の model 名に variant の Fields を union し、variant model は `models.tsp` に emit しない。mapconv root が同一のフィールドは重複とみなし先出し定義を採用する（例: `Icon.ID` と `Icon.Name` が別 op にあったら `Icon` 配下にまとめる）。
- **fat model との使い分け**: 共有エンドポイントグループ（appliance / commonserviceitem）は複数リソース横断のため別系統の `fat_model.go` で処理する。ここで述べた統合は単一リソース内の op 相乗りの話。

**例外的に生成から外したい op**: 単純な合流では API 仕様と整合が取れないケース向けに `excluded_ops.go` の `excludedOps` マップを残している。現時点で除外対象は無い。除外した場合は「実装しないエンドポイント」表に理由と共に記載する。

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
   * ただし、全ての op に個別で定義すると煩雑なため、Transport の RoundTrip で定義する。
   * そのようにする必要があることを main.tsp の `@doc` で宣言する。
 * **TypeSpec の `@doc` / `@summary` は日本語で記述し、API 利用者向けの外部ドキュメントとして完結させる**。
   * v2 TypeSpec は単体のドキュメントとして将来公開する想定。
   * Go 実装詳細、ogen・jx 等のライブラリ名、本プロジェクトの内部方針（oneOf 不使用など）、v1 DSL の型名や挙動といった「利用者が知る必要のない事情」は書かない。
   * 「仕様上はこうだが実 API の挙動はこう」という API 利用者に関係する情報は書いてよい（例: Success フィールドの型が複数ありうる説明）。
   * プロジェクト内部の判断理由・背景は AGENTS.md 側に書く。
 * GET の sort/include/exclude は定義しない。ページングとフィルタは定義する。
   * ページングはサーバー側で行った方が良いため。
   * include/exclude は複雑性が高すぎる。
 * oneOf は使用しない
 * 将来増えうる値は enum ではなく string にする
 * リソース単位で tag をつける。
 * **v2 のカバー範囲は「v1 DSL が実装しているもの」に一致させる**。
   * v2 TypeSpec は「実 API の完全な写像」ではなく「主要コンシューマ（terraform-provider-sakuracloud と usacloud）が必要とする範囲」をカバーすることを目指す。
   * 判定の近道として **v1 DSL を真とみなす**。Terraform/usacloud はすでに v1 で動いているため、v1 DSL に存在するエンドポイント・フィールドが「使われている」集合そのものであり、v1 に無いものは使われていないことが保証される。
   * **v1 DSL にないエンドポイントは v2 にも載せない**（具体的な除外リストは下記「実装しないエンドポイント」参照）。
   * **v1 DSL に無いフィールドは v2 TypeSpec に載せない**。API レスポンス JSON に余計なキーが混じっても ogen の decoder は TypeSpec 未定義フィールドを読み飛ばすので decode は失敗しない。
   * **v1 DSL のビューが過剰なフィールドを持ち、ネスト先の API 表現ではそれらが返らず required 違反で decode が失敗するケース** では、軽量ビュー（例: `internetInfoModel` → `InternetInfo`）を別モデルとして定義してネスト先で使わせる。

### HTTP ステータスコード

実 API を検証して判明したステータスコードの実態：

 * POST → **201 Created** がデフォルト。実 API の挙動に応じて **200 OK** / **202 Accepted** にエンドポイント単位で切り替える
 * GET（read/find）→ **200 OK**
 * PUT（update）→ **200 OK**
 * DELETE → **200 OK** with `{is_ok: boolean}`（204 ではない）

`ops.go` のテンプレートでこれらを自動的に生成している。

**POST ステータスコードのエンドポイント単位指定**: 公式マニュアルは 201 Created（同期作成）と 202 Accepted（非同期受付）を POST の正常系として定義しているが、実 API は sub-action 系 POST（例: `POST /internet/:id/ipv6net`）で 200 OK も返す。

TypeSpec で `201 | 202 | 200` のような union にすると ogen が複数の concrete 型（`*XxxOpCreateOK` / `*XxxOpCreateCreated` / `*XxxOpCreateAccepted`）+ interface 戻り値を生成してしまい、クライアント側で type switch が必要になる。これは実 API のシグネチャと乖離しており、API 仕様としても不正確。

そのためステータスコードはエンドポイント毎にひとつだけ指定する方針にしている（ogen は concrete な `*XxxResponseEnvelope` を戻り値に返す）。実装は `ops.go` の `postStatusCodeOverrides` マップで、201 以外を返すエンドポイントだけを解決済みパスで登録する:

```go
var postStatusCodeOverrides = map[string]int{
    "/{zone}/api/cloud/1.1/internet":              202, // 非同期受付
    "/{zone}/api/cloud/1.1/internet/{id}/ipv6net": 200, // sub-action
}
```

未登録の POST は 201 にフォールバックする。新規エンドポイントの統合テストで実 API が 201 以外を返したら decode が失敗するので、そのタイミングで observed コードを上の表に追記する。

OpenAPI YAML を生成後に書き換える（`2XX` にマージする等の）後処理系の解決策は **採用しない** 方針。パイプラインの中間成果物を書き換えると (1) 生成物とソースの対応が追えなくなる、(2) TypeSpec を真のソースとして公開する将来のユースケースと噛み合わない、(3) ツール更新で壊れやすい、といった副作用が大きい。

### レスポンスエンベロープの共通フィールド

公式マニュアルは「インターフェース種別ごとにレスポンス形式が決まっている」と規定しているが、**実 API は仕様より多くのフィールドを返してくる**（特に POST で `Success` も含む）。v2 TypeSpec は downstream コンシューマが実際に利用するフィールドのみに絞る方針。

| インターフェース | 仕様上の形（公式）| 実 API で返るフィールド | v2 TypeSpec の扱い |
|---|---|---|---|
| Find | `{Total, From, Count, <Resource>s}` | 同左 | `Total/From/Count` を常に出力、`<Resource>s` を list で |
| Get（read） | `{is_ok, <Resource>}` | 同左 | `is_ok: boolean`, `<Resource>` |
| POST（create） | `{is_ok, ...}` | `{is_ok, Success, <Resource>}` が実測で返る | `is_ok: boolean`, `<Resource>`（`Success` は未定義） |
| PUT（update） | `{Success}` | `{is_ok, Success, <Resource>}` が実測で返る | `is_ok: boolean`, `<Resource>`（`Success` は未定義） |
| DELETE | `{Success, is_ok}` | 同左（`is_ok` のみの op もあり） | `{is_ok: boolean}` のみ |

**Success フィールドを TypeSpec から除外している件**:
実 API はほぼ全オペレーションで `Success` を返すが、型が不安定（boolean のほか `"Created"` / `"Accepted"` などの文字列も返る — 公式仕様は boolean 限定と書きつつ実態は異なる）。一方で terraform-provider-sakuracloud（v2）・terraform-provider-sakura（v3）・usacloud を全 grep した結果、`Success` を読んでいる箇所は 0 件。v1 の `zz_envelopes.go` も定義だけで reader は無く、`types.APIResult` も `UnmarshalJSON` を実装しているが呼び元なし。
よって「downstream が利用しないフィールドは TypeSpec に載せない」ルール（上記）に従い v2 TypeSpec では `Success` を定義していない。実 API のレスポンス JSON に `Success` キーが含まれていても ogen の decoder は未知フィールドを読み飛ばすため decode は失敗しない。成功判定は `is_ok` を用いる。

### リクエストエンベロープの型

create/update のリクエストエンベロープは、レスポンス用のビュー型（`Icon` 等）ではなく、
オペレーション固有のリクエスト型（`IconCreateRequest`、`IconUpdateRequest` 等）を使う。
`envelopes.go` でオペレーションの Arguments から正しいモデル名を解決している。

### 実装しないエンドポイント

v1 DSL で未実装かつ Terraform/usacloud でも利用されていないため、v2 でも実装しないと決定したエンドポイント一覧。
新しいリソースのインテグレーションテストを書く際は、公式マニュアルと照合してここに漏れがないか確認し、未実装と決定したものは追記すること。

**共通パターン**: `GET /<resource>/tag`（タグ一覧）および `GET /<resource>/:id/tag`（個別リソースのタグ取得）は v1 全体で未実装であり、v2 でも実装しない。下表では個別に列挙する。

| リソース | メソッド・パス | 概要 | 備考 |
|---|---|---|---|
| Switch | GET `/switch/:id/appliance` | 接続中アプライアンス一覧 | v1 未実装 |
| Switch | GET `/switch/:id/tag` | スイッチのタグ取得 | v1 未実装（タグ系共通パターン） |
| Switch | GET `/switch/tag` | スイッチタグ一覧 | v1 未実装（タグ系共通パターン） |
| Icon | GET `/icon/:id?Size=[small\|medium\|large]` | アイコン画像データ取得 | v1 がサポート外と明言（`internal/define/icon.go` コメント） |
| Icon | GET `/icon/:id/tag` | アイコンのタグ取得 | v1 未実装（タグ系共通パターン） |
| Icon | GET `/icon/tag` | アイコンタグ一覧 | v1 未実装（タグ系共通パターン） |
| Note | GET `/note/:id/config` | スクリプトをテンプレートとして利用 | v1 未実装 |
| Note | GET `/note/:id/tag` | スクリプトのタグ取得 | v1 未実装（タグ系共通パターン） |
| Note | GET `/note/tag` | スクリプトタグ一覧 | v1 未実装（タグ系共通パターン） |
| Disk | PUT `/disk/:id/plan` | ディスクプラン変更 | v1 未実装 |
| Disk | GET `/disk/:id/tag` | ディスクのタグ取得 | v1 未実装（タグ系共通パターン） |
| Disk | GET `/disk/tag` | ディスクタグ一覧 | v1 未実装（タグ系共通パターン） |
| Server | GET `/server/:id/cdrom` | サーバに挿入された ISO 状態取得 | v1 未実装（挿入/排出のみ実装） |
| Server | GET `/server/:id/interface` | サーバ接続 NIC 一覧取得 | v1 未実装（Server view の `Interfaces` で取得可） |
| Server | PUT `/server/:id/mouse/:mouseindex` | マウス操作 | v1 未実装 |
| Server | GET `/server/:id/power` | 起動状態取得 | v1 未実装（Server view の `Instance.Status` で取得可） |
| Server | GET `/server/:id/tag` | サーバのタグ取得 | v1 未実装（タグ系共通パターン） |
| Server | PUT `/server/:id/to/plan/:planid` | プラン変更（alt path） | v1 未実装（`PUT /server/:id/plan` 側を使用） |
| Server | GET `/server/:id/vnc/size` | VNC 画面サイズ取得 | v1 未実装 |
| Server | GET `/server/:id/vnc/snapshot` | VNC スナップショット取得 | v1 未実装 |
| Server | GET `/server/tag` | サーバタグ一覧 | v1 未実装（タグ系共通パターン） |
| Archive | GET `/archive/:id/tag` | アーカイブのタグ取得 | v1 未実装（タグ系共通パターン） |
| Archive | GET `/archive/tag` | アーカイブタグ一覧 | v1 未実装（タグ系共通パターン） |
| Internet | GET `/internet/:id/tag` | ルータのタグ取得 | v1 未実装（タグ系共通パターン） |
| Internet | GET `/internet/tag` | ルータタグ一覧 | v1 未実装（タグ系共通パターン） |
| CDROM | GET `/cdrom/:id/tag` | ISO イメージのタグ取得 | v1 未実装（タグ系共通パターン） |
| CDROM | GET `/cdrom/tag` | ISO タグ一覧 | v1 未実装（タグ系共通パターン） |
| CDROM | GET `/zone/:zoneid/cdrom` | ゾーン内 CDROM 一覧（alt path） | v1 未実装（`GET /{zone}/.../cdrom` 側で取得可） |
| ServiceClass | GET `/public/price` | 価格表（ServiceClass 一覧） | v1 DSL は実装しているが v2 でまだ未対応。理由: (1) 実 API レスポンスの `ServiceClassID` JSON キーに対応する v2 の JSON name remapping 機構がまだ無い（v1 は naked 型の `json:"ServiceClassID"` で吸収）、(2) `Price` フィールドが `{}` か `[]` の polymorphic（v1 は `UnmarshalJSON` で吸収）。downstream は usacloud の list 表示のみなので優先度低。対応時は ServiceClass.ID の JSON name と Price の型を `unknown` に差し替える拡張が必要 |
| AuthStatus | GET `/auth-status` | 認証状態を取得 | v1 DSL / v2 TypeSpec は `{is_ok, AuthStatus: {...}}` で wrap されたレスポンスを想定しているが、実 API は AuthStatus のフィールドが envelope 直下にフラットで返る（`{is_ok, AuthMethod, AuthClass, Account: {...}, Member: {...}, ...}`）。v1 も同じ envelope 定義で `AuthStatus` フィールドは常に nil になっているが、v1 テストは `NotNil` assert だけなので気付かれていない。v2 でまともに decode するには envelope 生成時に単一 payload を wrap せず inline する対応が必要。downstream は usacloud の authstatus read 表示のみなので優先度低 |

## v2 テスト方針

v2 のテストは **実 API を使用した統合テストのみ** を実施する。

- ogen 生成されるクライアントコードの品質は ogen に担保してもらう（単体テスト不要）
- fake モックは作成しない（メンテナンス不可が増加するため）
- テストディレクトリ: `v2/integration/`
- 全リソースの CRUD 操作をテストする
- 最初のテスト対象は Icon（軽量・安全なリソース）から開始

### テスト設計時の参照先

新しいリソースのインテグレーションテストを書く前に、以下の2つを必ず参照する：

1. **さくらのクラウド API マニュアル** — そのリソースで提供されている全エンドポイントの把握に使う。
   例: Switch なら https://manual.sakura.ad.jp/cloud-api/1.1/switch/index.html
2. **terraform-provider-sakuracloud のドキュメント** — Terraform が実際に利用しているフィールド・オペレーションの把握に使う。テスト対象の要否判断（v1 DSL との照合）に使う。
   例: Switch なら https://registry.terraform.io/providers/sacloud/sakura/latest/docs/resources/switch

URL の `switch` 部分を対応するリソース名に置き換えれば、他リソースのマニュアル/docs も同じ構造で辿れる。

### テストカバレッジ監査

さくらのクラウド API マニュアル（https://manual.sakura.ad.jp/cloud-api/1.1/index.html）のサイドバー全ページに対する v2 インテグレーションテストの網羅状況。

**カバー済み（v2/integration/ にテスト実装済）**

| マニュアルページ | リソース | テストファイル |
|---|---|---|
| server | Server | `server_test.go` |
| disk | Disk | `disk_test.go` |
| switch | Switch, Switch-Bridge | `switch_test.go`, `bridge_test.go` |
| archive | Archive | `archive_test.go` |
| cdrom | CDROM | `cdrom_test.go` |
| bridge | Bridge | `bridge_test.go` |
| internet | Internet（Router）| `internet_test.go` |
| interface | Interface | `interface_test.go` |
| appliance | NFS / Database / LoadBalancer / VPCRouter / MobileGateway | `nfs_test.go`, `database_appliance_test.go`, `load_balancer_test.go`, `vpc_router_test.go`, `mobile_gateway_test.go`（MG は SIM 契約環境で skip） |
| icon | Icon | `icon_test.go` |
| note | Note（Script）| `note_test.go` |
| sshkey | SSHKey | `ssh_key_test.go` |
| facility | Region, Zone | `facility_test.go` |
| product | DiskPlan, InternetPlan, ServerPlan, LicenseInfo, PrivateHostPlan | `product_test.go`, `private_host_test.go` |
| account | License | `account_test.go` |

**テスト未実装（downstream 使用実績あり — 今後実装候補）**

サイドバーに独立ページは無いが、他ページからリンクされている or v1 DSL にあるリソース：

| リソース | マニュアル所在 | downstream 利用 | 優先度の目安 |
|---|---|---|---|
| DNS | commonserviceitem（appliance 系）| terraform-sakura / usacloud | 高 |
| GSLB | commonserviceitem | terraform-sakura / usacloud | 高 |
| ProxyLB（EnhancedLB）| commonserviceitem | terraform-sakura (enhanced_lb) / usacloud | 高 |
| PacketFilter | interface ページ | terraform-sakura / usacloud | 高 |
| SimpleMonitor | commonserviceitem | terraform-sakura / usacloud | 高 |
| LocalRouter | commonserviceitem | terraform-sakura / usacloud | 高 |
| EnhancedDB | commonserviceitem | terraform-sakura / usacloud | 中 |
| AutoBackup | commonserviceitem | terraform-sakura / usacloud | 中 |
| AutoScale | commonserviceitem | terraform-sakura / usacloud | 中 |
| ContainerRegistry | commonserviceitem | terraform-sakura / usacloud | 中 |
| Subnet | internet ページ | terraform-sakura / usacloud | 中 |
| IPAddress | internet ページ | terraform-sakura (ipv4_ptr) / usacloud | 中 |
| CertificateAuthority | commonserviceitem | usacloud / sakuracloud(old) | 中 |
| SimpleNotification（Group / Destination）| commonserviceitem | terraform-sakura | 中 |
| SIM | commonserviceitem | usacloud / sakuracloud(old) | 低（法人 SIM 契約必須）|
| ESME | commonserviceitem | usacloud | 低 |
| IPv6Addr | internet ページ | usacloud | 低 |
| IPv6Net | internet ページ | usacloud | 低 |
| Bill | bill ページ | usacloud | 低（課金情報、書き込み不可の read-only）|
| Coupon | bill ページ | usacloud | 低 |

**v2 ジェネレータ未対応として意図的にスキップ**

| リソース | 理由 | 状況 |
|---|---|---|
| AuthStatus | 実 API のレスポンスがフラット構造（envelope 直下にフィールド展開）で v2 TypeSpec の wrap envelope と不整合 | 「実装しないエンドポイント」表を参照 |
| ServiceClass | `ID` ↔ `ServiceClassID` の JSON name remapping と `Price` の polymorphic（`{}` / `[]`）を v2 ジェネレータが扱えない | 「実装しないエンドポイント」表を参照 |

**マニュアルページ単位の扱い（一覧）**

| サイドバー項目 | 状況 |
|---|---|
| はじめに | N/A（ドキュメント） |
| サーバ関連の API | ✓ カバー |
| ディスク関連の API | ✓ カバー |
| スイッチ関連の API | ✓ カバー |
| アーカイブ関連の API | ✓ カバー |
| ISO イメージ関連の API | ✓ カバー |
| ブリッジ関連の API | ✓ カバー |
| ルータ関連の API | ✓（Internet カバー。同ページの Subnet / IPAddress / IPv6Addr / IPv6Net は未実装） |
| インタフェース関連の API | ✓（Interface カバー。同ページの PacketFilter は未実装） |
| アプライアンス関連の API | ✓（NFS/DB/LB/VPCR カバー、MG は SIM env で skip。commonserviceitem 系は個別で未実装多数） |
| アイコン関連の API | ✓ カバー |
| スクリプト関連の API | ✓ カバー |
| SSH キー関連の API | ✓ カバー |
| 設備関連の API | ✓ カバー |
| 商品関連の API | ✓ カバー（`/public/price` ServiceClass は v2 非対応） |
| ユーザ・プロジェクト関連の API | △ License はカバー、AuthStatus は非対応 |
| 請求関連の API | ❌ Bill / Coupon 共に未実装（usacloud read-only のみ利用） |
| 列挙型一覧 | N/A（型定義） |
| 変更履歴 | N/A（changelog） |

### テスト保留中のエンドポイント（複数リソース連携が必要）

単一リソースの CRUD だけでは検証できず、他リソースと組み合わせた setup/teardown が必要なエンドポイントは、
**該当リソース側のインテグレーションテストが揃ってから** 追加する方針で、いったん保留する。
（例: Switch-Bridge の connect/disconnect は Bridge 側のテストが整ってから書く。）

TypeSpec / ogen クライアントには定義済みなのでコードは存在するが、`v2/integration/` にテストが無い状態。
各リソースのテスト整備が進み次第、この表から消していく。

| テスト対象 | 必要な前提リソース | 関連エンドポイント |
|---|---|---|
| Switch の接続サーバ一覧 | Server | GET `/switch/:id/server` |
| Disk ↔ Server 接続/切断 | Server | PUT `/disk/:id/to/server/:serverID`、DELETE `/disk/:id/to/server` |
| Disk の config（OS 初期設定） | Archive（SourceArchive 参照） | PUT `/disk/:id/config` |
| Disk の resize-partition | Archive（OS 入りディスクが前提） | PUT `/disk/:id/resize-partition` |
| Disk の monitor | 使用実績が必要（空ディスクではデータ取得不可） | GET `/disk/:id/monitor` |
| Server の power 操作 (boot/shutdown/reset) | Disk（OS 入りディスク接続が前提） | PUT `/server/:id/power`、DELETE `/server/:id/power`、PUT `/server/:id/reset` |
| Server の boot with variables (cloud-init) | Disk（OS 入りディスク接続が前提） | PUT `/server/:id/power`（`UserBootVariables` 付き） |
| Server への CDROM 挿入・排出 | CDROM | PUT `/server/:id/cdrom`、DELETE `/server/:id/cdrom` |
| Server の sendKey / sendNMI | 起動中サーバ（Disk が必要） | PUT `/server/:id/keyboard`、PUT `/server/:id/qemu/nmi` |
| Server の VNC プロキシ取得 | 起動中サーバ（Disk が必要） | GET `/server/:id/vnc/proxy` |
| Server の changePlan | 停止中サーバ（Disk 接続状態で制約あり） | PUT `/server/:id/plan` |
| Server の monitor | 使用実績が必要 | GET `/server/:id/monitor` |
| Archive の FTP 共有（Share/OpenFTP） | FTP クライアント | PUT `/archive/:id/ftp` |
| Archive の他ゾーン転送（Transfer / CreateFromShared） | 別ゾーンの既存 Archive + `SharedKey` の取得フロー | POST `/archive/:sourceArchiveID/to/zone/:destZoneID` |

### サンドボックス（tk1v）で動かないテスト

一部のリソースは sandbox `tk1v` に Plan が無いなどで動作しないため、v1 テストに倣って本番ゾーンをハードコードする。TEST_ACC 系テストはさくらインターネット社員が流す前提で、課金は問題にしていない。

| テスト | 固定ゾーン | 理由 |
|---|---|---|
| `TestPrivateHostPlanFind` / `TestPrivateHostCRUD` | `tk1a` | `tk1v` に PrivateHostPlan が存在しない。v1 `test/private_host_op_test.go` の `privateHostTestZone = "tk1a"` と同等 |
| `TestLicenseCRUD` | `tk1a` | `tk1v` で Create が `dont_create_in_sandbox` (403) を返す。License は課金対象かつサンドボックス禁止のため本番ゾーンで実行 |
| `TestBridgeCRUD` / `TestSwitchBridgeConnect` | `tk1a` | Bridge は `tk1v` で Create が `dont_create_in_sandbox` (403) を返すため本番ゾーン固定。`TestSwitchBridgeConnect` は tk1a の per-zone switch quota（アカウントにより 1）に既存 switch があると 409 で skip する（保険として `isLimitCountError` でガード） |

新たに sandbox で動かないリソースのテストを追加する際は、v1 側の既存テストのゾーン指定と合わせて定数化する。

### 契約・環境依存で実動作確認ができていないテスト

以下のテストは envelope / mapconv の生成結果としては v1 と同形の JSON を送り出すところまで実装・検証済みだが、契約面の制約により実 API との突合せ（Create 以降の CRUD）までは走らせられておらず **未検証** 扱い。フル CRUD が必要になった場合は下表の環境要件を満たした上で手動実行する。

| テスト | 未検証な CRUD の範囲 | 必要な環境 |
|---|---|---|
| `TestMobileGatewayApplianceCRUD` | Create 含めて全て | 法人契約 + SIM 契約。`SAKURACLOUD_SIM_ICCID` / `SAKURACLOUD_SIM_PASSCODE` が無いと skip する（v1 `test/mobile_gateway_op_test.go` と同じゲート） |

### 環境設定

テスト実行には以下の環境変数が必要：
- `SAKURA_ACCESS_TOKEN`
- `SAKURA_ACCESS_TOKEN_SECRET`
- `SAKURA_ZONE`（デフォルト: tk1v）

### テスト実行

```bash
cd v2
TEST_ACC=1 go test -v ./integration/...
```


