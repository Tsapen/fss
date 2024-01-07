package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Tsapen/fss/internal/config"
	"github.com/Tsapen/fss/internal/fss"
)

type server struct {
	cfg config.FSConfig
	s   *http.Server
}

func main() {
	cfg, err := config.GetForFS()
	if err != nil {
		log.Fatal().Err(err).Msg("read config")
	}

	fs := newFileServer(*cfg)

	fs.startServer()
}

func newFileServer(cfg config.FSConfig) *server {
	r := mux.NewRouter()
	s := &server{
		cfg: cfg,
		s: &http.Server{
			Addr:    cfg.HTTPCfg.Addr,
			Handler: r,
		},
	}

	r.HandleFunc("/file", s.storeHandler).Methods(http.MethodPost)
	r.HandleFunc("/file", s.getHandler).Methods(http.MethodGet)

	return s
}

func (s *server) startServer() error {
	log.Info().Msgf("HTTP server started to listen %s", s.cfg.HTTPCfg.Addr)

	return s.s.ListenAndServe()
}

func (s *server) storeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := fss.WithReqID(r.Context(), uuid.NewString())

	logger := log.With().Str("method", r.Method).Str("path", r.URL.String()).Str("request_id", fss.ReqIDFromCtx(ctx)).Logger()
	logger.Info().Msg("received request")

	if err := s.store(r); err != nil {
		renderErr(ctx, logger, err, w)

		return
	}

	w.WriteHeader(http.StatusOK)
	logger.Info().Msg("processed request")
}

func (s *server) store(r *http.Request) (err error) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		return fss.NewBadRequestError("filename is empty")
	}

	filePath := filepath.Join(".", "stored_files", filename)
	file, err := os.Create(filePath)
	if err != nil {
		return fss.NewInternalError("store file: %w", err)
	}

	defer func() {
		err = fss.HandleErrPair(file.Close(), err)
	}()

	if _, err = io.Copy(file, r.Body); err != nil {
		return fss.NewInternalError("copy into file: %w", err)
	}

	return nil
}

func (s *server) getHandler(w http.ResponseWriter, r *http.Request) {
	ctx := fss.WithReqID(r.Context(), uuid.NewString())

	logger := log.With().Str("method", r.Method).Str("path", r.URL.String()).Str("request_id", fss.ReqIDFromCtx(ctx)).Logger()
	logger.Info().Msg("received request")

	rc, err := s.get(r)
	if err != nil {
		renderErr(ctx, logger, err, w)
		return
	}

	defer rc.Close()

	if _, err = io.Copy(w, rc); err != nil {
		renderErr(ctx, logger, fss.NewInternalError("send file: %w", err), w)
		return
	}

	logger.Info().Msg("processed request")
}

func (s *server) get(r *http.Request) (rc io.ReadCloser, err error) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		return nil, fss.NewBadRequestError("filename is empty")
	}

	filePath := filepath.Join(".", "stored_files", filename)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fss.NewInternalError("open file: %w", err)
	}

	return file, nil
}

func httpStatus(err error) int {
	switch {
	case errors.As(err, &fss.ValidationError{}):
		return http.StatusBadRequest

	case errors.As(err, &fss.NotFoundError{}):
		return http.StatusNotFound

	case errors.As(err, &fss.ConflictError{}):
		return http.StatusConflict

	default:
		return http.StatusInternalServerError
	}
}

func renderErr(ctx context.Context, logger zerolog.Logger, err error, w http.ResponseWriter) {
	statusCode := httpStatus(err)
	w.WriteHeader(statusCode)
	logger.Info().Err(err).Int("status code", statusCode).Msg("failed to process message")
}
