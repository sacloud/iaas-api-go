# v2 設計方針

`iaas-api-go` v2 の設計判断・方針を 1 枚にまとめる。詳細な背景・運用手順・除外リスト等は
[AGENTS.md](../AGENTS.md) が引き続き一次ソース、既知課題は [docs/v2-issues.md](./v2-issues.md)
を参照する。本書は「v2 がどういう思想で作られているか」を把握したい人向けの俯瞰資料。

移行者向けの手順は [v2-migration-guide.md](./v2-migration-guide.md) を参照。

## 目次

1. [v2 の目的とスコープ](#1-v2-の目的とスコープ)
2. [生成パイプライン](#2-生成パイプライン)
3. [設計原則](#3-設計原則)
4. [主要な設計判断](#4-主要な設計判断)
5. [命名規則](#5-命名規則)
6. [API 規約](#6-api-規約)
7. [ラッパー層の構造](#7-ラッパー層の構造-v2_genovclientgovmiddlewarego)
8. [Error 型](#8-error-型)
9. [helper 層](#9-helper-層-v2helperpower-query-wait-plans-cleanup)
10. [テスト方針](#10-テスト方針)
11. [バージョン運用方針](#11-バージョン運用方針)
12. [将来像](#12-将来像)

---

## 1. v2 の目的とスコープ

- **目的**: クライアント生成を Go DSL 直生成から **OpenAPI → ogen** へ切り替え、OpenAPI 定義
  そのものをドキュメントサイトで公開できる状態にする。
- **中間フェーズ**: v1 DSL を真とみなして TypeSpec を生成する。TypeSpec が十分なクオリティに
  達した段階で DSL を廃止し、TypeSpec を手書き source of truth に昇格させる予定。
- **カバー範囲**: 実 API の完全写像ではなく、主要コンシューマ
  (`terraform-provider-sakuracloud`, `terraform-provider-sakura`, `usacloud`) が必要とする
  範囲に揃える。v1 DSL に無いエンドポイント／フィールドは v2 に載せない。
- **非スコープ**: 列挙型の日本語説明、認証方式の解説、運用ガイド等のドキュメント読み物は
  TypeSpec に含めない（OpenAPI/ogen の生成に寄与しないため）。

## 2. 生成パイプライン

```
Go DSL 定義
  → TypeSpec 生成  (internal/tools/gen-typespec)
      ※ fieldmanifest allowlist / excludedOps で downstream 未使用を pruning
  → OpenAPI 生成  (tsp compile)
  → Go クライアント生成  (ogen → v2/client/)
  → v2 ラッパー生成  (internal/tools/gen-v2-op → v2/<resource>_gen.go)
```

`cd spec && pnpm run build` で TypeSpec → OpenAPI → Go クライアントまで通る
(`generate` / `verify` / `format` / `lint` / `generate:client` / `generate:find-request`)。
`v2/<resource>_gen.go` の生成は別途 `go run ./internal/tools/gen-v2-op/`。

**重要ルール**: OpenAPI YAML を後処理で書き換えるステップは置かない。理由は (1) 生成物と
ソースの対応が追えなくなる、(2) TypeSpec を source of truth として公開する将来像と噛み合わない、
(3) ツール更新で壊れやすい。ステータスコードやエンベロープの齟齬は TypeSpec 側 (`ops.go` や
override マップ) で解決する。

### 生成物検証 (`verify-typespec`)

`internal/tools/verify-typespec/` が `spec/typespec/` を読み、以下の退行を検出する:

- 同一パス相乗り op の envelope merge が期待フィールドを維持しているか
- merge で吸収されるはずの variant model (`ArchiveCreateBlankRequest` 等) が再 emit されていないか
- Switch のレスポンス envelope が `BridgeInfo` に誤解決されていないか
- `fieldNullabilityOverrides` で nullable 化した submodel が退行していないか
  (`skipIfModelAbsent: true` ならモデル不在時はスキップ)

## 3. 設計原則

### 3.1 downstream 参照基準で絞る

v1 DSL は「過去に実装した全フィールド／全 op」の集合。v2 はそこから **実利用されているもの**
だけを emit する 2 層構成の pruning を入れる。

| レイヤー | ファイル | キー | 値 |
|---|---|---|---|
| フィールド allowlist | `internal/tools/fieldmanifest/manifest.go` | TypeSpec モデル名 | emit するフィールド名セット |
| op 除外 | `internal/tools/gen-typespec/excluded_ops.go` | API 名 | 除外する op 名セット |

- 未登録モデルは pass-through (全フィールド emit)。段階導入を許す
- 除外対象: `usacloud` / `terraform-provider-sakuracloud` / `terraform-provider-sakura` /
  `iaas-service-go` (廃止予定、3 downstream 経由分のみ) / `v2/integration` テストが
  参照しないもの
- 除外されたフィールドは `docs/typespec/excluded-fields.md` に自動レポート

### 3.2 v1 DSL を真とみなす

Terraform/usacloud は v1 で動いているので、v1 DSL = 実利用されている集合と見なせる。

- v1 に無いエンドポイントは v2 にも載せない (具体リストは AGENTS.md「実装しないエンドポイント」)
- v1 に無いフィールドは v2 TypeSpec に載せない (余計なキーは ogen が読み飛ばす)
- 例外: v1 のビューが過剰フィールドを持ちネスト先で required 違反になるケースは軽量ビュー
  モデルを別定義 (例: `InternetInfo`)

### 3.3 実 API の挙動を優先、公式仕様は参考

- POST のステータスコードは 201 デフォルト、200/202 は `postStatusCodeOverrides` で個別指定
- レスポンスの `Success` フィールドは型が不安定 (boolean / 文字列混在) かつ downstream 参照 0 件
  なので TypeSpec に載せない。成功判定は `is_ok` を使う
- `Remark.Switch.ID` / `Remark.Zone.ID` のように BigInt ヘッダに反して文字列で返るフィールドは
  top-level 側を使い、二重露出するフィールドは `modelFieldExclusions` で除外

### 3.4 Optional (`?`) と Nullable (`| null`) の分離

TypeSpec で別概念として扱い、`models.go` の `nakedFieldIsNullable` は **ポインタ型 `*T` のとき
のみ true** を返す。`omitempty` は optional (`?`) シグナルにのみ使う。

| v1 naked 宣言 | TypeSpec | ogen 出力 | 意味 |
|---|---|---|---|
| `*T` | `Foo?: T \| null` | `OptNilT` | null 許容 |
| `T` + `json:",omitempty"` | `Foo?: T` | `OptT` | 省略可能 |
| `T` (omitempty なし) | `Foo: T` | `T` | 必須 |
| `fieldNullabilityOverrides` | `Foo?: T \| null` | `OptNilT` | 実測で null 確認済み |

以前は `omitempty` も nullable 扱いしていたため、`Name` のような必須フィールドまで
`OptNilString` (`.Set`/`.Null`/`.Value` の 3 段アクセス) になり downstream の ergonomics が
劣化していた。現在は「実測で null が返ると確認できたフィールドのみ nullable」に揃えた。
実測で null を確認したら `fieldNullabilityOverrides` にエントリ追加。

## 4. 主要な設計判断

### 4.1 共有エンドポイントグループの fat model

`appliance` (Database/LoadBalancer/NFS/MobileGateway/VPCRouter) と `commonserviceitem`
(AutoBackup/AutoScale/DNS/GSLB/ProxyLB/…) は複数リソースが同一パスを共有する。union 型は
OpenAPI の `anyOf` になり ogen が "complex anyOf" 未実装扱いにしていたため、**全バリアントの
フィールドを 1 つのモデルに統合した fat model** を生成する (`fat_model.go`)。

- DSL の mapconv タグ (`"Remark.[]Servers.IPAddress"` 等) を解析してネスト復元
- 全バリアントに存在するフィールドは required、一部のみは optional
- create POST の fat model には `Class: string` を必須追加、update PUT には含めない
- 同パスに複数型が衝突したら `unknown` にフォールバック
- 実 API で省略される top-level (`Icon`/`Plan`/`Disk`/`Settings` 等) は
  `fatModelAlwaysOptionalTop` で強制 optional 化 (常時送信で 503 を避ける)
- Appliance の create/read/find response envelope は **Database** (アルファベット最初) を
  代表型として採用し、他 Appliance 固有フィールドは `fieldNullabilityOverrides` で optional 化

### 4.2 共有グループのリクエスト body ラッピング

DSL の `param:` 引数を TypeSpec にそのまま出すと body が `{"param": ...}` になるが、実 API は
`{"Appliance": ...}` を要求する。`ops.go` の `generateSharedGroupFile` で mapconv タグを見て:

- `"<payload>,recursive"` (Create/Update) → 専用 envelope (`ApplianceCreateRequestEnvelope
  { Appliance: ApplianceCreateRequest }`) を `@body body:` で渡す
- `,squash` (Shutdown の `ShutdownOption` 等) → wrap せず `@body body:` で直接渡す
- Find の `FindCondition` は例外 (後述)

### 4.3 同一 method+path 相乗り op の統合

v1 は 1 つの実 API を複数の Go メソッドに分けているケースがある (例: POST `/archive` の
Create / CreateBlank)。v2 は **実 API 1 個に 1 オペレーション** にまとめる。

- 名前が最短の op を primary (`primaryOpForKey`) として採用
- envelope payload は union、全 op にあれば required、一部のみなら optional
- variant の request model (`ArchiveCreateBlankRequest` 等) は primary に merge して単体では emit しない
- 例外的に生成から外したい op は `excludedOps` に登録

### 4.4 Find クエリの非標準フォーマット対応

実 API は `GET /bridge?{"Count":3,"Filter":{"Name":"foo"}}` のようにクエリ文字列直下に JSON を
置く非標準形式。OpenAPI では表現できないため:

- **TypeSpec 層**: Find op は `@get` + `@query q?: string` として記述 (将来形 `?q={json}` と一致)
- **ワイヤ層**: ogen は `?q=%7B...%7D` (URL-encoded) を送るので `v2/middleware.go` の
  `findQueryRewriteMiddleware` が `?{...}` に書き換えて現行サーバに適合
- **利用 API**: `internal/tools/gen-find-request/` が生成する `XxxFindRequest` /
  `XxxFindFilter` を `.ToOptString()` で変換して `Q` パラメータに渡す
- 対応フィルタフィールド (Name/Tags/Scope/Class/ProviderClass) は `gen-find-request/main.go`
  の `manifest` に allowlist 登録。Sort/Include/Exclude は **定義しない**
  (クライアント側で並べ替え可能、スキーマ駆動と相性が悪い)

### 4.5 convenient errors

全オペレーションの戻り値を `ReturnType | ApiError` にして ogen の convenient errors を有効化。
`spec/src/main.tsp` に `@error model ApiError` を定義。

### 4.6 enum デフォルト値の省略

enum 型のデフォルトを TypeSpec に入れると OpenAPI で `$ref + default` になり ogen が "complex
defaults" 未対応。`convertDefaultValue` で enum 型はコメントのみ出力する。

## 5. 命名規則

### 5.1 リソース名と TypeSpec ファイル

| レイヤー | 書式 | 例 |
|---|---|---|
| TypeSpec 内部の型名 | PascalCase の `TypeName` | `Server` / `PrivateHost` / `VPCRouter` |
| TypeSpec ファイル配置 | `spec/typespec/resources/<snake_case>/` | `server/` / `private_host/` / `vpc_router/` |
| 共有グループ | `resources/<group_snake>/ops.tsp` | `resources/appliance/ops.tsp` / `resources/common_service_item/ops.tsp` |

ディレクトリは `snake_case`、TypeSpec モデル名は `PascalCase`。共有グループの TypeSpec
グループ名は `pathNameToGroupName` マップで管理 (`commonserviceitem` → `CommonServiceItem`,
`appliance` → `Appliance`)。

### 5.2 TypeSpec interface / operation 名

**interface 名**: `<TypeName>Op` (単一リソース) / `<GroupName>Op` (共有グループ)。
共有グループではグループ共通 `ApplianceOp` に加え、リソース固有 op を持つ `DatabaseOp`
`VPCRouterOp` 等が同じ `ops.tsp` に同居する。

```typespec
@tag("Server")
interface ServerOp {
  @summary("Server 一覧取得")
  @get
  @route("/server")
  find(...): ServerFindResponseEnvelope | ApiError;
  ...
}

@tag("Database")
interface DatabaseOp { ... }
```

**op 名**: DSL operation 名を `lowerFirst` したもの (`find` / `create` / `read` /
`update` / `delete` / `boot` / `shutdown` / `changePlan` / `insertCDROM` 等)。
同一 method+path 相乗り op は `primaryOpForKey` で名前が最短のもの (`Create` < `CreateBlank`
→ `create`) を採用。共有グループの interface 内で op 名衝突が起きた場合:

- path 固有パラメータがあれば `<name>By<ParamSuffix>` (例: `changePlanBySid`)
- それ以外は `<name>2` `<name>3` と連番を付ける

**operation id**: TypeSpec は `<InterfaceName>_<opName>` を自動割り当て
(`ApplianceOp_create`, `ServerOp_find`)。これが OpenAPI の `operationId` となり、ogen の
生成メソッド名の元になる。

### 5.3 ogen 生成クライアントの命名 (`v2/client/`)

`<ResourceTypeName>Op<Action>` (action は PascalCase)。

| 種類 | 書式 | 例 |
|---|---|---|
| メソッド | `<TypeName>Op<Action>` | `ServerOpFind` / `ArchiveOpCreate` / `DatabaseOpSetParameter` |
| Params 構造体 | `<TypeName>Op<Action>Params` | `ServerOpReadParams` |
| DELETE レスポンス別名 | `<TypeName>Op<Action>OK` | `ServerOpDeleteOK` |

### 5.4 エンベロープ・リクエスト/レスポンスモデル名

DSL の `Operation.{Request,Response}EnvelopeStructName()` が
`<camelCase(resource)><opName>{Request,Response}Envelope` を返し、TypeSpec 側で upperFirst
した `<ResourceTypeName><OpName>{Request,Response}Envelope` が最終型名となる。

| 種類 | 書式 | 例 |
|---|---|---|
| Request envelope | `<TypeName><OpName>RequestEnvelope` | `ServerCreateRequestEnvelope` / `ServerBootRequestEnvelope` |
| Response envelope | `<TypeName><OpName>ResponseEnvelope` | `ServerFindResponseEnvelope` / `ServerUpdateResponseEnvelope` |
| Request body model | `<TypeName><Action>Request` | `ServerCreateRequest` / `ServerUpdateRequest` / `ServerChangePlanRequest` |
| View / レスポンス型 | `<TypeName>` | `Server` / `Archive` / `Disk` |
| ネスト submodel | `<TypeName><FieldPath>` | `ServerServerPlan` / `ServerInstance` / `ServerConnectedDisk` |
| 共有グループ request | `<GroupName><Action>Request` | `ApplianceCreateRequest` / `CommonServiceItemUpdateRequest` |
| 共有グループ response envelope | `<RepresentativeType><Op>ResponseEnvelope` | `DatabaseCreateResponseEnvelope` (Appliance 系は Database が代表) |
| Find request/filter (手書き生成) | `<TypeName>FindRequest` / `<TypeName>FindFilter` | `ServerFindRequest` / `ServerFindFilter` |

同一 method+path 相乗り op の variant request model (`ArchiveCreateBlankRequest` 等) は
primary (`ArchiveCreateRequest`) に merge され、単体では emit されない。envelope 合成は
primary を先に visit してから variant を処理することで payload 名の型解決を primary 型に
向けている (`visitOrder`)。

### 5.5 v2 ラッパー層 (`v2/<resource>_gen.go`) の命名

`internal/tools/gen-v2-op/` が ogen invoker をラップしたインターフェース + 構造体を生成する。

| 種類 | 書式 | 例 |
|---|---|---|
| public interface | `<TypeName>API` | `ServerAPI` / `DatabaseAPI` |
| 非公開実装構造体 | `<typeName>Op` (lowerFirst) | `serverOp` / `databaseOp` |
| コンストラクタ | `New<TypeName>Op(c *client.Client) <TypeName>API` | `NewServerOp` / `NewDatabaseOp` |
| メソッド | PascalCase action | `Boot` / `ChangePlan` / `Create` / `Delete` / `Read` / `Update` |

**メソッド名の例外**: Find op は Go のイディオムに合わせて **`List`** に rename される
(`op.List(ctx, req *FooFindRequest)`)。TypeSpec / ogen 側は `find` のまま。

### 5.6 フィールド名

- TypeSpec 側は **mapconv リネーム後の名前** を採用する (v1 naked の `IconID` は v2 で
  `Icon` として emit)
- `fieldmanifest.Manifest` のキー/値も TypeSpec 名 (mapconv 後) で記述
- `fieldNullabilityOverrides` `modelFieldExclusions` `modelFieldVisibility` も TypeSpec 名
  ベース

### 5.7 enum 型

- TypeSpec では `types.` 名前空間配下の enum として emit (例: `types.EInterfaceDriver`)
- Go DSL の `types/` パッケージを AST 解析して `spec/typespec/types.tsp` に生成
- enum 名は v1 の `E<Name>` プレフィックス規約を踏襲 (`EAvailability` / `EInterfaceDriver` /
  `EServerInstanceStatus` 等)

### 5.8 HTTP route とパスパラメータ

- 共通プレフィックス `/{zone}/api/cloud/1.1/` は `@server` で付与し、個別 route では省略
  (例: `@route("/server/{id}")`)
- パスパラメータ型は `pathParamDocs`/`pathParamSpec` で管理。数値 ID 系は `int64`
  (`id` / `serverID` / `switchID` / `bridgeID` / `packetFilterID` / `simID` /
  `sourceArchiveID` / `destZoneID` / `ipv6netID` / `accountID` / `subnetID`)
- `string` 据え置き: `clientID` (`cli_xxxx`) / `destination` / `ipAddress` / `MemberCode` /
  `username`
- 暫定 `string` 据え置き: `index` / `nicIndex` / `year` / `month` (ゼロ埋め書式の要確認)

### 5.9 summary/override マップのキー規約

- `summaryOverrides`: `<ResourceTypeName>.<actionLower>` (例: `Server.boot` / `Database.setParameter`)
- `actionSummaries`: action 名 (`lowerFirst` 済み) → 日本語動詞句
- `postStatusCodeOverrides`: TypeSpec `@route` の解決済みパス (`/internet` / `/appliance` /
  `/internet/{id}/ipv6net` 等)

## 6. API 規約

### 6.1 TypeSpec ドキュメンテーション

- `@doc` / `@summary` は **日本語**で記述し、API 利用者向け外部ドキュメントとして完結させる
- Go 実装詳細・ogen/jx ライブラリ名・プロジェクト内部方針・v1 DSL の型名などは書かない
  (これらは AGENTS.md 側に書く)
- op の `@summary` は `summaries.go` の `actionSummaries` で `<Resource> <動詞>` 形式に機械合成。
  日本語として不自然な場合は `summaryOverrides` にマニュアル URL をコメント付きで追加

### 6.2 HTTP ステータスコード

- POST → 201 Created デフォルト、個別で 200/202 を `postStatusCodeOverrides` に登録
- GET (read/find) / PUT (update) → 200 OK
- DELETE → 200 OK with `{is_ok: boolean}` (204 ではない)

union ステータス (`201 | 202`) は ogen が interface 戻り値 + type switch を生成してしまうため
エンドポイント毎にひとつだけ指定する。未登録は 201 fallback。

### 6.3 BigInt ヘッダ

全 API 呼び出しに `X-Sakura-Bigint-As-Int: 1` を付与。個別 op に書かず、saclient-go の
Transport middleware で一括付与する (利用者の必要操作は main.tsp の `@doc` に記載)。

### 6.4 レスポンスエンベロープ

| インターフェース | v2 TypeSpec の扱い |
|---|---|
| Find | `{Total, From, Count, <Resource>s}` |
| GET (read) | `{is_ok, <Resource>}` |
| POST (create) | `{is_ok, <Resource>}` (`Success` は未定義) |
| PUT (update) | `{is_ok, <Resource>}` (`Success` は未定義) |
| DELETE | `{is_ok}` |

`Success` を載せない理由: 実 API で型が不安定 (boolean / 文字列混在) かつ downstream 参照 0 件。
成功判定は `is_ok` を使う。

### 6.5 リクエストエンベロープ

create/update のリクエスト body は、レスポンス用ビュー型 (`Icon` 等) ではなく、オペレーション
固有の request 型 (`IconCreateRequest` / `IconUpdateRequest` 等) を使う
(`envelopes.go` で Arguments から解決)。

### 6.6 Password 等の機微情報

ユーザ設定値が平文 echo されるフィールド (VPCRouterRemoteAccessUser / Database の
UserPassword 等) は `modelFieldVisibility` で `@visibility(Lifecycle.Create, Lifecycle.Update)`
を付け、Read レスポンスから除外する。サーバ生成型 (`FTPServer.Password` / `VNCProxyInfo.Password`)
は性質上残置。

## 7. ラッパー層の構造 (v2/*_gen.go, v2/client.go, v2/middleware.go)

saclient-go の共通土台 (認証・プロファイル解決・ミドルウェアチェーン) の上にリソースごとの
Op インターフェースが載る薄いラッパー。使用感は `simple-notification-api-go` と揃える。

- `v2/client/` は ogen 生成物そのまま (手を入れない)
- `v2/client.go` の `iaas.NewClient(sc *saclient.Client)` が saclient ミドルウェアに
  `stripOgenAuthMiddleware` (ogen の空 Basic 認証を剥がす) と `findQueryRewriteMiddleware` を
  prepend する
- リソース Op (`v2/<resource>_gen.go`) は `internal/tools/gen-v2-op/` が自動生成。インタフェース
  を返すコンストラクタ + ogen 生成型をそのまま露出するメソッド
- `SAKURA_TRACE=1` は `saclient.WithTraceMode("all")` でトレース出力
- `op` のない bucket の stale ファイルは `gen-v2-op` が自動削除

### Op 層のシグネチャ

- zone は `iaas.NewClient(sc, zone)` に渡してクライアント構築時に固定。URL テンプレートの
  `{zone}` プレースホルダに埋め込まれる (v1 のようにメソッド毎に zone を取り回さない)
- ラッパーメソッドは `ctx` を第一引数、id (int64) を第二引数、必要なら request envelope を
  最後に取る
- 戻り値は ogen 生成型をそのまま露出 (v1 のような独自型詰め替えは行わない)
- エラーは `errors.As(err, &client.ApiErrorStatusCode)` で status code を取り出して
  `*iaas.Error` でラップ

## 8. Error 型

`v2/error.go` の `*iaas.Error` が v1 `APIError` interface 互換のアクセサを実装:

```go
func (e *Error) ResponseCode() int     // ApiErrorStatusCode.StatusCode
func (e *Error) Code() string          // Response.ErrorCode
func (e *Error) Message() string       // Response.ErrorMsg
func (e *Error) Serial() string        // Response.Serial
```

`iaas.IsNotFoundError(err)` は saclient-go の `IsNotFoundError` への alias として公開済み。
`IsStillCreatingError` 等の細かい判定は `err.Code() == "still_creating"` で代替する
(downstream 直 import を作らない方針)。

## 9. helper 層 (v2/helper/{power, query, wait, plans, cleanup})

v1 の `helper/*` を v2 型 (`int64` ID / ogen 生成型) で再実装。

- **v2 → v1 依存を作らない** (`iaas.*` / `types` / `search` / `ostype` を import しない)
- 各関数は narrow interface (必要最小限のメソッドだけを持つ interface) を引数に取り、テストで
  fake を書きやすくする
- `helper/api` は saclient-go + api-client-go でカバー済みなので port しない
- 将来モノレポ化のタイミングで `github.com/sacloud/helper-go` 相当の独立パッケージへ分離予定

主要 API は [v2-migration-guide.md](./v2-migration-guide.md) の helper 節を参照。

### 既知の制約 (v2 spec 未対応で省略中)

- `IsSwitchReferenced`: Database 系レスポンスの Interface から Switch 参照を検出できないため
  `Switch.GetServers` 経由の検査のみ
- `Switch.HybridConnectionID`: v2 spec 未公開のため参照検査から除外
- `ReadProxyLB`: `CommonServiceItemOpRead` の envelope が AutoBackup 固定型で ProxyLB fat model
  を返せず見送り

## 10. テスト方針

v2 は **実 API を使用した統合テストのみ**。

- ogen 生成コードの品質は ogen に担保してもらう (単体テスト不要)
- fake モックは作らない (メンテナンス不能が増えるため)
- テスト場所: `v2/integration/`
- 全リソース CRUD を対象。最初は Icon から開始
- 実行: `cd v2 && TEST_ACC=1 go test -v -timeout 30m ./integration/...`
  (Appliance 系は 1 ケース 1〜5 分かかるので `-timeout 30m` 以上必須)
- `SAKURA_TRACE=1` でリクエスト/レスポンスを tracer 出力
- `TEST_ACC_CLEANUP=1` で "test"/"integration" タグ付きリソースを削除

sandbox (`tk1v`) で動かない (Plan が無い / `dont_create_in_sandbox` 403 返却) テストは本番ゾーン
(`tk1a`) をハードコードする (PrivateHost / License / Bridge)。

## 11. バージョン運用方針

- **当面は同一 main branch で v1 と v2 を併存**させて開発する
- v1: 既存利用者向けの互換維持
- v2: 非互換変更の受け皿
- 生成物の出力先、ビルドスクリプト、CI ジョブは v1/v2 で分離
- `v2/go.mod` は独立モジュール
- 将来保守負荷でスプリットが必要になれば v1 maintenance ブランチを切る
- Go の import path / module path はメジャーバージョンの規約を維持 (cf. [Go Modules: v2 and
  Beyond](https://go.dev/blog/v2-go-modules))

v1.0 到達までは互換性のない変更があり得る (v2/README.md 参照)。

## 12. 将来像

- TypeSpec を source of truth に昇格させ、`spec/typespec/` を手書き編集、
  `internal/tools/gen-typespec/` を廃止、`fieldmanifest` は OpenAPI lint レベルのチェックへ
- `v2/helper/` をモノレポ外の `helper-go` 相当パッケージへ移動 (SEG 等他サービスと共用)
- `v2/client/` を手書きラッパーに書き換え、ogen 出力を `v2/client/oasgen/` サブパッケージへ
  分離 (Find RoundTripper の自動組み込み等のため)

切替条件の目安は「全 downstream が v2 に切り替わった時点で Go DSL を凍結」。詳細な移行 roadmap は
[docs/v2-issues.md](./v2-issues.md) の「低」項目を参照。
