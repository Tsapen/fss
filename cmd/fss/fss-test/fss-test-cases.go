package fsstest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tsapen/fss/internal/postgres"
	"github.com/Tsapen/fss/pkg/client"
)

const (
	sendFilePathTemplate = "/app/test_data/test_file_%d.txt"
	gotFilePathTemplate  = "/app/test_data/got_file_%d.txt"
)

type testData struct {
	addServerURI  string
	sendFilePaths []string
	gotFilePaths  []string

	db *postgres.DB
}

func newTestData(t *testing.T, addr string, db *postgres.DB) *testData {
	sendFilePaths := make([]string, 0, 3)
	gotFilePaths := make([]string, 0, 3)
	for i := 1; i < 4; i++ {
		sendFilePaths = append(sendFilePaths, fmt.Sprintf(sendFilePathTemplate, i))
		gotFilePaths = append(gotFilePaths, fmt.Sprintf(gotFilePathTemplate, i))
	}

	uri, err := url.Parse(addr)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}

	uri.Path = path.Join(uri.Path, "/api/v1/fs-server")

	return &testData{
		sendFilePaths: sendFilePaths,
		gotFilePaths:  gotFilePaths,
		addServerURI:  uri.String(),

		db: db,
	}
}

func (d *testData) test6Servers(ctx context.Context, t *testing.T, client *client.Client) {
	// 1. Add 2-6 servers.
	for i := 2; i < 7; i++ {
		d.addServer(i)
	}

	// 2. Create file 1.
	assert.NoError(t, client.SaveFile(ctx, "file_1", d.sendFilePaths[0]))

	// 3. Check file 1.
	assert.NoError(t, client.GetFile(ctx, "file_1", d.gotFilePaths[0]))
	d.equalFiles(t, d.sendFilePaths[0], d.gotFilePaths[0])
}

func (d *testData) test7Servers(ctx context.Context, t *testing.T, client *client.Client) {
	// 1. Add server.
	d.addServer(7)

	// 2. Check file 1.
	assert.NoError(t, client.GetFile(ctx, "file_1", d.gotFilePaths[0]))
	d.equalFiles(t, d.sendFilePaths[0], d.gotFilePaths[0])

	// 2. Create file 2.
	assert.NoError(t, client.SaveFile(ctx, "file_2", d.sendFilePaths[1]))

	// 3. Check file 2.
	assert.NoError(t, client.GetFile(ctx, "file_2", d.gotFilePaths[1]))
	d.equalFiles(t, d.sendFilePaths[1], d.gotFilePaths[1])
}

func (d *testData) test8Servers(ctx context.Context, t *testing.T, client *client.Client) {
	// 1. Add server.
	d.addServer(8)

	// 2. Check file 1.
	assert.NoError(t, client.GetFile(ctx, "file_1", d.gotFilePaths[0]))
	d.equalFiles(t, d.sendFilePaths[0], d.gotFilePaths[0])

	// 3. Check file 2.
	assert.NoError(t, client.GetFile(ctx, "file_2", d.gotFilePaths[1]))
	d.equalFiles(t, d.sendFilePaths[1], d.gotFilePaths[1])

	// 4. Create file 3.
	assert.NoError(t, client.SaveFile(ctx, "file_3", d.sendFilePaths[2]))

	// 5. Check file 3.
	assert.NoError(t, client.GetFile(ctx, "file_3", d.gotFilePaths[2]))
	d.equalFiles(t, d.sendFilePaths[2], d.gotFilePaths[2])
}

func (d *testData) testRepeatedSave(ctx context.Context, t *testing.T, client *client.Client) {
	q := `INSERT INTO files (name, last_server_id, last_committed_at) 
			VALUES ('file_4', 8, CURRENT_TIMESTAMP - INTERVAL '30 seconds')`
	_, err := d.db.ExecContext(ctx, q)
	assert.NoError(t, err)

	assert.NoError(t, client.SaveFile(ctx, "file_4", d.sendFilePaths[0]))

	assert.NoError(t, client.GetFile(ctx, "file_4", d.gotFilePaths[0]))
	d.equalFiles(t, d.sendFilePaths[0], d.gotFilePaths[0])
}

func (d *testData) testGetFileErrors(ctx context.Context, t *testing.T, client *client.Client) {
	tests := []struct {
		name     string
		modifier func(*testing.T, string, string) (string, string)
	}{
		{
			name: "file not found",
			modifier: func(t *testing.T, fileName, savingFilePath string) (string, string) {
				return "wrong_file_name", savingFilePath
			},
		},
		{
			name: "not committed file",
			modifier: func(t *testing.T, fileName, savingFilePath string) (string, string) {
				q := `UPDATE files f SET fragments = NULL WHERE name = 'file_1'`
				_, err := d.db.ExecContext(ctx, q)
				assert.NoError(t, err)

				return fileName, savingFilePath
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileName, savingFilePath := tt.modifier(t, "file_1", d.gotFilePaths[0])
			assert.Error(t, client.GetFile(ctx, fileName, savingFilePath))
		})
	}
}

func (d *testData) testSaveFileErrors(ctx context.Context, t *testing.T, client *client.Client) {
	assert.NoError(t, client.SaveFile(ctx, "file_5", d.sendFilePaths[0]))

	tests := []struct {
		name     string
		modifier func(*testing.T)
	}{
		{
			name: "saving is active",
			modifier: func(t *testing.T) {
				q := `UPDATE files f SET last_committed_at = CURRENT_TIMESTAMP, fragments = NULL WHERE name = 'file_5'`
				_, err := d.db.ExecContext(ctx, q)
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, client.SaveFile(ctx, "file_5", d.gotFilePaths[0]))
		})
	}
}

func (*testData) equalFiles(t *testing.T, filePath1, filePath2 string) {
	content1, err := os.ReadFile(filePath1)
	assert.NoError(t, err)

	content2, err := os.ReadFile(filePath2)
	assert.NoError(t, err)

	assert.Equal(t, string(content1), string(content2))
}
