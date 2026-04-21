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

// Package examples はオペレーションごとのレスポンス例を保持する。
//
// gen-typespec が各オペレーションの生成時に Manifest を参照し、該当エントリがあれば
// TypeSpec の @opExample デコレータを出力する。最終的に OpenAPI の examples に落ち、
// Redoc / Redocly のドキュメントに "Response samples" として表示される。
//
// # 追加の流れ
//
//  1. 実 API を叩いてレスポンスを収集する:
//
//     TESTACC=1 SAKURACLOUD_TRACE=1 go test ./test -run TestArchiveOpCRUD -v 2>&1 | tee /tmp/trace.log
//
//  2. ログから `[TRACE] <Resource>API.<Op> end` の直前にある `[TRACE] results: {...}` を抽出する。
//
//  3. 下記マスキング指針に沿ってダミー値に置換する。
//
//  4. Manifest に `"<Resource>.<Op>": {{Response: `{ ... }`}}` の形で追記する。
//     キーは DSL の resource.TypeName() + "." + op.Name（パスカルケース）。
//
// # マスキング指針
//
// 公開ドキュメントに載るため、実アカウント由来の値は必ずダミーに置換する。
//
//   - 数値フィールド（AccountID, MemberCode など）は文字列化せず数値のまま、ただし明らかに
//     ダミー感のある値にする。12 桁 ID なら 123456789012, 123456789013, ... のような連番。
//   - IPv4 はドキュメント用途の TEST-NET 帯を使う（RFC 5737）:
//     192.0.2.0/24 / 198.51.100.0/24 / 203.0.113.0/24。
//     プレースホルダ表記 (192.0.2.x) はやめて、実在しうるアドレスを具体値で書く（例 192.0.2.10）。
//   - IPv6 は 2001:db8::/32（RFC 3849）。
//   - ホスト名・ドメインは example.com / example.net / example.jp。
//   - MAC アドレスは 00:00:5E:00:53:00〜FF（RFC 7042）。
//   - パスワード・トークン系はダミーっぽい綴りで固定する:
//     "dummy-password", "dummy-ftp-password", "dummy-access-token" など。
//     本物の値をそのまま書かない。
//   - タイムスタンプは固定値で良い: "2025-01-01T00:00:00+09:00"。
//   - メールアドレスは user@example.com 系。
//
// # エントリの粒度
//
// 現状 1 op につき 1 example を想定しているが、Manifest の値はスライスなので将来
// 「空リスト」「フィルタヒット」などパターンを増やせる。Title を付けると Redoc 側の
// 表示タブ名になる。
package examples

// Example は単一のレスポンス例を表す。
type Example struct {
	// Title は Redoc のレスポンス例タブに表示される短いラベル（省略可）。
	// 同一オペレーションに複数 Example を登録するときは衝突しないよう一意にする。
	Title string

	// Description は補足説明（省略可）。OpenAPI の examples.<name>.description に落ちる。
	Description string

	// Response は実レスポンスの JSON 文字列（エンベロープ丸ごと）。
	// gen-typespec が TypeSpec の value literal に変換して @opExample(#{returnType: ...}) に流し込む。
	Response string
}

// Manifest はキー "<Resource>.<Op>" → Example 配列のマップ。
//
// キーの組み立て方:
//   - 個別リソース（Archive / Server 等）: "<resource.TypeName()>.<op.Name>"
//     例: "Archive.Read", "Server.Find", "Disk.Create"
//   - 共有グループ（CommonServiceItem / Appliance）: グループ名で登録する。
//     例: "Appliance.Find", "CommonServiceItem.Read"
//     グループ配下の個別リソース（Database, VPCRouter 等）固有エンドポイントは
//     "<Resource>.<Op>" を使う。例: "VPCRouter.Status"。
//
// 新規追加時はキーの書き方を既存エントリに倣うこと。キーミスはビルドエラーにならず、
// 単に example が出力されないだけなので注意。
var Manifest = map[string][]Example{
	// --- Icon ---
	// TestIconCRUD のトレース（TESTACC=1 SAKURA_TRACE=1）を元に作成。
	// ID は 12 桁ダミー（123456789012）、共有アイコンの例は実運用で常在するもの
	// （sakura cloud 全ユーザーに公開される category-feature アイコン）をそのまま載せている。
	"Icon.Find": {{
		Response: `{
			"Total": 3,
			"From": 0,
			"Count": 3,
			"Icons": [
				{"ID": 112300511380, "Name": "CGI", "Tags": ["category-feature"], "URL": "https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/icon/112300511380.png"},
				{"ID": 112300511382, "Name": "DNS", "Tags": ["category-feature"], "URL": "https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/icon/112300511382.png"},
				{"ID": 123456789012, "Name": "example-icon", "Tags": ["example", "dummy"], "URL": "https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/icon/123456789012.png"}
			]
		}`,
	}},
	"Icon.Create": {{
		Response: `{
			"is_ok": true,
			"Icon": {"ID": 123456789012, "Name": "example-icon", "Tags": ["example", "dummy"], "URL": "https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/icon/123456789012.png"}
		}`,
	}},
	"Icon.Read": {{
		Response: `{
			"is_ok": true,
			"Icon": {"ID": 123456789012, "Name": "example-icon", "Tags": ["example", "dummy"], "URL": "https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/icon/123456789012.png"}
		}`,
	}},
	"Icon.Update": {{
		Response: `{
			"is_ok": true,
			"Icon": {"ID": 123456789012, "Name": "example-icon-updated", "Tags": ["example", "dummy", "updated"], "URL": "https://secure.sakura.ad.jp/cloud/zone/tk1v/api/cloud/1.1/icon/123456789012.png"}
		}`,
	}},

	// --- Region ---
	// 東京 / 石狩 / Sandbox ゾーンは sakura cloud の公開情報なので実値をそのまま使う。
	"Region.Find": {{
		Response: `{
			"Total": 2,
			"From": 0,
			"Count": 2,
			"Regions": [
				{"ID": 210, "Name": "東京", "Description": "東京", "NameServers": ["210.188.224.10", "210.188.224.11"]},
				{"ID": 310, "Name": "石狩", "Description": "石狩", "NameServers": ["133.242.0.3", "133.242.0.4"]}
			]
		}`,
	}},
	"Region.Read": {{
		Response: `{
			"is_ok": true,
			"Region": {"ID": 310, "Name": "石狩", "Description": "石狩", "NameServers": ["133.242.0.3", "133.242.0.4"]}
		}`,
	}},

	// --- Zone ---
	"Zone.Find": {{
		Response: `{
			"Total": 2,
			"From": 0,
			"Count": 2,
			"Zones": [
				{"ID": 21001, "Name": "tk1a", "Description": "東京第1ゾーン", "Region": {"ID": 210, "Name": "東京", "Description": "東京", "NameServers": ["210.188.224.10", "210.188.224.11"]}},
				{"ID": 31001, "Name": "is1a", "Description": "石狩第1ゾーン", "Region": {"ID": 310, "Name": "石狩", "Description": "石狩", "NameServers": ["133.242.0.3", "133.242.0.4"]}}
			]
		}`,
	}},
	"Zone.Read": {{
		Response: `{
			"is_ok": true,
			"Zone": {"ID": 31001, "Name": "is1a", "Description": "石狩第1ゾーン", "Region": {"ID": 310, "Name": "石狩", "Description": "石狩", "NameServers": ["133.242.0.3", "133.242.0.4"]}}
		}`,
	}},

	// --- Bridge ---
	// TestBridgeCRUD のトレースを元に作成。
	"Bridge.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Bridges": [
				{"ID": 123456789012, "Name": "example-bridge", "Description": "ドキュメント用ダミーブリッジ"}
			]
		}`,
	}},
	"Bridge.Create": {{
		Response: `{
			"is_ok": true,
			"Bridge": {"ID": 123456789012, "Name": "example-bridge", "Description": "ドキュメント用ダミーブリッジ"}
		}`,
	}},
	"Bridge.Read": {{
		Response: `{
			"is_ok": true,
			"Bridge": {"ID": 123456789012, "Name": "example-bridge", "Description": "ドキュメント用ダミーブリッジ"}
		}`,
	}},
	"Bridge.Update": {{
		Response: `{
			"is_ok": true,
			"Bridge": {"ID": 123456789012, "Name": "example-bridge-updated", "Description": "更新後の説明"}
		}`,
	}},

	// --- SSHKey ---
	// TestSSHKeyCRUD のトレースを元に作成。PublicKey はダミーの ed25519 鍵。
	"SSHKey.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"SSHKeys": [
				{
					"ID": 123456789012,
					"Name": "example-sshkey",
					"Description": "ドキュメント用ダミー公開鍵",
					"PublicKey": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleDummyPublicKeyDoNotUseInProduction example@example.com",
					"Fingerprint": "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
				}
			]
		}`,
	}},
	"SSHKey.Create": {{
		Response: `{
			"is_ok": true,
			"SSHKey": {
				"ID": 123456789012,
				"Name": "example-sshkey",
				"Description": "ドキュメント用ダミー公開鍵",
				"PublicKey": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleDummyPublicKeyDoNotUseInProduction example@example.com",
				"Fingerprint": "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
			}
		}`,
	}},
	"SSHKey.Read": {{
		Response: `{
			"is_ok": true,
			"SSHKey": {
				"ID": 123456789012,
				"Name": "example-sshkey",
				"Description": "ドキュメント用ダミー公開鍵",
				"PublicKey": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleDummyPublicKeyDoNotUseInProduction example@example.com",
				"Fingerprint": "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
			}
		}`,
	}},
	"SSHKey.Update": {{
		Response: `{
			"is_ok": true,
			"SSHKey": {
				"ID": 123456789012,
				"Name": "example-sshkey-updated",
				"Description": "更新後の説明",
				"PublicKey": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExampleDummyPublicKeyDoNotUseInProduction example@example.com",
				"Fingerprint": "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
			}
		}`,
	}},

	// --- Switch ---
	// TestSwitchCRUD のトレースを元に作成。Switch モデルで保持しないネスト（Zone/Region/Internet 等）は除外。
	"Switch.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Switches": [
				{
					"ID": 123456789012,
					"Name": "example-switch",
					"Description": "ドキュメント用ダミースイッチ",
					"Tags": ["example"],
					"ServerCount": 0,
					"Subnets": [],
					"Icon": null,
					"Bridge": null
				}
			]
		}`,
	}},
	"Switch.Create": {{
		Response: `{
			"is_ok": true,
			"Switch": {
				"ID": 123456789012,
				"Name": "example-switch",
				"Description": "ドキュメント用ダミースイッチ",
				"Tags": ["example"],
				"ServerCount": 0,
				"Subnets": [],
				"Icon": null,
				"Bridge": null
			}
		}`,
	}},
	"Switch.Read": {{
		Response: `{
			"is_ok": true,
			"Switch": {
				"ID": 123456789012,
				"Name": "example-switch",
				"Description": "ドキュメント用ダミースイッチ",
				"Tags": ["example"],
				"ServerCount": 0,
				"Subnets": [],
				"Icon": null,
				"Bridge": null
			}
		}`,
	}},
	"Switch.Update": {{
		Response: `{
			"is_ok": true,
			"Switch": {
				"ID": 123456789012,
				"Name": "example-switch-updated",
				"Description": "更新後の説明",
				"Tags": ["example", "updated"],
				"ServerCount": 0,
				"Subnets": [],
				"Icon": null,
				"Bridge": null
			}
		}`,
	}},

	// --- Note (スタートアップスクリプト) ---
	// TestIaasNoteCRUD のトレースを元に作成（テスト自体は assertion 失敗したが wire response は成功）。
	"Note.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Notes": [
				{
					"ID": 123456789012,
					"Name": "example-note",
					"Description": "ドキュメント用ダミースクリプト",
					"Tags": ["example"],
					"Class": "shell",
					"Content": "#!/bin/bash\necho hello",
					"Icon": null
				}
			]
		}`,
	}},
	"Note.Create": {{
		Response: `{
			"is_ok": true,
			"Note": {
				"ID": 123456789012,
				"Name": "example-note",
				"Description": "",
				"Tags": ["example"],
				"Class": "shell",
				"Content": "#!/bin/bash\necho hello",
				"Icon": null
			}
		}`,
	}},
	"Note.Read": {{
		Response: `{
			"is_ok": true,
			"Note": {
				"ID": 123456789012,
				"Name": "example-note",
				"Description": "",
				"Tags": ["example"],
				"Class": "shell",
				"Content": "#!/bin/bash\necho hello",
				"Icon": null
			}
		}`,
	}},
	"Note.Update": {{
		Response: `{
			"is_ok": true,
			"Note": {
				"ID": 123456789012,
				"Name": "example-note-updated",
				"Description": "更新後の説明",
				"Tags": ["example", "updated"],
				"Class": "shell",
				"Content": "#!/bin/bash\necho updated",
				"Icon": null
			}
		}`,
	}},

	// --- Archive ---
	// TestArchiveCRUD / TestArchiveFindWithQuery のトレースを元に作成。
	// Description は HTML 埋め込みなど長大なので簡潔なダミー文にしている。
	"Archive.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Archives": [
				{
					"ID": 123456789012,
					"Name": "example-archive",
					"Description": "ドキュメント用ダミーアーカイブ",
					"Tags": ["example"],
					"Availability": "available",
					"SizeMB": 20480,
					"Icon": null
				}
			]
		}`,
	}},
	"Archive.Create": {{
		Response: `{
			"is_ok": true,
			"Archive": {
				"ID": 123456789012,
				"Name": "example-archive",
				"Description": "ドキュメント用ダミーアーカイブ",
				"Tags": ["example"],
				"Availability": "uploading",
				"SizeMB": 20480,
				"Icon": null
			}
		}`,
	}},
	"Archive.Read": {{
		Response: `{
			"is_ok": true,
			"Archive": {
				"ID": 123456789012,
				"Name": "example-archive",
				"Description": "ドキュメント用ダミーアーカイブ",
				"Tags": ["example"],
				"Availability": "available",
				"SizeMB": 20480,
				"Icon": null
			}
		}`,
	}},
	"Archive.Update": {{
		Response: `{
			"is_ok": true,
			"Archive": {
				"ID": 123456789012,
				"Name": "example-archive-updated",
				"Description": "更新後の説明",
				"Tags": ["example", "updated"],
				"Availability": "available",
				"SizeMB": 20480,
				"Icon": null
			}
		}`,
	}},

	// --- PrivateHostPlan ---
	// PrivateHostPlan は公開情報（プラン ID と CPU/メモリ情報）なので実 ID をそのまま使用。
	"PrivateHostPlan.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"PrivateHostPlans": [
				{
					"ID": 112900526366,
					"Name": "200Core 224GB 標準",
					"Class": "dynamic",
					"CPU": 200,
					"Dedicated": false,
					"MemoryMB": 229376,
					"Availability": "available"
				}
			]
		}`,
	}},
	"PrivateHostPlan.Read": {{
		Response: `{
			"is_ok": true,
			"PrivateHostPlan": {
				"ID": 112900526366,
				"Name": "200Core 224GB 標準",
				"Class": "dynamic",
				"CPU": 200,
				"Dedicated": false,
				"MemoryMB": 229376,
				"Availability": "available"
			}
		}`,
	}},

	// --- CDROM (ISO イメージ) ---
	"CDROM.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"CDROMs": [
				{
					"ID": 123456789012,
					"Name": "example-cdrom",
					"Description": "ドキュメント用ダミー ISO",
					"Tags": ["example"],
					"Availability": "available",
					"Icon": null
				}
			]
		}`,
	}},
	"CDROM.Create": {{
		Response: `{
			"is_ok": true,
			"CDROM": {
				"ID": 123456789012,
				"Name": "example-cdrom",
				"Description": "ドキュメント用ダミー ISO",
				"Tags": ["example"],
				"Availability": "uploading",
				"Icon": null
			},
			"FTPServer": {
				"HostName": "sac-tk1v-ftp.example.jp",
				"IPAddress": "192.0.2.1",
				"User": "cdrom123456789012",
				"Password": "dummy-ftp-password"
			}
		}`,
	}},
	"CDROM.Read": {{
		Response: `{
			"is_ok": true,
			"CDROM": {
				"ID": 123456789012,
				"Name": "example-cdrom",
				"Description": "ドキュメント用ダミー ISO",
				"Tags": ["example"],
				"Availability": "available",
				"Icon": null
			}
		}`,
	}},
	"CDROM.Update": {{
		Response: `{
			"is_ok": true,
			"CDROM": {
				"ID": 123456789012,
				"Name": "example-cdrom-updated",
				"Description": "更新後の説明",
				"Tags": ["example", "updated"],
				"Availability": "available",
				"Icon": null
			}
		}`,
	}},

	// --- DiskPlan ---
	// プラン情報は公開情報なので実 ID（2=標準, 4=SSD）をそのまま使用。
	"DiskPlan.Find": {{
		Response: `{
			"Total": 2,
			"From": 0,
			"Count": 2,
			"DiskPlans": [
				{
					"ID": 4,
					"Name": "SSDプラン",
					"StorageClass": "iscsi1204",
					"Availability": "available",
					"Size": [
						{"SizeMB": 20480, "DisplaySize": 20, "DisplaySuffix": "GB", "Availability": "available"},
						{"SizeMB": 40960, "DisplaySize": 40, "DisplaySuffix": "GB", "Availability": "available"}
					]
				},
				{
					"ID": 2,
					"Name": "標準プラン",
					"StorageClass": "iscsi1204",
					"Availability": "available",
					"Size": [
						{"SizeMB": 20480, "DisplaySize": 20, "DisplaySuffix": "GB", "Availability": "available"}
					]
				}
			]
		}`,
	}},
	"DiskPlan.Read": {{
		Response: `{
			"is_ok": true,
			"DiskPlan": {
				"ID": 4,
				"Name": "SSDプラン",
				"StorageClass": "iscsi1204",
				"Availability": "available",
				"Size": [
					{"SizeMB": 20480, "DisplaySize": 20, "DisplaySuffix": "GB", "Availability": "available"},
					{"SizeMB": 40960, "DisplaySize": 40, "DisplaySuffix": "GB", "Availability": "available"}
				]
			}
		}`,
	}},

	// --- InternetPlan (ルータ帯域プラン) ---
	"InternetPlan.Find": {{
		Response: `{
			"Total": 3,
			"From": 0,
			"Count": 3,
			"InternetPlans": [
				{"ID": 100, "Name": "100Mbps共有", "BandWidthMbps": 100, "Availability": "available"},
				{"ID": 250, "Name": "250Mbps共有", "BandWidthMbps": 250, "Availability": "available"},
				{"ID": 500, "Name": "500Mbps共有", "BandWidthMbps": 500, "Availability": "available"}
			]
		}`,
	}},
	"InternetPlan.Read": {{
		Response: `{
			"is_ok": true,
			"InternetPlan": {"ID": 100, "Name": "100Mbps共有", "BandWidthMbps": 100, "Availability": "available"}
		}`,
	}},

	// --- ServerPlan ---
	"ServerPlan.Find": {{
		Response: `{
			"Total": 2,
			"From": 0,
			"Count": 2,
			"ServerPlans": [
				{
					"ID": 100001001,
					"Name": "プラン/1Core-1GB",
					"CPU": 1,
					"MemoryMB": 1024,
					"GPU": 0,
					"GPUModel": "none",
					"CPUModel": "uncategorized",
					"Commitment": "standard",
					"Generation": 100,
					"Availability": "available"
				},
				{
					"ID": 100002001,
					"Name": "プラン/1Core-2GB",
					"CPU": 1,
					"MemoryMB": 2048,
					"GPU": 0,
					"GPUModel": "none",
					"CPUModel": "uncategorized",
					"Commitment": "standard",
					"Generation": 100,
					"Availability": "available"
				}
			]
		}`,
	}},
	"ServerPlan.Read": {{
		Response: `{
			"is_ok": true,
			"ServerPlan": {
				"ID": 100001001,
				"Name": "プラン/1Core-1GB",
				"CPU": 1,
				"MemoryMB": 1024,
				"GPU": 0,
				"GPUModel": "none",
				"CPUModel": "uncategorized",
				"Commitment": "standard",
				"Generation": 100,
				"Availability": "available"
			}
		}`,
	}},

	// --- LicenseInfo (ライセンスプラン) ---
	"LicenseInfo.Find": {{
		Response: `{
			"Total": 2,
			"From": 0,
			"Count": 2,
			"LicenseInfo": [
				{"ID": 10001, "Name": "Windows RDS SAL", "TermsOfUse": "1ライセンスにつき、1人のユーザが利用できます。"},
				{"ID": 10011, "Name": "Windows RDS SAL + Office SAL", "TermsOfUse": "1ライセンスにつき、1人のユーザが利用できます。"}
			]
		}`,
	}},
	"LicenseInfo.Read": {{
		Response: `{
			"is_ok": true,
			"LicenseInfo": {"ID": 10001, "Name": "Windows RDS SAL", "TermsOfUse": "1ライセンスにつき、1人のユーザが利用できます。"}
		}`,
	}},

	// --- License ---
	// TestLicenseCRUD のトレースを元に作成。
	"License.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Licenses": [
				{"ID": 123456789012, "Name": "example-license", "LicenseInfo": {"ID": 10001, "Name": "Windows RDS SAL"}}
			]
		}`,
	}},
	"License.Create": {{
		Response: `{
			"is_ok": true,
			"License": {"ID": 123456789012, "Name": "example-license", "LicenseInfo": {"ID": 10001, "Name": "Windows RDS SAL"}}
		}`,
	}},
	"License.Read": {{
		Response: `{
			"is_ok": true,
			"License": {"ID": 123456789012, "Name": "example-license", "LicenseInfo": {"ID": 10001, "Name": "Windows RDS SAL"}}
		}`,
	}},
	"License.Update": {{
		Response: `{
			"is_ok": true,
			"License": {"ID": 123456789012, "Name": "example-license-updated", "LicenseInfo": {"ID": 10001, "Name": "Windows RDS SAL"}}
		}`,
	}},

	// --- Disk ---
	// TestDiskCRUD のトレースを元に作成。巨大な nested（Storage/Zone/Region）は省略。
	"Disk.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Disks": [
				{
					"ID": 123456789012,
					"Name": "example-disk",
					"Description": "ドキュメント用ダミーディスク",
					"Tags": ["example"],
					"Availability": "available",
					"Connection": "virtio",
					"SizeMB": 20480,
					"Plan": {"ID": 4, "Name": "SSDプラン", "StorageClass": "iscsi1204", "Availability": "available"},
					"Icon": null
				}
			]
		}`,
	}},
	"Disk.Create": {{
		Response: `{
			"is_ok": true,
			"Disk": {
				"ID": 123456789012,
				"Name": "example-disk",
				"Description": "ドキュメント用ダミーディスク",
				"Tags": ["example"],
				"Availability": "available",
				"Connection": "virtio",
				"SizeMB": 20480,
				"Plan": {"ID": 4, "Name": "SSDプラン", "StorageClass": "iscsi1204", "Availability": "available"},
				"Icon": null
			}
		}`,
	}},
	"Disk.Read": {{
		Response: `{
			"is_ok": true,
			"Disk": {
				"ID": 123456789012,
				"Name": "example-disk",
				"Description": "ドキュメント用ダミーディスク",
				"Tags": ["example"],
				"Availability": "available",
				"Connection": "virtio",
				"SizeMB": 20480,
				"Plan": {"ID": 4, "Name": "SSDプラン", "StorageClass": "iscsi1204", "Availability": "available"},
				"Icon": null
			}
		}`,
	}},
	"Disk.Update": {{
		Response: `{
			"is_ok": true,
			"Disk": {
				"ID": 123456789012,
				"Name": "example-disk-updated",
				"Description": "更新後の説明",
				"Tags": ["example", "updated"],
				"Availability": "available",
				"Connection": "virtio",
				"SizeMB": 20480,
				"Plan": {"ID": 4, "Name": "SSDプラン", "StorageClass": "iscsi1204", "Availability": "available"},
				"Icon": null
			}
		}`,
	}},

	// --- Interface ---
	// TestInterfaceCRUD のトレースを元に作成。UserIPAddress は TEST-NET-1 (192.0.2.0/24)。
	"Interface.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Interfaces": [
				{
					"ID": 123456789012,
					"UserIPAddress": "192.0.2.10",
					"Switch": {"ID": 123456789013, "Scope": "user"},
					"Server": {"ID": 123456789014}
				}
			]
		}`,
	}},
	"Interface.Create": {{
		Response: `{
			"is_ok": true,
			"Interface": {
				"ID": 123456789012,
				"UserIPAddress": null,
				"Switch": null,
				"Server": {"ID": 123456789014}
			}
		}`,
	}},
	"Interface.Read": {{
		Response: `{
			"is_ok": true,
			"Interface": {
				"ID": 123456789012,
				"UserIPAddress": "192.0.2.10",
				"Switch": {"ID": 123456789013, "Scope": "user"},
				"Server": {"ID": 123456789014}
			}
		}`,
	}},
	"Interface.Update": {{
		Response: `{
			"is_ok": true,
			"Interface": {
				"ID": 123456789012,
				"UserIPAddress": "192.0.2.10",
				"Switch": {"ID": 123456789013, "Scope": "user"},
				"Server": {"ID": 123456789014}
			}
		}`,
	}},

	// --- Internet (ルータ+スイッチ) ---
	// TestInternetCRUD のトレースを元に作成。Switch のネストは最小限。
	"Internet.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Internet": [
				{
					"ID": 123456789012,
					"Name": "example-internet",
					"Description": "ドキュメント用ダミールータ",
					"Tags": ["example"],
					"BandWidthMbps": 100,
					"NetworkMaskLen": 28,
					"Switch": {"ID": 123456789013, "Name": "example-internet"},
					"Icon": null
				}
			]
		}`,
	}},
	"Internet.Create": {{
		Response: `{
			"is_ok": true,
			"Internet": {
				"ID": 123456789012,
				"Name": "example-internet",
				"Description": "ドキュメント用ダミールータ",
				"Tags": ["example"],
				"BandWidthMbps": 100,
				"NetworkMaskLen": 28,
				"Switch": {"ID": 123456789013, "Name": "example-internet"},
				"Icon": null
			}
		}`,
	}},
	"Internet.Read": {{
		Response: `{
			"is_ok": true,
			"Internet": {
				"ID": 123456789012,
				"Name": "example-internet",
				"Description": "ドキュメント用ダミールータ",
				"Tags": ["example"],
				"BandWidthMbps": 100,
				"NetworkMaskLen": 28,
				"Switch": {"ID": 123456789013, "Name": "example-internet"},
				"Icon": null
			}
		}`,
	}},
	"Internet.Update": {{
		Response: `{
			"is_ok": true,
			"Internet": {
				"ID": 123456789012,
				"Name": "example-internet-updated",
				"Description": "更新後の説明",
				"Tags": ["example", "updated"],
				"BandWidthMbps": 100,
				"NetworkMaskLen": 28,
				"Switch": {"ID": 123456789013, "Name": "example-internet-updated"},
				"Icon": null
			}
		}`,
	}},

	// --- Server ---
	// TestServerCRUD のトレースを元に作成。Instance/Disks/Interfaces 等の大きなネストは省略。
	"Server.Find": {{
		Response: `{
			"Total": 1,
			"From": 0,
			"Count": 1,
			"Servers": [
				{
					"ID": 123456789012,
					"Name": "example-server",
					"Description": "ドキュメント用ダミーサーバ",
					"Tags": ["example"],
					"Availability": "available",
					"HostName": "localhost",
					"InterfaceDriver": "virtio",
					"ServerPlan": {
						"ID": 200001001,
						"Name": "プラン/1Core-1GB(新プラン)",
						"CPU": 1,
						"MemoryMB": 1024,
						"GPU": 0,
						"GPUModel": "none",
						"CPUModel": "uncategorized",
						"Commitment": "standard",
						"Generation": 200,
						"ConfidentialVM": false
					},
					"Icon": null
				}
			]
		}`,
	}},
	"Server.Create": {{
		Response: `{
			"is_ok": true,
			"Server": {
				"ID": 123456789012,
				"Name": "example-server",
				"Description": "ドキュメント用ダミーサーバ",
				"Tags": ["example"],
				"Availability": "available",
				"HostName": "localhost",
				"InterfaceDriver": "virtio",
				"ServerPlan": {
					"ID": 200001001,
					"Name": "プラン/1Core-1GB(新プラン)",
					"CPU": 1,
					"MemoryMB": 1024,
					"GPU": 0,
					"GPUModel": "none",
					"CPUModel": "uncategorized",
					"Commitment": "standard",
					"Generation": 200,
					"ConfidentialVM": false
				},
				"Icon": null
			}
		}`,
	}},
	"Server.Read": {{
		Response: `{
			"is_ok": true,
			"Server": {
				"ID": 123456789012,
				"Name": "example-server",
				"Description": "ドキュメント用ダミーサーバ",
				"Tags": ["example"],
				"Availability": "available",
				"HostName": "localhost",
				"InterfaceDriver": "virtio",
				"ServerPlan": {
					"ID": 200001001,
					"Name": "プラン/1Core-1GB(新プラン)",
					"CPU": 1,
					"MemoryMB": 1024,
					"GPU": 0,
					"GPUModel": "none",
					"CPUModel": "uncategorized",
					"Commitment": "standard",
					"Generation": 200,
					"ConfidentialVM": false
				},
				"Icon": null
			}
		}`,
	}},
	"Server.Update": {{
		Response: `{
			"is_ok": true,
			"Server": {
				"ID": 123456789012,
				"Name": "example-server-updated",
				"Description": "更新後の説明",
				"Tags": ["example", "updated"],
				"Availability": "available",
				"HostName": "localhost",
				"InterfaceDriver": "virtio",
				"ServerPlan": {
					"ID": 200001001,
					"Name": "プラン/1Core-1GB(新プラン)",
					"CPU": 1,
					"MemoryMB": 1024,
					"GPU": 0,
					"GPUModel": "none",
					"CPUModel": "uncategorized",
					"Commitment": "standard",
					"Generation": 200,
					"ConfidentialVM": false
				},
				"Icon": null
			}
		}`,
	}},

	// --- Archive (action ops) ---
	"Archive.Share": {{
		Response: `{"is_ok": true}`,
	}},
	"Archive.Transfer": {{
		Response: `{
			"is_ok": true,
			"Archive": {"ID": 123456789012, "Name": "example-archive", "Description": "転送済みアーカイブ", "Tags": ["example"], "Availability": "available", "SizeMB": 20480, "Icon": null}
		}`,
	}},

	// --- AuthStatus ---
	"AuthStatus.Read": {{
		Response: `{
			"is_ok": true,
			"AuthStatus": {
				"Account": {"ID": 123456789012, "Name": "exampleアカウント", "Code": "example-account", "Class": "member"},
				"Member": {"Code": "example-member", "Class": "member"},
				"IsAPIKey": true
			}
		}`,
	}},

	// --- AutoBackup (commonserviceitem) ---
	"AutoBackup.Find": {{
		Response: `{
			"Total": 0,
			"From": 0,
			"Count": 0,
			"CommonServiceItems": []
		}`,
	}},
	"AutoBackup.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-auto-backup", "Description": "自動バックアップ例", "Tags": ["example"]}}`,
	}},
	"AutoBackup.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-auto-backup", "Description": "自動バックアップ例", "Tags": ["example"]}}`,
	}},
	"AutoBackup.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-auto-backup-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},

	// --- AutoScale (commonserviceitem) ---
	"AutoScale.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"AutoScale.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-auto-scale", "Description": "オートスケール例", "Tags": ["example"]}}`,
	}},
	"AutoScale.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-auto-scale", "Description": "オートスケール例", "Tags": ["example"]}}`,
	}},
	"AutoScale.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-auto-scale-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},
	"AutoScale.Status": {{
		Response: `{"is_ok": true, "AutoScale": {"LatestLogs": ["2025-01-01T00:00:00+09:00 scaled up"], "ResourcesText": "example-resources"}}`,
	}},

	// --- Bill ---
	// Bill 系の実 API は従量課金なので課金ゾーンで実動作。公開情報では無いためダミー値で代用。
	"Bill.Read": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Bills": []}`,
	}},
	"Bill.Details": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "BillDetails": []}`,
	}},
	"Bill.DetailsCSV": {{
		Response: `{"is_ok": true, "CSV": {"HeaderRow": [], "BodyRows": []}}`,
	}},
	"Bill.ByContract": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Bills": []}`,
	}},
	"Bill.ByContractYear": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Bills": []}`,
	}},
	"Bill.ByContractYearMonth": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Bills": []}`,
	}},

	// --- CDROM FTP ---
	"CDROM.OpenFTP": {{
		Response: `{
			"is_ok": true,
			"FTPServer": {"HostName": "sac-tk1v-ftp.example.jp", "IPAddress": "192.0.2.1", "User": "cdrom123456789012", "Password": "dummy-ftp-password"}
		}`,
	}},

	// --- CertificateAuthority (commonserviceitem) ---
	"CertificateAuthority.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"CertificateAuthority.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-ca", "Description": "マネージド CA 例", "Tags": ["example"]}}`,
	}},
	"CertificateAuthority.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-ca", "Description": "マネージド CA 例", "Tags": ["example"]}}`,
	}},
	"CertificateAuthority.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-ca-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},
	"CertificateAuthority.Detail": {{
		Response: `{"is_ok": true, "CertificateAuthority": {"Subject": "CN=example-ca", "CertificateData": {"CertificatePEM": "-----BEGIN CERTIFICATE-----\ndummy\n-----END CERTIFICATE-----"}}}`,
	}},
	"CertificateAuthority.AddClient": {{
		Response: `{"is_ok": true, "CertificateAuthority": {"ID": "cli_000000000001"}}`,
	}},
	"CertificateAuthority.AddServer": {{
		Response: `{"is_ok": true, "CertificateAuthority": {"ID": "srv_000000000001"}}`,
	}},
	"CertificateAuthority.ListClients": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CertificateAuthority": []}`,
	}},
	"CertificateAuthority.ListServers": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CertificateAuthority": []}`,
	}},
	"CertificateAuthority.ReadClient": {{
		Response: `{"is_ok": true, "CertificateAuthority": {"ID": "cli_000000000001", "Subject": "CN=example-client", "IssueState": "available"}}`,
	}},
	"CertificateAuthority.ReadServer": {{
		Response: `{"is_ok": true, "CertificateAuthority": {"ID": "srv_000000000001", "Subject": "CN=example-server", "IssueState": "available"}}`,
	}},

	// --- ContainerRegistry (commonserviceitem) ---
	"ContainerRegistry.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"ContainerRegistry.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-container-registry", "Description": "コンテナレジストリ例", "Tags": ["example"]}}`,
	}},
	"ContainerRegistry.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-container-registry", "Description": "コンテナレジストリ例", "Tags": ["example"]}}`,
	}},
	"ContainerRegistry.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-container-registry-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},
	"ContainerRegistry.ListUsers": {{
		Response: `{"is_ok": true, "ContainerRegistry": {"Users": []}}`,
	}},

	// --- Coupon ---
	"Coupon.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Coupon": []}`,
	}},

	// --- Database (appliance) ---
	// TestDatabaseApplianceCRUD のトレースを元に最小限に。
	"Database.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Appliances": []}`,
	}},
	"Database.Create": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-db", "Description": "データベース例", "Tags": ["example"], "Availability": "migrating", "Class": "database"}}`,
	}},
	"Database.Read": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-db", "Description": "データベース例", "Tags": ["example"], "Availability": "available", "Class": "database"}}`,
	}},
	"Database.Update": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-db-updated", "Description": "更新後", "Tags": ["example", "updated"], "Availability": "available", "Class": "database"}}`,
	}},
	"Database.Status": {{
		Response: `{"is_ok": true, "Appliance": {"SettingsResponse": {"Status": "up", "DBConf": {"MariaDB": {"Status": "running"}, "Postgres": {"Status": "running"}, "Version": {"LastModified": "2025-01-01T00:00:00+09:00", "Status": "running"}}, "IsFatal": false}, "Logs": [], "Backups": []}}`,
	}},
	"Database.MonitorCPU": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"Database.MonitorDatabase": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"Database.MonitorDisk": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"Database.MonitorInterface": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"Database.GetParameter": {{
		Response: `{"is_ok": true, "Database": {"Parameter": null, "MetaInfo": []}}`,
	}},

	// --- DNS (commonserviceitem) ---
	"DNS.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"DNS.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example.com", "Description": "DNS ゾーン例", "Tags": ["example"], "Records": []}}`,
	}},
	"DNS.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example.com", "Description": "DNS ゾーン例", "Tags": ["example"], "Records": [{"Name": "www", "Type": "A", "RData": "192.0.2.10", "TTL": 3600}]}}`,
	}},
	"DNS.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example.com", "Description": "更新後", "Tags": ["example", "updated"], "Records": []}}`,
	}},

	// --- Disk (monitor) ---
	"Disk.Monitor": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},

	// --- ESME (commonserviceitem) ---
	"ESME.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"ESME.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-esme", "Description": "ESME 例", "Tags": ["example"]}}`,
	}},
	"ESME.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-esme", "Description": "ESME 例", "Tags": ["example"]}}`,
	}},
	"ESME.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-esme-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},
	"ESME.Logs": {{
		Response: `{"is_ok": true, "ESME": {"MessageID": "msg-000000000001", "Status": "Delivered", "Destination": "09000000000", "RetryCount": 0}}`,
	}},
	"ESME.SendMessageWithGeneratedOTP": {{
		Response: `{"is_ok": true, "ESME": {"MessageID": "msg-000000000001", "Status": "Accepted", "OTP": "123456"}}`,
	}},
	"ESME.SendMessageWithInputtedOTP": {{
		Response: `{"is_ok": true, "ESME": {"MessageID": "msg-000000000002", "Status": "Accepted"}}`,
	}},

	// --- EnhancedDB (commonserviceitem) ---
	"EnhancedDB.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"EnhancedDB.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-tidb", "Description": "エンハンスド DB 例", "Tags": ["example"]}}`,
	}},
	"EnhancedDB.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-tidb", "Description": "エンハンスド DB 例", "Tags": ["example"]}}`,
	}},
	"EnhancedDB.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-tidb-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},
	"EnhancedDB.GetConfig": {{
		Response: `{"is_ok": true, "EnhancedDB": {"AllowedNetworks": []}}`,
	}},

	// --- GSLB (commonserviceitem) ---
	"GSLB.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"GSLB.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-gslb", "Description": "GSLB 例", "Tags": ["example"], "DestinationServers": []}}`,
	}},
	"GSLB.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-gslb", "Description": "GSLB 例", "Tags": ["example"], "DestinationServers": [{"IPAddress": "192.0.2.10", "Enabled": "True", "Weight": 1}]}}`,
	}},
	"GSLB.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-gslb", "Description": "更新後", "Tags": ["example", "updated"], "DestinationServers": []}}`,
	}},

	// --- IPAddress ---
	"IPAddress.List": {{
		Response: `{"Total": 1, "From": 0, "Count": 1, "IPAddress": [{"IPAddress": "192.0.2.10", "HostName": "example-host"}]}`,
	}},
	"IPAddress.Read": {{
		Response: `{"is_ok": true, "IPAddress": {"IPAddress": "192.0.2.10", "HostName": "example-host"}}`,
	}},
	"IPAddress.UpdateHostName": {{
		Response: `{"is_ok": true, "IPAddress": {"IPAddress": "192.0.2.10", "HostName": "updated-host"}}`,
	}},

	// --- IPv6Addr ---
	"IPv6Addr.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "IPv6Addrs": []}`,
	}},
	"IPv6Addr.Create": {{
		Response: `{"is_ok": true, "IPv6Addr": {"IPv6Addr": "2001:db8::1", "HostName": "example-host"}}`,
	}},
	"IPv6Addr.Read": {{
		Response: `{"is_ok": true, "IPv6Addr": {"IPv6Addr": "2001:db8::1", "HostName": "example-host"}}`,
	}},
	"IPv6Addr.Update": {{
		Response: `{"is_ok": true, "IPv6Addr": {"IPv6Addr": "2001:db8::1", "HostName": "updated-host"}}`,
	}},

	// --- IPv6Net ---
	"IPv6Net.List": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "IPv6Nets": []}`,
	}},
	"IPv6Net.Read": {{
		Response: `{"is_ok": true, "IPv6Net": {"ID": 123456789012, "IPv6Prefix": "2001:db8::", "IPv6PrefixLen": 64}}`,
	}},

	// --- Interface (monitor) ---
	"Interface.Monitor": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},

	// --- Internet actions ---
	"Internet.AddSubnet": {{
		Response: `{"is_ok": true, "Subnet": {"ID": 123456789013, "NetworkAddress": "192.0.2.0", "NetworkMaskLen": 28}}`,
	}},
	"Internet.UpdateSubnet": {{
		Response: `{"is_ok": true, "Subnet": {"ID": 123456789013, "NetworkAddress": "192.0.2.0", "NetworkMaskLen": 28}}`,
	}},
	"Internet.Monitor": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"Internet.UpdateBandWidth": {{
		Response: `{"is_ok": true, "Internet": {"ID": 123456789012, "Name": "example-internet", "Description": "帯域変更後", "Tags": ["example"], "BandWidthMbps": 250}}`,
	}},
	"Internet.EnableIPv6": {{
		Response: `{"is_ok": true, "IPv6Net": {"ID": 123456789014, "IPv6Prefix": "2001:db8::", "IPv6PrefixLen": 64}}`,
	}},

	// --- LoadBalancer (appliance) ---
	"LoadBalancer.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Appliances": []}`,
	}},
	"LoadBalancer.Create": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-lb", "Description": "ロードバランサ例", "Tags": ["example"], "IPAddresses": ["192.168.0.11", "192.168.0.12"], "VirtualIPAddresses": []}}`,
	}},
	"LoadBalancer.Read": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-lb", "Description": "ロードバランサ例", "Tags": ["example"], "IPAddresses": ["192.168.0.11", "192.168.0.12"], "VirtualIPAddresses": []}}`,
	}},
	"LoadBalancer.Update": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-lb-updated", "Description": "更新後", "Tags": ["example", "updated"], "IPAddresses": ["192.168.0.11", "192.168.0.12"], "VirtualIPAddresses": []}}`,
	}},
	"LoadBalancer.MonitorInterface": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},

	// --- LocalRouter (commonserviceitem) ---
	"LocalRouter.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"LocalRouter.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-local-router", "Description": "ローカルルータ例", "Tags": ["example"], "Peers": [], "StaticRoutes": []}}`,
	}},
	"LocalRouter.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-local-router", "Description": "ローカルルータ例", "Tags": ["example"], "Peers": [], "StaticRoutes": []}}`,
	}},
	"LocalRouter.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-local-router-updated", "Description": "更新後", "Tags": ["example", "updated"], "Peers": [], "StaticRoutes": []}}`,
	}},
	"LocalRouter.HealthStatus": {{
		Response: `{"is_ok": true, "LocalRouter": {"Peers": []}}`,
	}},
	"LocalRouter.MonitorLocalRouter": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},

	// --- MobileGateway (appliance) ---
	"MobileGateway.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Appliances": []}`,
	}},
	"MobileGateway.Create": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-mgw", "Description": "モバイルゲートウェイ例", "Tags": ["example"], "InterfaceSettings": [], "StaticRoutes": []}}`,
	}},
	"MobileGateway.Read": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-mgw", "Description": "モバイルゲートウェイ例", "Tags": ["example"], "InterfaceSettings": [], "StaticRoutes": []}}`,
	}},
	"MobileGateway.Update": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-mgw-updated", "Description": "更新後", "Tags": ["example", "updated"], "InterfaceSettings": [], "StaticRoutes": []}}`,
	}},
	"MobileGateway.GetDNS": {{
		Response: `{"is_ok": true, "SIMGroup": {"DNS1": "192.0.2.53", "DNS2": "198.51.100.53"}}`,
	}},
	"MobileGateway.GetSIMRoutes": {{
		Response: `{"is_ok": true, "SIMRoutes": []}`,
	}},
	"MobileGateway.GetTrafficConfig": {{
		Response: `{"is_ok": true, "TrafficMonitoring": {"TrafficQuotaInMB": 1024, "BandWidthLimitInKbps": 0, "EMailConfig": {"Enabled": false}, "SlackConfig": {"Enabled": false, "IncomingWebhooksURL": ""}, "AutoTrafficShaping": false}}`,
	}},
	"MobileGateway.ListSIM": {{
		Response: `{"is_ok": true, "SIM": []}`,
	}},
	"MobileGateway.Logs": {{
		Response: `{"is_ok": true, "Logs": []}`,
	}},
	"MobileGateway.MonitorInterface": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"MobileGateway.TrafficStatus": {{
		Response: `{"is_ok": true, "TrafficStatus": {"UplinkBytes": 0, "DownlinkBytes": 0, "TrafficShaping": false}}`,
	}},

	// --- NFS (appliance) ---
	"NFS.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Appliances": []}`,
	}},
	"NFS.Create": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-nfs", "Description": "NFS 例", "Tags": ["example"], "IPAddresses": ["192.168.0.11"]}}`,
	}},
	"NFS.Read": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-nfs", "Description": "NFS 例", "Tags": ["example"], "IPAddresses": ["192.168.0.11"]}}`,
	}},
	"NFS.Update": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-nfs-updated", "Description": "更新後", "Tags": ["example", "updated"], "IPAddresses": ["192.168.0.11"]}}`,
	}},
	"NFS.MonitorFreeDiskSize": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"NFS.MonitorInterface": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},

	// --- PacketFilter ---
	"PacketFilter.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "PacketFilters": []}`,
	}},
	"PacketFilter.Create": {{
		Response: `{"is_ok": true, "PacketFilter": {"ID": 123456789012, "Name": "example-pf", "Description": "パケットフィルタ例", "Expression": []}}`,
	}},
	"PacketFilter.Read": {{
		Response: `{"is_ok": true, "PacketFilter": {"ID": 123456789012, "Name": "example-pf", "Description": "パケットフィルタ例", "Expression": []}}`,
	}},
	"PacketFilter.Update": {{
		Response: `{"is_ok": true, "PacketFilter": {"ID": 123456789012, "Name": "example-pf-updated", "Description": "更新後", "Expression": []}}`,
	}},

	// --- PrivateHost ---
	"PrivateHost.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "PrivateHosts": []}`,
	}},
	"PrivateHost.Create": {{
		Response: `{"is_ok": true, "PrivateHost": {"ID": 123456789012, "Name": "example-private-host", "Description": "専有ホスト例", "Tags": ["example"]}}`,
	}},
	"PrivateHost.Read": {{
		Response: `{"is_ok": true, "PrivateHost": {"ID": 123456789012, "Name": "example-private-host", "Description": "専有ホスト例", "Tags": ["example"]}}`,
	}},
	"PrivateHost.Update": {{
		Response: `{"is_ok": true, "PrivateHost": {"ID": 123456789012, "Name": "example-private-host-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},

	// --- ProxyLB (commonserviceitem) ---
	"ProxyLB.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"ProxyLB.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-proxy-lb", "Description": "エンハンスド LB 例", "Tags": ["example"], "Plan": 100, "BindPorts": [], "Servers": [], "Rules": []}}`,
	}},
	"ProxyLB.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-proxy-lb", "Description": "エンハンスド LB 例", "Tags": ["example"], "Plan": 100, "BindPorts": [], "Servers": [], "Rules": []}}`,
	}},
	"ProxyLB.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-proxy-lb-updated", "Description": "更新後", "Tags": ["example", "updated"], "Plan": 100, "BindPorts": [], "Servers": [], "Rules": []}}`,
	}},
	"ProxyLB.ChangePlan": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-proxy-lb", "Description": "プラン変更後", "Tags": ["example"], "Plan": 500, "BindPorts": [], "Servers": [], "Rules": []}}`,
	}},
	"ProxyLB.GetCertificates": {{
		Response: `{"is_ok": true, "ProxyLB": {"PrimaryCert": null, "AdditionalCerts": []}}`,
	}},
	"ProxyLB.SetCertificates": {{
		Response: `{"is_ok": true, "ProxyLB": {"PrimaryCert": null, "AdditionalCerts": []}}`,
	}},
	"ProxyLB.HealthStatus": {{
		Response: `{"is_ok": true, "ProxyLB": {"ActiveConn": 0, "CurrentVIP": "192.0.2.10", "Servers": []}}`,
	}},
	"ProxyLB.MonitorConnection": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},

	// --- Server (actions) ---
	"Server.ChangePlan": {{
		Response: `{
			"is_ok": true,
			"Server": {
				"ID": 123456789012,
				"Name": "example-server",
				"Description": "プラン変更後",
				"Tags": ["example"],
				"Availability": "available",
				"HostName": "localhost",
				"InterfaceDriver": "virtio",
				"ServerPlan": {
					"ID": 200002001,
					"Name": "プラン/1Core-2GB(新プラン)",
					"CPU": 1,
					"MemoryMB": 2048,
					"GPU": 0,
					"GPUModel": "none",
					"CPUModel": "uncategorized",
					"Commitment": "standard",
					"Generation": 200,
					"ConfidentialVM": false
				},
				"Icon": null
			}
		}`,
	}},
	"Server.GetVNCProxy": {{
		Response: `{"is_ok": true, "VNCProxyInfo": {"Status": "available", "Host": "sac-tk1v-vnc.example.jp", "IOServerHost": "sac-tk1v-vnc.example.jp", "Port": 5900, "Password": "dummy-vnc-password", "VNCFile": ""}}`,
	}},
	"Server.Monitor": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},

	// --- ServiceClass ---
	"ServiceClass.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "ServiceClasses": []}`,
	}},

	// --- SIM (commonserviceitem) ---
	"SIM.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"SIM.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-sim", "Description": "SIM 例", "Tags": ["example"], "Class": "sim"}}`,
	}},
	"SIM.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-sim", "Description": "SIM 例", "Tags": ["example"], "Class": "sim"}}`,
	}},
	"SIM.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-sim-updated", "Description": "更新後", "Tags": ["example", "updated"], "Class": "sim"}}`,
	}},
	"SIM.Logs": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Logs": []}`,
	}},
	"SIM.MonitorSIM": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"SIM.GetNetworkOperator": {{
		Response: `{"is_ok": true, "NetworkOperationConfigs": []}`,
	}},

	// --- SimpleMonitor (commonserviceitem) ---
	"SimpleMonitor.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"SimpleMonitor.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example.com", "Description": "シンプル監視例", "Tags": ["example"], "Class": "simplemon"}}`,
	}},
	"SimpleMonitor.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example.com", "Description": "シンプル監視例", "Tags": ["example"], "Class": "simplemon"}}`,
	}},
	"SimpleMonitor.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example.com", "Description": "更新後", "Tags": ["example", "updated"], "Class": "simplemon"}}`,
	}},
	"SimpleMonitor.HealthStatus": {{
		Response: `{"is_ok": true, "SimpleMonitor": {"Health": "up", "LatestLogs": []}}`,
	}},
	"SimpleMonitor.MonitorResponseTime": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},

	// --- SimpleNotificationGroup (commonserviceitem) ---
	"SimpleNotificationGroup.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"SimpleNotificationGroup.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-notice-group", "Description": "シンプル通知グループ例", "Tags": ["example"]}}`,
	}},
	"SimpleNotificationGroup.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-notice-group", "Description": "シンプル通知グループ例", "Tags": ["example"]}}`,
	}},
	"SimpleNotificationGroup.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-notice-group-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},
	"SimpleNotificationGroup.History": {{
		Response: `{"is_ok": true, "NotificationHistories": {"NotificationHistories": []}}`,
	}},

	// --- SimpleNotificationDestination (commonserviceitem) ---
	"SimpleNotificationDestination.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "CommonServiceItems": []}`,
	}},
	"SimpleNotificationDestination.Create": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-notice-dest", "Description": "通知先例", "Tags": ["example"]}}`,
	}},
	"SimpleNotificationDestination.Read": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-notice-dest", "Description": "通知先例", "Tags": ["example"]}}`,
	}},
	"SimpleNotificationDestination.Update": {{
		Response: `{"is_ok": true, "CommonServiceItem": {"ID": 123456789012, "Name": "example-notice-dest-updated", "Description": "更新後", "Tags": ["example", "updated"]}}`,
	}},

	// --- Subnet ---
	"Subnet.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Subnets": []}`,
	}},
	"Subnet.Read": {{
		Response: `{"is_ok": true, "Subnet": {"ID": 123456789012, "NetworkAddress": "192.0.2.0", "NetworkMaskLen": 28}}`,
	}},

	// --- Switch (GetServers) ---
	"Switch.GetServers": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Servers": []}`,
	}},

	// --- VPCRouter (appliance) ---
	"VPCRouter.Find": {{
		Response: `{"Total": 0, "From": 0, "Count": 0, "Appliances": []}`,
	}},
	"VPCRouter.Create": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-vpc-router", "Description": "VPC ルータ例", "Tags": ["example"], "Class": "vpcrouter"}}`,
	}},
	"VPCRouter.Read": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-vpc-router", "Description": "VPC ルータ例", "Tags": ["example"], "Class": "vpcrouter"}}`,
	}},
	"VPCRouter.Update": {{
		Response: `{"is_ok": true, "Appliance": {"ID": 123456789012, "Name": "example-vpc-router-updated", "Description": "更新後", "Tags": ["example", "updated"], "Class": "vpcrouter"}}`,
	}},
	"VPCRouter.Status": {{
		Response: `{"is_ok": true, "Router": {"FirewallReceiveLogs": [], "FirewallSendLogs": [], "VPNLogs": [], "SessionCount": 0}}`,
	}},
	"VPCRouter.Ping": {{
		Response: `{"is_ok": true, "VPCRouter": {"Result": ["ok"]}}`,
	}},
	"VPCRouter.MonitorCPU": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
	"VPCRouter.MonitorInterface": {{
		Response: `{"is_ok": true, "Data": {}}`,
	}},
}
