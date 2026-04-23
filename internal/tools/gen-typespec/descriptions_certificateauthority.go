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
		"CertificateAuthority": {
			"Status": "CA の発行済証明書の状態・DN 情報",
		},
		"CertificateAuthorityCreateRequestStatus": {
			"Country":          "証明書 Subject の Country (国コード、例 \"JP\")",
			"Organization":     "証明書 Subject の Organization (組織名)",
			"OrganizationUnit": "証明書 Subject の OrganizationUnit (部署名)",
			"CommonName":       "証明書 Subject の CommonName",
			"NotAfter":         "証明書の有効期限 (ISO 8601)",
		},

		"CertificateAuthorityAddClientParam": {
			"CertificateSigningRequest": "CSR (PEM)。IssuanceMethod が \"csr\" のとき指定",
			"PublicKey":                 "公開鍵 (PEM)。IssuanceMethod が \"public_key\" のとき指定",
			"IssuanceMethod":            "発行方式 (\"csr\" / \"public_key\" / \"email\" / \"url\")",
			"EMail":                     "通知先メールアドレス",
			"Country":                   "Subject Country",
			"Organization":              "Subject Organization",
			"OrganizationUnit":          "Subject OrganizationUnit",
			"CommonName":                "Subject CommonName",
			"NotAfter":                  "証明書の有効期限",
		},
		"CertificateAuthorityAddServerParam": {
			"CertificateSigningRequest": "CSR (PEM)",
			"PublicKey":                 "公開鍵 (PEM)",
			"SANs":                      "Subject Alternative Name の配列",
			"Country":                   "Subject Country",
			"Organization":              "Subject Organization",
			"OrganizationUnit":          "Subject OrganizationUnit",
			"CommonName":                "Subject CommonName",
			"NotAfter":                  "証明書の有効期限",
		},

		"CertificateAuthorityClient": {
			"CertificateData": "発行済みクライアント証明書 (PEM)",
			"EMail":           "通知先メールアドレス",
			"IssuanceMethod":  "発行方式",
			"IssueState":      "発行状態 (\"approved\" / \"approving\" / \"denied\" / \"revoked\" 等)",
			"Subject":         "Subject DN",
			"URL":             "IssuanceMethod が \"url\" の場合のダウンロード URL",
		},
		"CertificateAuthorityAddClientOrServerResult": {
			"ID": "発行された証明書のID",
		},
	})
}
