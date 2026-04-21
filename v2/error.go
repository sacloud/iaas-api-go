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

package iaas

import (
	"errors"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/saclient-go"
)

// Error はパッケージ固有のエラー型。msg に操作名（例: "Note.List"）を、
// err に下位エラー（ogen 由来の *client.ApiErrorStatusCode や
// saclient.Error など）を保持する。errors.As / errors.Is で下位エラーを
// 取り出せるよう Unwrap を実装する。
type Error struct {
	msg string
	err error
}

func (e *Error) Error() string {
	if e.msg != "" {
		if e.err != nil {
			return "iaas: " + e.msg + ": " + e.err.Error()
		}
		return "iaas: " + e.msg
	}
	return "iaas: " + e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

// NewError は一般的なラップ用コンストラクタ。argument 検証失敗など
// HTTP 呼び出しを伴わないエラーに利用する。
func NewError(msg string, err error) *Error {
	return &Error{msg: msg, err: err}
}

// NewAPIError は HTTP ステータスコード付きのラップ用。
// 内部で saclient.NewError に委譲するため saclient.IsNotFoundError 等の
// 汎用ヘルパが errors.As 経由で利用できる。トランスポートエラーなどで
// ステータスコードが取れない場合は code=0 で呼ぶ。
func NewAPIError(method string, code int, err error) *Error {
	return &Error{msg: method, err: saclient.NewError(code, "", err)}
}

// wrapOpErr は ogen 生成エラーから *client.ApiErrorStatusCode を取り出し、
// HTTP ステータスコードを NewAPIError に渡す。取れなければ code=0 で
// NewAPIError を呼ぶ。各 Op メソッド共通のエラーラップ処理。
func wrapOpErr(methodName string, err error) error {
	var e *client.ApiErrorStatusCode
	if errors.As(err, &e) {
		return NewAPIError(methodName, e.StatusCode, err)
	}
	return NewAPIError(methodName, 0, err)
}
