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
	"net/http"
	"testing"

	"github.com/sacloud/iaas-api-go/v2/client"
)

func TestErrorAccessors(t *testing.T) {
	orig := &client.ApiErrorStatusCode{
		StatusCode: http.StatusConflict,
		Response: client.ApiError{
			Serial:    client.NewOptString("serial-abc"),
			ErrorCode: client.NewOptString("still_creating"),
			ErrorMsg:  client.NewOptString("resource is still being created"),
		},
	}
	err := wrapOpErr("Resource.Foo", orig).(*Error)

	if got := err.ResponseCode(); got != http.StatusConflict {
		t.Errorf("ResponseCode() = %d, want %d", got, http.StatusConflict)
	}
	if got := err.Code(); got != "still_creating" {
		t.Errorf("Code() = %q, want %q", got, "still_creating")
	}
	if got := err.Message(); got != "resource is still being created" {
		t.Errorf("Message() = %q, want %q", got, "resource is still being created")
	}
	if got := err.Serial(); got != "serial-abc" {
		t.Errorf("Serial() = %q, want %q", got, "serial-abc")
	}
}

func TestErrorAccessorsWithoutAPIError(t *testing.T) {
	// ネットワークエラー等で *client.ApiErrorStatusCode を持たないケース。
	err := NewError("Resource.Foo", errors.New("connection reset"))

	if got := err.ResponseCode(); got != 0 {
		t.Errorf("ResponseCode() = %d, want 0", got)
	}
	if got := err.Code(); got != "" {
		t.Errorf("Code() = %q, want empty", got)
	}
}

func TestIsNotFoundError(t *testing.T) {
	notFound := wrapOpErr("Resource.Foo", &client.ApiErrorStatusCode{
		StatusCode: http.StatusNotFound,
		Response:   client.ApiError{},
	})
	if !IsNotFoundError(notFound) {
		t.Error("IsNotFoundError should be true for 404")
	}

	conflict := wrapOpErr("Resource.Foo", &client.ApiErrorStatusCode{
		StatusCode: http.StatusConflict,
		Response:   client.ApiError{},
	})
	if IsNotFoundError(conflict) {
		t.Error("IsNotFoundError should be false for 409")
	}

	if IsNotFoundError(nil) {
		t.Error("IsNotFoundError(nil) should be false")
	}
}
