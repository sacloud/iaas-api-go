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

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sacloud/iaas-api-go/v2/client"
	"github.com/stretchr/testify/require"
)

func TestNoteCRUD(t *testing.T) {
	if os.Getenv("TEST_ACC") == "" {
		t.Skip("TEST_ACC=1 env var required")
	}

	c := newClient(t)
	ctx := context.Background()
	zone := getZone()

	// 1. Create - スクリプト作成
	createReq := &client.NoteCreateRequestEnvelope{
		Note: client.NoteCreateRequest{
			Name:    client.NewOptNilString("test-note"),
			Tags:    []string{"test", "integration"},
			Class:   client.NewOptNilString("shell"),
			Content: client.NewOptNilString("#!/bin/bash\necho hello"),
		},
	}

	createResp, err := c.NoteOpCreate(ctx, createReq, client.NoteOpCreateParams{Zone: zone})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	noteID := createResp.Note.ID.Value
	t.Logf("Created note ID: %d", noteID)
	require.Equal(t, "test-note", createResp.Note.Name.Value)
	require.Equal(t, "shell", createResp.Note.Class.Value)

	// 2. Read - スクリプト取得
	readResp, err := c.NoteOpRead(ctx, client.NoteOpReadParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", noteID),
	})
	require.NoError(t, err)
	require.Equal(t, "test-note", readResp.Note.Name.Value)
	require.Equal(t, noteID, readResp.Note.ID.Value)

	// 3. Update - スクリプト更新
	updateResp, err := c.NoteOpUpdate(ctx, &client.NoteUpdateRequestEnvelope{
		Note: client.NoteUpdateRequest{
			Name:    client.NewOptNilString("test-note-updated"),
			Tags:    []string{"test", "integration", "updated"},
			Class:   client.NewOptNilString("shell"),
			Content: client.NewOptNilString("#!/bin/bash\necho updated"),
		},
	}, client.NoteOpUpdateParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", noteID),
	})
	require.NoError(t, err)
	require.Equal(t, "test-note-updated", updateResp.Note.Name.Value)

	// 4. Find - スクリプト検索
	findResp, err := c.NoteOpFind(ctx, client.NoteOpFindParams{Zone: zone})
	require.NoError(t, err)
	require.Greater(t, len(findResp.Notes), 0)

	var found bool
	for _, note := range findResp.Notes {
		if note.ID.Value == noteID {
			found = true
			break
		}
	}
	require.True(t, found, "作成したスクリプトがリストに含まれていること")

	// 5. Delete - スクリプト削除
	_, err = c.NoteOpDelete(ctx, client.NoteOpDeleteParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", noteID),
	})
	require.NoError(t, err)

	// 削除後は 404 になることを確認
	_, err = c.NoteOpRead(ctx, client.NoteOpReadParams{
		Zone: zone,
		ID:   fmt.Sprintf("%d", noteID),
	})
	require.Error(t, err)
}
