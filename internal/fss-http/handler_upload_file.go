package fsshttp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/Tsapen/fss/internal/fss"
)

func (s *Server) uploadFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := fss.LoggerFromCtx(ctx)

	if err := s.saveFile(ctx, logger, r); err != nil {
		renderErr(ctx, logger, err, w)

		return
	}

	w.WriteHeader(http.StatusOK)
	logger.Info().Msg("finished")
}

func (s *Server) saveFile(ctx context.Context, logger zerolog.Logger, r *http.Request) (err error) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		return fss.NewBadRequestError("filename is empty")
	}

	file := r.Body
	defer func() {
		err = fss.HandleErrPair(file.Close(), err)
	}()

	serverURLs, err := s.dmService.StartSaving(ctx, filename)
	if err != nil {
		return fmt.Errorf("start saving: %w", err)
	}

	fragmentsNum, err := s.saveData(ctx, logger, serverURLs, filename, file)
	if err != nil {
		return fss.HandleErrPair(s.dmService.RollbackFile(ctx, filename), err)
	}

	if err := s.dmService.CommitFile(ctx, filename, fragmentsNum); err != nil {
		return fmt.Errorf("commit file: %w", err)
	}

	return nil
}

func (s *Server) saveData(ctx context.Context, logger zerolog.Logger, serverURLs []string, filename string, file io.Reader) (int, error) {
	var finalFragmentNum int
	var err error
	for fragmentNum := 0; ; fragmentNum += len(serverURLs) {
		finalFragmentNum, err = s.storeBatch(ctx, logger, serverURLs, filename, file, fragmentNum)
		if err != nil {
			return 0, fss.HandleErrPair(s.dmService.RollbackFile(ctx, filename), err)
		}

		if finalFragmentNum > 0 {
			break
		}

		if err := s.dmService.CommitBatch(ctx, filename); err != nil {
			return 0, fmt.Errorf("commit batch: %w", err)
		}
	}

	return finalFragmentNum, nil
}

func (s *Server) storeBatch(ctx context.Context, logger zerolog.Logger, serverURLs []string, filename string, file io.Reader, fragmentNum int) (int, error) {
	successCh := make(chan bool)
	batchStart := fragmentNum
	var finalFragmentNum int
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var last bool
	for _, serverURL := range serverURLs {
		buffer := make([]byte, s.maxFragmentSize)
		n, err := io.ReadFull(file, buffer)
		switch {
		case errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF):
			last = true

		case err != nil:
			return 0, err
		}

		go func(ctx context.Context, uri, fragmentName string, fragment []byte, successCh chan<- bool) {
			err := s.fsClient.StoreFragment(ctx, uri, fragmentName, fragment)
			if err != nil {
				logger.Info().Err(err).Msgf("store fragment '%s'", fragmentName)
			}

			successCh <- (err == nil)
		}(ctx, serverURL, getFragmentName(filename, fragmentNum), buffer[:n], successCh)

		fragmentNum++
		if last {
			finalFragmentNum = fragmentNum
			break
		}
	}

	timer := time.NewTimer(5 * time.Second)
	for i := batchStart; i < fragmentNum; i++ {
		select {
		case success := <-successCh:
			if success {
				continue
			}

		case <-timer.C:
		}

		return 0, fmt.Errorf("store batch")
	}

	return finalFragmentNum, nil
}

func getFragmentName(filename string, part int) string {
	return fmt.Sprintf("%s_%d", filename, part)
}
