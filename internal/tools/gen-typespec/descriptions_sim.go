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
		"SIMCreateRequest": {
			"Remark": "SIM の不変情報 (パスコード等)",
			"Status": "SIM の状態 (ICCID 等)",
		},
		"SIMCreateRequestRemark": {
			"PassCode": "SIM アクティベート用パスコード",
		},
		"SIMCreateRequestStatus": {
			"ICCID": "SIM カード裏面に記載の ICCID (19-20 桁の識別番号)",
		},
		"SIMAssignIPRequest": {
			"IP": "SIM に割り当てる IPv4 アドレス",
		},
		"SIMIMEILockRequest": {
			"IMEI": "ロック対象の IMEI (端末固有番号 15 桁)",
		},
		"SIMInfo": {
			"ICCID":           "SIM カードの ICCID",
			"IMEI":            "ロック中の IMEI",
			"ConnectedIMEI":   "現在接続中の端末の IMEI",
			"Activated":       "SIM が有効化されているかどうか",
			"ActivatedDate":   "有効化日時",
			"DeactivatedDate": "無効化日時",
		},
	})
}
