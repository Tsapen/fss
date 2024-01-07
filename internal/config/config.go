package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/caarlos0/env/v9"

	"github.com/Tsapen/fss/internal/fss"
)

type (
	fssEnvs struct {
		RootDir        string `env:"FSS_ROOT_DIR"`
		Config         string `env:"FSS_CONFIG"`
		MigrationsPath string `env:"FSS_MIGRATIONS_PATH"`
	}

	fsEnvs struct {
		RootDir string `env:"FSS_ROOT_DIR"`
		Config  string `env:"FS_CONFIG"`
	}

	httpClientEnvs struct {
		RootDir string `env:"FSS_ROOT_DIR"`
		Config  string `env:"FSS_CLIENT_CONFIG"`
	}

	FSSConfig struct {
		HTTPCfg *HTTPCfg `json:"http"`
		DB      *DBCfg   `json:"db"`

		MaxFragmentSize int64  `json:"max_fragment_size"`
		MigrationsPath  string `json:"-"`
	}

	FSConfig struct {
		HTTPCfg *HTTPCfg `json:"http"`

		FSDir string `json:"file_storage_directory"`
	}

	HTTPCfg struct {
		Addr string `json:"address"`
	}

	DBCfg struct {
		UserName    string `json:"username"`
		Password    string `json:"password"`
		Port        string `json:"port"`
		VirtualHost string `json:"virtual_host"`

		HostName string `json:"host"`
	}

	ClientConfig struct {
		Address string `json:"address"`
	}
)

func GetForFSS() (*FSSConfig, error) {
	envs := new(fssEnvs)
	if err := env.Parse(envs); err != nil {
		return nil, fmt.Errorf("get envs: %w", err)
	}

	cfg := new(FSSConfig)
	if err := readFromEnv(path.Join(envs.RootDir, envs.Config), cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg.MigrationsPath = path.Join(envs.RootDir, envs.MigrationsPath)

	return cfg, nil
}

func GetForFS() (*FSConfig, error) {
	envs := new(fsEnvs)
	if err := env.Parse(envs); err != nil {
		return nil, fmt.Errorf("get envs: %w", err)
	}

	cfg := new(FSConfig)
	if err := readFromEnv(path.Join(envs.RootDir, envs.Config), cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return cfg, nil
}

func GetForClient() (*ClientConfig, error) {
	envs := new(httpClientEnvs)
	if err := env.Parse(envs); err != nil {
		return nil, fmt.Errorf("get envs: %w", err)
	}

	cfg := new(ClientConfig)
	if err := readFromEnv(path.Join(envs.RootDir, envs.Config), cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return cfg, nil
}

func readFromEnv(filepath string, receiver any) (err error) {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("open file %s: %w", filepath, err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = fss.HandleErrPair(fmt.Errorf("close file: %w", closeErr), err)
		}
	}()

	if err = json.NewDecoder(file).Decode(receiver); err != nil {
		return fmt.Errorf("decode file: %w", err)
	}

	return
}
