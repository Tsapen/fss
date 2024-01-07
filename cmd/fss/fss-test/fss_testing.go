package fsstest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Tsapen/fss/pkg/client"
)

type tc struct {
	name     string
	testFunc func(ctx context.Context, t *testing.T, client *client.Client)
}

// TestFSS does integration testing.
func TestFSS(t *testing.T, addr string) {
	ctx := context.Background()

	s := newStorage(t, addr)
	s.waitRunning(t)

	client, err := client.New(client.Config{
		Address: addr,
	})
	if err != nil {
		t.Fatalf("create client: %v\n", err)
	}
	testcases := []tc{
		{name: "test 6 servers", testFunc: s.test6Servers},
		{name: "test 7 servers", testFunc: s.test7Servers},
		{name: "test 8 servers", testFunc: s.test8Servers},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			testcase.testFunc(ctx, t, client)
		})
	}
}

type addServerRequest struct {
	ServerURL string `json:"server_url"`
}

func (s *storage) waitRunning(t *testing.T) {
	const checkNum = 10
	const maxDelay = 100 * time.Millisecond

	for i := 0; i < checkNum; i++ {
		time.Sleep(maxDelay)

		if err := s.addServer(1); err == nil {
			return
		}
	}

	t.Fatalf("service could not start")
}

func (s *storage) addServer(num int) error {
	client := &http.Client{}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(addServerRequest{
		ServerURL: fmt.Sprintf("http://file-server-%d:43000/file", num),
	}); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.addServerURI, body)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got http status %d", resp.StatusCode)
	}

	return nil
}
