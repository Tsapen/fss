package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/Tsapen/fss/internal/fss"
)

// Config contains data for constructing client.
type Config struct {
	Address string
}

// Clients communicates with FSS http-server.
type Client struct {
	address string

	httpClient *http.Client
}

// New constructs a new FSS client.
func New(cfg Config) (*Client, error) {
	uri, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}

	uri.Path = path.Join(uri.Path, "/api/v1/file")
	return &Client{
		address:    uri.String(),
		httpClient: &http.Client{},
	}, nil
}

func (c *Client) SaveFile(ctx context.Context, savingFileName, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}

	defer func() {
		err = fss.HandleErrPair(file.Close(), err)
	}()

	uri, err := withFileName(c.address, savingFileName)
	if err != nil {
		return fmt.Errorf("add filename into url: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodPost, uri, file)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer func() {
		err = fss.HandleErrPair(resp.Body.Close(), err)
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get error http status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetFile(ctx context.Context, fileName, savingFilePath string) (err error) {
	file, err := os.Create(savingFilePath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %v", err)
	}

	defer func() {
		err = fss.HandleErrPair(file.Close(), err)
	}()

	uri, err := withFileName(c.address, fileName)
	if err != nil {
		return fmt.Errorf("add filename into url: %w", err)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer func() {
		err = fss.HandleErrPair(resp.Body.Close(), err)
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get error http status: %d", resp.StatusCode)
	}

	if _, err = io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("write file content: %w", err)
	}

	return nil
}

func withFileName(uri, fileName string) (string, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()
	q.Set("filename", fileName)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func (c *Client) doRequest(ctx context.Context, method, urlPath string, reqData io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, urlPath, reqData)
	if err != nil {
		return nil, fmt.Errorf("construct request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	return resp, nil
}
