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

### 1. `types` パッケージ相当が v2 に無い

**状況**
- v1 `github.com/sacloud/iaas-api-go/types` の `ID` / `StringNumber` / `StringFlag` / 各 enum
  (`EDayOfTheWeek`, `SimpleMonitorProtocols` 等) を downstream が多用。
- 例: `terraform-provider-sakuracloud` の `sakuracloud/` 配下だけで
  `types.ID` 36 回、`types.StringNumber` 24 回、`types.StringFlag` 18 回、
  `SimpleMonitorProtocols` 21 回 などを参照。
- v2 は ogen 生成の `client.OptNilString` / `client.OptNilInt64` を直接露出しており、
  downstream が v1 `types.ID` 形式に変換しなおすループが必要になる。

**影響**
- terraform v3 / usacloud の移行時、`types.ID(x.ID.Value)` のような詰め替えがほぼ全ファイルで発生。
- 新 enum (例: ストレージクラス追加) の追従が v1 と v2 で二重管理になる可能性。

**対応案**
- A. v1 `types` パッケージを v2 からも import し続けさせる（後方互換宣言）。
- B. `v2/types/` に v1 互換の alias を再定義し、v1 の enum 値を再 export する。
- いずれを採るかを AGENTS.md に宣言。現実的には A のほうがメンテ負担が小さい。

### 2. `iaas.IsNotFoundError` 等のエラー判定ヘルパが v2 に無い

**状況**
- v1 `iaas.IsNotFoundError(err)` / `IsStillCreatingError(err)` / `IsNoResultsError(err)` を
  downstream が多用（`terraform-provider-sakuracloud` のリソース削除ハンドラなどで 40+ 箇所）。
- v2 は `saclient.IsNotFoundError(err)` に委譲する設計 (`v2/error.go`) で、`iaas.` 名前空間に
  wrapper が無い。

**影響**
- 移行時に `iaas.IsNotFoundError` を `saclient.IsNotFoundError` に全置換する必要。
- `IsStillCreatingError` は saclient 側に相当物が無い（`error_code == "still_creating"` の判定）。

**対応案**
- `v2/error.go` に以下を追加:
  ```go
  func IsNotFoundError(err error) bool       // saclient.IsNotFoundError に委譲
  func IsStillCreatingError(err error) bool  // ApiErrorStatusCode の ErrorCode で判定
  func IsNoResultsError(err error) bool      // 必要なら v2 独自に定義
  ```
- 合わせて `v2.Error` に `Code() / Message() / Serial() / ResponseCode()` アクセサを実装し、
  `*client.ApiErrorStatusCode` を内部で `errors.As` で unwrap する。v1 `APIError` interface
  と同形にすれば downstream 移行が機械的になる。

### 3. `OptNilString` / `OptNilInt64` 氾濫の ergonomics ✅ 一次対応済み

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

### 4. `helper/{power, query, wait, plans, cleanup}` の v2 相当が未着手

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

### 5. ID パス引数の型不整合

**状況**
- TypeSpec: path parameter は `id: string`、レスポンスは `ID?: int64 | null`。
- ラッパー層の Read/Update/Delete も `id string` を受ける。
- Create 直後の ID 伝搬で毎回 `fmt.Sprintf("%d", createResp.Note.ID.Value)` が必要。

**影響**
- 利用側コードの毎メソッドに string 化が入る。
- `int64` と `string` を取り違えた型エラーが出ない分、引数順ミス等のバグが気付きにくい。

**対応案**
- TypeSpec の path `id` を `int64` に変更（実 API は数値として扱うので正しい）。
- ラッパーの引数も `int64` に統一。フォーマットが必要な表示用途は呼び出し側で処理。
- 破壊的変更なのでラッパー層の API 凍結前にやる。

---

## 【高】パイプライン整合性

### 6. integration tests が 2 系統並立 ✅ 対応済み

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

### 7. find query 書き換えの二重実装

**状況**
- `v2/client/find_transport.go`（http.RoundTripper）と `v2/middleware.go` の
  `findQueryRewriteMiddleware`（saclient.Middleware）が同一ロジックを 2 系統で持っている。

**影響**
- 書き換えロジックを修正する際の反映漏れリスク。
- 利用側が「どちらを使うべきか」判断に迷う。

**対応案**
- 項目 6 と同時に saclient ベースに統一し、`v2/client/find_transport.go` を削除。
- もし両方を残す判断なら AGENTS.md に「どちらをどの場面で使うか」を明記。

### 8. ogen `debug/example_tests` で v2/client が肥大

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

### 9. `gen-typespec` の override 系が散在

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

### 10. `fieldmanifest.Manifest` (893 行) の更新が全手作業

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

### 11. `@example` の source of truth が不明確

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

### 12. `ApiError.status` フィールドが未使用のまま残っている

**状況**
- `spec/typespec/main.tsp` の `ApiError` が `{is_fatal, serial, status, error_code, error_msg}`
  の 5 フィールド構成。
- `status` は実 API が "error" リテラルを入れるだけで、v1 downstream でも分岐に使われていない。

**影響**
- TypeSpec / ogen 生成物に必要性の薄いフィールドが残る。

**対応案**
- 削除する。削除の影響は ogen 生成の `ApiError` 構造体のみで、downstream には波及しない。

### 13. Error 型のアクセサ欠落

**状況**
- v1 `APIError` は `Code() / Message() / Serial() / ResponseCode()` を interface で公開。
- v2 `*iaas.Error` は `Error()` と `Unwrap()` のみ。エラーコード取得は
  `errors.As(*client.ApiErrorStatusCode)` 経由で `e.Response.ErrorCode.Value` まで辿る必要。

**影響**
- downstream が「iaas パッケージのエラー判定」を書く際の記述量が増える。
- v1 からの移行時に per-callsite での書き換えが必要。

**対応案**
- `v2/error.go` にアクセサを実装:
  ```go
  func (e *Error) ResponseCode() int
  func (e *Error) Code() string
  func (e *Error) Message() string
  func (e *Error) Serial() string
  ```
- 内部では `*client.ApiErrorStatusCode` を `errors.As` で unwrap。

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
特に最優先の 5 項目が決まるまでは terraform v3 / usacloud の v2 移行で per-resource PR が
肥大する。

推奨する作業順序:

1. 最優先 1〜5 の方針を決定（特に 1 `types`, 4 `helper/*` は意思決定のみ）
2. 高 6〜8（test / middleware / ogen config）を整理し、v2 公開前提の線を固める
3. 生成器側（中 9〜11）は後回しでよい。面積が広がっても移行ブロッカーにはならない
4. ドキュメント（低 14〜17）はリリース判断前にまとめて更新

新 resource の integration test 追加は上記 1〜2 が片付いた後に再開したほうが、
後戻りコストが小さい。
