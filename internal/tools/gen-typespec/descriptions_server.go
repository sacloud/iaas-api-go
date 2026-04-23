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

func init() {
	registerFieldDescriptions(map[string]map[string]string{
		"Server": {
			"HostName":        "サーバに設定されるホスト名 (ゲスト OS 内部の hostname とは独立した API 上のラベル)",
			"InterfaceDriver": "NIC 仮想デバイスのドライバ種別。\"virtio\" (推奨) または \"e1000\"",
			"ServerPlan":      "サーバプラン (CPU/メモリ/世代)。IDで特定するかプラン条件で指定する",
			"PrivateHost":     "配置する専用ホスト。null または未指定なら共有ホストに配置される",
			"Instance":        "サーバインスタンスの実行状態 (稼働状態や起動/停止情報)",
			"Interfaces":      "NIC (ネットワークインターフェース) の配列",
			"Disks":           "接続されているディスクの配列",
			"Zone":            "サーバが所属するゾーンの情報",
			"BundleInfo":      "サーバに付随するバンドル情報 (ライセンス同梱等)",
		},

		"ServerCreateRequest": {
			"ServerPlan":        "希望するサーバプラン。CPU/メモリ/世代の条件指定でマッチするプランが割り当てられる",
			"ConnectedSwitches": "作成と同時に接続するスイッチ/ネットワークの配列 (0 番目が共有セグメントや特定スイッチ)",
			"PrivateHost":       "配置する専用ホスト。指定しない場合は共有ホストに配置される",
			"InterfaceDriver":   "NIC ドライバ種別。指定しない場合は \"virtio\"",
		},
		"ServerCreateRequestServerPlan": {
			"CPU":            "CPU コア数",
			"MemoryMB":       "メモリ容量 (MB)",
			"GPU":            "GPU 数 (GPU プランの場合)",
			"GPUModel":       "GPU モデル名",
			"CPUModel":       "CPU モデル (\"uncategorized\" や \"amd_epyc_7713p_standard\" 等)",
			"Commitment":     "CPU 割当コミット方式。\"standard\" (共有) または \"dedicatedcpu\" (CPU 専有)",
			"Generation":     "プラン世代。100 = 第1世代、200 = 第2世代(新プラン) 等",
			"ConfidentialVM": "Confidential VM (機密計算対応プラン) を有効化するかどうか",
		},

		"ServerServerPlan": {
			"CPU":            "CPU コア数",
			"MemoryMB":       "メモリ容量 (MB)",
			"GPU":            "GPU 数",
			"GPUModel":       "GPU モデル名",
			"CPUModel":       "CPU モデル名",
			"Commitment":     "CPU 割当コミット方式 (\"standard\" / \"dedicatedcpu\")",
			"Generation":     "プラン世代",
			"ConfidentialVM": "Confidential VM プランであるか",
		},

		"ServerChangePlanRequest": {
			"CPU":        "変更後の CPU コア数",
			"MemoryMB":   "変更後のメモリ容量 (MB)",
			"GPU":        "変更後の GPU 数",
			"GPUModel":   "変更後の GPU モデル",
			"CPUModel":   "変更後の CPU モデル",
			"Generation": "変更後のプラン世代",
			"Commitment": "変更後の CPU 割当コミット方式",
		},
		"ServerChangePlanRequestEnvelope": {
			"CPU":        "変更後の CPU コア数",
			"MemoryMB":   "変更後のメモリ容量 (MB)",
			"GPU":        "変更後の GPU 数",
			"GPUModel":   "変更後の GPU モデル",
			"CPUModel":   "変更後の CPU モデル",
			"Generation": "変更後のプラン世代",
			"Commitment": "変更後の CPU 割当コミット方式",
		},

		"ServerDeleteRequestEnvelope": {
			"WithDisk": "サーバと一緒に削除するディスクのIDの配列。省略時はディスクを削除しない",
		},
		"ServerDeleteWithDisksRequest": {
			"WithDisk": "サーバと一緒に削除するディスクのIDの配列。省略時はディスクを削除しない",
		},

		"ServerShutdownRequestEnvelope": {
			"Force": "true を指定すると強制停止 (ACPI シャットダウンではなく電源断相当)",
		},
		"ShutdownOption": {
			"Force": "true を指定すると強制停止 (ACPI シャットダウンではなく電源断相当)",
		},

		"ServerSendKeyRequestEnvelope": {
			"Key":  "VNC コンソールに送る単一キー (\"ctrl-alt-delete\" 等)",
			"Keys": "VNC コンソールに送るキーのシーケンス",
		},
		"SendKeyRequest": {
			"Key":  "VNC コンソールに送る単一キー",
			"Keys": "VNC コンソールに送るキーのシーケンス",
		},

		"ServerBootRequestEnvelope": {
			"UserBootVariables": "起動時に cloud-init 等へ渡すユーザ変数",
		},
		"ServerBootVariables": {
			"CloudInit": "cloud-init 用データ",
		},
		"ServerBootVariablesCloudInit": {
			"UserData": "cloud-init user-data (YAML などの文字列)",
		},

		"VNCProxyInfo": {
			"Host":         "VNC プロキシのホスト名",
			"IOServerHost": "VNC IO サーバのホスト名",
			"Password":     "VNC 接続時のパスワード",
			"Port":         "VNC プロキシのポート番号",
			"Status":       "VNC プロキシの状態",
			"VNCFile":      "VNC クライアント用の設定ファイル内容 (.vnc)",
		},
		"VNCProxy": {
			"HostName":  "VNC プロキシのホスト名",
			"IPAddress": "VNC プロキシの IPv4 アドレス",
		},

		"ServerInstance": {
			"Status":          "インスタンスの稼働状態 (\"up\" = 起動中、\"down\" = 停止中、\"cleaning\" 等)",
			"BeforeStatus":    "直前の稼働状態 (直前に起動/停止操作があった場合)",
			"StatusChangedAt": "稼働状態が最後に変化した日時",
			"Host":            "現在サーバが配置されているホストの情報",
			"CDROM":           "挿入されている CDROM (ISO イメージ)",
			"Warnings":        "インスタンスに関する警告メッセージ",
			"WarningsValue":   "警告の数値コード",
		},
		"ServerInstanceHost": {
			"InfoURL": "ホスト情報の参照 URL",
			"Name":    "ホスト名",
		},

		"InterfaceView": {
			"UserIPAddress": "ユーザが手動で割り当てた IPv4 アドレス",
			"IPAddress":     "自動または静的に割り当てられた IPv4 アドレス",
			"Switch":        "接続先スイッチ/ルータ+スイッチ/共有セグメントの情報",
			"PacketFilter":  "割り当てられているパケットフィルタ",
		},
		"InterfaceViewSwitch": {
			"Subnet":     "スイッチに紐付くルータ+スイッチのサブネット情報",
			"UserSubnet": "ユーザが指定した論理サブネット (mask/デフォルトルート)",
		},
		"InterfaceViewSwitchSubnet": {
			"NetworkAddress": "サブネットのネットワークアドレス",
			"NetworkMaskLen": "サブネットのマスク長",
			"DefaultRoute":   "サブネットのデフォルトルート",
			"Internet":       "ルータ+スイッチの場合の回線情報",
		},
		"InterfaceViewSwitchSubnetInternet": {
			"BandWidthMbps": "ルータの帯域幅 (Mbps)",
		},

		"ServerConnectedDisk": {
			"Availability":        "ディスクの利用状態",
			"Connection":          "接続方式 (\"virtio\" / \"ide\")",
			"ConnectionOrder":     "接続順序 (0 始まり、小さいほど先頭)",
			"EncryptionAlgorithm": "ディスク暗号化アルゴリズム。\"none\" または \"aes256_xts\"",
			"ReinstallCount":      "OS 再インストール回数",
			"SizeMB":              "ディスク容量 (MB)",
			"Plan":                "ディスクプラン",
			"Storage":             "配置ストレージ",
		},
		"Storage": {
			"Class":      "ストレージクラス",
			"Generation": "ストレージ世代",
		},

		"InsertCDROMRequest": {
			"ID": "挿入する CDROM (ISO イメージ) のID",
		},
		"EjectCDROMRequest": {
			"ID": "排出する CDROM のID (現在挿入されている CDROM のID)",
		},

		"FTPServerInfo": {
			"HostName":  "FTP サーバのホスト名",
			"IPAddress": "FTP サーバの IPv4 アドレス",
		},
		"ConnectedSwitch": {
			"ID":    "接続先スイッチのID (shared の場合は 0 相当)",
			"Scope": "\"shared\" で共有セグメント、\"user\" でユーザスイッチ",
		},

		"ZoneInfo": {
			"IsDummy":   "テスト用のダミーゾーンかどうか",
			"FTPServer": "当該ゾーンの FTP サーバ情報",
			"Region":    "所属リージョン情報",
			"VNCProxy":  "VNC プロキシ情報",
		},
	})
}
