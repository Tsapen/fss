package keeper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/Tsapen/fss/internal/fss"
)

type Keeper struct {
	httpClient *http.Client
}

func New() *Keeper {
	return &Keeper{
		httpClient: &http.Client{},
	}
}

func (k *Keeper) GetFragment(ctx context.Context, uri, fragmentName string) (res io.ReadCloser, err error) {
	uri, err = k.withFilename(uri, fragmentName)
	if err != nil {
		return nil, fmt.Errorf("add filename into url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("construct a request: %w", err)
	}

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got '%d' response http status", resp.StatusCode)
	}

	return resp.Body, nil
}

func (k *Keeper) StoreFragment(ctx context.Context, uri, fragmentName string, fragment []byte) (err error) {
	uri, err = k.withFilename(uri, fragmentName)
	if err != nil {
		return fmt.Errorf("add filename into url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uri, bytes.NewReader(fragment))
	if err != nil {
		return fmt.Errorf("construct a request: %w", err)
	}

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	defer func() {
		err = fss.HandleErrPair(resp.Body.Close(), err)
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("got '%d' response http status", resp.StatusCode)
	}

	return nil
}

func (*Keeper) withFilename(uri, filename string) (string, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	q := parsed.Query()
	q.Set("filename", filename)
	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}
