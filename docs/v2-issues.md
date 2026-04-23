# v2 課題一覧

`feat/typespec-gen` ブランチを main に対して評価した際の改善点を、継続メンテナンスと downstream
(usacloud / terraform-provider-sakuracloud v2 / terraform-provider-sakura v3) 移行を前提に列挙する。

扱うスコープ:

- v2 生成パイプライン（gen-typespec / ogen / gen-v2-op）の保守性
- v2 公開 API の downstream 移行容易性
- ラッパー層・統合テスト・ドキュメント整合

各項目は「状況」「影響」「対応案」の 3 点セットで記述する。

---

## 【最優先】downstream 移行の前提が未整備

terraform v3 / usacloud を v2 に寄せるなら、ここが決まらないと先に進めない。
これらが解決されないまま resource を追加しても downstream 側で per-resource の書き換え PR
が肥大するため、生成器の拡張より先に片付けたい。

### 1. `iaas.IsNotFoundError` が v2 に無い ✅ 対応済み

**対応内容**
- `v2/error.go` に `var IsNotFoundError = saclient.IsNotFoundError` を追加。
  downstream 78 箇所の `iaas.IsNotFoundError(err)` が v1 名のまま通る。
- `IsStillCreatingError` は追加しない（downstream 未使用、helper/* 移行時に判断）。
- `IsNoResultsError` は追加しない（v2 の Find は空リストを返すだけで `NoResultsError`
  相当を吐かないため、定義しても常に false になり誤解を招く）。
- 細かい条件判定 (`error_code == "still_creating"` 等) は #12 で実装した `*iaas.Error`
  アクセサ `err.Code()` で吸収する方針。
- 回帰防止テスト `v2/error_test.go` `TestIsNotFoundError` を追加。

### 2. `OptNilString` / `OptNilInt64` 氾濫の ergonomics ✅ 一次対応済み

**対応内容（2026-04-22）**
- `gen-typespec` の `nakedFieldIsNullable` から `omitempty` 判定を除去し、`*T`
  （ポインタ型）のみを nullable シグナルとして扱うよう変更。optional (`?`) と
  nullable (`| null`) を別フラグで管理する。
- 結果: v2/client/ の OptNilString **765 → 33**、OptNilInt **399 → 0**、
  OptNilBool **33 → 0**。`Name?: string | null` などの過剰 nullable が
  `Name?: string` に落ち、downstream の 3 段アクセスが 2 段に減った。
- 新方針は AGENTS.md「Optional と Nullable の分離」節に記述。実測で null が返ると
  確認できたフィールドは `fieldNullabilityOverrides` に追記するフローで吸収する。
  初回の追加例: `Interface.UserIPAddress`（未割り当て NIC で実 API が null を返す）。

**残タスク（継続課題）**
- v1 naked が `*T`（ポインタ）で宣言している 162 件の `OptNilResourceRef` 等は依然
  nullable のまま。これも「ポインタ宣言」ではなく「実測 null 観測」で判定したいが、
  全フィールドの trace 採取が必要なので段階対応。
- downstream の `.Value` は依然多い（`OptString.Value` に変わっただけ）。
  ID 型の統一（項目 5）と合わせて更なる削減を検討する。
- verify-typespec に「このフィールドは non-nullable のまま」退行チェックを追加すると
  将来の誤差し戻しを防げる。

### 3. `helper/{power, query, wait, plans, cleanup}` の v2 相当が未着手

**状況**
- downstream 3 プロジェクトすべてが v1 `helper/power.BootRetry` / `helper/query.FindProductDisk`
  などを常用（import 行単位で全リソースに散らばる）。
- v2/AGENTS.md の「要実装判断」節で方針保留のまま。

**影響**
- v2 に切り替えると helper 経由の操作が全部壊れる（import path 違い + 型違い）。
- helper 内部で v1 型（`types.ID` 等）を受け渡している箇所があり、v2 型への移植は機械的ではない。

**対応案**
- 方針 A: v1 helper を v1 型のまま維持し、downstream は v1 helper + v2 client の
  ハイブリッド構成で一定期間運用。helper は v1 タグに固定しメンテナンスのみ。
- 方針 B: `v2/helper/` を新設して v2 型で書き直す（最大のコスト）。
- 方針 C: helper の責務を downstream 側に吸収させる（helper を廃止）。
- 現状は B/C の痕跡が無いので A が現実解。ただし v1 廃止タイミングが決まらないと helper だけ
  永続する。v1 サポート期限の宣言とセットで合意が必要。

### 4. ID パス引数の型不整合 ✅ 対応済み

**対応内容**
- `internal/tools/gen-typespec/ops.go` の `pathParamDocs` を `map[string]pathParamSpec` (Doc + Type)
  に拡張し、数値リソース ID 系の `@path` を `int64` で出すように変更。
- int64 化したキー: `id`, `accountID`, `bridgeID`, `destZoneID`, `ipv6netID`,
  `packetFilterID`, `serverID`, `simID`, `sourceArchiveID`, `subnetID`, `switchID`。
  `pathParamDocs` 未登録だった `ipv6netID` を追加登録。
- string 据え置き: `clientID` (`cli_xxxx`)、`destination` (IP/hostname)、`ipAddress`、
  `MemberCode`、`username`。
- 暫定的に string のまま据え置き: `index`, `nicIndex`, `year`, `month`（小さい int / ゼロ埋め
  書式の懸念。混在を doc コメントで補足）。別 issue として留保。
- 結果: ogen 生成の `*Params` の ID フィールドが `int64` になり、`v2/*_gen.go` ラッパーの
  Read/Update/Delete 等のシグネチャも `id int64` に自動追従。
- `v2/integration/*_test.go` の `fmt.Sprintf("%d", ...ID.Value)` 50 箇所を直接渡しに置換。
  helper の wait 関数 (`waitArchiveAvailable`, `waitCDROMAvailable`, `waitDiskAvailable`,
  `waitInternetSwitchReady`, `waitApplianceAvailable`, `waitApplianceShutdown` 等) のシグネチャも
  `id int64` に変更。残った 3 箇所の `fmt.Sprintf` は Appliance Body の `Remark.Switch.ID`
  を文字列で送る必要があるためで、コメント付きで意図的に残してある。
- `v2/README.md` のサンプルを `id := s.ID.Value` で直接渡せる形に更新。

**残タスク（継続課題）**
- `index`, `nicIndex`, `year`, `month` の int64 化は実 API の書式仕様（ゼロ埋め可否）を
  確認した上で別 issue で対応。
- `Remark.Switch.ID` を body で文字列として要求している実装（database / nfs / load_balancer）は、
  サーバ側が数値 ID を受け付けるなら fat model 側を直して fmt.Sprintf を消せる。

---

## 【高】パイプライン整合性

### 5. integration tests が 2 系統並立 ✅ 対応済み

**状況（旧）**
- `v2/integration/helper_test.go` の `newClient`（saclient 未経由、`securitySource` +
  `baseTransport` + `findQueryRewriteTransport` を手組み）が 22 テストで使われる。
- `iaas_note_test.go` の `newIaasClient`（`iaas.NewClient` + ラッパー op 経由）は 1 件のみ。
- README は `iaas.NewClient` の使い方を推すのに検証側はそちらを使っていない。

**対応内容**
- `helper_test.go` の `newClient` を `iaas.NewClient(&sc)` 経由に差し替え。
- `securitySource` / `baseTransport` / `dumpTransport` を削除。認証・UA・BigInt ヘッダ・
  find query 書き換えは saclient-go のミドルウェアと `iaas.NewClient` の自動組み込みに委譲。
- `SAKURA_TRACE=1` は `saclient.WithTraceMode("all")` に置換（出力形式は `[TRACE]` prefix の
  `log.Printf` に変わるが機能的に等価）。
- `iaas_note_test.go` の `newIaasClient` を削除し、`TestIaasNoteCRUD` は共通 `newClient` を利用。

**スコープ外（残課題）**
- 各テスト本体で `c.NoteOpCreate` 等の raw ogen 呼び出しはそのまま。wrapper メソッド
  (`noteOp.Create`) への本格移行は段階的に別 PR で対応する。
- `v2/client/find_transport.go` の削除は項目 7 で扱う。

### 6. find query 書き換えの二重実装 ✅ 対応済み

**対応内容**
- `v2/client/find_transport.go` と `v2/client/find_transport_test.go` を削除。
  `WithFindQueryRewrite` / `NewFindQueryRewriteTransport` は repo 内に呼び出し元が無く、
  項目 6 で integration test を `iaas.NewClient` 経由（saclient ミドルウェア chain）に
  寄せた段階で実質デッドコードだった。
- 退行検知のため `v2/middleware_test.go` を新設し、`findQueryRewriteMiddleware` を
  saclient.Middleware として直接呼び出す形でテストケースを移植。
- `internal/tools/gen-typespec/{ops.go, main.go}` のコメントを `findQueryRewriteMiddleware`
  / `v2/middleware.go` に更新し、再生成して main.tsp / openapi.yaml に反映。

### 7. ogen `debug/example_tests` で v2/client が肥大

**状況**
- `spec/ogen.yml` が `debug/example_tests` を有効化。
- 生成物: `oas_faker_gen.go` 12,926 行 + `oas_test_examples_gen_test.go` 11,623 行 ≈ 24k 行。
- `v2/client/` 全体の 12.7% を占める。

**影響**
- 公開クライアントとしては不要な肥大。
- IDE のコード補完やビルド時間への影響。

**対応案**
- `ogen.yml` で `debug/example_tests` を disable。
- 生成テストが必要であれば別 build tag（`//go:build ogen_example`）で隔離。

---

## 【中】生成器の保守性

### 8. `gen-typespec` の override 系が散在

**状況**
- `internal/tools/gen-typespec/` の主要ファイル行数:
  `models.go` 849 / `ops.go` 1,017 / `envelopes.go` 576 / `fat_model.go` 277 / `main.go` 409。
  計 ~4,900 行を単一 `main` パッケージに保持。
- override レイヤー: `fieldNullabilityOverrides` / `modelFieldExclusions` /
  `fatModelAlwaysOptionalTop` / `excludedOps` / `fieldmanifest` / `summaryOverrides` /
  `postStatusCodeOverrides` と散在。

**影響**
- 新規リソース追加時、どの override を検討すべきかのチェックリストが AGENTS.md 依存。
- 同一リソース向けの override が複数ファイルにまたがり、全量を把握するには複数ファイルを grep する必要。

**対応案**
- リソース単位の `ResourceOverride` 構造体に集約して 1 箇所で確認可能にする。
- 例:
  ```go
  var resourceOverrides = map[string]ResourceOverride{
      "Database": {
          FieldNullability: ...,
          FieldExclusions:  ...,
          ExcludedOps:      ...,
      },
  }
  ```
- verify-typespec のチェック追加と合わせて移植する。

### 9. `fieldmanifest.Manifest` (893 行) の更新が全手作業

**状況**
- downstream が新フィールドを使い始めるたびに手動で allowlist に追記。
- 現状 893 行の手書き。新規リソース追加時に確実に漏れが出る。

**影響**
- 「downstream で使われ始めたのに emit されない」不整合の検出が遅れる（ビルドが通ってしまう）。
- メンテナンスコストが時間経過とともに単調増加。

**対応案**
- `gen-fieldmanifest` ツールを新設。downstream 3 リポジトリを AST 走査し、
  `iaas.<Type>.Get<Field>()` 呼び出しから参照フィールドを抽出して allowlist の差分を提案する。
- manifest は引き続き手書きだが、差分提案を CI で回すことで追従漏れを検知。

### 10. `@example` の source of truth が不明確

**状況**
- 直近コミット `7f2b528 add examples` で `@example` を追加。
- 一方で `internal/tools/gen-typespec/examples/examples.go` は 1,618 行の生成器。
- どちらが真のソースか、および追加・更新フローが AGENTS.md に明記されていない。

**影響**
- DSL 側が変更された際、example が陳腐化していても気付かない。
- 手動追加の example が再生成で上書きされる / されないの判断が揺れる。

**対応案**
- `gen-typespec/examples/` の責務境界（何を自動生成し、何が手動か）を AGENTS.md に追記。
- 生成で賄えない example は `examples/manual/` のような別ディレクトリに分離。

---

## 【中】API 品質

### 11. `ApiError.status` フィールドが未使用のまま残っている ✅ 対応済み

**対応内容**
- `spec/typespec/main.tsp` の `ApiError` から `status?` を削除。
- `pnpm run lint` で openapi.yaml を再生成、`pnpm run generate:client` で
  `v2/client/oas_schemas_gen.go` の `ApiError.Status` / `GetStatus` / `SetStatus` が消えた。
- v1 `APIError` interface は `Status` accessor を持たず、fake サーバの書き込み以外に consumer が
  無かったため、削除の波及なし。

### 12. Error 型のアクセサ欠落 ✅ 対応済み

**対応内容**
- `v2/error.go` の `*Error` に v1 `APIError` interface 互換のアクセサを実装:
  ```go
  func (e *Error) ResponseCode() int  // *client.ApiErrorStatusCode.StatusCode
  func (e *Error) Code() string       // Response.ErrorCode.Value
  func (e *Error) Message() string    // Response.ErrorMsg.Value
  func (e *Error) Serial() string     // Response.Serial.Value
  ```
- 内部共通ヘルパ `apiErrorStatusCode()` で `errors.As(*client.ApiErrorStatusCode)` を
  一度だけ実施。ネットワークエラー等で取れないケースは 0 / 空文字列を返す。
- 回帰防止テスト `v2/error_test.go` `TestErrorAccessors` /
  `TestErrorAccessorsWithoutAPIError` を追加。
- これにより #1 で提供しなかった `IsStillCreatingError` 相当も downstream 側で
  `err.Code() == "still_creating"` と書けば等価の判定が可能。

### 13. レスポンスに含まれる平文 Password フィールドの整理 ✅ 対応済み

**状況**
- v1 のレスポンスモデルにユーザ設定値の平文 Password が echo されているフィールドが複数残っていた。
  サーバ生成型 (FTPServer, VNCProxyInfo) は性質上残すしかないが、設定値 echo 型は response に
  載せるべきではない。
- 対応した一覧 (gen-typespec の `modelFieldVisibility` で `Lifecycle.Create + Update` に絞り、
  Read レスポンスから除外):

  | フィールド | downstream の response 参照 |
  |---|---|
  | `VPCRouterRemoteAccessUser.Password` | response からは読まれていない (provider は config 由来で復元) |
  | `DatabaseSettingCommon.UserPassword` / `ReplicaPassword` | 参照なし |
  | `DatabaseRemarkDBConfCommon.UserPassword` | 参照なし |
  | `DatabaseReplicationSetting.Password` | 参照なし |
  | `SimpleMonitorHealthCheck.BasicAuthPassword` | terraform-provider-sakuracloud / sakura が response から読んで state へ反映 (v2 移行と合わせて provider 側を config 由来へ書き換え予定) |

  サーバ生成型 (`FTPServer.Password`, `VNCProxyInfo.Password`) は残置。

**実装**
- `internal/tools/gen-typespec/models.go` に `modelFieldVisibility` を追加。
  該当フィールドに `@visibility(Lifecycle.Create, Lifecycle.Update)` を付与する。
- `@typespec/openapi3` は visibility を見て Read 用と Create / Update 用の別 schema
  (`XxxCreate`, `XxxCreateOrUpdate`) を発行する。Read 側は Password を持たない。
- 同 emitter は visibility 変種を別名で emit する都合上、参照不能な orphan model から派生する
  schema 名と衝突する。`spec/tspconfig.yaml` で `omit-unreachable-types: true` を有効にして回避。
  (副作用として、operations から到達不能な request 系 model は OpenAPI / v2 client から落ちる。
  downstream は v2 移行時に request payload を `unknown` 経由で組み立てる前提なので問題なし。)

**残課題**
- `SimpleMonitorHealthCheck.BasicAuthPassword` の Read 反映を terraform-provider-sakuracloud /
  sakura が今でも行っている。v2 移行 PR で provider 側を `VPCRouterRemoteAccessUser` 等と同じく
  config 由来で復元する方式へ揃える必要がある。

---

## 【低】ドキュメント・運用

### 14. リリース・バージョニング方針が README 記述のみ

**状況**
- `v2/go.mod` は独立モジュール。
- semver tagging (`v2.0.0-alpha.N` 等) / CHANGELOG 分離 / deprecation window / v1 サポート期限
  が README レベルの記述のみ。

**影響**
- downstream が pinning する際の目安が無い。
- v1 の保守期限が曖昧なため helper 層の継続可否が決まらない（項目 4 に連動）。

**対応案**
- `RELEASING.md` または AGENTS.md にリリース方針と v1 サポート期限を明記。
- `v2.0.0-alpha.1` 等の tag をまず切って downstream が import 可能にする。

### 15. Experimental 宣言が README だけ

**状況**
- `v2/README.md` 冒頭で「v1.0 に達するまで互換性のない変更があり得る」と注意。
- Go doc コメント（`v2/doc.go` / 各 public API）には明示が無い。

**影響**
- godoc 経由で API を参照するユーザに警告が届かない。

**対応案**
- `v2/doc.go`（新設 or 既存）に `// Experimental` または `// Stability: beta` を記載。
- 主要 public API（`NewClient`, `New<Resource>Op`）のコメントにも同文脈を追加。

### 16. AGENTS.md が 717 行の単一ファイル

**状況**
- 内容: パイプライン概要 / 設計判断 / wire 規約 / テスト方針 / 実装しないエンドポイント一覧 /
  ラッパー参考設計 / ogen 警告事典 などが混在。

**影響**
- 新規参画者・自分自身の再参照時に読み戻しコストが高い。
- 今後各トピックが独立に育つと 1,000 行を超えて編集衝突が増える。

**対応案**
- 分割案:
  - `docs/arch/pipeline.md` … DSL → TypeSpec → OpenAPI → ogen
  - `docs/arch/wire-format.md` … Find `?q=` 書き換え、BigInt ヘッダ、ステータスコード
  - `docs/arch/downstream.md` … fieldmanifest / excluded_ops / iaas-service-go 扱い
  - `docs/arch/testing.md` … integration テスト、sandbox 制約
  - `docs/arch/wrapper.md` … saclient-go 統合・Op 層方針
- AGENTS.md は各文書への index + 最重要方針のみに絞る。

### 17. TypeSpec を source of truth に昇格するロードマップ未定

**状況**
- AGENTS.md は「v1 DSL が現時点の真」「将来 TypeSpec を外に出す」の両方を記載。
- 切替条件・時期が未記述。

**影響**
- 「どちらを編集するか」の議論が繰り返される。
- fieldmanifest / excluded_ops の位置付け（暫定か恒久か）が定まらない。

**対応案**
- 切替条件（例: 「全 downstream が v2 の公開 API に切り替わった時点で Go DSL を凍結」）を明記。
- 切替後の構成: `spec/typespec/` を手書き編集、`internal/tools/gen-typespec/` を廃止、
  `fieldmanifest` は OpenAPI lint レベルのチェックに置換、といった像まで描いておく。

---

## 総括

生成パイプラインと設計判断は筋が通っており、AGENTS.md に理由が残っている点で、ツール面の
メンテナンス性は高い。一方で **downstream が実際に触る API 表面の ergonomics が未整備** で、
特に最優先の 4 項目が決まるまでは terraform v3 / usacloud の v2 移行で per-resource PR が
肥大する。

推奨する作業順序:

1. 最優先 1〜4 の方針を決定（特に 3 `helper/*` は意思決定のみ）
2. 高 5〜7（test / middleware / ogen config）を整理し、v2 公開前提の線を固める
3. 生成器側（中 8〜10）は後回しでよい。面積が広がっても移行ブロッカーにはならない
4. ドキュメント（低 14〜17）はリリース判断前にまとめて更新

新 resource の integration test 追加は上記 1〜2 が片付いた後に再開したほうが、
後戻りコストが小さい。
