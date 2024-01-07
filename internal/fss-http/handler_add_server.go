package fsshttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Tsapen/fss/internal/fss"
)

type addServerRequest struct {
	ServerURL string `json:"server_url"`
}

func (s *Server) addServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := fss.LoggerFromCtx(ctx)
	req := new(addServerRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		renderErr(ctx, logger, err, w)
		return
	}

	logger.Info().Any("request", req).Msg("request body")
	if err := s.createFileServer(ctx, req.ServerURL); err != nil {
		renderErr(ctx, logger, err, w)
		return
	}

	logger.Info().Msg("success")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) createFileServer(ctx context.Context, uri string) error {
	if err := s.dmService.CreateServer(ctx, uri); err != nil {
		return fmt.Errorf("get fragment by url '%s': %w", uri, err)
	}

	return nil
}
