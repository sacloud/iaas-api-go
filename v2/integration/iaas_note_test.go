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
	"os"
	"testing"

	iaas "github.com/sacloud/iaas-api-go/v2"
	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

// TestIaasNoteCRUD はラッパー層（iaas.NewNoteOp）経由で Note の CRUD を通す。
// integration/note_test.go が raw ogen メソッドを直接呼ぶのに対し、本テストは
// Op インターフェース経由での CRUD を検証する wrapper 層のスモークテスト。
func TestIaasNoteCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()

	noteOp := iaas.NewNoteOp(c)

	// Create
	createResp, err := noteOp.Create(ctx,&client.NoteCreateRequestEnvelope{
		Note: client.NoteCreateRequest{
			Name:    client.NewOptString("test-note-wrapper"),
			Tags:    []string{"test", "integration", "wrapper"},
			Class:   client.NewOptString("shell"),
			Content: client.NewOptString("#!/bin/bash\necho hello from wrapper"),
		},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	noteID := createResp.Note.ID.Value
	t.Logf("Created note ID (wrapper): %d", noteID)

	// Read
	readResp, err := noteOp.Read(ctx, noteID)
	require.NoError(t, err)
	require.Equal(t, "test-note-wrapper", readResp.Note.Name.Value)

	// Update
	updateResp, err := noteOp.Update(ctx, noteID, &client.NoteUpdateRequestEnvelope{
		Note: client.NoteUpdateRequest{
			Name:    client.NewOptString("test-note-wrapper-updated"),
			Tags:    []string{"test", "integration", "wrapper", "updated"},
			Class:   client.NewOptString("shell"),
			Content: client.NewOptString("#!/bin/bash\necho updated"),
		},
	})
	require.NoError(t, err)
	require.Equal(t, "test-note-wrapper-updated", updateResp.Note.Name.Value)

	// List（ラッパーの Find クエリ書き換えミドルウェアの動作確認。
	//        メソッド名は List だが内部は ogen の NoteOpFind を呼んでいる。
	//        アカウント内 Note の累積に影響されないよう Name で絞り込む）
	findResp, err := noteOp.List(ctx, &client.NoteFindRequest{
		Count:  50,
		Filter: client.NoteFindFilter{Name: "test-note-wrapper-updated"},
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
	err = noteOp.Delete(ctx, noteID)
	require.NoError(t, err)

	// 削除後は 404
	_, err = noteOp.Read(ctx, noteID)
	require.Error(t, err)
}
