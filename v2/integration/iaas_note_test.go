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

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/sacloud/saclient-go"
	"github.com/stretchr/testify/require"
)

// newIaasClient は iaas.NewClient 経由で ogen クライアントを組み立てる。
// saclient-go のプロファイル/環境変数解決をそのまま使う。
func newIaasClient(t *testing.T) *client.Client {
	t.Helper()

	if os.Getenv("SAKURA_ACCESS_TOKEN") == "" || os.Getenv("SAKURA_ACCESS_TOKEN_SECRET") == "" {
		t.Skip("SAKURA_ACCESS_TOKEN and SAKURA_ACCESS_TOKEN_SECRET must be set")
	}

	var sc saclient.Client
	if err := sc.SetEnviron(os.Environ()); err != nil {
		t.Fatalf("failed to set environment: %v", err)
	}

	c, err := iaas.NewClient(&sc)
	if err != nil {
		t.Fatalf("failed to create iaas client: %v", err)
	}
	return c
}

// TestIaasNoteCRUD はラッパー層（iaas.NewNoteOp）経由で Note の CRUD を通す。
// 比較用に integration/note_test.go は直接 ogen クライアントを使っている。
func TestIaasNoteCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newIaasClient(t)
	ctx := context.Background()
	zone := getZone()

	noteOp := iaas.NewNoteOp(c)

	// Create
	createResp, err := noteOp.Create(ctx, zone, &client.NoteCreateRequestEnvelope{
		Note: client.NoteCreateRequest{
			Name:    client.NewOptNilString("test-note-wrapper"),
			Tags:    []string{"test", "integration", "wrapper"},
			Class:   client.NewOptNilString("shell"),
			Content: client.NewOptNilString("#!/bin/bash\necho hello from wrapper"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	noteID := createResp.Note.ID.Value
	t.Logf("Created note ID (wrapper): %d", noteID)
	idStr := fmt.Sprintf("%d", noteID)

	// Read
	readResp, err := noteOp.Read(ctx, zone, idStr)
	require.NoError(t, err)
	require.Equal(t, "test-note-wrapper", readResp.Note.Name.Value)

	// Update
	updateResp, err := noteOp.Update(ctx, zone, idStr, &client.NoteUpdateRequestEnvelope{
		Note: client.NoteUpdateRequest{
			Name:    client.NewOptNilString("test-note-wrapper-updated"),
			Tags:    []string{"test", "integration", "wrapper", "updated"},
			Class:   client.NewOptNilString("shell"),
			Content: client.NewOptNilString("#!/bin/bash\necho updated"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, "test-note-wrapper-updated", updateResp.Note.Name.Value)

	// List（ラッパーの Find クエリ書き換えミドルウェアの動作確認。
	//        メソッド名は List だが内部は ogen の NoteOpFind を呼んでいる）
	findResp, err := noteOp.List(ctx, zone, &client.NoteFindRequest{
		Count: 50,
	})
	require.NoError(t, err)
	var found bool
	for _, n := range findResp.Notes {
		if n.ID.Value == noteID {
			found = true
			break
		}
	}
	require.True(t, found, "作成した Note が Find 結果に含まれていること")

	// Delete
	err = noteOp.Delete(ctx, zone, idStr)
	require.NoError(t, err)

	// 削除後は 404
	_, err = noteOp.Read(ctx, zone, idStr)
	require.Error(t, err)
}
