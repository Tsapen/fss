package fsshttp

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	dm "github.com/Tsapen/fss/internal/download-manager"
	"github.com/Tsapen/fss/internal/fss"
	"github.com/Tsapen/fss/internal/keeper"
)

type Server struct {
	cfg             Config
	maxFragmentSize int64
	s               *http.Server
	dmService       *dm.Service
	fsClient        *keeper.Keeper
}

type Config struct {
	Addr string
}

func NewServer(cfg Config, maxFragmentSize int64, dmService *dm.Service) (*Server, error) {
	r := mux.NewRouter()
	s := &Server{
		cfg:       cfg,
		dmService: dmService,
		s: &http.Server{
			Addr:    cfg.Addr,
			Handler: r,
		},
		fsClient:        keeper.New(),
		maxFragmentSize: maxFragmentSize,
	}

	r = r.PathPrefix("/api/v1").Subrouter()
	r.HandleFunc("/file", s.withMW(s.uploadFile)).Methods(http.MethodPost)
	r.HandleFunc("/file", s.withMW(s.downloadFile)).Methods(http.MethodGet)

	r.HandleFunc("/fs-server", s.withMW(s.addServer)).Methods(http.MethodPost)

	return s, nil
}

// Start runs server.
func (s *Server) StartServer() error {
	log.Info().Msgf("HTTP server started to listen %s", s.cfg.Addr)

	return s.s.ListenAndServe()
}

func (s *Server) withMW(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := fss.WithReqID(r.Context(), uuid.NewString())

		logger := log.With().Str("method", r.Method).Str("path", r.URL.String()).Str("request_id", fss.ReqIDFromCtx(ctx)).Logger()
		logger.Info().Msg("received request")

		ctx = fss.WithLogger(ctx, logger)
		r = r.WithContext(ctx)

		f(w, r)
	}
}

func renderErr(ctx context.Context, logger zerolog.Logger, err error, w http.ResponseWriter) {
	statusCode := httpStatus(err)
	w.WriteHeader(statusCode)

	logger.Info().Err(err).Int("status code", statusCode).Msg("failed to process message")
}
