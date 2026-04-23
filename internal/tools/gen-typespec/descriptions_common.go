// Copyright 2022-2025 The sacloud/iaas-api-go Authors
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

// commonFieldDescriptions は全リソースのモデルで共通に使える汎用フィールドの説明。
// lookupFieldDescription はモデル固有の説明を優先し、そこに該当が無ければここを参照する。
// 「型だけ見れば意味が自明」なフィールドでも、UI やドキュメント上で文言が揃っている方が
// 読み手の負担が減るため、ごく基本的な共通フィールドはここで面倒を見る。
var commonFieldDescriptions = map[string]string{
	// リソース共通のメタ情報
	"ID":           "リソースを一意に識別するID",
	"Name":         "リソースの名前",
	"Description":  "説明",
	"Tags":         "タグの配列。リソースのグルーピングや検索に利用する",
	"Icon":         "アイコン (ResourceRef で ID を指定、null で解除)",
	"CreatedAt":    "作成日時 (ISO 8601 / UTC)",
	"ModifiedAt":   "最終更新日時 (ISO 8601 / UTC)",
	"Availability": "リソースの利用状態。\"available\" で利用可能、\"migrating\" で移行中、\"uploading\" でアップロード中など",
	"Class":        "リソースの内部分類 (API 種別ごとの識別子)",
	"Index":        "配列内での順序を表すインデックス",
	"Scope":        "スコープ。\"shared\" = 共有、\"user\" = ユーザ占有",
	"SettingsHash": "設定ハッシュ値。楽観的排他制御のキーとして更新時にクライアントが引き継いで送る",
	"ServiceClass": "サービスクラス名。課金単価を特定する分類",
	"DisplayOrder": "表示順序",

	// ページング/検索系 (FindRequestEnvelope 共通)
	"Count":  "取得する件数の上限 (ページングサイズ)",
	"From":   "取得開始位置 (0 始まりのオフセット)",
	"Filter": "検索条件を表す JSON オブジェクト。キーは \"{フィールド名}\" の形式で、値に完全一致や部分一致の条件を指定する",
	"Total":  "検索にヒットした総件数",

	// Appliance / CommonServiceItem 系共通 (wrapped in request envelope)
	"CommonServiceItem": "CommonServiceItem (各種付加サービス) の設定ペイロード",
	"Appliance":         "アプライアンス (VPC ルータ、ロードバランサ等) の設定ペイロード",

	// ネットワーク共通
	"IPAddress":      "IPv4 アドレス",
	"IPAddresses":    "IPv4 アドレスの配列",
	"MACAddress":     "MAC アドレス",
	"NetworkAddress": "ネットワークアドレス (IPv4)",
	"NetworkMaskLen": "ネットワークマスク長 (プレフィックス長)",
	"DefaultRoute":   "デフォルトルートの IP アドレス",
	"IPv6Prefix":     "IPv6 プレフィックス",
	"IPv6PrefixLen":  "IPv6 プレフィックス長",
	"NextHop":        "ネクストホップ IP アドレス",
	"StaticRoute":    "静的ルート先 IP アドレス",
	"Hostname":       "ホスト名",
	"HostName":       "ホスト名",
	"Port":           "ポート番号",
	"Protocol":       "プロトコル (\"tcp\", \"udp\", \"http\", \"https\" など)",

	// リクエスト/レスポンス envelope 共通
	"is_ok": "オペレーションが成功したかどうかを示すフラグ。成功判定にはこのフィールドを用いること",

	// FTP 接続情報共通
	"User":     "FTP 接続のユーザ名",
	"Password": "FTP 接続のパスワード",
	"ChangePassword": "FTP のパスワードを再発行するかどうか。true で新しいパスワードを採番する",
}
