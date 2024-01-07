package dm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/Tsapen/fss/internal/fss"
)

type Metadata struct {
	ServerURLs []string
	PartNum    int
}

type Storage interface {
	CreateFile(ctx context.Context, filename string) (int64, error)
	File(ctx context.Context, name string) (*fss.File, error)
	UpdateFile(ctx context.Context, f *fss.File) (err error)
	DeleteFile(ctx context.Context, name string) error
	Servers(ctx context.Context, last int64) ([]fss.Server, error)
	CreateServer(ctx context.Context, uri string) error
}

type Service struct {
	storage Storage
}

func New(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (s *Service) Metadata(ctx context.Context, filename string) (*Metadata, error) {
	f, err := s.storage.File(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	serverURLs, err := s.orderedServerURLs(ctx, f.Name, f.LastServerID)
	if err != nil {
		return nil, fmt.Errorf("get ordered server urls: %w", err)
	}

	return &Metadata{
		ServerURLs: serverURLs,
		PartNum:    *f.Fragments,
	}, nil
}

func (s *Service) StartSaving(ctx context.Context, filename string) ([]string, error) {
	lastServerID, err := s.storage.CreateFile(ctx, filename)
	if err != nil {
		return nil, fmt.Errorf("create file metadata: %w", err)
	}

	serverURLS, err := s.orderedServerURLs(ctx, filename, lastServerID)
	if err != nil {
		return nil, fmt.Errorf("create file metadata: %w", err)
	}

	return serverURLS, nil
}

func (s *Service) RollbackFile(ctx context.Context, filename string) error {
	return s.storage.DeleteFile(ctx, filename)
}

func (s *Service) CommitBatch(ctx context.Context, filename string) error {
	now := time.Now()

	return s.storage.UpdateFile(ctx, &fss.File{
		Name:            filename,
		LastCommittedAt: &now,
	})
}

func (s *Service) CommitFile(ctx context.Context, filename string, fragmentsNum int) error {
	return s.storage.UpdateFile(ctx, &fss.File{
		Name:            filename,
		LastCommittedAt: nil,
		Fragments:       &fragmentsNum,
	})
}

func (s *Service) orderedServerURLs(ctx context.Context, filename string, lastServerID int64) ([]string, error) {
	servers, err := s.storage.Servers(ctx, lastServerID)
	if err != nil {
		return nil, fmt.Errorf("get servers: %w", err)
	}

	if len(servers) == 0 {
		return nil, fmt.Errorf("empty servers list")
	}

	filenameHash, err := s.hash(filename, len(servers))
	if err != nil {
		return nil, fmt.Errorf("get hash: %w", err)
	}

	serverURLs := make([]string, 0, len(servers))
	for i := 0; i < len(servers); i++ {
		serverURLs = append(serverURLs, servers[(i+filenameHash)%len(servers)].URL)
	}

	return serverURLs, nil
}

func (s *Service) hash(inputString string, leng int) (int, error) {
	hasher := sha256.New()
	hasher.Write([]byte(inputString))

	result, err := hex.DecodeString(hex.EncodeToString(hasher.Sum(nil)))
	if err != nil {
		return 0, err
	}

	return int(result[0]) % leng, nil
}

func (s *Service) CreateServer(ctx context.Context, uri string) error {
	return s.storage.CreateServer(ctx, uri)
}
