package fsshttp

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Tsapen/fss/internal/fss"
)

func (s *Server) downloadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := fss.LoggerFromCtx(ctx)
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		renderErr(ctx, logger, fss.NewBadRequestError("filename is empty"), w)
		return
	}

	m, err := s.dmService.Metadata(ctx, filename)
	if err != nil {
		renderErr(ctx, logger, err, w)
		return
	}

	for partNumber := 0; partNumber < m.PartNum; partNumber++ {
		uri := m.ServerURLs[partNumber%len(m.ServerURLs)]

		if err := s.writeFragment(ctx, uri, getFragmentName(filename, partNumber), w); err != nil {
			logger.Info().Err(err).Msg("failed to get file")
			http.Error(w, "storage error", http.StatusInternalServerError)

			return
		}
	}

	logger.Info().Msg("finished")
}

func (s *Server) writeFragment(ctx context.Context, uri, fragmentName string, w http.ResponseWriter) (err error) {
	fragment, err := s.fsClient.GetFragment(ctx, uri, fragmentName)
	if err != nil {
		return fmt.Errorf("get fragment '%s' by url '%s': %w", fragmentName, uri, err)
	}

	defer func() {
		err = fss.HandleErrPair(err, fragment.Close())
	}()

	if _, err = io.Copy(w, fragment); err != nil {
		return fmt.Errorf("write fragment '%s': %w", fragmentName, err)
	}

	return nil
}
