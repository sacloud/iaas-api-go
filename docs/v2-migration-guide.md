# v1 → v2 移行ガイド

`iaas-api-go` v1 から v2 への移行で、downstream (usacloud / terraform-provider-sakuracloud /
terraform-provider-sakura) が書き換える必要のある主要な差分を整理する。

- 設計思想は [v2-design.md](./v2-design.md) を参照
- 継続中の課題・未完了項目は [v2-issues.md](./v2-issues.md) を参照
- v2.0 GA 前のため、互換性のない変更が入る可能性あり

## 目次

1. [モジュールパスとインポート](#1-モジュールパスとインポート)
2. [クライアント初期化](#2-クライアント初期化)
3. [リソース Op の使い方](#3-リソース-op-の使い方)
4. [Find 検索の書き方](#4-find-検索の書き方)
5. [ID の型 (`types.ID` → `int64`)](#5-id-の型-typesid--int64)
6. [Optional / Nullable フィールドのアクセス](#6-optional--nullable-フィールドのアクセス)
7. [エラーハンドリング](#7-エラーハンドリング)
8. [helper 層の移行](#8-helper-層の移行)
9. [リソース単位の注意点](#9-リソース単位の注意点)
10. [既知のギャップ・未対応項目](#10-既知のギャップ未対応項目)

---

## 1. モジュールパスとインポート

v2 は独立モジュール (`v2/go.mod`)。downstream は v1 と同じ repo の v2 サブパッケージを参照する。

```go
// v1
import "github.com/sacloud/iaas-api-go"
import "github.com/sacloud/iaas-api-go/types"
import "github.com/sacloud/iaas-api-go/helper/power"

// v2
import iaas "github.com/sacloud/iaas-api-go/v2"
import "github.com/sacloud/iaas-api-go/v2/client"
import "github.com/sacloud/iaas-api-go/v2/helper/power"
import "github.com/sacloud/saclient-go"
```

- **`iaas` alias**: v2 トップレベルのパッケージ名は `iaas`。`github.com/sacloud/iaas-api-go/v2`
  を `iaas` として alias するのが慣習。
- **`client`**: ogen 生成型 (リクエスト/レスポンス envelope、モデル、Params 構造体) は
  `v2/client/` にまとまっている。
- **`types` パッケージは廃止**: `types.ID` / `types.APIResult` 等の独自型は使わない。ID は
  `int64` (後述)。enum 相当は ogen 生成型 (`client.EAvailability` など)。

## 2. クライアント初期化

認証・リトライ・ヘッダ付与は [`saclient-go`](https://github.com/sacloud/saclient-go)
(共通土台) に一元化されている。

```go
package main

import (
    "context"
    "os"

    iaas "github.com/sacloud/iaas-api-go/v2"
    "github.com/sacloud/saclient-go"
)

func main() {
    var sc saclient.Client
    if err := sc.SetEnviron(os.Environ()); err != nil {
        panic(err)
    }
    // `Populate` はクライアント初回利用時に遅延実行される

    c, err := iaas.NewClient(&sc, "tk1v")
    if err != nil {
        panic(err)
    }

    ctx := context.Background()
    _ = ctx
    _ = c
}
```

### 1 クライアント = 1 ゾーン

v2 では **zone をクライアント構築時に決める**。`iaas.NewClient(sc, zone)` が URL テンプレートの
`{zone}` を埋め込んだ API root URL を組み立てて ogen クライアントを返す。複数ゾーンを跨ぐ場合は
ゾーンごとに `NewClient` を呼んで別インスタンスを持つ (v1 のように毎呼び出しで `zone` を渡す
形式から変わる)。

### 認証情報の渡し方

- **環境変数** (`SAKURA_ACCESS_TOKEN` / `SAKURA_ACCESS_TOKEN_SECRET` / `SAKURA_ZONE`):
  `sc.SetEnviron(os.Environ())`
- **usacloud プロファイル**: `sc.SetUsacloudProfile(profileName)`
- **HCL (Terraform)**: Terraform provider は HCL 由来の設定を saclient の設定源に注入する

### 自動で組み込まれるミドルウェア

`iaas.NewClient` 内で以下が saclient のミドルウェアチェーンに prepend される:

- `stripOgenAuthMiddleware`: ogen が自動注入する空 Basic 認証ヘッダを剥がす
  (saclient の `middlewareAuthorization` が正しいヘッダを付ける)
- `findQueryRewriteMiddleware`: ogen 生成の `?q={encoded-json}` を実 API が期待する
  `?{json}` 形式に書き換える

saclient-go 側で `X-Sakura-Bigint-As-Int: 1` ヘッダ付与・User-Agent・tracer・retry 等は一括
提供されるため、v1 のような個別 Transport 構築は不要。

### トレース

`SAKURA_TRACE=1` でリクエスト/レスポンスを標準エラー出力にダンプする
(内部的に `saclient.WithTraceMode("all")`)。

## 3. リソース Op の使い方

各リソースに対する CRUD は `v2/<resource>_gen.go` が生成する `<Resource>API` インターフェース
経由で行う。

```go
serverOp := iaas.NewServerOp(c)

// Read
resp, err := serverOp.Read(ctx, id)

// Create
resp, err := serverOp.Create(ctx, &client.ServerCreateRequestEnvelope{
    Server: client.ServerCreateRequest{
        Name:        "my-server",
        Description: client.NewOptString("hello"),
        ServerPlan:  client.ServerCreateRequestServerPlan{ID: client.NewOptInt64(100001001)},
        Tags:        []string{"test"},
    },
})

// Update / Delete / Power 操作
_, err = serverOp.Update(ctx, id, &client.ServerUpdateRequestEnvelope{...})
err = serverOp.Delete(ctx, id, &client.ServerDeleteRequestEnvelope{})
err = serverOp.Boot(ctx, id, &client.ServerBootRequestEnvelope{})
err = serverOp.Shutdown(ctx, id, &client.ServerShutdownRequestEnvelope{Force: client.NewOptBool(false)})
```

### シグネチャのルール

- 第 1 引数: `ctx context.Context`
- 第 2 引数: `id int64` (対象リソース ID) — `read` / `update` / `delete` / sub-action
- 最後の引数: `request *client.<Resource><Op>RequestEnvelope` (body を伴う op のみ)
- 戻り値: `(*client.<Resource><Op>ResponseEnvelope, error)` または `error`
- **zone 引数は取らない**: クライアント構築時に確定済み

### Find op は `List` にリネームされる

TypeSpec / ogen 側の op 名は `find` のままだが、v2 ラッパーでは Go イディオムに合わせて
**`List`** に rename されている。引数も envelope ではなく `<Resource>FindRequest` を渡す
(後述)。

### 共有エンドポイントグループ (Appliance / CommonServiceItem)

v1 では Database / LoadBalancer / NFS / VPCRouter / MobileGateway を個別の Op で扱っていた
が、実 API は同一パス。v2 では:

- 共通 CRUD (`create` / `read` / `update` / `find` / `delete` / power 操作) は
  `iaas.NewDatabaseOp(c)` 等に対して呼ぶが、request/response は **共有 fat model**
  (`client.ApplianceCreateRequest` / `client.DatabaseCreateResponseEnvelope` 等) が使われる
- Appliance 系の GET レスポンス envelope は **Database** 型を代表として採用している
  (アルファベット順で最初)。Database 固有フィールドは optional になっているため
  `.IsSet()` チェックが必要
- リソース固有の sub-action (例: `DatabaseOp.GetParameter` / `VPCRouterOp.MonitorInterface`)
  は個別の Op に分離されている

CommonServiceItem (AutoBackup / AutoScale / DNS / GSLB / ProxyLB / SimpleMonitor /
CertificateAuthority 等) も同じ構造。

## 4. Find 検索の書き方

v1 の `search.Filter` + `types.SortBy` は廃止。v2 は `<Resource>FindRequest` +
`<Resource>FindFilter` を組み立てて `List()` に渡す。

```go
req := &client.ServerFindRequest{
    Count: 50,
    From:  0,
    Filter: client.ServerFindFilter{
        Name: "my-server",        // スペース区切りで部分一致 AND
        Tags: []string{"prod"},   // 完全一致 AND
    },
}
resp, err := serverOp.List(ctx, req)
```

### フィルタフィールドのサポート

対応フィールドは `internal/tools/gen-find-request/main.go` の `manifest` で allowlist 化
されている。全リソース共通ではなく、リソースごとに必要なものだけ生成される。

| フィールド | 対象リソース |
|---|---|
| `Name` | ほぼ全リソース |
| `Tags` | ユーザ作成リソース / CommonServiceItem / Appliance |
| `Scope` | Archive / CDROM / Disk / Icon / Note / Switch |
| `Class` | Appliance 個別 (Database / LoadBalancer / …) / PrivateHostPlan |
| `ProviderClass` | CommonServiceItem 系 (DNS / GSLB / ProxyLB 等) |

**定義しない**: Sort / Include / Exclude。クライアント側で並べ替え可能・スキーマ駆動と相性が
悪いため意図的に除外している。どうしても必要ならベタに
`client.ServerOpFindParams{Q: client.NewOptString(jsonStr)}` を ogen 直接で叩くしかない。

### 生の ogen クライアント直叩き

`iaas.NewClient` は `*client.Client` (ogen 生成) を直接返す。ラッパー Op が提供しない
形のアクセスが必要な場合はそのまま利用できる:

```go
c, _ := iaas.NewClient(&sc, "tk1v")
resp, err := c.ServerOpFind(ctx, client.ServerOpFindParams{
    Q: req.ToOptString(),
})
```

`findQueryRewriteMiddleware` がトランスポート層で `?q=...` を `?{...}` に自動書き換えするので、
ユーザ側で RoundTripper を組み立てる必要はない。zone はクライアントの URL テンプレートに
埋め込まれているので `Params` 側に zone フィールドは無い。

## 5. ID の型 (`types.ID` → `int64`)

v2 では `types.ID` を廃止し、**数値 ID パスパラメータはすべて `int64`** に統一した。
対応済みのキー: `id` / `accountID` / `bridgeID` / `destZoneID` / `ipv6netID` /
`packetFilterID` / `serverID` / `simID` / `sourceArchiveID` / `subnetID` / `switchID`。

```go
// v1
var id types.ID = types.StringID("113000000123")
s, err := iaas.NewServerOp(caller).Read(ctx, zone, id)
fmt.Println(s.ID.String())  // "113000000123"
fmt.Println(s.ID.Int64())

// v2
var id int64 = 113000000123
resp, err := serverOp.Read(ctx, id)
fmt.Println(resp.Server.ID.Value)  // int64 直値
```

### 例外

- `string` のまま据え置き: `clientID` (`cli_xxxx` 等) / `destination` (IP/hostname) /
  `ipAddress` / `MemberCode` / `username`
- 暫定 `string`: `index` / `nicIndex` / `year` / `month` (ゼロ埋め書式の要確認、別 issue 残)
- `Remark.Switch.ID` を body で文字列として渡す実装 (Database / NFS / LoadBalancer) は
  旧 API 仕様に従い `fmt.Sprintf("%d", switchID)` で埋めている箇所が残る。該当箇所は
  `helper/cleanup` / integration test で `// 文字列必須` のコメント付き

## 6. Optional / Nullable フィールドのアクセス

ogen は TypeSpec の `?` / `| null` を Go の Opt 型で表現する。2026-04-22 の調整で優先される
パターンが整理された (詳細: [v2-issues.md #2](./v2-issues.md) / v2-design.md 3.4)。

| v1 naked 宣言 | v2 Go 型 | アクセス | 書き込み |
|---|---|---|---|
| `T` (omitempty なし) | `T` | そのまま | そのまま |
| `T` + `json:",omitempty"` | `client.OptT` | `.Value` / `.IsSet()` | `client.NewOptT(v)` |
| `*T` (ポインタ) | `client.OptNilT` | `.Value` / `.IsSet()` / `.Null` | `client.NewOptNilT(v)` / `client.NewOptNilT[T]{Null: true}` |
| 実測で null 観測 | `client.OptNilT` | 同上 | 同上 |

```go
// v1
fmt.Println(server.Name)                 // string
fmt.Println(server.Description)          // string（omitempty 吸収）

// v2
fmt.Println(resp.Server.Name)            // string（non-nullable）
if resp.Server.Description.IsSet() {
    fmt.Println(resp.Server.Description.Value)
}
if resp.Server.HostName.IsSet() && !resp.Server.HostName.Null {
    fmt.Println(resp.Server.HostName.Value)
}
```

### 以前より nullable が減っている

2026-04-22 以前は `omitempty` も nullable 扱いしていたため、`Name` のような必須フィールドまで
`OptNilString` (3 段アクセス) になっていた。現在は `Name` のような実測 non-null フィールドは
`string` 直値、`Description` などは `OptString` (2 段)、実測で null が返るフィールドだけ
`OptNilString` (3 段)。

新しく null を観測した場合は `internal/tools/gen-typespec/models.go` の
`fieldNullabilityOverrides` に追加する (実測根拠をコメントで残す)。

## 7. エラーハンドリング

### 404 判定

v1 の `iaas.IsNotFoundError(err)` は v2 でも同名で使える (saclient-go への alias)。
**downstream 78 箇所の書き換えは不要**。

```go
resp, err := serverOp.Read(ctx, id)
if err != nil {
    if iaas.IsNotFoundError(err) {
        return nil // not found
    }
    return err
}
```

### `*iaas.Error` のアクセサ

v1 `APIError` インターフェース互換のアクセサを実装している:

```go
var ierr *iaas.Error
if errors.As(err, &ierr) {
    fmt.Println(ierr.ResponseCode())  // HTTP status code (int)
    fmt.Println(ierr.Code())          // API error code (string, 例: "still_creating")
    fmt.Println(ierr.Message())       // 人向けメッセージ
    fmt.Println(ierr.Serial())        // リクエスト識別シリアル
}
```

ネットワークエラー等で `*client.ApiErrorStatusCode` に `errors.As` できない場合は `0` /
空文字が返る。

### `IsStillCreatingError` 相当

v1 の `iaas.IsStillCreatingError` は提供していない (downstream 未使用のため)。判定したい
場合は `err.Code() == "still_creating"` と書く。`IsNoResultsError` は v2 Find が空リストを
返すだけで `NoResultsError` を吐かないため、提供しない。

## 8. helper 層の移行

v1 `helper/{power, query, wait, plans, cleanup}` は v2 でも同構成で提供する。v2 型
(`int64` ID / ogen 生成型) で再実装されており、v2 → v1 依存は持たない。

```go
import (
    "github.com/sacloud/iaas-api-go/v2/helper/power"
    "github.com/sacloud/iaas-api-go/v2/helper/cleanup"
    "github.com/sacloud/iaas-api-go/v2/helper/wait"
    "github.com/sacloud/iaas-api-go/v2/helper/query"
    "github.com/sacloud/iaas-api-go/v2/helper/plans"
)
```

### 主要 API

| パッケージ | 代表関数 | 備考 |
|---|---|---|
| `wait` | `StateWaiter` / `SimpleStateWaiter` / `UntilServerIs{Up,Down}` / `UntilApplianceIs{Up,Down,Ready}` / `UntilArchiveIsReady` / `UntilDiskIsReady` | 各関数は narrow interface を引数に取る |
| `power` | `BootServer` / `ShutdownServer` / `BootAppliance` / `ShutdownAppliance` | `still_creating` リトライ + 逆方向状態時の再送あり |
| `query` | `FindArchiveByOSType` (v2 内 `ArchiveOSType` enum) / `FindServerPlan` / `ReadServer` / `ReadRouter` / `Is{Disk,CDROM,Switch,Bridge,PrivateHost,PacketFilter,SIM}Referenced` / `WaitWhile*IsReferenced` | `previous-id` fallback + `ErrNoResults` を返す |
| `cleanup` | `DeleteServer` (withDisks) / `DeleteDisk` / `DeleteCDROM` / `DeleteSwitch` / `DeleteBridge` / `DeletePacketFilter` / `DeletePrivateHost` / `DeleteSIM` / `DeleteInternet` / `DeleteMobileGateway` | 依存リソースの detach を含む |
| `plans` | `AppendPreviousIDTagIfAbsent` / `ChangeServerPlan` / `ChangeRouterPlan` | プラン変更時に旧 ID を tag 保持 |

### narrow interface パターン

v2 helper は `*iaas.Client` を直接受け取らず、必要なメソッドだけを持つ interface を引数に
とる。例:

```go
type archiveReader interface {
    Read(ctx context.Context, zone string, id int64) (*client.ArchiveReadResponseEnvelope, error)
}

func UntilArchiveIsReady(ctx context.Context, reader archiveReader, zone string, id int64, timeout time.Duration) error
```

これによりテストで fake を書きやすく、将来モノレポ分離したときの依存も最小化している。

### v1 との差分

- `helper/api` は移植しない (saclient-go + api-client-go でカバー済)
- `ostype.ArchiveOSType` は v2 helper 内で自前定義 (v1 `ostype` パッケージに依存しない)
- 関数シグネチャの ID 引数は `int64` 化。`types.ID` は使わない

### 既知の制約

- `IsSwitchReferenced`: Database 系レスポンスに Interface ごとの Switch 情報が無いため、
  v2 では `Switch.GetServers` 経由の検査のみ (v1 にあった Appliance Interface 走査は省略)
- `Switch.HybridConnectionID`: v2 spec 未公開のため参照検査から除外
- `ReadProxyLB`: ProxyLB fat model を返す envelope が v2 spec 未対応で見送り中

これらは v2 spec 側の追加と合わせて後続で拡張する。

## 9. リソース単位の注意点

### Password フィールドの Read レスポンス除外

以下の設定値 echo 型 Password は、Read レスポンスに含まれなくなった
(`modelFieldVisibility` で `@visibility(Lifecycle.Create, Lifecycle.Update)` を付与):

- `VPCRouterRemoteAccessUser.Password`
- `DatabaseSettingCommon.UserPassword` / `ReplicaPassword`
- `DatabaseRemarkDBConfCommon.UserPassword`
- `DatabaseReplicationSetting.Password`
- `SimpleMonitorHealthCheck.BasicAuthPassword`

downstream (`terraform-provider-sakuracloud` / `terraform-provider-sakura`) が state 復元に
Read レスポンスの Password を使っているケース (`SimpleMonitorHealthCheck.BasicAuthPassword`)
は、provider 側を **config 由来で復元する方式** に書き換えが必要。

残置されているのは以下のサーバ生成型 (性質上除外不可):

- `FTPServer.Password`
- `VNCProxyInfo.Password`

### `Success` レスポンスフィールド

v1 の `zz_envelopes.go` が定義していた `Success` フィールドは v2 では定義しない。成功判定は
`is_ok` を使う。実 API が `Success` を返しても ogen decoder が読み飛ばすため問題なし。

### v2 で未実装のエンドポイント

v1 DSL にもなく downstream も使っていないエンドポイントは v2 に載せていない。代表例:

- タグ系: `GET /<resource>/tag` / `GET /<resource>/:id/tag` (全リソース未実装)
- Switch: `GET /switch/:id/appliance`
- Icon: `GET /icon/:id?Size=...` (画像データ)
- Server: `GET /server/:id/cdrom` (挿入/排出のみ実装) / `/mouse/` / `/vnc/size` /
  `/vnc/snapshot`
- Appliance: 個別サーバ/ディスクの monitor / 個別プラン変更
- Database: `database/plugin` / `database/syslog` / `database/slaves` 等

完全な一覧は [AGENTS.md](../AGENTS.md) の「実装しないエンドポイント」表を参照。

### ジェネレータ未対応 (見送り中)

- **AuthStatus**: 実 API のレスポンスが envelope 直下にフラット展開されるため、wrap envelope
  と不整合。AuthStatus を使っている usacloud の表示系は v1 経由で当面残す必要あり。
- **ServiceClass (`GET /public/price`)**: `ID ↔ ServiceClassID` の JSON name remapping と
  `Price` の polymorphic (`{}` / `[]`) 対応が必要なため未対応。
- **ProxyLB Read**: `CommonServiceItemOpRead` の envelope が AutoBackup 固定型で ProxyLB
  fat model を返せず見送り。

## 10. 既知のギャップ・未対応項目

[v2-issues.md](./v2-issues.md) が一次ソース。主な未完了項目:

- **`OptNilResourceRef` 系の段階縮小** (issue #2 残タスク): `*T` ポインタ宣言 162 件は
  依然 nullable。実測 null ベースへ段階移行中
- **ID 型の統一拡張** (issue #4 残タスク): `index` / `nicIndex` / `year` / `month` の int64 化
- **ogen `debug/example_tests` の肥大** (issue #7): `v2/client/` の 12.7% (24k 行) を占める
  生成テストの無効化検討
- **`gen-typespec` の override 散在** (issue #8): リソース単位の `ResourceOverride` 構造体に
  集約する提案
- **`fieldmanifest.Manifest` の追従漏れ検知** (issue #9): downstream AST 走査で allowlist
  差分を CI 提案する `gen-fieldmanifest` ツール案
- **Experimental 宣言** (issue #15): `v2/doc.go` に `// Experimental` を記載
- **リリース・バージョニング方針** (issue #14): `v2.0.0-alpha.N` の tag 運用と CHANGELOG 整備

v2.0 GA までは互換性のない変更が入りうる。v2/README.md の ⚠️ 警告を参照すること。

---

## 付録: v1 → v2 対応表 (よく使うもの)

| v1 | v2 | メモ |
|---|---|---|
| `iaas.NewClient(token, secret)` | `iaas.NewClient(&saclient.Client{...}, zone)` | saclient-go 経由、zone はクライアント単位 |
| `iaas.NewServerOp(client)` | `iaas.NewServerOp(c *client.Client)` | `c` は `iaas.NewClient` の戻り値 |
| `op.Find(ctx, zone, cond)` | `op.List(ctx, req)` | Find → List 改称、zone 引数は無し |
| `op.Read(ctx, zone, id types.ID)` | `op.Read(ctx, id int64)` | ID は int64、zone 引数は無し |
| `iaas.IsNotFoundError(err)` | `iaas.IsNotFoundError(err)` | 同名で使える |
| `iaas.IsStillCreatingError(err)` | `err.Code() == "still_creating"` | 直接の helper は非提供 |
| `apiErr.ResponseCode()` | `ierr.ResponseCode()` | `errors.As(err, &ierr)` で取り出し |
| `server.Name` (string) | `resp.Server.Name` (string) | 必須フィールドは直値 |
| `server.Description` | `resp.Server.Description.Value` | `OptString.Value` 経由 |
| `server.HostName` (`*string`) | `resp.Server.HostName.Value` + `.IsSet()` + `.Null` | `OptNilString` の 3 段 |
| `helper/power.BootServer` | `v2/helper/power.BootServer` | narrow interface 引数 |
| `search.Filter{...}` | `client.ServerFindFilter{...}` | リソース別 typed filter |
