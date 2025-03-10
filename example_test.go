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

package iaas_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/power"
	"github.com/sacloud/iaas-api-go/naked"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/size"
)

func Example_basic() {
	// APIクライアントの基本的な使い方の例

	// APIキー
	token := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	secret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")

	// クライアントの作成
	client := iaas.NewClient(token, secret)

	// スイッチの作成
	swOp := iaas.NewSwitchOp(client)
	sw, err := swOp.Create(context.Background(), "is1a", &iaas.SwitchCreateRequest{
		Name:        "libsacloud-example",
		Description: "description",
		Tags:        types.Tags{"tag1", "tag2"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s", sw.Name)
}

func Example_serverCRUD() {
	// ServerのCRUDを行う例

	// Note: サーバの作成を行いたい場合、通常はgithub.com/libsacloud/v2/utils/serverパッケージを利用してください。
	// この例はServer APIを直接利用したい場合向けです。

	// APIキー
	token := os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
	secret := os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")

	// クライアントの作成
	client := iaas.NewClient(token, secret)

	// サーバの作成(ディスクレス)
	ctx := context.Background()
	serverOp := iaas.NewServerOp(client)
	server, err := serverOp.Create(ctx, "is1a", &iaas.ServerCreateRequest{
		CPU:                  1,
		MemoryMB:             1 * size.GiB,
		ServerPlanCommitment: types.Commitments.Standard,
		ServerPlanGeneration: types.PlanGenerations.Default,
		ConnectedSwitches:    []*iaas.ConnectedSwitch{{Scope: types.Scopes.Shared}},
		InterfaceDriver:      types.InterfaceDrivers.VirtIO,
		Name:                 "libsacloud-example",
		Description:          "description",
		Tags:                 types.Tags{"tag1", "tag2"},
		//IconID:               0,
		WaitDiskMigration: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 更新
	server, err = serverOp.Update(ctx, "is1a", server.ID, &iaas.ServerUpdateRequest{
		Name:        "libsacloud-example-updated",
		Description: "description-updated",
		Tags:        types.Tags{"tag1-updated", "tag2-updated"},
		// IconID:      0,
	})
	if err != nil {
		log.Fatal(err)
	}

	// 起動
	if err := power.BootServer(ctx, serverOp, "is1a", server.ID); err != nil {
		log.Fatal(err)
	}

	// シャットダウン(force)
	if err := power.ShutdownServer(ctx, serverOp, "is1a", server.ID, true); err != nil {
		log.Fatal(err)
	}

	// 削除
	if err := serverOp.Delete(ctx, "is1a", server.ID); err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_Do_direct() {
	// iaas.Clientを直接利用する例
	// Note: 通常はiaas.xxxOpを通じて操作してください。

	// クライアントの作成
	if os.Getenv("SAKURACLOUD_ACCESS_TOKEN") == "" ||
		os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET") == "" {
		log.Fatal("required: SAKURACLOUD_ACCESS_TOKEN and SAKURACLOUD_ACCESS_TOKEN_SECRET")
	}
	client := iaas.NewClientFromEnv()

	// ゾーン一覧を取得する例
	url := "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/zone"
	data, err := client.Do(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	var zones map[string]interface{}
	err = json.Unmarshal(data, &zones)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(zones)
}

func ExampleClient_Do_withNaked() {
	// iaas.Clientを直接利用する例
	// レスポンスとしてnakedパッケージを利用する
	// Note: 通常はiaas.xxxOpを通じて操作してください。

	// クライアントの作成
	if os.Getenv("SAKURACLOUD_ACCESS_TOKEN") == "" ||
		os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET") == "" {
		log.Fatal("required: SAKURACLOUD_ACCESS_TOKEN and SAKURACLOUD_ACCESS_TOKEN_SECRET")
	}
	client := iaas.NewClientFromEnv()

	// ゾーン一覧を取得する例
	url := "https://secure.sakura.ad.jp/cloud/zone/is1a/api/cloud/1.1/zone"
	data, err := client.Do(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	// レスポンスを受けるためのstruct
	type searchResult struct {
		Zones []*naked.Zone
	}
	result := &searchResult{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Fatal(err)
	}

	for _, zone := range result.Zones {
		fmt.Printf("ID: %v Name: %v\n", zone.ID, zone.Name)
	}
}
