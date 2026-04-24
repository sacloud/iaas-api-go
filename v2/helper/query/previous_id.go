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

package query

import (
	"context"
	"fmt"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
)

// PreviousIDTagName はリソース再作成時に旧 ID を保持するためのタグ名プレフィックス。
// v1 helper/plans.PreviousIDTagName と同じ文字列。
const PreviousIDTagName = "@previous-id"

// ErrNoResults は `previous-id` タグでも対象が見つからなかった場合に返されるエラー。
// v1 iaas.NoResultsError の v2 相当。
var ErrNoResults = fmt.Errorf("query: no results")

// previousIDTag は id を保持する @previous-id タグ文字列を返す。
func previousIDTag(id int64) string {
	return fmt.Sprintf("%s=%d", PreviousIDTagName, id)
}

// ServerReadFinder は ReadServer に必要な I/F。iaas.ServerAPI が満たす。
type ServerReadFinder interface {
	ServerReader
	ServerFinder
}

// ReadServer は指定 ID で Server を Read し、失敗したら @previous-id タグで Find して返す。
// どちらでも見つからなければ ErrNoResults を返す。
func ReadServer(ctx context.Context, op ServerReadFinder, id int64) (*client.Server, error) {
	resp, err := op.Read(ctx, id)
	if err == nil {
		s := resp.Server
		return &s, nil
	}
	if !iaas.IsNotFoundError(err) {
		return nil, err
	}

	findResp, err := op.List(ctx, &client.ServerFindRequest{
		Filter: client.ServerFindFilter{Tags: []string{previousIDTag(id)}},
		Count:  1,
	})
	if err != nil {
		return nil, err
	}
	if len(findResp.Servers) == 0 {
		return nil, ErrNoResults
	}
	s := findResp.Servers[0]
	return &s, nil
}

// InternetReadFinder は ReadRouter に必要な I/F。
type InternetReadFinder interface {
	Read(ctx context.Context, id int64) (*client.InternetReadResponseEnvelope, error)
	List(ctx context.Context, req *client.InternetFindRequest) (*client.InternetFindResponseEnvelope, error)
}

// ReadRouter は指定 ID で Internet (ルータ+スイッチ) を Read し、失敗したら @previous-id で Find する。
func ReadRouter(ctx context.Context, op InternetReadFinder, id int64) (*client.Internet, error) {
	resp, err := op.Read(ctx, id)
	if err == nil {
		r := resp.Internet
		return &r, nil
	}
	if !iaas.IsNotFoundError(err) {
		return nil, err
	}

	findResp, err := op.List(ctx, &client.InternetFindRequest{
		Filter: client.InternetFindFilter{Tags: []string{previousIDTag(id)}},
		Count:  1,
	})
	if err != nil {
		return nil, err
	}
	if len(findResp.Internet) == 0 {
		return nil, ErrNoResults
	}
	r := findResp.Internet[0]
	return &r, nil
}

// 注: v1 の ReadProxyLB は v2 では未提供。
// v2 では ProxyLB は CommonServiceItem の一種として扱われ、ProxyLBAPI には Read が
// 存在しない。CommonServiceItem.Read / List は AutoBackup 型のエンベロープで返るため、
// ProxyLB 固有の fat model を取り出す helper は別途設計が必要。downstream で必要に
// なった時点で設計し直す。
