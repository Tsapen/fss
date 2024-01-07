package main

import (
	"github.com/rs/zerolog/log"

	"github.com/Tsapen/fss/internal/config"
	dm "github.com/Tsapen/fss/internal/download-manager"
	fsshttp "github.com/Tsapen/fss/internal/fss-http"
	"github.com/Tsapen/fss/internal/migrator"
	"github.com/Tsapen/fss/internal/postgres"
)

func main() {
	cfg, err := config.GetForFSS()
	if err != nil {
		log.Fatal().Err(err).Msg("read config")
	}

	db, err := postgres.New(postgres.Config(*cfg.DB))
	if err != nil {
		log.Fatal().Err(err).Msg("init storage")
	}

	if err = migrator.ApplyMigrations(cfg.MigrationsPath, db.DB.DB); err != nil {
		log.Fatal().Err(err).Msg("apply migrations")
	}

	dmService := dm.New(db)

	httpService, err := fsshttp.NewServer(fsshttp.Config(*cfg.HTTPCfg), cfg.MaxFragmentSize, dmService)
	if err != nil {
		log.Fatal().Err(err).Msg("init http server")
	}

	if err = httpService.StartServer(); err != nil {
		log.Fatal().Err(err).Msg("run tcp server")
	}
}
