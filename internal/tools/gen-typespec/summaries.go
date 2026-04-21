// Copyright 2022-2026 The sacloud/iaas-api-go Authors
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

import (
	"fmt"
	"log"
	"sync"
)

// actionSummaries は TypeSpec op のアクション名（= lowerFirst(DSL Op 名)）→ 日本語の動詞句。
// buildSummary() が "<Resource> <verbJa>" 形式で @summary を合成するときの語源になる。
// 未登録のアクションは buildSummary() 側で英語そのままの fallback を使い、
// stderr に警告を出して後から手当てできるようにする。
var actionSummaries = map[string]string{
	// 基本 CRUD / 一覧系
	"find":       "一覧取得",
	"list":       "一覧取得",
	"read":       "取得",
	"create":     "作成",
	"update":     "更新",
	"delete":     "削除",
	"detail":     "詳細取得",
	"details":    "明細取得",
	"detailsCSV": "明細 CSV 取得",
	"history":    "履歴取得",
	"logs":       "ログ取得",

	// 状態・電源系
	"boot":         "起動",
	"shutdown":     "シャットダウン",
	"reset":        "リセット",
	"status":       "ステータス取得",
	"healthStatus": "ヘルスステータス取得",
	"config":       "設定反映",
	"getConfig":    "設定取得",
	"setConfig":    "設定更新",
	"activate":     "有効化",
	"deactivate":   "無効化",

	// 請求系
	"byContract":          "契約別取得",
	"byContractYear":      "契約・年別取得",
	"byContractYearMonth": "契約・年月別取得",

	// モニター情報
	"monitor":                 "モニター情報取得",
	"monitorCPU":              "CPU モニター情報取得",
	"monitorConnection":       "コネクションモニター情報取得",
	"monitorDatabase":         "データベースモニター情報取得",
	"monitorDisk":             "ディスクモニター情報取得",
	"monitorInterface":        "インターフェイスモニター情報取得",
	"monitorInterfaceByIndex": "インターフェイス別モニター情報取得",
	"monitorLocalRouter":      "ローカルルータモニター情報取得",
	"monitorResponseTime":     "応答時間モニター情報取得",
	"monitorSIM":              "SIM モニター情報取得",

	// ネットワーク系
	"assignIP":                   "IP アドレス割り当て",
	"clearIP":                    "IP アドレスクリア",
	"addSubnet":                  "サブネット追加",
	"deleteSubnet":               "サブネット削除",
	"updateSubnet":               "サブネット更新",
	"disableIPv6":                "IPv6 無効化",
	"enableIPv6":                 "IPv6 有効化",
	"connectToBridge":            "ブリッジ接続",
	"disconnectFromBridge":       "ブリッジ切断",
	"connectToPacketFilter":      "パケットフィルタ接続",
	"disconnectFromPacketFilter": "パケットフィルタ切断",
	"connectToServer":            "サーバー接続",
	"disconnectFromServer":       "サーバー切断",
	"connectToSharedSegment":     "共有セグメント接続",
	"connectToSwitch":            "スイッチ接続",
	"disconnectFromSwitch":       "スイッチ切断",

	// SIM / MobileGateway 系
	"addSIM":               "SIM 追加",
	"deleteSIM":            "SIM 削除",
	"listSIM":              "SIM 一覧取得",
	"getDNS":               "DNS 取得",
	"setDNS":               "DNS 設定",
	"getNetworkOperator":   "通信キャリア取得",
	"setNetworkOperator":   "通信キャリア設定",
	"getSIMRoutes":         "SIM ルート取得",
	"setSIMRoutes":         "SIM ルート設定",
	"getTrafficConfig":     "トラフィック設定取得",
	"setTrafficConfig":     "トラフィック設定更新",
	"deleteTrafficConfig":  "トラフィック設定削除",
	"trafficStatus":        "トラフィック状況取得",
	"imeiLock":             "IMEI ロック",
	"imeiUnlock":           "IMEI ロック解除",

	// Server 系
	"changePlan":     "プラン変更",
	"insertCDROM":    "CD-ROM 挿入",
	"ejectCDROM":     "CD-ROM 取り出し",
	"sendKey":        "キー送信",
	"sendNMI":        "NMI 送信",
	"getVNCProxy":    "VNC プロキシ取得",
	"scaleDown":      "スケールダウン",
	"scaleUp":        "スケールアップ",
	"updateHostName": "ホスト名更新",
	"updateBandWidth": "帯域更新",

	// Disk 系
	"resizePartition": "パーティションリサイズ",

	// Archive 系
	"closeFTP": "FTP クローズ",
	"openFTP":  "FTP オープン",
	"share":    "共有",
	"transfer": "移管",

	// Internet 系
	"ping": "Ping",

	// 証明書系
	"getCertificates":      "証明書取得",
	"setCertificates":      "証明書設定",
	"deleteCertificates":   "証明書削除",
	"renewLetsEncryptCert": "Let's Encrypt 証明書更新",

	// LoadBalancer / VPCRouter の server/user/client 管理系
	"addServer":    "サーバー追加",
	"addClient":    "クライアント追加",
	"addUser":      "ユーザー追加",
	"getServers":   "サーバー一覧取得",
	"listServers":  "サーバー一覧取得",
	"listClients":  "クライアント一覧取得",
	"listUsers":    "ユーザー一覧取得",
	"readClient":   "クライアント取得",
	"readServer":   "サーバー取得",
	"revokeServer": "サーバー破棄",
	"revokeClient": "クライアント破棄",
	"holdServer":   "サーバー保留",
	"holdClient":   "クライアント保留",
	"resumeServer": "サーバー再開",
	"resumeClient": "クライアント再開",
	"denyClient":   "クライアント拒否",
	"deleteUser":   "ユーザー削除",
	"updateUser":   "ユーザー更新",

	// パスワード
	"setPassword": "パスワード設定",

	// メッセージ / OTP
	"postMessage":                 "メッセージ送信",
	"sendMessageWithGeneratedOTP": "自動生成 OTP 付きメッセージ送信",
	"sendMessageWithInputtedOTP":  "入力 OTP 付きメッセージ送信",

	// Database パラメータ
	"getParameter": "パラメータ取得",
	"setParameter": "パラメータ設定",
}

// summaryOverrides は "<ResourceName>.<actionName>" キーで合成結果を差し替える。
// 基本は actionSummaries による "<Resource> <verbJa>" の合成で済ませるが、
//   - ローマ字リソース名のままでは意味が通らない
//   - 語順・助詞が不自然
//   - マニュアル表記と合わせたい
// というケースでここに登録する。マニュアル参照ページ URL をコメントで残すこと。
// 参照: https://manual.sakura.ad.jp/cloud/
var summaryOverrides = map[string]string{
	// 請求情報: "Bill 取得" では語感が弱い
	// https://manual.sakura.ad.jp/cloud/api/billing.html
	"Bill.read":                "請求情報取得",
	"Bill.details":             "請求明細取得",
	"Bill.detailsCSV":          "請求明細 CSV 取得",
	"Bill.byContract":          "契約別請求情報取得",
	"Bill.byContractYear":      "契約・年別請求情報取得",
	"Bill.byContractYearMonth": "契約・年月別請求情報取得",

	// ESME: SMS メッセージ送信サービス。"ESME 〜メッセージ送信" は意味が重複するのでマニュアル表記に寄せる
	// https://manual.sakura.ad.jp/cloud/appliance/esme/
	"ESME.logs":                        "SMS 送信ログ取得",
	"ESME.sendMessageWithGeneratedOTP": "自動生成 OTP 付き SMS メッセージ送信",
	"ESME.sendMessageWithInputtedOTP":  "指定 OTP 付き SMS メッセージ送信",

	// Coupon: クーポン情報取得
	"Coupon.find": "クーポン情報取得",

	// AuthStatus: 認証情報（API キー検証結果）取得
	// https://manual.sakura.ad.jp/cloud/api/auth.html
	"AuthStatus.read": "認証情報取得",

	// Zone / Region: 一覧用 find のみ
	"Zone.find":   "ゾーン一覧取得",
	"Zone.read":   "ゾーン情報取得",
	"Region.find": "リージョン一覧取得",
	"Region.read": "リージョン情報取得",

	// ServiceClass: 料金プラン情報取得
	"ServiceClass.find": "ServiceClass 一覧取得",

	// Disk.config は「ディスクの修正」(OS/ユーザ等の反映) を指す
	// https://manual.sakura.ad.jp/cloud/server/disk-modify.html
	"Disk.config": "ディスクの修正反映",

	// Switch.connectToBridge / disconnectFromBridge
	"Switch.connectToBridge":      "ブリッジ接続",
	"Switch.disconnectFromBridge": "ブリッジ切断",

	// SIM: assignIP/clearIP/imeiLock 等の粒度を細かく出す
	// https://manual.sakura.ad.jp/cloud/appliance/sim/
	"SIM.assignIP":   "SIM への IP アドレス割り当て",
	"SIM.clearIP":    "SIM の IP アドレス割り当て解除",
	"SIM.imeiLock":   "SIM の IMEI ロック設定",
	"SIM.imeiUnlock": "SIM の IMEI ロック解除",
	"SIM.logs":       "SIM セッションログ取得",
	// monitorSIM アクションはリソース名 "SIM" と重複するので省略
	"SIM.monitorSIM": "SIM モニター情報取得",

	// LocalRouter.monitorLocalRouter: "LocalRouter ローカルルータモニター情報取得" は冗長
	"LocalRouter.monitorLocalRouter": "LocalRouter モニター情報取得",
}

// unknownActionsOnce は未登録 action の警告出力を一度だけに抑える。
var unknownActionsOnce sync.Map

// buildSummary はリソース名（PascalCase）とアクション名（lowerCamel）から
// "@summary(...)" に入れる日本語文字列を返す。
//
// 優先順:
//  1. summaryOverrides["<Resource>.<action>"] があればそれ
//  2. actionSummaries[action] があれば "<Resource> <verbJa>"
//  3. fallback: "<Resource> <action>"（英語そのまま）＋ stderr に警告
func buildSummary(resource, action string) string {
	if override, ok := summaryOverrides[resource+"."+action]; ok {
		return override
	}
	if verb, ok := actionSummaries[action]; ok {
		return resource + " " + verb
	}
	if _, loaded := unknownActionsOnce.LoadOrStore(action, true); !loaded {
		log.Printf("gen-typespec: warning: unknown action %q (resource=%s); emitting raw fallback. Add it to actionSummaries or summaryOverrides.", action, resource)
	}
	return fmt.Sprintf("%s %s", resource, action)
}
