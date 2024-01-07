package fsstest

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tsapen/fss/pkg/client"
)

const (
	sendFilePathTemplate = "/app/test_data/test_file_%d.txt"
	gotFilePathTemplate  = "/app/test_data/got_file_%d.txt"
)

type storage struct {
	addServerURI  string
	sendFilePaths []string
	gotFilePaths  []string
}

func newStorage(t *testing.T, addr string) *storage {
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

	return &storage{
		sendFilePaths: sendFilePaths,
		gotFilePaths:  gotFilePaths,
		addServerURI:  uri.String(),
	}
}

func (s *storage) test6Servers(ctx context.Context, t *testing.T, client *client.Client) {
	// 1. Add 2-6 servers.
	for i := 2; i < 7; i++ {
		s.addServer(i)
	}

	// 2. Create file 1.
	assert.NoError(t, client.SaveFile(ctx, "file_1", s.sendFilePaths[0]))

	// 3. Check file 1.
	assert.NoError(t, client.GetFile(ctx, "file_1", s.gotFilePaths[0]))
	s.equalFiles(t, s.sendFilePaths[0], s.gotFilePaths[0])
}

func (s *storage) test7Servers(ctx context.Context, t *testing.T, client *client.Client) {
	// 1. Add server.
	s.addServer(7)

	// 2. Check file 1.
	assert.NoError(t, client.GetFile(ctx, "file_1", s.gotFilePaths[0]))
	s.equalFiles(t, s.sendFilePaths[0], s.gotFilePaths[0])

	// 2. Create file 2.
	assert.NoError(t, client.SaveFile(ctx, "file_2", s.sendFilePaths[1]))

	// 3. Check file 2.
	assert.NoError(t, client.GetFile(ctx, "file_2", s.gotFilePaths[1]))
	s.equalFiles(t, s.sendFilePaths[1], s.gotFilePaths[1])
}

func (s *storage) test8Servers(ctx context.Context, t *testing.T, client *client.Client) {
	// 1. Add server.
	s.addServer(8)

	// 2. Check file 1.
	assert.NoError(t, client.GetFile(ctx, "file_1", s.gotFilePaths[0]))
	s.equalFiles(t, s.sendFilePaths[0], s.gotFilePaths[0])

	// 3. Check file 2.
	assert.NoError(t, client.GetFile(ctx, "file_2", s.gotFilePaths[1]))
	s.equalFiles(t, s.sendFilePaths[1], s.gotFilePaths[1])

	// 4. Create file 3.
	assert.NoError(t, client.SaveFile(ctx, "file_3", s.sendFilePaths[2]))

	// 5. Check file 3.
	assert.NoError(t, client.GetFile(ctx, "file_3", s.gotFilePaths[2]))
	s.equalFiles(t, s.sendFilePaths[2], s.gotFilePaths[2])
}

func (*storage) equalFiles(t *testing.T, filePath1, filePath2 string) {
	content1, err := os.ReadFile(filePath1)
	assert.NoError(t, err)

	content2, err := os.ReadFile(filePath2)
	assert.NoError(t, err)

	assert.Equal(t, string(content1), string(content2))
}
